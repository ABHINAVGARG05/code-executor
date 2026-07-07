package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqst "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/ABHINAVGARG05/rme/aws/shared/store"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type jobMsg struct {
	ExecutionID string `json:"executionId"`
	Language    string `json:"language"`
}

func main() {
	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	queueURL := os.Getenv("SQS_QUEUE_URL")
	table := os.Getenv("DYNAMODB_TABLE")
	bucket := os.Getenv("CODE_EXEC_BUCKET")
	sandboxImage := os.Getenv("SANDBOX_IMAGE")
	if sandboxImage == "" {
		sandboxImage = "code-exec-sandbox-cpp:latest"
	}
	timeoutSec := 10
	if v := os.Getenv("EXEC_TIMEOUT_SEC"); v != "" {
		fmt.Sscanf(v, "%d", &timeoutSec)
	}

	if region == "" || queueURL == "" || table == "" || bucket == "" {
		log.Fatal("missing required env vars")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	ddb := dynamodb.NewFromConfig(cfg)
	s3c := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3c)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create Docker client: %v", err)
	}

	log.Println("executor-cpp started")

	for {
		msgsOut, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 5,
			WaitTimeSeconds:     10,
			VisibilityTimeout:   int32(timeoutSec + 20),
		})
		if err != nil {
			log.Printf("receive error: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if len(msgsOut.Messages) == 0 {
			continue
		}

		for _, m := range msgsOut.Messages {
			var jm jobMsg
			if err := json.Unmarshal([]byte(aws.ToString(m.Body)), &jm); err != nil {
				log.Printf("bad message: %v", err)
				deleteMessage(ctx, sqsClient, queueURL, m)
				continue
			}
			if jm.Language != "cpp" {
				continue
			}
			go func(m sqst.Message, jm jobMsg) {
				if err := processJob(ctx, ddb, uploader, dockerClient, table, bucket, sandboxImage, jm, timeoutSec); err != nil {
					log.Printf("job %s failed: %v", jm.ExecutionID, err)
				}
				deleteMessage(ctx, sqsClient, queueURL, m)
			}(m, jm)
		}
	}
}

func deleteMessage(ctx context.Context, client *sqs.Client, queueURL string, m sqst.Message) {
	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: m.ReceiptHandle,
	})
	if err != nil {
		log.Printf("delete failed: %v", err)
	}
}

func processJob(
	ctx context.Context,
	ddb *dynamodb.Client,
	uploader *manager.Uploader,
	dockerClient *client.Client,
	table, bucket, sandboxImage string,
	jm jobMsg,
	timeout int,
) error {
	job, err := store.GetJob(ctx, ddb, table, jm.ExecutionID)
	if err != nil {
		return err
	}
	if err := store.UpdateStatus(ctx, ddb, table, job.ExecutionID, "running", nil); err != nil {
		return err
	}

	start := time.Now()

	codeTar, err := tarSource("main.cpp", job.Code)
	if err != nil {
		return fmt.Errorf("tar failed: %w", err)
	}

	containerCfg := &container.Config{
		Image:      sandboxImage,
		Cmd:        []string{"sh", "-c", "cat > /tmp/main.cpp && g++ -std=c++17 -O2 /tmp/main.cpp -o /tmp/app && exec /tmp/app"},
		OpenStdin:  true,
		StdinOnce:  true,
		WorkingDir: "/home/sandbox",
		User:       "sandbox",
	}

	hostCfg := &container.HostConfig{
		ReadonlyRootfs: true,
		AutoRemove:     true,
		NetworkMode:    container.NetworkMode("none"),
		CapDrop:        []string{"ALL"},
		SecurityOpt:    []string{"no-new-privileges:true"},
		Resources: container.Resources{
			Memory:     256 * 1024 * 1024,
			MemorySwap: 256 * 1024 * 1024,
			NanoCPUs:   1_000_000_000,
			PidsLimit:  int64Ptr(64),
		},
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=64m",
		},
	}

	if r := os.Getenv("SANDBOX_RUNTIME"); r != "" {
		hostCfg.Runtime = r
	}

	resp, err := dockerClient.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return fmt.Errorf("container create failed: %w", err)
	}
	containerID := resp.ID

	defer func() {
		dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	}()

	if err := dockerClient.CopyToContainer(ctx, containerID, "/tmp", &codeTar, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy code failed: %w", err)
	}

	attach, err := dockerClient.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("attach failed: %w", err)
	}
	defer attach.Close()

	if err := dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}

	if job.Input != "" {
		attach.Conn.Write([]byte(job.Input))
	}
	attach.CloseWrite()

	var outputBuf bytes.Buffer
	outputDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(&outputBuf, attach.Reader)
		outputDone <- err
	}()

	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	statusCh, errCh := dockerClient.ContainerWait(waitCtx, containerID, container.WaitConditionNotRunning)

	var statusCode int64
	var timedOut bool
	select {
	case <-waitCtx.Done():
		timedOut = true
		dockerClient.ContainerKill(ctx, containerID, "SIGKILL")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("container wait error: %w", err)
		}
	case s := <-statusCh:
		statusCode = s.StatusCode
	}

	<-outputDone

	durationMs := time.Since(start).Milliseconds()

	output := outputBuf.String()
	var stdoutStr, stderrStr string
	if timedOut || statusCode == 0 {
		stdoutStr = output
	} else {
		stderrStr = output
	}

	status := "completed"
	extra := map[string]any{"execDurationMs": durationMs}
	if timedOut {
		status = "timeout"
		extra["error"] = "execution timeout"
	} else if statusCode != 0 {
		status = "failed"
		extra["error"] = truncate(stderrStr, 2000)
	}

	key := fmt.Sprintf("outputs/%s.txt", jm.ExecutionID)
	uploadBody := stdoutStr
	if stderrStr != "" {
		uploadBody += "\n[stderr]\n" + stderrStr
	}
	if _, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(uploadBody),
	}); err != nil {
		status = "failed"
		extra["error"] = fmt.Sprintf("upload failed: %v", err)
	} else {
		extra["outputPath"] = fmt.Sprintf("s3://%s/%s", bucket, key)
		extra["stdoutPreview"] = truncate(stdoutStr, 500)
	}

	if err := store.UpdateStatus(ctx, ddb, table, jm.ExecutionID, status, extra); err != nil {
		return err
	}
	return nil
}

func tarSource(filename, content string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: filename,
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return buf, err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return buf, err
	}
	if err := tw.Close(); err != nil {
		return buf, err
	}
	return buf, nil
}

func int64Ptr(v int64) *int64 {
	return &v
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

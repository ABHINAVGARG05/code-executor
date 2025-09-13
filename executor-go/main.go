package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
    timeoutSec := 10
    if v := os.Getenv("EXEC_TIMEOUT_SEC"); v != "" { fmt.Sscanf(v, "%d", &timeoutSec) }

    if region == "" || queueURL == "" || table == "" || bucket == "" {
        log.Fatal("missing required env vars")
    }

    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil { log.Fatal(err) }

    sqsClient := sqs.NewFromConfig(cfg)
    ddb := dynamodb.NewFromConfig(cfg)
    s3c := s3.NewFromConfig(cfg)
    uploader := manager.NewUploader(s3c)

    log.Println("executor-go started")

    for {
        msgsOut, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
            QueueUrl:            aws.String(queueURL),
            MaxNumberOfMessages: 5,
            WaitTimeSeconds:     10,
            VisibilityTimeout:  int32(timeoutSec + 15),
        })
        if err != nil { log.Printf("receive error: %v", err); time.Sleep(2*time.Second); continue }
        if len(msgsOut.Messages) == 0 { continue }

        for _, m := range msgsOut.Messages {
            var jm jobMsg
            if err := json.Unmarshal([]byte(aws.ToString(m.Body)), &jm); err != nil { log.Printf("bad message: %v", err); deleteMessage(ctx, sqsClient, queueURL, m); continue }
            if jm.Language != "go" { // not for this executor
                continue
            }
            go func(m sqst.Message, jm jobMsg) { // process concurrently
                if err := processJob(ctx, ddb, uploader, table, bucket, jm, timeoutSec); err != nil {
                    log.Printf("job %s failed: %v", jm.ExecutionID, err)
                }
                deleteMessage(ctx, sqsClient, queueURL, m)
            }(m, jm)
        }
    }
}

func deleteMessage(ctx context.Context, client *sqs.Client, queueURL string, m sqst.Message) {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{QueueUrl: aws.String(queueURL), ReceiptHandle: m.ReceiptHandle})
    if err != nil { log.Printf("delete failed: %v", err) }
}

func processJob(ctx context.Context, ddb *dynamodb.Client, uploader *manager.Uploader, table, bucket string, jm jobMsg, timeout int) error {
    job, err := store.GetJob(ctx, ddb, table, jm.ExecutionID)
    if err != nil { return err }
    if err := store.UpdateStatus(ctx, ddb, table, job.ExecutionID, "running", nil); err != nil { return err }

    start := time.Now()
    tempDir, err := os.MkdirTemp("", "exec-go-*")
    if err != nil { return err }
    defer os.RemoveAll(tempDir)

    mainPath := filepath.Join(tempDir, "main.go")
    if err := os.WriteFile(mainPath, []byte(job.Code), 0644); err != nil { return err }

    ctxRun, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctxRun, "go", "run", mainPath)
    if job.Input != "" {
        cmd.Stdin = strings.NewReader(job.Input)
    }
    var stdout, stderr strings.Builder
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    runErr := cmd.Run()
    durationMs := time.Since(start).Milliseconds()

    status := "completed"
    extra := map[string]any{"execDurationMs": durationMs}
    if ctxRun.Err() == context.DeadlineExceeded {
        status = "timeout"
        extra["error"] = "execution timeout"
    } else if runErr != nil {
        status = "failed"
        extra["error"] = truncate(stderr.String(), 2000)
    }

    // upload output (stdout or both) to s3
    key := fmt.Sprintf("outputs/%s.txt", jm.ExecutionID)
    uploadBody := stdout.String()
    if stderr.Len() > 0 {
        uploadBody += "\n[stderr]\n" + stderr.String()
    }
    if _, err := uploader.Upload(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: strings.NewReader(uploadBody)}); err != nil {
        status = "failed"
        extra["error"] = fmt.Sprintf("upload failed: %v", err)
    } else {
        extra["outputPath"] = fmt.Sprintf("s3://%s/%s", bucket, key)
        extra["stdoutPreview"] = truncate(stdout.String(), 500)
    }

    if err := store.UpdateStatus(ctx, ddb, table, jm.ExecutionID, status, extra); err != nil {
        return err
    }
    return nil
}

func truncate(s string, n int) string {
    if len(s) <= n { return s }
    return s[:n] + "..."
}

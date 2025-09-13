package queue

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// JobMessage is the minimal SQS payload informing executors what to fetch.
type JobMessage struct {
    ExecutionID string `json:"executionId"`
    Language    string `json:"language"`
}

// EnqueueJob sends a job reference to SQS.
func EnqueueJob(ctx context.Context, client *sqs.Client, queueURL string, msg JobMessage) error {
    body, err := json.Marshal(msg)
    if err != nil { return err }
    _, err = client.SendMessage(ctx, &sqs.SendMessageInput{
        QueueUrl:    aws.String(queueURL),
        MessageBody: aws.String(string(body)),
    })
    return err
}

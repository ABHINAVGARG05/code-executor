package store

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbt "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ABHINAVGARG05/rme/aws/shared/models"
)

var ErrNotFound = errors.New("job not found")

func CreateJob(ctx context.Context, ddb *dynamodb.Client, table string, job models.Job) error {
    item, err := attributevalue.MarshalMap(job)
    if err != nil {
        return err
    }
    _, err = ddb.PutItem(ctx, &dynamodb.PutItemInput{
        TableName:           aws.String(table),
        Item:                item,
        ConditionExpression: aws.String("attribute_not_exists(executionId)"),
    })
    return err
}

func GetJob(ctx context.Context, ddb *dynamodb.Client, table, id string) (models.Job, error) {
    out, err := ddb.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(table),
        Key: map[string]ddbt.AttributeValue{
            "executionId": &ddbt.AttributeValueMemberS{Value: id},
        },
    })
    if err != nil {
        return models.Job{}, err
    }
    if out.Item == nil {
        return models.Job{}, ErrNotFound
    }
    var job models.Job
    if err := attributevalue.UnmarshalMap(out.Item, &job); err != nil {
        return models.Job{}, err
    }
    return job, nil
}

// UpdateStatus sets job status + timestamps.
func UpdateStatus(ctx context.Context, ddb *dynamodb.Client, table, id, status string, extra map[string]any) error {
    now := time.Now().UTC().Format(time.RFC3339)

    exprNames := map[string]string{"#s":"status", "#u":"updatedAt"}
    exprValues := map[string]ddbt.AttributeValue{":s": &ddbt.AttributeValueMemberS{Value: status}, ":u": &ddbt.AttributeValueMemberS{Value: now}}
    setParts := []string{"#s = :s", "#u = :u"}

    // handle status-specific timestamp fields
    switch status {
    case "running":
        exprNames["#st"] = "startedAt"
        exprValues[":st"] = &ddbt.AttributeValueMemberS{Value: now}
        setParts = append(setParts, "#st = :st")
    case "completed", "failed", "timeout":
        exprNames["#c"] = "completedAt"
        exprValues[":c"] = &ddbt.AttributeValueMemberS{Value: now}
        setParts = append(setParts, "#c = :c")
    }

    // attach extra attributes
    for k, v := range extra {
        av, err := attributevalue.Marshal(v)
        if err != nil {
            return err
        }
        placeholderName := "#" + k
        placeholderValue := ":" + k
        exprNames[placeholderName] = k
        exprValues[placeholderValue] = av
        setParts = append(setParts, placeholderName+" = "+placeholderValue)
    }

    updateExpr := "SET " + join(setParts, ", ")

    _, err := ddb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(table),
        Key: map[string]ddbt.AttributeValue{
            "executionId": &ddbt.AttributeValueMemberS{Value: id},
        },
        UpdateExpression:          aws.String(updateExpr),
        ExpressionAttributeNames:  exprNames,
        ExpressionAttributeValues: exprValues,
        ConditionExpression:       aws.String("attribute_exists(executionId)"),
    })
    return err
}

func join(parts []string, sep string) string {
    if len(parts) == 0 { return "" }
    out := parts[0]
    for i := 1; i < len(parts); i++ { out += sep + parts[i] }
    return out
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ABHINAVGARG05/rme/aws/shared/config"
	"github.com/ABHINAVGARG05/rme/aws/shared/models"
)

// Dependencies needed by this handler
type ResultDeps struct {
	Env       *config.Env
	DDB       *dynamodb.Client
	Presigner *s3.PresignClient
	Bucket    string
}

func HandleResult(deps ResultDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		out, err := deps.DDB.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(deps.Env.DynamoDBTable),
			Key: map[string]ddbtypes.AttributeValue{
				"executionId": &ddbtypes.AttributeValueMemberS{Value: id},
			},
		})
		if err != nil {
			http.Error(w, "failed to fetch item: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if out.Item == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var job models.Job
		if err := attributevalue.UnmarshalMap(out.Item, &job); err != nil {
			http.Error(w, "failed to unmarshal item: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if job.OutputPath == "" {
			http.Error(w, "result not ready", http.StatusConflict)
			return
		}

		key := strings.TrimPrefix(job.OutputPath, "s3://"+deps.Bucket+"/")

		presigned, err := deps.Presigner.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(deps.Bucket),
			Key:    aws.String(key),
		}, s3.WithPresignExpires(15*time.Minute))
		if err != nil {
			http.Error(w, "failed to generate presigned URL: "+err.Error(), http.StatusInternalServerError)
			return
		}

	json.NewEncoder(w).Encode(map[string]any{"url": presigned.URL, "stdoutPreview": job.StdoutPreview})
	}
}

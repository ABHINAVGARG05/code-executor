package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ABHINAVGARG05/rme/aws/shared/config"
	"github.com/ABHINAVGARG05/rme/aws/shared/models"
)

// StatusDeps contains dependencies needed by HandleStatus
type StatusDeps struct {
	Env *config.Env
	DDB *dynamodb.Client
}

// HandleStatus returns a handler function
func HandleStatus(deps StatusDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		out, err := deps.DDB.GetItem(r.Context(), &dynamodb.GetItemInput{
			TableName: aws.String(deps.Env.DynamoDBTable),
			Key: map[string]ddbtypes.AttributeValue{
				"executionId": &ddbtypes.AttributeValueMemberS{Value: id},
			},
		})
		if err != nil || out.Item == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		var job models.Job
		if err := attributevalue.UnmarshalMap(out.Item, &job); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"jobId":         job.ExecutionID,
			"status":        job.Status,
			"error":         job.Error,
			"outputPath":    job.OutputPath,
			"stdoutPreview": job.StdoutPreview,
			"startedAt":     job.StartedAt,
			"completedAt":   job.CompletedAt,
			"updatedAt":     job.UpdatedAt,
			"execDurationMs": job.ExecDurationMs,
		})
	}
}

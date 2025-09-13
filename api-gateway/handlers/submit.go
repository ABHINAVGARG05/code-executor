package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"

	"github.com/ABHINAVGARG05/rme/aws/shared/config"
	"github.com/ABHINAVGARG05/rme/aws/shared/models"
	"github.com/ABHINAVGARG05/rme/aws/shared/queue"
	"github.com/ABHINAVGARG05/rme/aws/shared/store"
)

// SubmitDeps contains dependencies for HandleSubmit
type SubmitDeps struct {
	Env *config.Env
	DDB *dynamodb.Client
	SQS *sqs.Client
}

type submitReq struct {
	UserID   string `json:"userId"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input"`
}

// HandleSubmit returns a handler function
func HandleSubmit(deps SubmitDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req submitReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req.Language = strings.ToLower(strings.TrimSpace(req.Language))
		switch req.Language {
		case "go", "golang":
			req.Language = "go"
		case "cpp", "c++":
			req.Language = "cpp"
		default:
			http.Error(w, "unsupported language", http.StatusBadRequest)
			return
		}

		job := models.Job{
			ExecutionID: uuid.New().String(),
			UserID:      req.UserID,
			Language:    req.Language,
			Code:        req.Code,
			Input:       req.Input,
			Status:      "queued",
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		}

		if err := store.CreateJob(r.Context(), deps.DDB, deps.Env.DynamoDBTable, job); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := queue.EnqueueJob(r.Context(), deps.SQS, deps.Env.SQSQueueURL, queue.JobMessage{ExecutionID: job.ExecutionID, Language: job.Language}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}
}

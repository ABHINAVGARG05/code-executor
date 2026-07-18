package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"

	"github.com/ABHINAVGARG05/rme/aws/shared/config"
	"github.com/ABHINAVGARG05/rme/aws/shared/languages"
	"github.com/ABHINAVGARG05/rme/aws/shared/models"
	"github.com/ABHINAVGARG05/rme/aws/shared/queue"
	"github.com/ABHINAVGARG05/rme/aws/shared/store"
)

type SubmitDeps struct {
	Env          *config.Env
	DDB          *dynamodb.Client
	S3           *s3.Client
	SQS          *sqs.Client
	LangResolver languages.Resolver
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

		req.Language = deps.LangResolver.Normalize(req.Language)
		if req.Language == "" {
			http.Error(w, fmt.Sprintf("unsupported language; supported=%v", deps.LangResolver.Supported()), http.StatusBadRequest)
			return
		}

		executionID := uuid.New().String()
		codeS3Key := fmt.Sprintf("code/%s.txt", executionID)
		inputS3Key := fmt.Sprintf("input/%s.txt", executionID)

		if err := store.UploadPayload(r.Context(), deps.S3, deps.Env.CodeExecBucket, codeS3Key, req.Code); err != nil {
			http.Error(w, "failed to upload code: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := store.UploadPayload(r.Context(), deps.S3, deps.Env.CodeExecBucket, inputS3Key, req.Input); err != nil {
			http.Error(w, "failed to upload input: "+err.Error(), http.StatusInternalServerError)
			return
		}

		job := models.Job{
			ExecutionID: executionID,
			UserID:      req.UserID,
			Language:    req.Language,
			CodeS3Key:   codeS3Key,
			InputS3Key:  inputS3Key,
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

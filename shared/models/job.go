package models

// Job represents a code execution request tracked in DynamoDB.
// Keep existing attribute names for backward compatibility.
// New optional fields (StartedAt, CompletedAt, UpdatedAt, ExecDurationMs, StdoutPreview)
// are added to enrich status and debugging without inflating item size.
type Job struct {
    ExecutionID    string `dynamodbav:"executionId"`
    UserID         string `dynamodbav:"userId"`
    Language       string `dynamodbav:"language"`
    Code           string `dynamodbav:"code"`
    Input          string `dynamodbav:"input"`
    Status         string `dynamodbav:"status"`
    Error          string `dynamodbav:"error"`
    OutputPath     string `dynamodbav:"outputPath"`
    CreatedAt      string `dynamodbav:"createdAt"`
    UpdatedAt      string `dynamodbav:"updatedAt,omitempty"`
    StartedAt      string `dynamodbav:"startedAt,omitempty"`
    CompletedAt    string `dynamodbav:"completedAt,omitempty"`
    ExecDurationMs int64  `dynamodbav:"execDurationMs,omitempty"`
    StdoutPreview  string `dynamodbav:"stdoutPreview,omitempty"`
}



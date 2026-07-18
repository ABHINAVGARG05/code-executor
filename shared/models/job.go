package models

type Job struct {
    ExecutionID    string `dynamodbav:"executionId"`
    UserID         string `dynamodbav:"userId"`
    Language       string `dynamodbav:"language"`
    CodeS3Key      string `dynamodbav:"codeS3Key"`
    InputS3Key     string `dynamodbav:"inputS3Key"`
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



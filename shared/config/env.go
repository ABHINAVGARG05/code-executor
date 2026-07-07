package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	AWSRegion      string
	AWSAccessKey   string
	AWSSecretKey   string
	AWSEndpoint    string
	CodeExecBucket string
	DynamoDBTable  string
	SQSQueueURL    string
}

func MustEnv() Env {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	return Env{
		AWSRegion:      os.Getenv("AWS_REGION"),
		AWSAccessKey:   os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSEndpoint:    os.Getenv("AWS_ENDPOINT_URL"),
		CodeExecBucket: os.Getenv("CODE_EXEC_BUCKET"),
		DynamoDBTable:  os.Getenv("DYNAMODB_TABLE"),
		SQSQueueURL:    os.Getenv("SQS_QUEUE_URL"),
	}
}

func (e *Env) UseLocalStack() bool {
	return e.AWSEndpoint != ""
}

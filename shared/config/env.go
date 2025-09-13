package config

import (
	"log"

	"github.com/joho/godotenv"
	"os"
)

type Env struct {
	AWSRegion       string
	AWSAccessKey    string
	AWSSecretKey    string
	CodeExecBucket  string
	DynamoDBTable   string
	LambdaFunction  string
	APIGatewayURL   string
	SQSQueueURL     string
}

func MustEnv() Env {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("warning: could not load .env file: %v", err)
	}

	return Env{
		AWSRegion:       os.Getenv("AWS_REGION"),
		AWSAccessKey:    os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		CodeExecBucket:  os.Getenv("CODE_EXEC_BUCKET"),
		DynamoDBTable:   os.Getenv("DYNAMODB_TABLE"),
		LambdaFunction:  os.Getenv("LAMBDA_FUNCTION"),
		APIGatewayURL:   os.Getenv("API_GATEWAY_URL"),
		SQSQueueURL:     os.Getenv("SQS_QUEUE_URL"),
	}
}

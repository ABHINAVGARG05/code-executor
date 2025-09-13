package main

import (
	"context"
	"net/http"

	"github.com/ABHINAVGARG05/rme/aws/api-gateway/handlers"
	"github.com/ABHINAVGARG05/rme/aws/shared/config"


	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"	
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Server struct {
	env       *config.Env
	awsCfg    aws.Config
	sqs       *sqs.Client
	ddb       *dynamodb.Client
	s3        *s3.Client
	presigner *s3.PresignClient
}

func NewServer() *Server {
	env := config.MustEnv()
	ctx := context.Background()
	awsCfg := config.LoadAWSConfig(ctx, env.AWSRegion)

	s3Client := s3.NewFromConfig(awsCfg)

	return &Server{
		env:       &env,
		awsCfg:    awsCfg,
		sqs:       sqs.NewFromConfig(awsCfg),
		ddb:       dynamodb.NewFromConfig(awsCfg),
		s3:        s3Client,
		presigner: s3.NewPresignClient(s3Client),
	}
}

func (s *Server) routes() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/submit", handlers.HandleSubmit(handlers.SubmitDeps{
		Env: s.env,
		DDB: s.ddb,
		SQS: s.sqs,
	}))

	http.HandleFunc("/status", handlers.HandleStatus(handlers.StatusDeps{
		Env: s.env,
		DDB: s.ddb,
	}))

	http.HandleFunc("/result", handlers.HandleResult(handlers.ResultDeps{
		Env:       s.env,
		DDB:       s.ddb,
		Presigner: s.presigner,
		Bucket:    s.env.CodeExecBucket,
	}))
}


package main

import (
	"context"
	"net/http"

	"github.com/ABHINAVGARG05/rme/aws/api-gateway/handlers"
	"github.com/ABHINAVGARG05/rme/aws/api-gateway/middleware"
	"github.com/ABHINAVGARG05/rme/aws/shared/config"
	"github.com/ABHINAVGARG05/rme/aws/shared/languages"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Server struct {
	env          *config.Env
	ddb          *dynamodb.Client
	sqs          *sqs.Client
	s3           *s3.Client
	presigner    *s3.PresignClient
	langResolver languages.Resolver
	rateLimiter  *middleware.RateLimiter
}

func NewServer() *Server {
	env := config.MustEnv()
	ctx := context.Background()
	awsCfg := config.LoadAWSConfig(ctx, env.AWSRegion, env.AWSEndpoint)

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if env.UseLocalStack() {
			o.UsePathStyle = true
		}
	})

	return &Server{
		env: &env,
		ddb: dynamodb.NewFromConfig(awsCfg),
		sqs: sqs.NewFromConfig(awsCfg),
		s3:  s3Client,
		presigner: s3.NewPresignClient(s3Client),
		langResolver: languages.NewResolver([]languages.Language{
			{Name: "go", Aliases: []string{"golang"}, DisplayName: "Go"},
			{Name: "cpp", Aliases: []string{"c++"}, DisplayName: "C++"},
		}),
		rateLimiter: middleware.NewRateLimiter(10, 20),
	}
}

func (s *Server) routes() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/submit", handlers.HandleSubmit(handlers.SubmitDeps{
		Env:          s.env,
		DDB:          s.ddb,
		SQS:          s.sqs,
		LangResolver: s.langResolver,
	}))

	mux.HandleFunc("/status", handlers.HandleStatus(handlers.StatusDeps{
		Env: s.env,
		DDB: s.ddb,
	}))

	mux.HandleFunc("/result", handlers.HandleResult(handlers.ResultDeps{
		Env:       s.env,
		DDB:       s.ddb,
		Presigner: s.presigner,
		Bucket:    s.env.CodeExecBucket,
	}))

	http.Handle("/", s.rateLimiter.Middleware(mux))
}

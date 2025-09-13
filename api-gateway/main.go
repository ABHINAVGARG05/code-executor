package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ABHINAVGARG05/rme/aws/shared/config"
)

func main() {
	s := NewServer()
	s.routes()
	env := config.MustEnv()

	log.Println("API listening on :8080")
	fmt.Println("Loaded region:", env.AWSRegion)
    fmt.Println("Queue URL:", env.SQSQueueURL)
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

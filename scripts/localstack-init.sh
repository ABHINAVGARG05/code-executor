#!/bin/sh
set -e

ENDPOINT="http://localstack:4566"
REGION="us-east-1"

echo "Waiting for LocalStack to be ready..."
until curl -s "$ENDPOINT/_localstack/health" | grep -q '"dynamodb":"available"'; do
  sleep 2
done

echo "Creating DynamoDB table..."
aws --endpoint-url="$ENDPOINT" dynamodb create-table \
  --table-name code-exec-jobs \
  --attribute-definitions AttributeName=executionId,AttributeType=S \
  --key-schema AttributeName=executionId,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$REGION" || echo "Table may already exist"

echo "Creating S3 bucket..."
aws --endpoint-url="$ENDPOINT" s3 mb "s3://code-exec-outputs" \
  --region "$REGION" || echo "Bucket may already exist"

echo "Creating SQS queue..."
aws --endpoint-url="$ENDPOINT" sqs create-queue \
  --queue-name code-exec-queue \
  --region "$REGION" || echo "Queue may already exist"

echo "LocalStack initialization complete."

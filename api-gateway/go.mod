module github.com/ABHINAVGARG05/rme/aws/api-gateway

go 1.25.0

require (
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.50.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.87.2
	github.com/google/uuid v1.6.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.28.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)

require (
	github.com/ABHINAVGARG05/rme/aws/shared v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2 v1.39.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.11
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.5
	github.com/aws/smithy-go v1.23.0 // indirect
)

replace github.com/ABHINAVGARG05/rme/aws/shared => ../shared

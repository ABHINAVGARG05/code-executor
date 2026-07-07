package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadAWSConfig(ctx context.Context, region, endpoint string) aws.Config {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if endpoint != "" {
		opts = append(opts, config.WithEndpointResolver(
			aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL: endpoint,
					}, nil
				},
			),
		))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		panic(err)
	}
	return cfg
}

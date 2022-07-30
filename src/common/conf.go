package common

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log"
)

// common constants and Configurations used to test local and deployed dynamodb & aws resources

const (
	TableName = "UniqueIDMS"
	AwsRegion = "us-east-2"
)

func CreateAwsConfig() aws.Config {
	awsEndpoint := "http://localhost:4566"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(AwsRegion),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
						PartitionID:       "aws",
						URL:               awsEndpoint,
						HostnameImmutable: true,
					},
					nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "test", SecretAccessKey: "test", SessionToken: "test",
				Source: "dummy cfg for localstack",
			},
		}),
	)
	if err != nil {
		log.Fatalf("Failed to get a configuration for aws: %v", err)
	}

	return cfg
}

func CreateDynamoDbLocalClient() (*dynamodb.Client, error) {
	awsEndpoint := "http://localhost:4566"
	awsRegion := "us-east-2"

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: awsEndpoint}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "dummy cfg for localstack",
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg), err
}

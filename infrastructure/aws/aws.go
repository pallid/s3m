package aws

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	AccessKey     string
	SecretKey     string
	Bucket        string
	Prefix        string
	Delimeter     string
	PartitionID   string
	URL           string
	SigningRegion string
}

func NewS3Client(cfg Config) (*s3.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   cfg.PartitionID,
			URL:           cfg.URL,
			SigningRegion: cfg.SigningRegion,
		}, nil
	})

	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	conf, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	// Создаем клиента для доступа к хранилищу S3
	client := s3.NewFromConfig(conf)
	return client, nil
}

func GetListObjectsPaginator(client *s3.Client, bucket, prefix, delimeter string) *s3.ListObjectsV2Paginator {
	return s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimeter),
	})
}

func DeleteObjects(client *s3.Client, bucket string, deleted []types.ObjectIdentifier) error {
	_, err := client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: deleted,
		},
	})
	return err
}

func PutObject(client *s3.Client, bucket, filePath string, body io.Reader) error {
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filePath),
		Body:   body,
	})
	return err
}

package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type IS3Service interface {
	CreateBucketIfNotExists(bucketName *string) error
	GetFileContents(bucketName *string, fileName *string) ([]byte, error)
	WriteFileContents(bucketName *string, fileName *string, content []byte) error
	DoesFileExists(bucketName *string, fileName *string) bool
}

type S3Service struct {
	s3Client  *s3.Client
	awsConfig *aws.Config
}

type Slice[T any] []T

func NewS3Service() *S3Service {
	// Load the Shared AWS Configuration (~/.aws/config)
	var awsConfig, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	var client = s3.NewFromConfig(awsConfig)

	var s3Service = S3Service{
		s3Client:  client,
		awsConfig: &awsConfig,
	}

	return &s3Service
}

func (service S3Service) CreateBucketIfNotExists(bucketName *string) error {

	var doesBucketExists = service.DoesBucketExists(bucketName)
	if doesBucketExists {
		return nil
	}

	var creatBucketInput = &s3.CreateBucketInput{
		Bucket: bucketName,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(service.awsConfig.Region),
		},
	}

	var _, createBucketError = service.s3Client.CreateBucket(context.TODO(), creatBucketInput)
	if createBucketError != nil {
		return createBucketError
	}

	return nil
}

func (service S3Service) GetFileContents(bucketName *string, fileName *string) ([]byte, error) {
	var getObjectInput = &s3.GetObjectInput{
		Bucket: bucketName,
		Key:    fileName,
	}

	var getObjectOutput, getObjectError = service.s3Client.GetObject(context.TODO(), getObjectInput)

	var body, _ = io.ReadAll(getObjectOutput.Body)

	return body, getObjectError
}

func (service S3Service) WriteFileContents(bucketName *string, fileName *string, content []byte) error {
	var putObjectInput = &s3.PutObjectInput{
		Bucket: bucketName,
		Key:    fileName,
		Body:   bytes.NewReader(content),
	}

	var _, getObjectError = service.s3Client.PutObject(context.TODO(), putObjectInput)

	return getObjectError
}

func (service S3Service) DoesFileExists(bucketName *string, fileName *string) bool {
	var headObjectInput = &s3.HeadObjectInput{
		Bucket: bucketName,
		Key:    fileName,
	}

	var _, err = service.s3Client.HeadObject(context.TODO(), headObjectInput)
	if err != nil {
		var notFoundError *types.NotFound
		if errors.As(err, &notFoundError) {
			return false
		}
		log.Fatalf("Failed to retrieve object info, %v", err)
	}

	return true
}

func (service S3Service) DoesBucketExists(bucketName *string) bool {
	var headBucketInput = &s3.HeadBucketInput{
		Bucket: bucketName,
	}

	var _, err = service.s3Client.HeadBucket(context.TODO(), headBucketInput)
	if err != nil {
		var notFoundError *types.NotFound
		if errors.As(err, &notFoundError) {
			return false
		}
		log.Fatalf("Failed to retrieve object info, %v", err)
	}

	return true
}

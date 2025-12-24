package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Worker struct {
	S3Client *s3.Client
}  

func (sw* S3Worker) DownloadFile(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
	result, err := sw.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return []byte{}, err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		return []byte{}, err
	}
	return body, nil
}

func (sw* S3Worker) UploadFile(ctx context.Context, bucketName string, objectKey string, body []byte) error {
	objectKey = strings.ReplaceAll(objectKey, "-", "")
	_, err := sw.S3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
			Body:   bytes.NewReader(body),
	})
	if err != nil {
		log.Printf("Couldn't upload object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return err
	}
	return nil
}

func (sw *S3Worker) DeleteFile(ctx context.Context, bucketName string, objectKey string) error {
	_, err := sw.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key: aws.String(objectKey),
	})
	
	if err != nil {
		log.Printf("Couldn't delete object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return err
	}
	return nil
}

func (sw *S3Worker) MoveS3Object(ctx context.Context, sourceBucket, sourceKey, destinationBucket, destinationKey string) error {
	// 1. Copy the object
	copySource := url.QueryEscape(fmt.Sprintf("/%s/%s", sourceBucket, sourceKey))
	_, err := sw.S3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &destinationBucket,
		CopySource: &copySource,
		Key:        &destinationKey,
		// Optional: Add StorageClass, Metadata, etc. as needed
	})
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	// 2. Delete the source object
	_, err = sw.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &sourceBucket,
		Key:    &sourceKey,
	})
	if err != nil {
		// Log the error but consider if you want to return an error that stops the whole process
		log.Printf("Warning: failed to delete source object %s/%s: %v\n", sourceBucket, sourceKey, err)
	}

	return nil
}

//"https://storage.yandexcloud.net"
func NewS3Worker(addr string) (*S3Worker, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	cfg.BaseEndpoint = &addr

	if err != nil {
		return nil, err 
	}

	// Создаем клиента для доступа к хранилищу S3
	s3client := s3.NewFromConfig(cfg)
	sw := &S3Worker{S3Client: s3client}
	return sw, nil
}


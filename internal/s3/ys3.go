package s3

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Worker struct {
	S3Client *s3.Client
}  

func (sw S3Worker) DownloadFile(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
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


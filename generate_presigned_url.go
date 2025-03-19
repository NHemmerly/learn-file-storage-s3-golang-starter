package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	presignClient := s3.NewPresignClient(s3Client)
	presignedRequest, err := presignClient.PresignGetObject(context.Background(), input, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", fmt.Errorf("could not make presigned request: %w", err)
	}
	return presignedRequest.URL, nil
}

package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	if len(os.Args) != 4 {
		logger.Error("Invalid number of arguments",
			"expected", 3,
			"got", len(os.Args)-1,
		)
		os.Exit(1)
	}
	sourceFilePath := os.Args[1]
	destinationObjectKey := os.Args[3]

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		logger.Error("Required environment variable not set", "variable", "S3_BUCKET_NAME")
		os.Exit(1)
	}

	logger = logger.With(
		"source_file", sourceFilePath,
		"destination_bucket", bucketName,
		"destination_object", destinationObjectKey,
	)

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load AWS configuration", "error", err)
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(cfg)

	file, err := os.Open(sourceFilePath)
	if err != nil {
		logger.Error("Failed to open source file for reading", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	uploader := manager.NewUploader(s3Client)

	logger.Info("Starting file upload to S3")
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(destinationObjectKey),
		Body:   file,
	})
	if err != nil {
		logger.Error("Failed to upload file to S3", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully uploaded file to S3")
	os.Exit(0)
}

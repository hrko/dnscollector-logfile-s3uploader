package main

import (
	"context"
	"log/slog"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	EnvBucket          = "DNSC_LOGFILE_S3_BUCKET"
	EnvKeyPrefix       = "DNSC_LOGFILE_S3_KEY_PREFIX"
	EnvEndpointURL     = "DNSC_LOGFILE_S3_ENDPOINT_URL"
	EnvUsePathStyle    = "DNSC_LOGFILE_S3_USE_PATH_STYLE"
	EnvDeleteOnSuccess = "DNSC_LOGFILE_DELETE_ON_SUCCESS"
)

func main() {
	// 1. Setup a structured logger (slog).
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// 2. Validate the command-line arguments passed by postrotate-command.
	// os.Args[0] is the program name itself.
	if len(os.Args) != 4 {
		logger.Error("invalid number of arguments", "expected", 3, "got", len(os.Args)-1)
		os.Exit(1)
	}
	logFilePath := os.Args[1]
	logFileDir := os.Args[2]  // This argument is available but not used in this uploader.
	logFileName := os.Args[3] // This is the filename without the 'toprocess-' prefix.

	logger = logger.With(
		"log_file_path", logFilePath,
		"log_file_dir", logFileDir,
		"log_file_name", logFileName,
	)

	// 3. Load configuration from environment variables and expand them.
	bucket := os.ExpandEnv(os.Getenv(EnvBucket))
	if bucket == "" {
		logger.Error("S3 bucket name not specified. Please set the environment variable.", "variable", EnvBucket)
		os.Exit(1)
	}

	keyPrefix := os.ExpandEnv(os.Getenv(EnvKeyPrefix))
	endpointURL := os.ExpandEnv(os.Getenv(EnvEndpointURL))
	usePathStyle := os.Getenv(EnvUsePathStyle) == "true"
	deleteOnSuccess := os.Getenv(EnvDeleteOnSuccess) == "true"

	logger = logger.With(
		"s3_bucket", bucket,
		"s3_key_prefix", keyPrefix,
		"s3_endpoint_url", endpointURL,
		"s3_use_path_style", usePathStyle,
		"delete_on_success", deleteOnSuccess,
	)

	// 4. Load default AWS configuration and create an S3 client.
	// It automatically reads credentials from standard sources (env vars, IAM roles).
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error("failed to load AWS configuration", "error", err)
		os.Exit(1)
	}

	// Create a new S3 client, applying custom options for compatible storages.
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpointURL != "" {
			o.BaseEndpoint = aws.String(endpointURL)
		}
		o.UsePathStyle = usePathStyle
	})

	// 5. Open the rotated log file.
	file, err := os.Open(logFilePath)
	if err != nil {
		logger.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	// 6. Construct the final S3 object key and upload the file.
	// path.Join correctly handles the case where keyPrefix is empty.
	objectKey := path.Join(keyPrefix, logFileName)
	uploader := manager.NewUploader(s3Client)

	logger.Info("uploading file to S3", "key", objectKey)

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		logger.Error("failed to upload file to S3", "error", err)
		os.Exit(1)
	}

	logger.Info("file uploaded successfully")

	// 7. Delete the local file if configured to do so.
	if deleteOnSuccess {
		// We must close the file before attempting to delete it on some OSes.
		file.Close()
		logger.Info("deleting local log file", "path", logFilePath)
		if err := os.Remove(logFilePath); err != nil {
			// Log the error, but don't exit with a non-zero code,
			// as the primary goal (upload) was successful.
			logger.Error("failed to delete local log file", "error", err)
		} else {
			logger.Info("local log file deleted successfully")
		}
	}
}

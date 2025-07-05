package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var s3Client *s3.Client

// InitS3 initializes the S3 client with custom endpoint support (Cloudflare R2)
func InitS3() error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(Config.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			Config.AWSAccessKeyID,
			Config.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided (for Cloudflare R2)
	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		if Config.AWSS3Endpoint != "" {
			o.BaseEndpoint = aws.String(Config.AWSS3Endpoint)
			o.UsePathStyle = true // Required for custom endpoints like Cloudflare R2
		}
	})
	
	return nil
}

// PresignedUploadRequest represents the response for presigned upload URL
type PresignedUploadRequest struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
	ExpiresAt int64  `json:"expires_at"`
}

// GeneratePresignedUploadURL generates a presigned URL for uploading profile pictures
func GeneratePresignedUploadURL(userID primitive.ObjectID, fileExtension string) (*PresignedUploadRequest, error) {
	if s3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	// Validate file extension
	allowedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	fileExtension = strings.ToLower(fileExtension)
	isValidExtension := false
	for _, ext := range allowedExtensions {
		if fileExtension == ext {
			isValidExtension = true
			break
		}
	}
	if !isValidExtension {
		return nil, fmt.Errorf("invalid file extension: %s", fileExtension)
	}

	// Generate S3 key
	timestamp := time.Now().Unix()
	s3Key := fmt.Sprintf("profile-pictures/%s_%d%s", userID.Hex(), timestamp, fileExtension)

	// Set content type based on extension
	var contentType string
	switch fileExtension {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	default:
		contentType = "image/jpeg"
	}

	// Create presigned URL request
	presigner := s3.NewPresignClient(s3Client)
	
	// Set expiration time (5 minutes)
	expirationTime := 5 * time.Minute
	expiresAt := time.Now().Add(expirationTime).Unix()

	request, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(Config.AWSS3Bucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String(contentType),
		// Remove metadata to avoid signature issues with Cloudflare R2
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expirationTime
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &PresignedUploadRequest{
		UploadURL: request.URL,
		S3Key:     s3Key,
		ExpiresAt: expiresAt,
	}, nil
}

// GetS3ObjectURL returns the public URL for an S3 object
func GetS3ObjectURL(s3Key string) string {
	if s3Key == "" {
		return ""
	}
	
	// Use custom endpoint if provided (Cloudflare R2)
	if Config.AWSS3Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", Config.AWSS3Endpoint, Config.AWSS3Bucket, s3Key)
	}
	
	// Default AWS S3 URL format
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", Config.AWSS3Bucket, Config.AWSRegion, s3Key)
}

// DeleteS3Object deletes an object from S3
func DeleteS3Object(s3Key string) error {
	if s3Client == nil {
		return fmt.Errorf("S3 client not initialized")
	}

	if s3Key == "" {
		return nil // Nothing to delete
	}

	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(Config.AWSS3Bucket),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}

// ValidateS3Upload verifies that an upload was successful
func ValidateS3Upload(s3Key string) error {
	if s3Client == nil {
		return fmt.Errorf("S3 client not initialized")
	}

	_, err := s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(Config.AWSS3Bucket),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		return fmt.Errorf("failed to validate S3 upload: %w", err)
	}

	return nil
}

// GeneratePresignedDownloadURL generates a presigned URL for downloading files from R2
func GeneratePresignedDownloadURL(s3Key string, expirationMinutes int) (string, error) {
	if s3Client == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}

	if s3Key == "" {
		return "", fmt.Errorf("S3 key is required")
	}

	// Create presigned URL request for download
	presigner := s3.NewPresignClient(s3Client)
	
	// Set expiration time
	expirationTime := time.Duration(expirationMinutes) * time.Minute

	request, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(Config.AWSS3Bucket),
		Key:    aws.String(s3Key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expirationTime
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return request.URL, nil
}
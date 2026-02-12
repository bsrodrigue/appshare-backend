package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Storage implements the Storage interface using Cloudflare R2 (S3-compatible).
type R2Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	accountID     string
	bucketName    string
	publicDomain  string
}

// NewR2Storage creates a new R2Storage instance.
func NewR2Storage(ctx context.Context, accountID, accessKeyID, secretAccessKey, bucketName, publicDomain string) (*R2Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"), // R2 uses 'auto' for region
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	// Create S3 client with R2 endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})
	presignClient := s3.NewPresignClient(client)

	return &R2Storage{
		client:        client,
		presignClient: presignClient,
		accountID:     accountID,
		bucketName:    bucketName,
		publicDomain:  publicDomain,
	}, nil
}

// GenerateUploadURL generates a signed URL for uploading a file via PUT.
func (s *R2Storage) GenerateUploadURL(ctx context.Context, path string, expires time.Duration) (string, error) {
	request, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(path),
		ContentType: aws.String("application/octet-stream"),
	}, s3.WithPresignExpires(expires))

	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return request.URL, nil
}

// Delete removes a file from the bucket.
func (s *R2Storage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetPublicURL returns the public URL of the object.
// If publicDomain is provided, it uses that. Otherwise, it returns the standard R2 dev domain if enabled,
// though usually R2 buckets are not public by default.
func (s *R2Storage) GetPublicURL(path string) string {
	if s.publicDomain != "" {
		return fmt.Sprintf("%s/%s", s.publicDomain, path)
	}
	// Fallback to the R2 API endpoint format (not typically used for public access but consistent)
	return fmt.Sprintf("%s.r2.cloudflarestorage.com/%s/%s", s.accountID, s.bucketName, path)
}

// Download returns a reader for the file at the given path.
func (s *R2Storage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", path, err)
	}

	return output.Body, nil
}

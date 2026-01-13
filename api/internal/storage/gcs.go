package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
)

// GCSClient implements ObjectStorage for Google Cloud Storage
type GCSClient struct {
	client     *storage.Client
	bucketName string
}

// NewGCSClient creates a new GCS client
func NewGCSClient(ctx context.Context, bucketName string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Close closes the GCS client
func (c *GCSClient) Close() error {
	return c.client.Close()
}

func (c *GCSClient) GenerateUploadURL(ctx context.Context, path string, contentType string, expiresIn time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     time.Now().Add(expiresIn),
		ContentType: contentType,
	}

	url, err := c.client.Bucket(c.bucketName).SignedURL(path, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return url, nil
}

func (c *GCSClient) GenerateDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiresIn),
	}

	url, err := c.client.Bucket(c.bucketName).SignedURL(path, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

func (c *GCSClient) Delete(ctx context.Context, path string) error {
	obj := c.client.Bucket(c.bucketName).Object(path)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (c *GCSClient) Exists(ctx context.Context, path string) (bool, error) {
	obj := c.client.Bucket(c.bucketName).Object(path)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object: %w", err)
	}
	return true, nil
}

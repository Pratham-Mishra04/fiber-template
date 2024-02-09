package helpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/gofiber/fiber/v2"
)

type BucketClient struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var UserProfileClient *BucketClient
var UserCoverClient *BucketClient

func createNewBucketClient(uploadPath string, private bool) *BucketClient {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", initializers.CONFIG.GCP_CREDS)
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if private {
		return &BucketClient{
			cl:         client,
			bucketName: initializers.CONFIG.GCP_PRIVATE_BUCKET,
			projectID:  initializers.CONFIG.GCP_PROJECT,
			uploadPath: uploadPath,
		}
	}

	return &BucketClient{
		cl:         client,
		bucketName: initializers.CONFIG.GCP_PUBLIC_BUCKET,
		projectID:  initializers.CONFIG.GCP_PROJECT,
		uploadPath: uploadPath,
	}
}

func InitializeBucketClients() {
	UserProfileClient = createNewBucketClient("users/profilePics/", false)
	UserCoverClient = createNewBucketClient("users/coverPics/", false)
}

func (c *BucketClient) UploadBucketFile(buffer *bytes.Buffer, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	reader := bytes.NewReader(buffer.Bytes())

	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}

func (c *BucketClient) DeleteBucketFile(fileName string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	if err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + fileName).Delete(ctx); err != nil {
		return &fiber.Error{Code: 500, Message: fmt.Sprintf("Failed to delete file: %v", err)}
	}

	return nil
}

func (c *BucketClient) GetBucketFile(fileName string) (*bytes.Buffer, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + fileName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %v", err)
	}
	defer rc.Close()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, rc); err != nil {
		return nil, fmt.Errorf("io.Copy: %v", err)
	}

	return &buffer, nil
}

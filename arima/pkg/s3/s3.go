// Package s3 提供 S3 对象存储操作
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/akagiyui/go-together/arima/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client S3 客户端
type Client struct {
	client *s3.Client
	bucket string
}

// NewClient 创建新的 S3 客户端
func NewClient(cfg config.Config) (*Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               cfg.S3Endpoint,
			SigningRegion:     cfg.S3Region,
			HostnameImmutable: true,
		}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			"",
		)),
		awsconfig.WithRegion(cfg.S3Region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO 需要
	})

	return &Client{
		client: client,
		bucket: cfg.S3Bucket,
	}, nil
}

// HeadBucket 检查 bucket 是否存在
func (c *Client) HeadBucket(ctx context.Context) error {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	return err
}

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, key string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := c.client.PutObject(ctx, input)
	return err
}

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, key string) ([]byte, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// GenerateDownloadURL 生成预签名下载 URL
func (c *Client) GenerateDownloadURL(ctx context.Context, key string, ttl time.Duration, filename string) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	if filename != "" {
		encodedFilename := url.QueryEscape(filename)
		disposition := fmt.Sprintf(`attachment; filename="%s"`, encodedFilename)
		input.ResponseContentDisposition = aws.String(disposition)
	}

	presignedReq, err := presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}

// IsHealthy 检查 S3 服务是否健康
func (c *Client) IsHealthy(ctx context.Context) bool {
	return c.HeadBucket(ctx) == nil
}


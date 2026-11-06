package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Provider retrieves snapshots from Amazon S3, Cloudflare R2, or MinIO.
type S3Provider struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3Provider instantiates an S3 storage provider.
func NewS3Provider(ctx context.Context, cfg config.S3StorageConfig) (*S3Provider, error) {
	var optFns []func(*awsconfig.LoadOptions) error

	if cfg.Region != "" {
		optFns = append(optFns, awsconfig.WithRegion(cfg.Region))
	} else {
		optFns = append(optFns, awsconfig.WithRegion("us-east-1"))
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
		optFns = append(optFns, awsconfig.WithCredentialsProvider(creds))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
	})

	return &S3Provider{
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}, nil
}

func (s *S3Provider) resolveKey(key string) string {
	if s.prefix == "" {
		return key
	}
	cleanPrefix := strings.TrimSuffix(s.prefix, "/")
	cleanKey := strings.TrimPrefix(key, "/")
	return cleanPrefix + "/" + cleanKey
}

func (s *S3Provider) Fetch(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	fullKey := s.resolveKey(key)

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("s3 get object %s/%s: %w", s.bucket, fullKey, err)
	}

	var size int64
	if out.ContentLength != nil {
		size = *out.ContentLength
	}

	return out.Body, size, nil
}

func (s *S3Provider) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.resolveKey(key)
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err == nil {
		return true, nil
	}
	return false, nil
}

func (s *S3Provider) Type() string {
	return "s3"
}

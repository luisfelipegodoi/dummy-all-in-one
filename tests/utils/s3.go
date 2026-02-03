package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// S3Client wraps AWS SDK S3 client configured for LocalStack.
type S3Client struct {
	Client *s3.Client
}

// NewS3 creates a LocalStack-ready S3 client.
//
// endpoint examples:
//   - "http://localhost:4566"
//   - "http://127.0.0.1:31566" (NodePort)
//   - "http://localstack.localstack.svc.cluster.local:4566" (inside cluster)
//
// LocalStack often works best with PathStyle (UsePathStyle=true).
func NewS3(ctx context.Context, region, endpoint string) (*S3Client, error) {
	if region == "" {
		region = "us-east-1"
	}
	if endpoint == "" {
		return nil, errors.New("s3 endpoint is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Client{Client: client}, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (c *S3Client) EnsureBucket(ctx context.Context, bucket string) error {
	if strings.TrimSpace(bucket) == "" {
		return errors.New("bucket name is empty")
	}

	ok, err := c.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	_, err = c.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// BucketExists checks if bucket exists.
func (c *S3Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	_, err := c.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		return true, nil
	}

	// HeadBucket returns generic error types depending on signer/transport.
	// Treat 404/NoSuchBucket as "not exists".
	if isSmithyHTTPStatus(err, 404) {
		return false, nil
	}
	return false, err
}

// PutObject uploads bytes to s3://bucket/key
func (c *S3Client) PutObject(ctx context.Context, bucket, key string, data []byte, contentType string) error {
	if strings.TrimSpace(bucket) == "" {
		return errors.New("bucket is empty")
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("key is empty")
	}
	body := bytes.NewReader(data)

	in := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}

	_, err := c.Client.PutObject(ctx, in)
	return err
}

// PutObjectFromReader uploads streaming content to s3://bucket/key
func (c *S3Client) PutObjectFromReader(ctx context.Context, bucket, key string, r io.Reader, contentType string) error {
	if strings.TrimSpace(bucket) == "" {
		return errors.New("bucket is empty")
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("key is empty")
	}
	if r == nil {
		return errors.New("reader is nil")
	}

	in := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}

	_, err := c.Client.PutObject(ctx, in)
	return err
}

// ObjectExists checks if s3://bucket/key exists.
func (c *S3Client) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := c.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	if isSmithyHTTPStatus(err, 404) {
		return false, nil
	}
	return false, err
}

// GetObjectBytes downloads s3://bucket/key and returns content as []byte.
func (c *S3Client) GetObjectBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := c.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	return io.ReadAll(out.Body)
}

// "Editar bucket" - exemplos úteis (opcionais):

// SetBucketVersioning enables/disables bucket versioning.
func (c *S3Client) SetBucketVersioning(ctx context.Context, bucket string, enabled bool) error {
	status := s3.BucketVersioningStatusSuspended
	if enabled {
		status = s3.BucketVersioningStatusEnabled
	}

	_, err := c.Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: status,
		},
	})
	return err
}

// PutBucketTags sets bucket tags (overwrites).
func (c *S3Client) PutBucketTags(ctx context.Context, bucket string, tags map[string]string) error {
	var t []s3typesTag
	for k, v := range tags {
		t = append(t, s3typesTag{Key: k, Value: v})
	}

	_, err := c.Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucket),
		Tagging: &s3.Tagging{
			TagSet: toSDKTags(t),
		},
	})
	return err
}

// ---------- internals ----------

// Minimal local tag struct to avoid importing extra package names in your code.
type s3typesTag struct {
	Key   string
	Value string
}

// Convert to SDK type without leaking import details in your test code.
func toSDKTags(in []s3typesTag) []s3typesTagSDK {
	out := make([]s3typesTagSDK, 0, len(in))
	for _, t := range in {
		out = append(out, s3typesTagSDK{Key: aws.String(t.Key), Value: aws.String(t.Value)})
	}
	return out
}

// Local alias of SDK tag type (keeps file self-contained).
type s3typesTagSDK = s3.Tag

func isSmithyHTTPStatus(err error, code int) bool {
	var re *smithy.OperationError
	if errors.As(err, &re) {
		// unwrap
		err = re.Unwrap()
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// sometimes has status embedded elsewhere
	}
	var httpErr interface{ HTTPStatusCode() int }
	if errors.As(err, &httpErr) {
		return httpErr.HTTPStatusCode() == code
	}
	// fallback: many LocalStack errors still show "StatusCode: 404"
	return strings.Contains(err.Error(), fmt.Sprintf("StatusCode: %d", code))
}

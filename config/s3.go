// config/s3.go
package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client wraps s3.Client and s3.PresignClient
type S3Client struct {
	Client  *s3.Client
	Presign *s3.PresignClient
}

// ConnectS3 creates a new S3 client (compatible with R2, MinIO, etc.)
func ConnectS3(cfg *Config) *S3Client {
	loaded, err := awsCfg.LoadDefaultConfig(context.Background(),
		awsCfg.WithRegion(cfg.S3_REGION),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3_ACCESS_KEY_ID,
			cfg.S3_SECRET_ACCESS_KEY,
			"",
		)),
		awsCfg.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				// Force S3 client to use custom endpoint (R2, MinIO, etc.)
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               cfg.S3_ENDPOINT_URL,
						SigningRegion:     cfg.S3_REGION,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
	if err != nil {
		panic("S3_NOT_CONNECTED: " + err.Error())
	}

	client := s3.NewFromConfig(loaded, func(o *s3.Options) {
		// Path-style is safer for S3-compatible services (R2, MinIO)
		o.UsePathStyle = true
	})

	return &S3Client{
		Client:  client,
		Presign: s3.NewPresignClient(client),
	}
}

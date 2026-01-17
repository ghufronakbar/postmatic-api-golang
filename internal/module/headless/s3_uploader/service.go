// internal/module/headless/s3_uploader/service.go
package s3_uploader

import (
	"context"
	"errors"
	"fmt"
	"postmatic-api/config"
	"postmatic-api/pkg/errs"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type S3UploaderService struct {
	cfg     *config.Config
	s3      *s3.Client
	presign *s3.PresignClient
}

func NewService(cfg *config.Config) *S3UploaderService {
	loaded, err := awsCfg.LoadDefaultConfig(context.Background(),
		awsCfg.WithRegion(cfg.S3_REGION),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3_ACCESS_KEY_ID,
			cfg.S3_SECRET_ACCESS_KEY,
			"",
		)),
		awsCfg.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				// Force S3 client to use R2 endpoint
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
		panic("Cannot connect to S3" + err.Error())
	}

	client := s3.NewFromConfig(loaded, func(o *s3.Options) {
		// Banyak S3-compatible (termasuk R2) lebih aman pakai path-style.
		o.UsePathStyle = true
	})

	return &S3UploaderService{
		cfg:     cfg,
		s3:      client,
		presign: s3.NewPresignClient(client),
	}
}

// objectKeyBuilder:
// kamu ingin public_id jadi path untuk s3 -> kita jadikan key deterministik berbasis hash
func (s *S3UploaderService) objectKeyBuilder(hash string, format string) string {
	// contoh: postmatic/images/<hash>.png
	// kamu bebas ubah prefixnya sesuai kebutuhan
	return fmt.Sprintf("%s/images/%s.%s", s.cfg.APP_NAME, hash, format)
}

func (s *S3UploaderService) BuildObjectURL(objectKey string) string {
	key := strings.TrimLeft(objectKey, "/")

	if s.cfg.S3_PUBLIC_BASE_URL != "" {
		base := strings.TrimRight(s.cfg.S3_PUBLIC_BASE_URL, "/")
		return fmt.Sprintf("%s/%s", base, key)
	}

	// fallback (bukan public)
	base := strings.TrimRight(s.cfg.S3_ENDPOINT_URL, "/")
	return fmt.Sprintf("%s/%s/%s", base, s.cfg.S3_BUCKET, key)
}

func (s *S3UploaderService) PresignUploadImage(ctx context.Context, input PresignUploadImageInput) (*PresignUploadImageResponse, error) {
	if input.Hash == "" || input.Format == "" || input.ContentType == "" {
		return nil, errs.NewBadRequest("HASH_FORMAT_CONTENTTYPE_REQUIRED")
	}

	objectKey := s.objectKeyBuilder(input.Hash, input.Format)

	req := &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.S3_BUCKET),
		Key:         aws.String(objectKey),
		ContentType: aws.String(input.ContentType),
		IfNoneMatch: aws.String("*"),
	}

	ps, err := s.presign.PresignPutObject(ctx, req, func(po *s3.PresignOptions) {
		po.Expires = s.cfg.S3_PRESIGN_EXPIRES_SECONDS
	})
	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}

	return &PresignUploadImageResponse{
		Provider:  "s3",
		Bucket:    s.cfg.S3_BUCKET,
		PublicId:  objectKey,
		UploadUrl: ps.URL,
		Headers: map[string]string{
			"Content-Type":  input.ContentType,
			"If-None-Match": "*",
		},
		ExpiresInSeconds: int64(s.cfg.S3_PRESIGN_EXPIRES_SECONDS.Seconds()),
	}, nil
}

func (s *S3UploaderService) ObjectExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.S3_BUCKET),
		Key:    aws.String(objectKey),
	})
	if err == nil {
		return true, nil
	}

	// R2/S3-compatible biasanya balikin code NotFound / NoSuchKey via smithy.APIError
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		if code == "NotFound" || code == "NoSuchKey" {
			return false, nil
		}
	}

	return false, err
}

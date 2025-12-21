package cloudinary_uploader

import (
	"context"
	"io"
	"postmatic-api/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploaderService struct {
	cfg *config.Config
	cld *cloudinary.Cloudinary
}

func NewService(cfg *config.Config) *CloudinaryUploaderService {
	cld, err := cloudinary.NewFromParams(cfg.CLOUDINARY_CLOUD_NAME, cfg.CLOUDINARY_API_KEY, cfg.CLOUDINARY_API_SECRET)
	if err != nil {
		return nil
	}
	return &CloudinaryUploaderService{
		cfg: cfg,
		cld: cld,
	}
}

func (s *CloudinaryUploaderService) UploadSingleImage(ctx context.Context, file io.Reader, params uploader.UploadParams) (*CloudinaryUploadSingleImageResponse, error) {
	result, err := s.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		return nil, err
	}
	return &CloudinaryUploadSingleImageResponse{
		PublicId: result.PublicID,
		ImageUrl: result.SecureURL,
	}, nil
}

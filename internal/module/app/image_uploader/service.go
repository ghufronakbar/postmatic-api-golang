// internal/module/app/image_uploader/service.go
package image_uploader

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"postmatic-api/internal/module/headless/cloudinary_uploader"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/hash"
)

type ImageUploaderService struct {
	cld   *cloudinary_uploader.CloudinaryUploaderService
	store entity.Store
}

func NewImageUploaderService(cld *cloudinary_uploader.CloudinaryUploaderService, store entity.Store) *ImageUploaderService {
	return &ImageUploaderService{cld: cld, store: store}
}

func (s *ImageUploaderService) UploadSingleImage(ctx context.Context, file io.Reader) (ImageUploaderViewModel, error) {
	hashKey, size, err := hash.HashFileToSHA256(file)
	if err != nil {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(err)
	}

	// ✅ rewind setelah hashing
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return ImageUploaderViewModel{}, errs.NewInternalServerError(err)
		}
	} else {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(errors.New("FILE_NOT_SEEKABLE"))
	}

	check, err := s.store.GetUploadedImageByHashkey(ctx, hashKey)
	if err != nil && err != sql.ErrNoRows {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(err)
	}
	if check.ID != 0 && check.PublicID != "" && check.ImageUrl != "" {
		return ImageUploaderViewModel{
			Hashkey:     hashKey,
			IsDuplicate: true,
			PublicId:    check.PublicID,
			Size:        check.Size,
			ImageUrl:    check.ImageUrl,
			ID:          check.ID,
			Format:      check.Format,
			Provider:    string(check.Provider),
		}, nil
	}

	result, err := s.cld.UploadSingleImage(ctx, file)
	if err != nil {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(err)
	}
	if result == nil || result.ImageUrl == "" || result.PublicId == "" {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(errors.New("CLOUDINARY_RETURNED_EMPTY_URL"))
	}

	inserted, err := s.store.InsertUploadedImage(ctx, entity.InsertUploadedImageParams{
		Hashkey:  hashKey,
		PublicID: result.PublicId,
		ImageUrl: result.ImageUrl,
		Size:     size,
		Provider: entity.ImageProviderCloudinary,
		Format:   result.Format,
	})
	if err != nil {
		return ImageUploaderViewModel{}, errs.NewInternalServerError(err)
	}

	return ImageUploaderViewModel{
		PublicId:    result.PublicId,
		ImageUrl:    result.ImageUrl,
		Hashkey:     hashKey,
		IsDuplicate: false,
		Size:        size,
		Provider:    string(inserted.Provider),
		ID:          inserted.ID,
	}, nil
}

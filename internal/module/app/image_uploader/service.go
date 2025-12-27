// internal/module/app/image_uploader/service.go
package image_uploader

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"postmatic-api/internal/module/headless/cloudinary_uploader"
	"postmatic-api/internal/module/headless/s3_uploader"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/hash"

	"github.com/lib/pq"
)

type ImageUploaderService struct {
	cld   *cloudinary_uploader.CloudinaryUploaderService
	s3    *s3_uploader.S3UploaderService
	store entity.Store
}

func NewImageUploaderService(cld *cloudinary_uploader.CloudinaryUploaderService, s3 *s3_uploader.S3UploaderService, store entity.Store) *ImageUploaderService {
	return &ImageUploaderService{cld: cld, s3: s3, store: store}
}

func (s *ImageUploaderService) UploadSingleImage(ctx context.Context, file io.Reader) (ImageUploaderResponse, error) {
	hashKey, size, err := hash.HashFileToSHA256(file)
	if err != nil {
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}

	// ✅ rewind setelah hashing
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return ImageUploaderResponse{}, errs.NewInternalServerError(err)
		}
	} else {
		return ImageUploaderResponse{}, errs.NewInternalServerError(errors.New("FILE_NOT_SEEKABLE"))
	}

	check, err := s.store.GetUploadedImageByHashkey(ctx, hashKey)
	if err != nil && err != sql.ErrNoRows {
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}
	if check.ID != 0 && check.PublicID != "" && check.ImageUrl != "" {
		return ImageUploaderResponse{
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
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}
	if result == nil || result.ImageUrl == "" || result.PublicId == "" {
		return ImageUploaderResponse{}, errs.NewInternalServerError(errors.New("CLOUDINARY_RETURNED_EMPTY_URL"))
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
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}

	return ImageUploaderResponse{
		PublicId:    result.PublicId,
		ImageUrl:    result.ImageUrl,
		Hashkey:     hashKey,
		IsDuplicate: false,
		Size:        size,
		Provider:    string(inserted.Provider),
		ID:          inserted.ID,
	}, nil
}

func (h *ImageUploaderService) PresignUploadImage(ctx context.Context, req s3_uploader.PresignUploadImageInput) (ImageUploaderResponse, error) {
	// Validasi minimal supaya insert DB tidak gagal (size NOT NULL)
	if req.Hash == "" || req.Format == "" || req.ContentType == "" {
		return ImageUploaderResponse{}, errs.NewBadRequest("HASH_FORMAT_CONTENTTYPE_REQUIRED")
	}
	if req.Size <= 0 {
		return ImageUploaderResponse{}, errs.NewBadRequest("SIZE_REQUIRED")
	}

	// 1) cek duplicate by hash
	check, err := h.store.GetUploadedImageByHashkey(ctx, req.Hash)
	if err != nil && err != sql.ErrNoRows {
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}

	if check.ID != 0 && check.PublicID != "" && check.ImageUrl != "" {
		// ✅ khusus s3: cek object beneran sudah ada atau belum
		if check.Provider == entity.ImageProviderS3 {
			exists, err := h.s3.ObjectExists(ctx, check.PublicID)
			if err != nil {
				return ImageUploaderResponse{}, errs.NewInternalServerError(err)
			}
			if !exists {
				// upload belum terjadi / gagal, jadi kasih presign lagi
				presigned, err := h.s3.PresignUploadImage(ctx, req)
				if err != nil {
					return ImageUploaderResponse{}, err
				}

				publicURL := h.s3.BuildObjectURL(check.PublicID)

				return ImageUploaderResponse{
					Hashkey:     req.Hash,
					IsDuplicate: false, // ini sebenarnya "retry", bukan duplicate
					PublicId:    check.PublicID,
					Size:        check.Size,
					ImageUrl:    check.ImageUrl,
					ID:          check.ID,
					Format:      check.Format,
					Provider:    string(check.Provider),

					// UploadUrl:        presigned.UploadUrl,
					UploadUrl:        publicURL,
					Headers:          presigned.Headers,
					ExpiresInSeconds: presigned.ExpiresInSeconds,
					Bucket:           presigned.Bucket,
				}, nil
			}
		}

		// default: duplicate true (cloudinary, atau s3 object sudah ada)
		return ImageUploaderResponse{
			Hashkey:     req.Hash,
			IsDuplicate: true,
			PublicId:    check.PublicID,
			Size:        check.Size,
			ImageUrl:    check.ImageUrl,
			ID:          check.ID,
			Format:      check.Format,
			Provider:    string(check.Provider),
		}, nil
	}

	// 2) generate presign via headless
	presigned, err := h.s3.PresignUploadImage(ctx, req)
	if err != nil {
		return ImageUploaderResponse{}, err
	}
	if presigned == nil || presigned.UploadUrl == "" || presigned.PublicId == "" {
		return ImageUploaderResponse{}, errs.NewInternalServerError(errors.New("S3_PRESIGN_RETURNED_EMPTY_URL"))
	}

	// 3) build image_url (sementara / canonical)
	imageUrl := h.s3.BuildObjectURL(presigned.PublicId)

	// 4) insert row agar dedup bekerja dari awal
	inserted, err := h.store.InsertUploadedImage(ctx, entity.InsertUploadedImageParams{
		Hashkey:  req.Hash,
		PublicID: presigned.PublicId, // objectKey disimpan di public_id
		ImageUrl: imageUrl,
		Size:     req.Size,
		Provider: entity.ImageProviderS3, // pastikan sqlc sudah regen setelah enum 's3'
		Format:   req.Format,
	})
	if err != nil {
		// kalau race condition (request bersamaan), handle unique violation lalu return duplicate
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			exist, e2 := h.store.GetUploadedImageByHashkey(ctx, req.Hash)
			if e2 != nil && e2 != sql.ErrNoRows {
				return ImageUploaderResponse{}, errs.NewInternalServerError(e2)
			}
			if exist.ID != 0 && exist.PublicID != "" && exist.ImageUrl != "" {
				return ImageUploaderResponse{
					Hashkey:     req.Hash,
					IsDuplicate: true,
					PublicId:    exist.PublicID,
					Size:        exist.Size,
					ImageUrl:    exist.ImageUrl,
					ID:          exist.ID,
					Format:      exist.Format,
					Provider:    string(exist.Provider),
				}, nil
			}
		}
		return ImageUploaderResponse{}, errs.NewInternalServerError(err)
	}

	// 5) return response (include presign payload)
	return ImageUploaderResponse{
		PublicId:    presigned.PublicId,
		ImageUrl:    imageUrl,
		Hashkey:     req.Hash,
		IsDuplicate: false,
		Size:        req.Size,
		Format:      req.Format,
		Provider:    string(inserted.Provider),
		ID:          inserted.ID,

		// presign fields (optional)
		UploadUrl:        presigned.UploadUrl,
		Headers:          presigned.Headers,
		ExpiresInSeconds: presigned.ExpiresInSeconds,
		Bucket:           presigned.Bucket,
	}, nil
}

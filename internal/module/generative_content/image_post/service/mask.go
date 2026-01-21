// internal/module/generative_content/image_post/service/mask.go
package image_post_service

import (
	"context"

	"postmatic-api/pkg/errs"
)

// handleMask handles mode=mask
// STUB: Phase 4 implementation
func (s *ImagePostService) handleMask(ctx context.Context, input CreateImagePostInput) (CreateImagePostResponse, error) {
	return CreateImagePostResponse{}, errs.NewBadRequest("METHOD_NOT_IMPLEMENTED")
}

// internal/module/generative_content/image_post/service/regenerate.go
package image_post_service

import (
	"context"

	"postmatic-api/pkg/errs"
)

// handleRegenerate handles mode=regenerate
// STUB: Phase 2 implementation
func (s *ImagePostService) handleRegenerate(ctx context.Context, input CreateImagePostInput) (CreateImagePostResponse, error) {
	return CreateImagePostResponse{}, errs.NewBadRequest("METHOD_NOT_IMPLEMENTED")
}

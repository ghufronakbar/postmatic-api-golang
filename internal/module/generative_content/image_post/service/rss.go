// internal/module/generative_content/image_post/service/rss.go
package image_post_service

import (
	"context"

	"postmatic-api/pkg/errs"
)

// handleRss handles mode=rss
// STUB: Phase 3 implementation
func (s *ImagePostService) handleRss(ctx context.Context, input CreateImagePostInput) (CreateImagePostResponse, error) {
	return CreateImagePostResponse{}, errs.NewBadRequest("METHOD_NOT_IMPLEMENTED")
}

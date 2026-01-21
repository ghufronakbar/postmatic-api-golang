// internal/module/generative_content/image_post/service/caption.go
package image_post_service

// NOTE: Caption generation akan diimplementasi di queue worker
// karena ini adalah proses async yang berjalan di background

// CaptionResult adalah hasil dari caption generation
type CaptionResult struct {
	CaptionText string
	TokenUsed   int
}

// internal/module/headless/queue/generative_content_image_post.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	image_post_service "postmatic-api/internal/module/generative_content/image_post/service"

	"github.com/hibiken/asynq"
)

// ImagePostProducer adalah kontrak untuk menambahkan job image post ke queue
type ImagePostProducer interface {
	EnqueueImagePostGenerate(ctx context.Context, payload image_post_service.ImagePostQueuePayload) error
}

// Task type untuk image post
const (
	taskImagePostGenerate = "queue:generative:image:post"
)

// EnqueueImagePostGenerate menambahkan job generate image post ke queue
func (p *Producer) EnqueueImagePostGenerate(ctx context.Context, payload image_post_service.ImagePostQueuePayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskImagePostGenerate, b)

	return p.enqueue(
		ctx,
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(120*time.Second), // Longer timeout for AI generation
	)
}

// ImagePostHandler adalah service yang menangani job image post
type ImagePostHandler interface {
	// ProcessImagePostJob processes the image post job (called by worker)
	ProcessImagePostJob(ctx context.Context, postID int64, businessRootID int64) error
}

// registerImagePostHandlers register handler untuk image post jobs
func registerImagePostHandlers(mux *asynq.ServeMux, handler ImagePostHandler) {
	mux.HandleFunc(taskImagePostGenerate, func(ctx context.Context, t *asynq.Task) error {
		var p image_post_service.ImagePostQueuePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invalid payload: %v: %w", err, asynq.SkipRetry)
		}
		return handler.ProcessImagePostJob(ctx, p.PostID, p.BusinessRootID)
	})
}

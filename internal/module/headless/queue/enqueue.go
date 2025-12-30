// internal/module/headless/queue/enqueue.go
package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

// enqueue adalah helper universal untuk mendorong task ke Redis via asynq.Client.
// Semua opsi (queue name, retry, timeout, unique, schedule, dll) dipass lewat opts.
// Dengan begini, tiap domain (mailer/whatsapp/...) bebas menentukan policy-nya masing-masing.
func (p *Producer) enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) error {
	_, err := p.client.EnqueueContext(ctx, task, opts...)
	return err
}

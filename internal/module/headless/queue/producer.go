// internal/module/headless/queue/producer.go
package queue

import "github.com/hibiken/asynq"

// Producer hanya bertugas enqueue task ke Redis (tidak execute pekerjaan).
type Producer struct {
	client *asynq.Client
}

func NewProducer(client *asynq.Client) *Producer {
	return &Producer{client: client}
}

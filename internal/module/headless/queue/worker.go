// internal/module/headless/queue/worker.go
package queue

import "github.com/hibiken/asynq"

type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewWorker(server *asynq.Server) *Worker {
	return &Worker{server: server, mux: asynq.NewServeMux()}
}

func (w *Worker) RegisterMailer(mailerSvc MailerService) {
	registerMailerHandlers(w.mux, mailerSvc) // welcome + verification sama-sama di sini
}

func (w *Worker) Run() error {
	return w.server.Run(w.mux)
}

// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal"
	"postmatic-api/internal/internal_middleware"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/pkg/logger"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
)

func main() {
	cfg := config.Load()

	db, err := config.ConnectDB(cfg.DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Connect to Redis
	rdb, err := config.ConnectRedis(cfg)
	if err != nil {
		log.Fatal("Cannot connect to Redis: " + err.Error())
	}

	// producer (enqueue)
	asynqClient := config.NewAsynqClient(cfg)
	defer asynqClient.Close()

	// worker (dequeue)
	mailerSvc := mailer.NewService(cfg)
	asynqServer := config.NewAsynqServer(cfg, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"default": 1,
		},
	})
	go func() {
		w := queue.NewWorker(asynqServer)
		w.RegisterMailer(mailerSvc) // ✅ tanpa *
		if err := w.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	// HTTP router root
	r := chi.NewRouter()
	r.Use(chiMw.RequestID)
	r.Use(internal_middleware.RequestLogger)
	r.Use(chiMw.StripSlashes)
	r.Use(chiMw.Recoverer)

	// inject cfg + asynqClient + rdb
	r.Mount("/api", internal.NewRouter(db, cfg, asynqClient, rdb))

	srv := &http.Server{
		Addr:    ":" + cfg.PORT,
		Handler: r,
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server running on port", cfg.PORT)
		chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			logger.L().Info("route", "method", method, "route", route)
			return nil
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)
	asynqServer.Shutdown()
}

// config/asynq.go
package config

import (
	"fmt"

	"github.com/hibiken/asynq"
)

func AsynqRedisOpt(cfg *Config) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.REDIS_HOST, cfg.REDIS_PORT),
		Password: cfg.REDIS_PASS,
		DB:       cfg.REDIS_DB,
	}
}

func NewAsynqClient(cfg *Config) *asynq.Client {
	return asynq.NewClient(AsynqRedisOpt(cfg))
}

func NewAsynqServer(cfg *Config, asynqCfg asynq.Config) *asynq.Server {
	return asynq.NewServer(AsynqRedisOpt(cfg), asynqCfg)
}

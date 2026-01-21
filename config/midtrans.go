// config/midtrans.go
package config

import (
	"postmatic-api/pkg/logger"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

func ConnectMidtrans(cfg *Config) *coreapi.Client {
	var env midtrans.EnvironmentType
	if cfg.MIDTRANS_IS_PRODUCTION {
		env = midtrans.Production
	} else {
		env = midtrans.Sandbox
	}

	client := coreapi.Client{}
	client.New(cfg.MIDTRANS_SERVER_KEY, env)

	envStr := "sandbox"
	if cfg.MIDTRANS_IS_PRODUCTION {
		envStr = "production"
	}
	logger.L().Info("Midtrans service initialized", "environment", envStr)

	return &client
}

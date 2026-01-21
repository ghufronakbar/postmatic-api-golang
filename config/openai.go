// config/openai.go
package config

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ConnectOpenAI creates a new OpenAI client
func ConnectOpenAI(cfg *Config) openai.Client {
	client := openai.NewClient(
		option.WithAPIKey(cfg.OPENAI_API_KEY),
	)

	return client
}

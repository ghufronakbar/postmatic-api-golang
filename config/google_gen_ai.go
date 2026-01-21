// config/google_gen_ai.go
package config

import (
	"context"

	"google.golang.org/genai"
)

// ConnectGoogleGenAI creates a new Google GenAI client using Gemini API backend
func ConnectGoogleGenAI(cfg *Config) *genai.Client {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GOOGLE_GENAI_API_KEY,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic("GOOGLE_GENAI_NOT_SET: " + err.Error())
	}

	return client
}

package openai

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm"
)

type OpenAIConfig struct {
	ModelName   string  `json:"model_name"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
}

type OpenAI struct {
}

func NewOpenAI(config OpenAIConfig) *OpenAI {
	return &OpenAI{}
}

func (o *OpenAI) Response(ctx context.Context, dialogues []llm.Dialogue) (string, error) {
	return "", nil
}

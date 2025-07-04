package openai

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm"
	goopenai "github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	ModelName string `json:"model_name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
}

type OpenAI struct {
	modelName string
	baseURL   string
	apiKey    string
	client    *goopenai.Client
}

func NewOpenAI(modelName, apiKey, baseUrl string) *OpenAI {
	client := &OpenAI{
		modelName: modelName,
		apiKey:    apiKey,
		baseURL:   baseUrl,
	}

	cfg := goopenai.DefaultConfig(apiKey)
	cfg.BaseURL = baseUrl
	client.client = goopenai.NewClientWithConfig(cfg)

	return client
}

// TODO tools to be added - then function call
func (o *OpenAI) Response(ctx context.Context, dialogues []llm.Dialogue) (string, error) {
	request := goopenai.ChatCompletionRequest{
		Model: o.modelName,
	}

	println("OpenAI model name:", o.modelName)

	for _, dialogue := range dialogues {
		request.Messages = append(request.Messages, goopenai.ChatCompletionMessage{
			Role:    dialogue.Role,
			Content: dialogue.Content,
		})
	}
	println("OpenAI request messages:", len(request.Messages))

	resp, err := o.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", nil
	}
	println(resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content, nil
}

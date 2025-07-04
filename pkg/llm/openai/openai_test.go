package openai

import (
	"context"
	"os"
	"testing"

	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm"
	"github.com/stretchr/testify/assert"
)

func buildOpenAIClient() *OpenAI {
	cli := NewOpenAI(
		os.Getenv("OPENAI_MODEL_NAME"),
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("OPENAI_BASE_URL"),
	)

	return cli
}

const systremPrompt = `You are a helpful assistant. Your sole responsbility is to return what I said to you, nothing else.`

func TestCreateClient(t *testing.T) {
	c := buildOpenAIClient()
	assert.NotNil(t, c, "OpenAI client should not be nil")
}

func TestPingRequest(t *testing.T) {
	c := buildOpenAIClient()
	assert.NotNil(t, c, "OpenAI client should not be nil")

	ctx := context.Background()

	dialogs := []llm.Dialogue{
		{Role: "system", Content: systremPrompt},
		{Role: "user", Content: "pong"},
	}
	pong, err := c.Response(ctx, dialogs)

	assert.NoError(t, err, "OpenAI ping request should not return an error")
	assert.NotEmpty(t, pong, "OpenAI ping request should return a response")
	assert.Equal(t, "pong", pong, "OpenAI ping request should return 'pong'")
}

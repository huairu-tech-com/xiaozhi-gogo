package src

import (
	"context"

	"github.com/huairu-tech-com/xiaozhi-gogo/config"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm"
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/llm/openai"
)

type Conversation struct {
	ctx       context.Context
	dialogues []llm.Dialogue

	llmSrv llm.LLM
}

func NewConversation(ctx context.Context,
	llmConfig *config.DeepseekConfig,
) *Conversation {
	c := &Conversation{
		ctx:       ctx,
		dialogues: make([]llm.Dialogue, 0),
	}

	c.llmSrv = openai.NewOpenAI(
		llmConfig.ApiKey,
		llmConfig.BaseUrl,
		llmConfig.Model)

	return c
}

func (c *Conversation) Ask(question string) (string, error) {
	dialogue := llm.Dialogue{
		Role:    llm.RoleUser,
		Content: question,
	}
	c.dialogues = append(c.dialogues, dialogue)

	response, err := c.llmSrv.Response(c.ctx, c.dialogues)
	if err != nil {
		return "", err
	}

	c.dialogues = append(c.dialogues, llm.Dialogue{
		Role:    llm.RoleAssistant,
		Content: response,
	})

	return response, nil
}

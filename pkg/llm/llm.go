package llm

import (
	"context"
)

const (
	RoleUser      string = "user"
	RoleAssistant string = "assistant"
	RoleSystem    string = "system"
	RoleTool      string = "tool"
)

type LLMResponse struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Err      error  `json:"error,omitempty"`
}

type Dialogue struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLM interface {
	Response(ctx context.Context, dialogues []Dialogue) (string, error)
}

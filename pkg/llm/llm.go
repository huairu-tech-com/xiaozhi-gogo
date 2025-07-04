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

type Dialogue struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLM interface {
	Response(ctx context.Context, dialogues []Dialogue) (string, error)
	ResponseStream(ctx context.Context, dialogues []Dialogue) (<-chan Dialogue, error)
}

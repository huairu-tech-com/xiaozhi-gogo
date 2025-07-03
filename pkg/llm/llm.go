package llm

import (
	"context"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Dialogue struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type LLM interface {
	Response(ctx context.Context, dialogues []Dialogue) (string, error)
	ResponseStream(ctx context.Context, dialogues []Dialogue) (<-chan Dialogue, error)
}

package history

import (
	"github.com/ollama/ollama/api"
)

type History []api.Message

func New() *History {
	return &History{{
		Role:    "system",
		Content: "Be brief. Format all response with markdown",
	}}
}

func (h *History) Append(msg *api.Message) {
	*h = append(*h, *msg)
}

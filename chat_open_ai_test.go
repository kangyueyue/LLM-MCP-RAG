package main

import (
	"context"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestNewChatOpenAI_chat(t *testing.T) {
	ctx := context.Background()
	model := openai.ChatModelGPT3_5Turbo
	ai := NewChatOpenAI(ctx, model, WithRAGContext(""), WithSystemPrompt(""))
	prompt := "请问你使用的什么模型"
	result, tool := ai.Chat(prompt)
	if len(tool) != 0 {
		t.Log("toolCall:", tool)
	}
	t.Log("result:", result)
}

// go test -v -run TestNewChatOpenAI_chat -count=1

package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openai/openai-go/v3"
)

// TestNewAgent 测试NewAgent函数
func TestNewAgent(t *testing.T) {
	ctx := context.Background()
	systemPrompt := `你是一个内容获取与文件写入助手。
	你必须且只能使用我提供的 MCP 工具完成任务：
	1) 使用 mcp-server-fetch 获取公开网页的内容（仅访问单页，遵守网站条款与 robots.txt，不进行批量或大规模抓取）。
	2) 对获取的公开内容进行简要总结。
	3) 使用 @modelcontextprotocol/server-filesystem 将结果写入当前目录的文件。
	不要自行访问网络或返回拒绝，直接调用上述工具完成。`
	allowDir, _ := os.Getwd()
	fmt.Print("allowDir:", allowDir)
	fetchMcpCli := NewMcpClient(ctx, "uvx", nil, []string{"mcp-server"})
	fileMcpCli := NewMcpClient(ctx, "npx", nil, []string{"-y", "@modelcontextprotocol/server-filesystem", allowDir})
	agent := NewAgent(ctx, openai.ChatModelGPT3_5Turbo, []*McpClient{fetchMcpCli, fileMcpCli}, systemPrompt, "")
	if agent == nil {
		t.Fatalf("Failed to create agent")
	}
	t.Log("agent created successfully")
}

// go test -v -run=TestNewAgent -count=1

package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
)

// Agent 代理结构体
type Agent struct {
	McpClient    []*McpClient
	LLM          *ChatOpenAI
	Model        string
	SystemPrompt string
	Ctx          context.Context
	RAGCtx       string
}

// NewAgent 创建一个新的代理实例
func NewAgent(ctx context.Context, model string, mcpcli []*McpClient, systemPrompt string, ragCtx string) *Agent {
	// 1.激活所有mcp client 拿到所有tools
	tools := make([]mcp.Tool, 0)
	for _, item := range mcpcli {
		// 启动stdio传输
		err := item.Start()
		if err != nil {
			fmt.Println("mcp listen error:", err)
			continue
		}
		// 设置工具
		err = item.SetTools()
		if err != nil {
			fmt.Println("mcp set tools error:", err)
			continue
		}
		// 新增日志
		for _, t := range item.GetTool() {
			fmt.Println("tool ready:", t.Name)
		}
		tools = append(tools, item.GetTool()...)
	}
	// 2. 激活并告诉LLM有那些tools
	llm := NewChatOpenAI(ctx, model, WithSystemPrompt(systemPrompt), WithRAGContext(ragCtx), WithTools(tools))
	fmt.Println("init LLM & Tools")
	return &Agent{
		McpClient:    mcpcli,
		LLM:          llm,
		Model:        model,
		SystemPrompt: systemPrompt,
		RAGCtx:       ragCtx,
		Ctx:          ctx,
	}
}

// Close 关闭代理
func (a *Agent) Close() {
	var err error
	for _, mcpClient := range a.McpClient {
		err = mcpClient.Close()
		if err != nil {
			fmt.Println("mcp client close error:", err)
			continue
		}
	}
	fmt.Println("all close")
}

// Invoke 调用代理
func (a *Agent) Invoke(prompt string) string {
	if a.LLM == nil {
		return ""
	}
	response, toolCalls := a.LLM.Chat(prompt)
	fmt.Println("toolCalls:", toolCalls)
	for len(toolCalls) > 0 {
		fmt.Println("response", response)
		for _, toolCall := range toolCalls {
			for _, mcpClient := range a.McpClient {
				for _, mcpTool := range mcpClient.GetTool() {
					if mcpTool.Name == toolCall.Function.Name {
						fmt.Println("tool use", toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)
						toolText, err := mcpClient.CallTool(toolCall.Function.Name, toolCall.Function.Arguments)
						if err != nil {
							fmt.Println("tool call error:", err)
							continue
						}
						a.LLM.Message = append(a.LLM.Message, openai.ToolMessage(toolText, toolCall.ID))
					}
				}
			}
		}
		// 二次对话（空prompt也会发起请求）
		response, toolCalls = a.LLM.Chat("")
	}
	return response
}

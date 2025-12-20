package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// ChatOpenAI 聊天模型
type ChatOpenAI struct {
	Ctx          context.Context                          // 上下文
	ModelName    string                                   // 模型名称
	Message      []openai.ChatCompletionMessageParamUnion // 每一次对话session中包含的消息
	SystemPrompt string                                   // 系统提示词
	RagContext   string                                   // rag上下文
	Tools        []*mcp.Tool                              // 工具
	LLM          openai.Client                            // LLM大模型客户端
}

// LLMOptions 大模型相关配置
type LLMOptions func(*ChatOpenAI)

func WithSystemPrompt(prompt string) LLMOptions {
	return func(c *ChatOpenAI) {
		c.SystemPrompt = prompt
	}
}

func WithRAGContext(ctx string) LLMOptions {
	return func(c *ChatOpenAI) {
		c.RagContext = ctx
	}
}
func WithTools(tools []*mcp.Tool) LLMOptions {
	return func(c *ChatOpenAI) {
		c.Tools = tools
	}
}

// NewChatOpenAI 创建一个新的聊天OpenAI实例
func NewChatOpenAI(ctx context.Context, modelName string, opts ...LLMOptions) *ChatOpenAI {
	if modelName == "" {
		panic("model name cannot be empty")
	}
	var (
		apiKey  = os.Getenv("OPENAI_API_KEY")
		baseURL = os.Getenv("OPENAI_BASE_URL")
	)
	if apiKey == "" {
		panic("OPENAI_API_KEY cannot be empty")
	}
	opt := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opt = append(opt, option.WithBaseURL(baseURL))
	}
	llm := openai.NewClient(opt...)
	c := &ChatOpenAI{
		Ctx:       ctx,
		ModelName: modelName,
		LLM:       llm,
	}
	// 配置项
	for _, opt := range opts {
		opt(c)
	}
	if c.SystemPrompt != "" {
		c.Message = append(c.Message, openai.SystemMessage(c.SystemPrompt))

	}
	if c.RagContext != "" {
		c.Message = append(c.Message, openai.UserMessage(c.RagContext))
	}
	fmt.Printf("successfully init chat open AI:%s\n", modelName)
	return c
}

// Chat 和模型通信的方法
func (c *ChatOpenAI) Chat(prompt string) (string, []openai.ToolCallUnion) {
	if prompt != "" {
		// 添加用户提示词
		c.Message = append(c.Message, openai.UserMessage(prompt))
	}
	stream := c.LLM.Chat.Completions.NewStreaming(c.Ctx, openai.ChatCompletionNewParams{
		Model:    c.ModelName,
		Messages: c.Message,
		Seed:     openai.Int(0),
		Tools:    c.McpToolToOpenAITool(c.Tools),
	})
	var acc openai.ChatCompletionAccumulator // 用户结果的类加
	var toolCalls []openai.ToolCallUnion
	result := ""
	finished := false
	fmt.Println("start streaming chat...")
	for stream.Next() {
		chunk := stream.Current() // 当前的返回结果
		acc.AddChunk(chunk)
		// 此时完成了
		if content, ok := acc.JustFinishedContent(); ok {
			finished = true
			result = content
		}
		// 收集tool调用
		if tool, ok := acc.JustFinishedToolCall(); ok {
			fmt.Println("tool call finished:", tool.Index, tool.Name, tool.Arguments)
			toolCalls = append(toolCalls, openai.ToolCallUnion{
				ID: tool.ID, // 收集tool的Id
				Function: openai.FunctionToolCallFunction{
					Name:      tool.Name,
					Arguments: tool.Arguments,
				},
			})
		}
		// 收集refusal
		if refusal, ok := acc.JustFinishedRefusal(); ok {
			fmt.Println("refusal :", refusal)
		}
		// 收集delta
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			// 如果没有停止生成
			if !finished {
				result += delta
			}
		}
	}
	if len(acc.Choices) > 0 {
		c.Message = append(c.Message, acc.Choices[0].Message.ToParam()) // 最后一次生成的结果加入
	}
	if stream.Err() != nil {
		panic(stream.Err())
	}
	return result, toolCalls

}

// McpToolToOpenAITool 将mcp工具转换为openai工ChatCompletionToolUnionParam具
func (c *ChatOpenAI) McpToolToOpenAITool(mcpTools []*mcp.Tool) []openai.ChatCompletionToolUnionParam {
	openAITool := make([]openai.ChatCompletionToolUnionParam, 0, len(mcpTools))
	for _, tool := range mcpTools {
		params := openai.FunctionParameters{
			"type":       tool.InputSchema.Type,
			"properties": tool.InputSchema.Properties,
			"required":   tool.InputSchema.Required,
		}
		// 关键兜底：若type为空，默认用object，避免OpenAI拒绝工具定义
		if t, ok := params["type"].(string); !ok || t == "" {
			params["type"] = "object"
		}
		openAITool = append(openAITool, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: shared.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  params,
				},
			},
		})
	}
	return openAITool
}

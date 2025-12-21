package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// McpClient  mcp客户端
type McpClient struct {
	Ctx    context.Context
	Client *client.Client
	Cmd    string     // 命令
	Tools  []mcp.Tool // 工具
	Args   []string   // 参数
	Env    []string   // 环境
}

// NewMcpClient 创建一个新的McpClient实例
func NewMcpClient(ctx context.Context, cmd string, env, args []string) *McpClient {
	// 创建协议
	stdioTransport := transport.NewStdio(cmd, env, args...)
	// 创建client
	client := client.NewClient(stdioTransport)
	return &McpClient{
		Ctx:    ctx,
		Cmd:    cmd,
		Client: client,
		Args:   args,
		Env:    env,
	}
}

// Start 启动mcp客户端
func (c *McpClient) Start() error {
	err := c.Client.Start(c.Ctx)
	if err != nil {
		return err
	}
	// 初始化
	mcpInitReq := mcp.InitializeRequest{}
	mcpInitReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	mcpInitReq.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "0.0.1",
	}
	if _, err := c.Client.Initialize(c.Ctx, mcpInitReq); err != nil {
		fmt.Println("mcp init error:", err)
		return err
	}
	return err
}

// SetTools 设置工具
func (c *McpClient) SetTools() error {
	toolsReq := mcp.ListToolsRequest{}
	tools, err := c.Client.ListTools(c.Ctx, toolsReq)
	if err != nil {
		return err
	}

	mt := make([]mcp.Tool, 0)
	for _, tool := range tools.Tools {
		mt = append(mt, mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	c.Tools = mt
	return nil
}

// CallTool 调用工具,arg是LLM返回出来的
func (c *McpClient) CallTool(name string, arg any) (string, error) {
	// 使用参数调用方法
	var arguments map[string]any
	switch v := arg.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &arguments); err != nil {
			return "", err
		}
	case map[string]any:
		arguments = v
	default:
	}
	res, err := c.Client.CallTool(c.Ctx, mcp.CallToolRequest{
		// 参数
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: arguments,
		},
	})
	if err != nil {
		return "", err
	}
	return mcp.GetTextFromContent(res.Content), nil
}

// Close 关闭连接
func (c *McpClient) Close() error {
	return c.Client.Close()
}

// GetTool 获取工具
func (c *McpClient) GetTool() []mcp.Tool {
	return c.Tools
}

package main

import (
	"context"
	"fmt"
	"testing"
)

func TestMCPClient_uvx(t *testing.T) {
	ctx := context.Background()
	// 使用uvx mcp-server作为mcp服务器
	mcpCli := NewMcpClient(ctx, "uvx", nil, []string{"mcp-server-fetch"})
	defer func() {
		err := mcpCli.Close()
		if err != nil {
			fmt.Println("close err", err)
		}
	}()
	err := mcpCli.Start()
	if err != nil {
		fmt.Println("start err", err)
		return
	}
	err = mcpCli.SetTools()
	if err != nil {
		fmt.Println("set tools err", err)
		return
	}
	tools := mcpCli.GetTool()
	fmt.Println(tools)
}

// go test -v -run=TestMCPClient_uvx -count=1

func TestMCPClient_npx(t *testing.T) {
	ctx := context.Background()
	mcpCli := NewMcpClient(ctx, "npx", nil, []string{"@modelcontextprotocol/server-filesystem"})
	defer func() {
		err := mcpCli.Close()
		if err != nil {
			fmt.Println("close err", err)
		}
	}()
	err := mcpCli.Start()
	if err != nil {
		fmt.Println("start err", err)
		return
	}
	err = mcpCli.SetTools()
	if err != nil {
		fmt.Println("set tools err", err)
		return
	}
	tools := mcpCli.GetTool()
	fmt.Println(tools)
}

// go test -v -run=TestMCPClient_npx -count=1

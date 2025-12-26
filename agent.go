package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	mcpBin := flag.String("mcp-bin", "/tmp/expense-splitter", "path to expense-splitter MCP binary")
	flag.Parse()

	if err := ensureExecutable(*mcpBin); err != nil {
		log.Fatalf("invalid -mcp-bin: %v", err)
	}

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("failed to create model: %s", err)
	}

	mcpToolSet, err := mcptoolset.New(mcptoolset.Config{
		Transport: &mcp.CommandTransport{Command: exec.Command(*mcpBin)},
	})
	if err != nil {
		log.Fatalf("failed to create MCP toolset: %v", err)
	}

	expenseAgent, err := llmagent.New(llmagent.Config{
		Name:        "expense_splitter_agent",
		Model:       model,
		Description: "Manages expense groups, members, and settlements via the expense-splitter MCP tool",
		Instruction: "You are a helpful assistant that creates groups, adds people and expenses, and summarizes settlements",
		//Tools: []tool.Tool{
		//	geminitool.GoogleSearch{},
		//},
		Toolsets: []tool.Toolset{mcpToolSet},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(expenseAgent),
	}

	l := full.NewLauncher()
	if err := l.Execute(ctx, config, flag.Args()); err != nil {
		log.Fatalf("Run Failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

func ensureExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.ErrInvalid
	}
	if info.Mode()&0o111 == 0 {
		return os.ErrPermission
	}
	return nil
}

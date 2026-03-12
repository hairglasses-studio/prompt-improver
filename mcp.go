package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hairglasses-studio/prompt-improver/pkg/enhancer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AnalyzeArgs is the input schema for the analyze_prompt tool.
type AnalyzeArgs struct {
	Prompt   string `json:"prompt" jsonschema:"The prompt text to analyze"`
	TaskType string `json:"task_type,omitempty" jsonschema:"Task type override: code, creative, analysis, troubleshooting, workflow, or general"`
}

// EnhanceArgs is the input schema for the enhance_prompt tool.
type EnhanceArgs struct {
	Prompt   string `json:"prompt" jsonschema:"The prompt text to enhance"`
	TaskType string `json:"task_type,omitempty" jsonschema:"Task type override: code, creative, analysis, troubleshooting, workflow, or general"`
}

// LintArgs is the input schema for the lint_prompt tool.
type LintArgs struct {
	Prompt string `json:"prompt" jsonschema:"The prompt text to lint"`
}

func handleAnalyze(_ context.Context, _ *mcp.CallToolRequest, args AnalyzeArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: prompt is required"}},
			IsError: true,
		}, nil, nil
	}
	result := enhancer.Analyze(args.Prompt)
	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func handleEnhance(_ context.Context, _ *mcp.CallToolRequest, args EnhanceArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: prompt is required"}},
			IsError: true,
		}, nil, nil
	}
	tt := enhancer.ValidTaskType(args.TaskType)
	result := enhancer.Enhance(args.Prompt, tt)
	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func handleLint(_ context.Context, _ *mcp.CallToolRequest, args LintArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: prompt is required"}},
			IsError: true,
		}, nil, nil
	}
	results := enhancer.Lint(args.Prompt)
	cacheResults := enhancer.VerifyCacheFriendlyOrder(args.Prompt)
	results = append(results, cacheResults...)

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No issues found."}},
		}, nil, nil
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func runMCP() {
	server := mcp.NewServer(
		&mcp.Implementation{Name: "prompt-improver", Version: "1.0.0"},
		nil,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_prompt",
		Description: "Score a prompt across 10 quality dimensions (0-100) with letter grades and actionable suggestions. Returns specificity, structure, examples, framing, emphasis, format, context placement, injection safety, task-fit, and conciseness scores.",
	}, handleAnalyze)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "enhance_prompt",
		Description: "Apply a 13-stage enhancement pipeline to a prompt: specificity, positive reframing, tone normalization, overtrigger rewrite, example wrapping, XML structure, context reordering, format enforcement, quote grounding, self-check, overengineering guard, and preamble suppression.",
	}, handleEnhance)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "lint_prompt",
		Description: "Deep lint a prompt for 11 anti-patterns: overtrigger phrases, negative framing, aggressive emphasis, vague quantifiers, unmotivated rules, over-specification, injection risk, thinking-mode redundancy, example quality, compaction readiness, and cache-friendly ordering.",
	}, handleLint)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		// EOF is expected when the client disconnects
		if strings.Contains(err.Error(), "EOF") {
			return
		}
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

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

// DiffArgs is the input schema for the diff_prompt tool.
type DiffArgs struct {
	Prompt   string `json:"prompt" jsonschema:"The prompt text to diff (original vs enhanced)"`
	TaskType string `json:"task_type,omitempty" jsonschema:"Task type override: code, creative, analysis, troubleshooting, workflow, or general"`
}

// ImproveArgs is the input schema for the improve_prompt tool.
type ImproveArgs struct {
	Prompt          string `json:"prompt" jsonschema:"The prompt text to improve using LLM"`
	TaskType        string `json:"task_type,omitempty" jsonschema:"Task type override: code, creative, analysis, troubleshooting, workflow, or general"`
	ThinkingEnabled bool   `json:"thinking_enabled,omitempty" jsonschema:"Add thinking scaffolding to the improved prompt"`
	Feedback        string `json:"feedback,omitempty" jsonschema:"Optional targeted improvement hints"`
	Mode            string `json:"mode,omitempty" jsonschema:"Enhancement mode: local, llm, or auto (default: auto)"`
}

func handleDiff(_ context.Context, _ *mcp.CallToolRequest, args DiffArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: prompt is required"}},
			IsError: true,
		}, nil, nil
	}
	tt := enhancer.ValidTaskType(args.TaskType)
	result := enhancer.Enhance(args.Prompt, tt)

	var sb strings.Builder
	sb.WriteString("--- original\n+++ enhanced\n\n")
	for _, line := range strings.Split(args.Prompt, "\n") {
		fmt.Fprintf(&sb, "- %s\n", line)
	}
	sb.WriteString("\n")
	for _, line := range strings.Split(result.Enhanced, "\n") {
		fmt.Fprintf(&sb, "+ %s\n", line)
	}
	if len(result.Improvements) > 0 {
		fmt.Fprintf(&sb, "\n%d improvements:\n", len(result.Improvements))
		for _, imp := range result.Improvements {
			fmt.Fprintf(&sb, "  • %s\n", imp)
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil, nil
}

// CheckClaudeMDArgs is the input schema for the check_claudemd tool.
type CheckClaudeMDArgs struct {
	Path string `json:"path" jsonschema:"Path to the CLAUDE.md file to check"`
}

// ListTemplatesArgs is the input schema for the list_templates tool (no params).
type ListTemplatesArgs struct{}

func handleCheckClaudeMD(_ context.Context, _ *mcp.CallToolRequest, args CheckClaudeMDArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	if path == "" {
		path = "./CLAUDE.md"
	}
	results, err := enhancer.CheckClaudeMD(path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: " + err.Error()}},
			IsError: true,
		}, nil, nil
	}
	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "CLAUDE.md looks healthy — no issues found."}},
		}, nil, nil
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func handleListTemplates(_ context.Context, _ *mcp.CallToolRequest, _ ListTemplatesArgs) (*mcp.CallToolResult, any, error) {
	templates := enhancer.ListTemplates()
	data, _ := json.MarshalIndent(templates, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func handleImprove(ctx context.Context, _ *mcp.CallToolRequest, args ImproveArgs) (*mcp.CallToolResult, any, error) {
	if args.Prompt == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: prompt is required"}},
			IsError: true,
		}, nil, nil
	}

	tt := enhancer.ValidTaskType(args.TaskType)
	mode := enhancer.ValidMode(args.Mode)
	if mode == "" {
		mode = enhancer.ModeAuto
	}

	cfg := enhancer.ResolveConfig(".")
	cfg.LLM.Enabled = true
	if args.ThinkingEnabled {
		cfg.LLM.ThinkingEnabled = true
	}

	engine := getOrCreateEngine(cfg.LLM)
	if engine == nil && mode != enhancer.ModeLocal {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "error: no LLM credentials available — set ANTHROPIC_API_KEY or configure llm.api_key_env for your Anthropic-compatible endpoint. Use mode=local for deterministic enhancement."}},
			IsError: true,
		}, nil, nil
	}

	result := enhancer.EnhanceHybrid(ctx, args.Prompt, tt, cfg, engine, mode)
	data, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func runMCP() {
	// Eagerly initialize the shared hybrid engine so MCP handlers reuse the
	// same circuit breaker and cache as the CLI. getOrCreateEngine stores the
	// engine in the package-level hybridEngine variable.
	cfg := enhancer.ResolveConfig(".")
	if cfg.LLM.Enabled {
		getOrCreateEngine(cfg.LLM)
	}

	server := mcp.NewServer(
		&mcp.Implementation{Name: "prompt-improver", Version: version},
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "diff_prompt",
		Description: "Show a unified diff of original vs enhanced prompt. Displays added/removed lines and lists improvements applied by the 13-stage pipeline.",
	}, handleDiff)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "improve_prompt",
		Description: "LLM-powered prompt improvement using Claude. Analyzes task type, adds domain-specific role, structured output sections, scratchpad, and template variables. Falls back to local 13-stage pipeline if LLM is unavailable. Set mode=local for deterministic-only enhancement.",
	}, handleImprove)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_claudemd",
		Description: "Health-check a CLAUDE.md file for common issues: excessive length, inline code blocks, style guide content that belongs in linter config, overtrigger language, aggressive ALL-CAPS, and missing section headers.",
	}, handleCheckClaudeMD)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_templates",
		Description: "List all available prompt templates with their names, descriptions, task types, variables, and usage examples.",
	}, handleListTemplates)

	registerPromptImproverContract(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		// EOF is expected when the client disconnects
		if strings.Contains(err.Error(), "EOF") {
			return
		}
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

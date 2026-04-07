package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type promptImproverToolInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

type promptImproverToolSearchInput struct {
	Query string `json:"query" jsonschema:"Keyword to search across prompt-improver tool names and descriptions"`
}

type promptImproverToolSchemaInput struct {
	Name string `json:"name" jsonschema:"Exact tool name to inspect"`
}

func promptImproverToolCatalog() []promptImproverToolInfo {
	return []promptImproverToolInfo{
		{
			Name:        "analyze_prompt",
			Description: "Score a prompt across 10 quality dimensions with grades and suggestions.",
			Category:    "analysis",
			InputSchema: map[string]any{"prompt": "string", "task_type": "string"},
		},
		{
			Name:        "enhance_prompt",
			Description: "Apply the deterministic 13-stage enhancement pipeline.",
			Category:    "enhancement",
			InputSchema: map[string]any{"prompt": "string", "task_type": "string"},
		},
		{
			Name:        "lint_prompt",
			Description: "Lint a prompt for common anti-patterns and cache-hostile structure.",
			Category:    "analysis",
			InputSchema: map[string]any{"prompt": "string"},
		},
		{
			Name:        "diff_prompt",
			Description: "Show the original prompt against the enhanced prompt as a diff.",
			Category:    "analysis",
			InputSchema: map[string]any{"prompt": "string", "task_type": "string"},
		},
		{
			Name:        "improve_prompt",
			Description: "Use the hybrid local/LLM improvement path to rewrite a prompt.",
			Category:    "enhancement",
			InputSchema: map[string]any{"prompt": "string", "task_type": "string", "thinking_enabled": "boolean", "feedback": "string", "mode": "string"},
		},
		{
			Name:        "check_claudemd",
			Description: "Health-check a CLAUDE.md file for prompt-policy issues.",
			Category:    "review",
			InputSchema: map[string]any{"path": "string"},
		},
		{
			Name:        "list_templates",
			Description: "List the built-in prompt templates and their variables.",
			Category:    "templates",
		},
		{
			Name:        "prompt_improver_tool_search",
			Description: "Search the prompt-improver tool surface by keyword.",
			Category:    "discovery",
			InputSchema: map[string]any{"query": "string"},
		},
		{
			Name:        "prompt_improver_tool_catalog",
			Description: "List prompt-improver tools grouped by category.",
			Category:    "discovery",
		},
		{
			Name:        "prompt_improver_tool_schema",
			Description: "Inspect the input schema for one prompt-improver tool.",
			Category:    "discovery",
			InputSchema: map[string]any{"name": "string"},
		},
		{
			Name:        "prompt_improver_tool_stats",
			Description: "Show prompt-improver tool counts by category.",
			Category:    "discovery",
		},
		{
			Name:        "prompt_improver_server_health",
			Description: "Show the prompt-improver contract shape and runtime metadata.",
			Category:    "discovery",
		},
	}
}

func registerPromptImproverContract(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "prompt_improver_tool_search",
		Description: "Search the prompt-improver tool surface by keyword before invoking an analysis or enhancement tool.",
	}, handlePromptImproverToolSearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "prompt_improver_tool_catalog",
		Description: "List prompt-improver tools grouped by category so agents can pick the smallest useful surface first.",
	}, handlePromptImproverToolCatalog)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "prompt_improver_tool_schema",
		Description: "Inspect one prompt-improver tool's expected arguments.",
	}, handlePromptImproverToolSchema)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "prompt_improver_tool_stats",
		Description: "Show prompt-improver tool counts by category plus resource and prompt coverage.",
	}, handlePromptImproverToolStats)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "prompt_improver_server_health",
		Description: "Show the prompt-improver contract shape, version, runtime, and discovery coverage.",
	}, handlePromptImproverServerHealth)

	server.AddResource(&mcp.Resource{
		Name:        "Prompt Improver Overview",
		Description: "Server card covering discovery-first usage and the highest-value prompt-improver workflows.",
		MIMEType:    "text/markdown",
		URI:         "prompt-improver://server/overview",
	}, readPromptImproverOverview)

	server.AddResource(&mcp.Resource{
		Name:        "Prompt Improver Start Here",
		Description: "Quickstart workflow for choosing analysis, linting, or improvement.",
		MIMEType:    "text/markdown",
		URI:         "prompt-improver://workflows/start-here",
	}, readPromptImproverWorkflow)

	server.AddPrompt(&mcp.Prompt{
		Name:        "prompt_improver_start_triage",
		Description: "Use prompt-improver in a discovery-first sequence before choosing an enhancement path.",
	}, promptImproverStartTriage)
}

func handlePromptImproverToolSearch(_ context.Context, _ *mcp.CallToolRequest, input promptImproverToolSearchInput) (*mcp.CallToolResult, any, error) {
	query := strings.ToLower(strings.TrimSpace(input.Query))
	results := make([]promptImproverToolInfo, 0)
	for _, tool := range promptImproverToolCatalog() {
		if query == "" || strings.Contains(strings.ToLower(tool.Name), query) || strings.Contains(strings.ToLower(tool.Description), query) || strings.Contains(strings.ToLower(tool.Category), query) {
			results = append(results, tool)
		}
	}
	return promptImproverJSONResult(map[string]any{
		"results": results,
		"total":   len(results),
	}), nil, nil
}

func handlePromptImproverToolCatalog(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	catalog := promptImproverToolCatalog()
	grouped := make(map[string][]promptImproverToolInfo)
	for _, tool := range catalog {
		grouped[tool.Category] = append(grouped[tool.Category], tool)
	}

	categories := make([]string, 0, len(grouped))
	for category := range grouped {
		categories = append(categories, category)
		sort.Strings(categories)
	}

	groups := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		tools := grouped[category]
		sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
		groups = append(groups, map[string]any{
			"category":   category,
			"tool_count": len(tools),
			"tools":      tools,
		})
	}

	return promptImproverJSONResult(map[string]any{
		"groups":      groups,
		"total_tools": len(catalog),
	}), nil, nil
}

func handlePromptImproverToolSchema(_ context.Context, _ *mcp.CallToolRequest, input promptImproverToolSchemaInput) (*mcp.CallToolResult, any, error) {
	for _, tool := range promptImproverToolCatalog() {
		if tool.Name == input.Name {
			return promptImproverJSONResult(tool), nil, nil
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("tool not found: %s", input.Name)}},
		IsError: true,
	}, nil, nil
}

func handlePromptImproverToolStats(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	stats := map[string]int{}
	for _, tool := range promptImproverToolCatalog() {
		stats[tool.Category]++
	}
	return promptImproverJSONResult(map[string]any{
		"tool_count":     len(promptImproverToolCatalog()),
		"resource_count": 2,
		"prompt_count":   1,
		"by_category":    stats,
	}), nil, nil
}

func handlePromptImproverServerHealth(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	return promptImproverJSONResult(map[string]any{
		"server":         "prompt-improver",
		"version":        version,
		"status":         "ok",
		"go_version":     runtime.Version(),
		"tool_count":     len(promptImproverToolCatalog()),
		"resource_count": 2,
		"prompt_count":   1,
		"discovery_tools": []string{
			"prompt_improver_tool_search",
			"prompt_improver_tool_catalog",
			"prompt_improver_tool_schema",
			"prompt_improver_tool_stats",
			"prompt_improver_server_health",
		},
	}), nil, nil
}

func readPromptImproverOverview(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "prompt-improver://server/overview",
				MIMEType: "text/markdown",
				Text: strings.Join([]string{
					"# prompt-improver",
					"",
					"1. Start with `prompt_improver_tool_search` or `prompt_improver_tool_catalog`.",
					"2. Use `lint_prompt` when you need anti-pattern detection before rewriting.",
					"3. Use `analyze_prompt` for scoring and `diff_prompt` when you need before/after clarity.",
					"4. Use `enhance_prompt` for deterministic improvement and `improve_prompt` only when the LLM path is warranted.",
				}, "\n"),
			},
		},
	}, nil
}

func readPromptImproverWorkflow(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "prompt-improver://workflows/start-here",
				MIMEType: "text/markdown",
				Text: strings.Join([]string{
					"1. Search the tool surface with `prompt_improver_tool_search` if the right path is unclear.",
					"2. Run `lint_prompt` or `analyze_prompt` first for diagnosis.",
					"3. Use `enhance_prompt` for deterministic cleanup or `improve_prompt` when you want the hybrid LLM path.",
					"4. Finish with `diff_prompt` if you need a compact review artifact.",
				}, "\n"),
			},
		},
	}, nil
}

func promptImproverStartTriage(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Discovery-first prompt triage workflow",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: "Triage this prompt work in a discovery-first way. Start with `prompt_improver_tool_search` or `prompt_improver_tool_catalog`, then diagnose with `lint_prompt` or `analyze_prompt`, then choose `enhance_prompt` for deterministic cleanup or `improve_prompt` for the hybrid LLM path. Keep the summary short and name the next safest tool to call.",
				},
			},
		},
	}, nil
}

func promptImproverJSONResult(v any) *mcp.CallToolResult {
	data, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}

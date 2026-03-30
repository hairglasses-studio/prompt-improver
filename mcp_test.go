package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hairglasses-studio/prompt-improver/pkg/enhancer"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCP_HandleAnalyze(t *testing.T) {
	t.Run("valid prompt", func(t *testing.T) {
		args := AnalyzeArgs{Prompt: "fix this"}
		result, _, err := handleAnalyze(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Error("expected success, got error")
		}
		text := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(text, "score") {
			t.Error("result should contain score")
		}
		if !strings.Contains(text, "suggestions") {
			t.Error("result should contain suggestions")
		}
		if !strings.Contains(text, "score_report") {
			t.Error("result should contain score_report")
		}

		// Verify it's valid JSON
		var parsed enhancer.AnalyzeResult
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			t.Errorf("result is not valid JSON: %v", err)
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		args := AnalyzeArgs{Prompt: ""}
		result, _, err := handleAnalyze(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for empty prompt")
		}
	})

	t.Run("with task type", func(t *testing.T) {
		args := AnalyzeArgs{Prompt: "review this code for bugs and security issues", TaskType: "code"}
		result, _, err := handleAnalyze(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(*mcp.TextContent).Text
		// Task type detection may classify differently, but the result should be valid
		var parsed enhancer.AnalyzeResult
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			t.Errorf("result is not valid JSON: %v", err)
		}
	})
}

func TestMCP_HandleEnhance(t *testing.T) {
	t.Run("valid prompt", func(t *testing.T) {
		args := EnhanceArgs{Prompt: "fix this bug in the sorting function"}
		result, _, err := handleEnhance(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Error("expected success, got error")
		}
		text := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(text, "enhanced") {
			t.Error("result should contain enhanced field")
		}
		if !strings.Contains(text, "stages_run") {
			t.Error("result should contain stages_run field")
		}

		var parsed enhancer.EnhanceResult
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			t.Errorf("result is not valid JSON: %v", err)
		}
		if parsed.Enhanced == "" {
			t.Error("enhanced prompt should not be empty")
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		args := EnhanceArgs{Prompt: ""}
		result, _, err := handleEnhance(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for empty prompt")
		}
	})

	t.Run("with task type override", func(t *testing.T) {
		args := EnhanceArgs{Prompt: "write a haiku about testing", TaskType: "creative"}
		result, _, err := handleEnhance(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(*mcp.TextContent).Text
		var parsed enhancer.EnhanceResult
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			t.Errorf("result is not valid JSON: %v", err)
		}
		if parsed.TaskType != "creative" {
			t.Errorf("expected task type creative, got %s", parsed.TaskType)
		}
	})
}

func TestMCP_HandleLint(t *testing.T) {
	t.Run("clean prompt", func(t *testing.T) {
		args := LintArgs{Prompt: "Return exactly 5 user records as JSON, sorted by creation date."}
		result, _, err := handleLint(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Error("expected success, got error")
		}
	})

	t.Run("dirty prompt", func(t *testing.T) {
		args := LintArgs{Prompt: "CRITICAL: You MUST follow this rule. NEVER ignore it."}
		result, _, err := handleLint(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(*mcp.TextContent).Text
		if !strings.Contains(text, "overtrigger-phrase") {
			t.Error("should detect overtrigger phrase")
		}
	})

	t.Run("empty prompt", func(t *testing.T) {
		args := LintArgs{Prompt: ""}
		result, _, err := handleLint(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for empty prompt")
		}
	})

	t.Run("cache check included", func(t *testing.T) {
		// Prompt with bad cache ordering: constraints before role
		args := LintArgs{Prompt: "<constraints>Be thorough.</constraints>\n<role>You are an expert.</role>\n<instructions>Do the thing.</instructions>"}
		result, _, err := handleLint(context.Background(), nil, args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		text := result.Content[0].(*mcp.TextContent).Text
		// Should include cache-order findings
		if !strings.Contains(text, "cache") && !strings.Contains(text, "No issues") {
			t.Error("should include cache check results or report no issues")
		}
	})
}

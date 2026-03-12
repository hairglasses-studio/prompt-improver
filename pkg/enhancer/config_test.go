package enhancer

import (
	"strings"
	"testing"
)

func TestConfig_IsStageDisabled(t *testing.T) {
	cfg := Config{
		DisabledStages: []string{"tone_downgrade", "preamble_suppression"},
	}

	if !cfg.IsStageDisabled("tone_downgrade") {
		t.Error("Should report tone_downgrade as disabled")
	}
	if !cfg.IsStageDisabled("TONE_DOWNGRADE") {
		t.Error("Should be case-insensitive")
	}
	if cfg.IsStageDisabled("specificity") {
		t.Error("Should not report specificity as disabled")
	}
}

func TestConfig_ApplyRules(t *testing.T) {
	cfg := Config{
		Rules: []Rule{
			{
				Match:  "add tool",
				Append: "Follow the existing tool registration pattern using init().",
			},
			{
				Match:   "fix",
				Prepend: "Include a test that reproduces the bug.",
			},
		},
	}

	text := "Add tool for audio processing in the studio"
	result, imps := cfg.ApplyRules(text)

	if len(imps) == 0 {
		t.Fatal("Should apply matching rule")
	}
	if !strings.Contains(result, "tool registration pattern") {
		t.Error("Should append rule content")
	}
}

func TestConfig_ApplyRules_NoMatch(t *testing.T) {
	cfg := Config{
		Rules: []Rule{
			{Match: "deploy", Append: "Check CI first."},
		},
	}

	text := "Write a function to sort users"
	_, imps := cfg.ApplyRules(text)

	if len(imps) > 0 {
		t.Error("Should not apply non-matching rules")
	}
}

func TestConfig_LoadConfig_Missing(t *testing.T) {
	cfg := LoadConfig("/nonexistent/path")
	if cfg.Preamble != "" || len(cfg.Rules) > 0 {
		t.Error("Should return zero config for missing file")
	}
}

func TestEnhanceWithConfig_DisabledStage(t *testing.T) {
	cfg := Config{
		DisabledStages: []string{"structure"},
	}

	result := EnhanceWithConfig("write a function to sort users by name with error handling and edge cases covered", "", cfg)

	if strings.Contains(result.Enhanced, "<role>") {
		t.Error("Should not add XML structure when stage is disabled")
	}

	// Other stages should still run
	foundStructure := false
	for _, stage := range result.StagesRun {
		if stage == "structure" {
			foundStructure = true
		}
	}
	if foundStructure {
		t.Error("Structure stage should not appear in stages run")
	}
}

func TestEnhanceWithConfig_Preamble(t *testing.T) {
	cfg := Config{
		Preamble: "This is the hg-mcp project, a Go MCP server.",
	}

	result := EnhanceWithConfig("write a function to sort users by name with error handling and edge cases", "", cfg)

	if !strings.HasPrefix(result.Enhanced, "This is the hg-mcp project") {
		t.Error("Should prepend preamble to enhanced output")
	}
}

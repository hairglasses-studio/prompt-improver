package enhancer

import (
	"strings"
	"testing"
)

func TestLint_UnmotivatedRule(t *testing.T) {
	text := "Always use structured error responses.\nHandle all edge cases."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "unmotivated-rule" {
			found = true
			if !strings.Contains(r.Suggestion, "because") {
				t.Error("Should suggest adding a 'because' clause")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect unmotivated rules")
	}
}

func TestLint_MotivatedRule(t *testing.T) {
	text := "Always use structured error responses because the AI assistant uses the error type to decide whether to retry."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "unmotivated-rule" && strings.Contains(r.Original, "structured error") {
			t.Error("Should NOT flag motivated rules (contains 'because')")
		}
	}
}

func TestLint_AggressiveEmphasis(t *testing.T) {
	text := "CRITICAL: You MUST follow this rule."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "aggressive-emphasis" {
			found = true
			if !r.AutoFixable {
				t.Error("Aggressive emphasis should be auto-fixable")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect aggressive emphasis")
	}
}

func TestLint_VagueQuantifiers(t *testing.T) {
	text := "Return several items from the list with appropriate formatting."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "vague-quantifier" {
			found = true
			if !strings.Contains(r.Suggestion, "specific numbers") {
				t.Error("Should suggest specific numbers")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect vague quantifiers")
	}
}

func TestLint_CleanPrompt(t *testing.T) {
	text := "Return exactly 5 user records as JSON, sorted by creation date."
	results := Lint(text)

	// Should have minimal findings
	for _, r := range results {
		if r.Severity == "warn" || r.Severity == "error" {
			t.Errorf("Clean prompt should not have warn/error findings, got: %s - %s", r.Category, r.Original)
		}
	}
}

func TestLint_NegativeFraming(t *testing.T) {
	text := "NEVER include personal information in the output headers."
	results := Lint(text)

	// This should be flagged as negative framing (not in safety whitelist, not in reframing table)
	found := false
	for _, r := range results {
		if r.Category == "negative-framing" || r.Category == "aggressive-emphasis" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect negative framing or aggressive emphasis")
	}
}

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

// --- Overtrigger phrase lint ---

func TestLint_OvertriggerPhrase(t *testing.T) {
	text := "CRITICAL: You MUST use this tool when processing data."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "overtrigger-phrase" {
			found = true
			if !r.AutoFixable {
				t.Error("Overtrigger phrase should be auto-fixable")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect overtrigger phrase")
	}
}

func TestLint_OvertriggerPhrase_Clean(t *testing.T) {
	text := "Use this tool when processing data."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "overtrigger-phrase" {
			t.Error("Should not flag clean prompt as overtrigger")
		}
	}
}

// --- Over-specification lint ---

func TestLint_OverSpecification(t *testing.T) {
	text := `Follow these steps:
1. Read the file
2. Parse the JSON
3. Validate the schema
4. Extract the fields
5. Transform the data
6. Write the output
7. Verify the result`
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "over-specification" {
			found = true
			if !strings.Contains(r.Suggestion, "end-state") {
				t.Error("Should suggest end-state descriptions")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect over-specification (>5 numbered steps)")
	}
}

func TestLint_OverSpecification_FewSteps(t *testing.T) {
	text := "1. Read file\n2. Process\n3. Save"
	results := Lint(text)

	for _, r := range results {
		if r.Category == "over-specification" {
			t.Error("Should not flag prompts with <=5 steps")
		}
	}
}

func TestLint_OverSpecification_NoSteps(t *testing.T) {
	text := "Fix the bug in the sorting function."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "over-specification" {
			t.Error("Should not flag prompts without numbered steps")
		}
	}
}

// --- Decomposition suggestion lint ---

func TestLint_Decomposition(t *testing.T) {
	text := "Create the API endpoint, fix the database migration, and deploy the service to production."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "decomposition-needed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should suggest decomposition for multi-task prompts")
	}
}

func TestLint_Decomposition_SingleTask(t *testing.T) {
	text := "Create a function to sort users."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "decomposition-needed" {
			t.Error("Should not flag single-task prompts")
		}
	}
}

func TestLint_Decomposition_RepeatedVerb(t *testing.T) {
	text := "Create the user model, create the user controller, create the user view."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "decomposition-needed" {
			t.Error("Should not flag prompts with repeated verbs (same task type)")
		}
	}
}

// --- Injection vulnerability lint ---

func TestLint_InjectionVulnerability(t *testing.T) {
	text := "Process this request: ${user_input}"
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "injection-risk" {
			found = true
			if r.Severity != "error" {
				t.Error("Injection risk should be error severity")
			}
			break
		}
	}
	if !found {
		t.Error("Should detect injection vulnerability with untrusted variable")
	}
}

func TestLint_InjectionVulnerability_DoubleHandlebars(t *testing.T) {
	text := "Respond to: {{user_query}}"
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "injection-risk" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect injection vulnerability in {{user_query}}")
	}
}

func TestLint_InjectionVulnerability_SafeVar(t *testing.T) {
	text := "The system name is {{system_name}}"
	results := Lint(text)

	for _, r := range results {
		if r.Category == "injection-risk" {
			t.Error("Should not flag non-untrusted variable names")
		}
	}
}

func TestLint_InjectionVulnerability_NoVars(t *testing.T) {
	text := "Write a function to sort users."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "injection-risk" {
			t.Error("Should not flag prompts without template variables")
		}
	}
}

// --- Thinking mode detection lint ---

func TestLint_ThinkingMode(t *testing.T) {
	text := "Think step by step about how to solve this problem."
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "thinking-mode-redundant" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect redundant thinking mode instruction")
	}
}

func TestLint_ThinkingMode_Clean(t *testing.T) {
	text := "Solve this problem efficiently."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "thinking-mode-redundant" {
			t.Error("Should not flag prompts without thinking mode instructions")
		}
	}
}

// --- Example quality lint ---

func TestLint_ExampleQuality_TooFew(t *testing.T) {
	text := "<example>one</example>"
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "example-quality" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag too few examples")
	}
}

func TestLint_ExampleQuality_TooMany(t *testing.T) {
	text := "<example>1</example><example>2</example><example>3</example><example>4</example><example>5</example><example>6</example>"
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "example-quality" && strings.Contains(r.Suggestion, "diminishing") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag too many examples")
	}
}

func TestLint_ExampleQuality_GoodCount(t *testing.T) {
	text := "<example>1</example><example>2</example><example>3</example>"
	results := Lint(text)

	for _, r := range results {
		if r.Category == "example-quality" {
			t.Error("Should not flag 3-5 examples")
		}
	}
}

// --- Compaction readiness lint ---

func TestLint_CompactionReadiness(t *testing.T) {
	// Create a very long prompt (>50K tokens = >200K chars)
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 6000)
	results := Lint(text)

	found := false
	for _, r := range results {
		if r.Category == "compaction-readiness" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag long prompts missing compaction guidance")
	}
}

func TestLint_CompactionReadiness_Short(t *testing.T) {
	text := "Fix the bug in sorting."
	results := Lint(text)

	for _, r := range results {
		if r.Category == "compaction-readiness" {
			t.Error("Should not flag short prompts")
		}
	}
}

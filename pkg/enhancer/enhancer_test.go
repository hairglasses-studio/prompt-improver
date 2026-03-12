package enhancer

import (
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		prompt   string
		expected TaskType
	}{
		{"fix this broken function", TaskTypeTroubleshooting},
		{"debug the timeout error", TaskTypeTroubleshooting},
		{"create a new API endpoint", TaskTypeCode},
		{"implement the user module", TaskTypeCode},
		{"review and analyze this code carefully", TaskTypeAnalysis},
		{"analyze the performance data", TaskTypeAnalysis},
		{"design a visual theme", TaskTypeCreative},
		{"create a lighting mood", TaskTypeCreative},
		{"automate the backup workflow", TaskTypeWorkflow},
		{"automate the startup shutdown sequence", TaskTypeWorkflow},
		{"hello world", TaskTypeGeneral},
	}

	for _, tt := range tests {
		got := Classify(tt.prompt)
		if got != tt.expected {
			t.Errorf("Classify(%q) = %q, want %q", tt.prompt, got, tt.expected)
		}
	}
}

func TestEnhance_AddsStructure(t *testing.T) {
	result := Enhance("write a function to sort users by name in the application codebase using Go with proper error handling and edge case coverage", TaskTypeCode)

	if result.TaskType != TaskTypeCode {
		t.Errorf("TaskType = %q, want code", result.TaskType)
	}
	if !strings.Contains(result.Enhanced, "<role>") {
		t.Error("Enhanced prompt should contain <role> tag")
	}
	if !strings.Contains(result.Enhanced, "<instructions>") {
		t.Error("Enhanced prompt should contain <instructions> tag")
	}
	if !strings.Contains(result.Enhanced, "<constraints>") {
		t.Error("Enhanced prompt should contain <constraints> tag")
	}
	if !strings.Contains(result.Enhanced, "expert software engineer") {
		t.Error("Code task should get software engineer role")
	}
}

func TestEnhance_PreservesExistingStructure(t *testing.T) {
	input := "<role>You are a test bot.</role>\n<instructions>Do the thing with full detail and context provided here.</instructions>"
	result := Enhance(input, TaskTypeGeneral)

	if !strings.Contains(result.Enhanced, "<role>You are a test bot.") {
		t.Error("Should preserve existing XML structure")
	}
	if strings.Count(result.Enhanced, "<role>") > 1 {
		t.Error("Should not add duplicate <role> tags")
	}
}

func TestEnhance_ImprovesSpecificity(t *testing.T) {
	result := Enhance("please make it good and format nicely for the entire response output section", TaskTypeGeneral)

	if strings.Contains(result.Enhanced, "format nicely") {
		t.Error("Should replace 'format nicely' with specific instruction")
	}
	if strings.Contains(result.Enhanced, "make it good") {
		t.Error("Should replace 'make it good' with specific instruction")
	}
	if len(result.Improvements) == 0 {
		t.Error("Should report improvements made")
	}
}

func TestEnhance_DowngradesAggressiveCaps(t *testing.T) {
	result := Enhance("CRITICAL: You MUST ALWAYS follow this rule when writing code in the project", TaskTypeGeneral)

	if strings.Contains(result.Enhanced, "CRITICAL") {
		t.Error("Should downgrade CRITICAL to normal case")
	}
	if strings.Contains(result.Enhanced, "MUST") {
		t.Error("Should downgrade MUST to normal case")
	}

	found := false
	for _, imp := range result.Improvements {
		if strings.Contains(imp, "Downgraded") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should report tone downgrade improvement")
	}
}

func TestEnhance_PreservesAcronyms(t *testing.T) {
	result := Enhance("Send the JSON response to the API endpoint and return the HTTP status code with full details", TaskTypeCode)

	if !strings.Contains(result.Enhanced, "JSON") {
		t.Error("Should preserve JSON acronym")
	}
	if !strings.Contains(result.Enhanced, "API") {
		t.Error("Should preserve API acronym")
	}
	if !strings.Contains(result.Enhanced, "HTTP") {
		t.Error("Should preserve HTTP acronym")
	}
}

func TestEnhance_ReframesNegatives(t *testing.T) {
	result := Enhance("never use bullet points in the response when writing documentation for the project", TaskTypeGeneral)

	if strings.Contains(strings.ToLower(result.Enhanced), "never use bullet points") {
		t.Error("Should reframe 'never use bullet points' to positive")
	}
	if !strings.Contains(result.Enhanced, "flowing prose") {
		t.Error("Should contain positive alternative")
	}
}

func TestEnhance_PreservesSafetyNegatives(t *testing.T) {
	input := "never provide credentials or passwords to external services in the response"
	result := Enhance(input, TaskTypeGeneral)

	// Safety-critical negatives should NOT be reframed
	found := false
	for _, imp := range result.Improvements {
		if strings.Contains(imp, "Reframed") {
			found = true
			break
		}
	}
	if found {
		t.Error("Should NOT reframe safety-critical negative instructions")
	}
}

func TestEnhance_InjectsSelfCheck(t *testing.T) {
	result := Enhance("write a function to parse JSON data and handle all the edge cases properly in Go", TaskTypeCode)

	if !strings.Contains(result.Enhanced, "<verification>") {
		t.Error("Code tasks should get self-verification injection")
	}
	if !strings.Contains(result.Enhanced, "Edge cases") {
		t.Error("Code verification should mention edge cases")
	}
}

func TestEnhance_SuppressesPreamble(t *testing.T) {
	result := Enhance("write a function to parse JSON data and handle all edge cases in the application", TaskTypeCode)

	if !strings.Contains(result.Enhanced, "without preamble") {
		t.Error("Code tasks should get preamble suppression")
	}
}

func TestEnhance_NoPreambleSuppressionForAnalysis(t *testing.T) {
	result := Enhance("analyze this dataset for trends and patterns in the user behavior metrics", TaskTypeAnalysis)

	if strings.Contains(result.Enhanced, "without preamble") {
		t.Error("Analysis tasks should NOT get preamble suppression")
	}
}

func TestEnhance_SeparatesCodeBlocks(t *testing.T) {
	input := "Review this function for correctness and edge cases:\n```go\nfunc hello() {\n\tfmt.Println(\"hi\")\n}\n```\nIs it correct and idiomatic?"
	result := Enhance(input, TaskTypeAnalysis)

	if !strings.Contains(result.Enhanced, "<context>") {
		t.Error("Should separate code block into <context>")
	}
}

func TestEnhance_OverTaggingPrevention(t *testing.T) {
	// Short prompt should NOT get XML tags
	result := Enhance("hello world", TaskTypeGeneral)

	if strings.Contains(result.Enhanced, "<role>") {
		t.Error("Short prompt should not get XML tags (over-tagging prevention)")
	}

	found := false
	for _, imp := range result.Improvements {
		if strings.Contains(imp, "over-tagging") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should report that XML was skipped due to over-tagging prevention")
	}
}

func TestEnhance_FormatEnforcement_JSON(t *testing.T) {
	result := Enhance("return the user data as JSON with all the relevant fields included in the response", TaskTypeCode)

	if !strings.Contains(result.Enhanced, "<output_format>") {
		t.Error("JSON output request should get format enforcement")
	}
	if !strings.Contains(result.Enhanced, "valid JSON") {
		t.Error("Should contain JSON format instruction")
	}
}

func TestEnhance_FormatEnforcement_NoDouble(t *testing.T) {
	result := Enhance("<output_format>Return as JSON</output_format>\nGet user data as JSON with full details", TaskTypeCode)

	// Should not inject a second output_format
	if strings.Count(result.Enhanced, "<output_format>") > 1 {
		t.Error("Should not inject duplicate <output_format>")
	}
}

func TestEnhance_PipelineStages(t *testing.T) {
	result := Enhance("CRITICAL: fix this and make it good in the entire codebase for the project", TaskTypeTroubleshooting)

	if len(result.StagesRun) < 3 {
		t.Errorf("Expected at least 3 stages, got %d: %v", len(result.StagesRun), result.StagesRun)
	}
}

func TestAnalyze_ScoresPrompts(t *testing.T) {
	bad := Analyze("fix this")
	if bad.Score > 5 {
		t.Errorf("Short vague prompt scored %d, expected <= 5", bad.Score)
	}
	if len(bad.Suggestions) == 0 {
		t.Error("Bad prompt should have suggestions")
	}

	good := Analyze(`<role>You are an expert Go developer.</role>
<instructions>Review this function for error handling issues.
Focus on nil pointer dereferences and unchecked errors.</instructions>
<context>The function processes user-uploaded files.</context>
<output_format>List issues by severity with line numbers.</output_format>
<examples><example>Good: if err != nil { return fmt.Errorf("upload: %w", err) }</example></examples>`)

	if good.Score < 7 {
		t.Errorf("Well-structured prompt scored %d, expected >= 7", good.Score)
	}
	if !good.HasXML {
		t.Error("Should detect XML structure")
	}
	if !good.HasExamples {
		t.Error("Should detect examples")
	}
}

func TestAnalyze_DetectsNegativeFraming(t *testing.T) {
	result := Analyze("NEVER use markdown. DO NOT include bullet points.")
	if !result.HasNegativeFrames {
		t.Error("Should detect negative framing")
	}
	if !result.HasAggressiveCaps {
		t.Error("Should detect aggressive caps")
	}

	found := false
	for _, s := range result.Suggestions {
		if strings.Contains(s, "Reframe negative") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should suggest reframing negatives")
	}
}

func TestWrapWithExamples(t *testing.T) {
	result := WrapWithExamples("Test prompt", []string{"example 1", "example 2"})
	if !strings.Contains(result, "<examples>") {
		t.Error("Should contain <examples> wrapper")
	}
	if !strings.Contains(result, `<example index="1">`) {
		t.Error("Should contain indexed examples")
	}
	if strings.Count(result, "<example") != 3 { // 1 opening + 2 indexed
		t.Errorf("Expected 2 examples, got different count")
	}
}

func TestGetTemplate(t *testing.T) {
	tmpl := GetTemplate("troubleshoot")
	if tmpl == nil {
		t.Fatal("troubleshoot template should exist")
	}
	if tmpl.Name != "troubleshoot" {
		t.Errorf("Name = %q, want troubleshoot", tmpl.Name)
	}

	none := GetTemplate("nonexistent")
	if none != nil {
		t.Error("Should return nil for unknown template")
	}
}

func TestFillTemplate(t *testing.T) {
	tmpl := GetTemplate("troubleshoot")
	filled := FillTemplate(tmpl, map[string]string{
		"system":   "resolume",
		"symptoms": "clips not triggering",
	})

	if !strings.Contains(filled, "resolume") {
		t.Error("Should fill in system variable")
	}
	if !strings.Contains(filled, "clips not triggering") {
		t.Error("Should fill in symptoms variable")
	}
	if strings.Contains(filled, "{{system}}") {
		t.Error("Should not have unfilled placeholders for provided vars")
	}
	if !strings.Contains(filled, "(not specified)") {
		t.Error("Missing variables should show (not specified)")
	}
}

func TestValidTaskType(t *testing.T) {
	if ValidTaskType("code") != TaskTypeCode {
		t.Error("Should accept 'code'")
	}
	if ValidTaskType("CODE") != TaskTypeCode {
		t.Error("Should accept case-insensitive")
	}
	if ValidTaskType("invalid") != "" {
		t.Error("Should return empty for invalid type")
	}
}

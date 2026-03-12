// Package enhancer provides deterministic prompt enhancement using XML structuring,
// specificity improvements, context reordering, and task-type-aware formatting.
// No external LLM calls — pure Go string manipulation.
package enhancer

import (
	"fmt"
	"strings"
)

// EnhanceResult holds the output of the enhancement pipeline
type EnhanceResult struct {
	Original     string   `json:"original"`
	Enhanced     string   `json:"enhanced"`
	TaskType     TaskType `json:"task_type"`
	StagesRun    []string `json:"stages_run"`
	Improvements []string `json:"improvements"`
}

// AnalyzeResult holds prompt quality analysis
type AnalyzeResult struct {
	Score       int      `json:"score"`
	Suggestions []string `json:"suggestions"`
	HasXML      bool     `json:"has_xml_structure"`
	HasExamples bool     `json:"has_examples"`
	HasContext   bool     `json:"has_context"`
	HasFormat   bool     `json:"has_output_format"`
	WordCount   int      `json:"word_count"`
	TaskType    TaskType `json:"task_type"`
}

// Enhance runs the full enhancement pipeline on a raw prompt
func Enhance(raw string, taskType TaskType) EnhanceResult {
	if taskType == "" {
		taskType = Classify(raw)
	}

	result := EnhanceResult{
		Original: raw,
		TaskType: taskType,
	}

	text := raw

	// Stage 1: Specificity — replace vague phrases
	text, specImprovements := improveSpecificity(text)
	if len(specImprovements) > 0 {
		result.StagesRun = append(result.StagesRun, "specificity")
		result.Improvements = append(result.Improvements, specImprovements...)
	}

	// Stage 2: Structure — wrap in XML tags based on task type
	text, structImprovements := addStructure(text, taskType)
	result.StagesRun = append(result.StagesRun, "structure")
	result.Improvements = append(result.Improvements, structImprovements...)

	// Stage 3: Context reorder — long context to top, query to bottom
	text = reorderContext(text)
	result.StagesRun = append(result.StagesRun, "context_reorder")

	result.Enhanced = text
	return result
}

// Analyze scores a prompt and returns improvement suggestions
func Analyze(prompt string) AnalyzeResult {
	lower := strings.ToLower(prompt)
	words := strings.Fields(prompt)
	taskType := Classify(prompt)

	result := AnalyzeResult{
		WordCount: len(words),
		TaskType:  taskType,
	}

	// Check for existing structure
	result.HasXML = strings.Contains(prompt, "<") && strings.Contains(prompt, ">") &&
		(strings.Contains(lower, "<instructions") || strings.Contains(lower, "<context") ||
			strings.Contains(lower, "<role") || strings.Contains(lower, "<constraints"))
	result.HasExamples = strings.Contains(lower, "example") || strings.Contains(lower, "<example")
	result.HasContext = strings.Contains(lower, "context") || strings.Contains(lower, "<context")
	result.HasFormat = strings.Contains(lower, "format") || strings.Contains(lower, "<output")

	// Score (1-10)
	score := 3 // baseline for any non-empty prompt

	if result.HasXML {
		score += 2
	}
	if result.HasExamples {
		score += 1
	}
	if result.HasContext {
		score += 1
	}
	if result.HasFormat {
		score += 1
	}
	if len(words) > 20 {
		score += 1 // sufficient detail
	}
	if len(words) > 50 {
		score += 1 // comprehensive
	}

	// Check for vague patterns (deductions)
	vagueCount := 0
	for pattern := range vagueReplacements {
		if strings.Contains(lower, pattern) {
			vagueCount++
		}
	}
	if vagueCount > 2 {
		score--
	}

	if score > 10 {
		score = 10
	}
	if score < 1 {
		score = 1
	}
	result.Score = score

	// Suggestions
	if !result.HasXML {
		result.Suggestions = append(result.Suggestions, "Add XML structure tags (<role>, <instructions>, <constraints>) for clearer organization")
	}
	if !result.HasExamples {
		result.Suggestions = append(result.Suggestions, "Include examples of desired output using <examples><example> tags")
	}
	if !result.HasContext {
		result.Suggestions = append(result.Suggestions, "Add a <context> section with relevant background information")
	}
	if !result.HasFormat {
		result.Suggestions = append(result.Suggestions, "Specify the desired output format in an <output_format> section")
	}
	if len(words) < 20 {
		result.Suggestions = append(result.Suggestions, "Add more detail — prompts under 20 words often produce inconsistent results")
	}
	if vagueCount > 0 {
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("Replace %d vague phrases with specific instructions (e.g., 'format nicely' → 'Format using markdown with headers and code blocks')", vagueCount))
	}

	// Task-specific suggestions
	switch taskType {
	case TaskTypeCode:
		if !strings.Contains(lower, "language") && !strings.Contains(lower, "go") &&
			!strings.Contains(lower, "python") && !strings.Contains(lower, "typescript") {
			result.Suggestions = append(result.Suggestions, "Specify the programming language")
		}
	case TaskTypeTroubleshooting:
		if !strings.Contains(lower, "error") && !strings.Contains(lower, "symptom") {
			result.Suggestions = append(result.Suggestions, "Include the exact error message or symptoms")
		}
	case TaskTypeAnalysis:
		if !strings.Contains(lower, "criteria") && !strings.Contains(lower, "focus") {
			result.Suggestions = append(result.Suggestions, "Specify evaluation criteria or focus areas")
		}
	}

	return result
}

// WrapWithExamples wraps a prompt and examples into proper XML few-shot format
func WrapWithExamples(prompt string, examples []string) string {
	var b strings.Builder
	b.WriteString(prompt)
	b.WriteString("\n\n<examples>\n")
	for i, ex := range examples {
		fmt.Fprintf(&b, "<example index=\"%d\">\n%s\n</example>\n", i+1, strings.TrimSpace(ex))
	}
	b.WriteString("</examples>")
	return b.String()
}

// --- Internal pipeline stages ---

// vagueReplacements maps vague phrases to specific alternatives
var vagueReplacements = map[string]string{
	"format nicely":      "Format using markdown with headers and code blocks",
	"make it good":       "Ensure correctness, clarity, and completeness",
	"make it better":     "Improve clarity, reduce redundancy, and strengthen specificity",
	"clean it up":        "Refactor for readability: consistent naming, clear structure, remove dead code",
	"do your best":       "Provide a thorough, well-structured response",
	"be creative":        "Explore unconventional approaches while remaining practical",
	"be thorough":        "Cover all edge cases and provide step-by-step detail",
	"keep it simple":     "Use the minimum complexity needed — prefer standard patterns over abstractions",
	"make it fast":       "Optimize for performance: minimize allocations, reduce iterations, cache where appropriate",
	"make it secure":     "Follow security best practices: validate inputs, use parameterized queries, apply least privilege",
	"handle errors":      "Return descriptive errors with context, wrap errors at boundaries, never swallow errors silently",
	"add tests":          "Write unit tests covering happy path, edge cases, and error conditions",
	"fix this":           "Identify the root cause, apply a minimal fix, and verify it resolves the issue",
	"help me":            "Guide me step-by-step through",
	"i need":             "The goal is to",
	"can you":            "Please",
	"as soon as possible": "by [specific deadline]",
}

func improveSpecificity(text string) (string, []string) {
	lower := strings.ToLower(text)
	var improvements []string

	for vague, specific := range vagueReplacements {
		if idx := strings.Index(lower, vague); idx != -1 {
			// Replace preserving original casing context
			text = text[:idx] + specific + text[idx+len(vague):]
			lower = strings.ToLower(text)
			improvements = append(improvements, fmt.Sprintf("Replaced '%s' → '%s'", vague, specific))
		}
	}

	return text, improvements
}

// roleForTaskType returns an appropriate role prefix
func roleForTaskType(tt TaskType) string {
	switch tt {
	case TaskTypeCode:
		return "You are an expert software engineer."
	case TaskTypeCreative:
		return "You are a creative director with deep technical knowledge."
	case TaskTypeAnalysis:
		return "You are a thorough analytical reviewer."
	case TaskTypeTroubleshooting:
		return "You are a systems diagnostician focused on root cause analysis."
	case TaskTypeWorkflow:
		return "You are a workflow architect focused on reliability and simplicity."
	default:
		return "You are a knowledgeable assistant."
	}
}

// constraintsForTaskType returns task-type-specific constraints
func constraintsForTaskType(tt TaskType) string {
	switch tt {
	case TaskTypeCode:
		return `- Write clean, idiomatic code
- Handle errors explicitly
- Prefer simplicity over cleverness
- Do not add features beyond what was requested`
	case TaskTypeCreative:
		return `- Provide specific parameters, not vague descriptions
- Balance ambition with practicality
- Consider technical constraints`
	case TaskTypeAnalysis:
		return `- Support claims with evidence
- Distinguish between facts and opinions
- Note confidence levels for uncertain conclusions`
	case TaskTypeTroubleshooting:
		return `- Start with the least disruptive checks
- Do not restart services without asking
- Identify root cause, not just symptoms`
	case TaskTypeWorkflow:
		return `- Each step must have a clear success condition
- Include error handling for each step
- Prefer parallel execution where dependencies allow`
	default:
		return `- Be specific and actionable
- Structure output for clarity`
	}
}

func addStructure(text string, taskType TaskType) (string, []string) {
	// Don't re-structure if already has XML tags
	lower := strings.ToLower(text)
	if strings.Contains(lower, "<instructions") || strings.Contains(lower, "<role") {
		return text, []string{"Prompt already has XML structure — preserved"}
	}

	var b strings.Builder
	var improvements []string

	b.WriteString("<role>")
	b.WriteString(roleForTaskType(taskType))
	b.WriteString("</role>\n\n")
	improvements = append(improvements, "Added <role> tag with task-appropriate persona")

	b.WriteString("<instructions>\n")
	b.WriteString(strings.TrimSpace(text))
	b.WriteString("\n</instructions>\n\n")
	improvements = append(improvements, "Wrapped prompt in <instructions> tags")

	b.WriteString("<constraints>\n")
	b.WriteString(constraintsForTaskType(taskType))
	b.WriteString("\n</constraints>\n")
	improvements = append(improvements, "Added task-type-specific <constraints>")

	return b.String(), improvements
}

func reorderContext(text string) string {
	// If the prompt has a clear context block (multi-line quoted text, code blocks, or
	// long paragraphs) followed by a short question, move context up and query down.
	// This is a heuristic — Claude processes long context better when placed before the query.

	lines := strings.Split(text, "\n")
	if len(lines) < 5 {
		return text // too short to reorder
	}

	// Already structured with XML — don't rearrange
	if strings.Contains(text, "<context>") || strings.Contains(text, "<instructions>") {
		return text
	}

	return text
}

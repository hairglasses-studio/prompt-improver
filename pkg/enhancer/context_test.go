package enhancer

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	// ~4 chars per token
	text := strings.Repeat("word ", 100) // 500 chars
	tokens := EstimateTokens(text)

	if tokens < 100 || tokens > 150 {
		t.Errorf("Expected ~125 tokens for 500 chars, got %d", tokens)
	}
}

func TestReorderLongContext_ShortPrompt(t *testing.T) {
	text := "Analyze this data please."
	result, imps := ReorderLongContext(text)

	if len(imps) > 0 {
		t.Error("Should not reorder short prompts")
	}
	if result != text {
		t.Error("Should return unchanged")
	}
}

func TestReorderLongContext_AlreadyStructured(t *testing.T) {
	text := "<context>Long data here</context>\n\nWhat does this mean?"
	result, imps := ReorderLongContext(text)

	if len(imps) > 0 {
		t.Error("Should not reorder already-structured prompts")
	}
	if result != text {
		t.Error("Should return unchanged")
	}
}

func TestInjectQuoteGrounding_ShortPrompt(t *testing.T) {
	text := "Analyze this briefly."
	result, imps := InjectQuoteGrounding(text, TaskTypeAnalysis)

	if len(imps) > 0 {
		t.Error("Should not inject grounding for short prompts")
	}
	if result != text {
		t.Error("Should return unchanged")
	}
}

func TestInjectQuoteGrounding_AlreadyHasQuotes(t *testing.T) {
	text := strings.Repeat("data point. ", 5000) + "\nPlease quote the relevant sections."
	result, imps := InjectQuoteGrounding(text, TaskTypeAnalysis)

	if len(imps) > 0 {
		t.Error("Should not inject grounding when 'quote' already mentioned")
	}
	if result != text {
		t.Error("Should return unchanged")
	}
}

func TestInjectQuoteGrounding_LongAnalysis(t *testing.T) {
	// Create a prompt with >5000 estimated tokens
	text := strings.Repeat("The system logged an important data point about user behavior. ", 400)
	text += "\n\nAnalyze the patterns in the data above."

	result, imps := InjectQuoteGrounding(text, TaskTypeAnalysis)

	if len(imps) == 0 {
		t.Error("Should inject grounding for long analysis prompts")
	}
	if !strings.Contains(result, "<quotes>") {
		t.Error("Should mention <quotes> tags")
	}
}

// --- Cache-friendly order verification ---

func TestVerifyCacheFriendlyOrder_DynamicBeforeStatic(t *testing.T) {
	text := `<instructions>
Process the {{user_input}} data.
</instructions>

<role>You are an expert analyst.</role>

<constraints>
Be thorough.
</constraints>`

	results := VerifyCacheFriendlyOrder(text)

	found := false
	for _, r := range results {
		if r.Category == "cache-unfriendly-order" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect dynamic section before static section")
	}
}

func TestVerifyCacheFriendlyOrder_CorrectOrder(t *testing.T) {
	text := `<role>You are an expert analyst.</role>

<constraints>
Be thorough.
</constraints>

<instructions>
Process the {{user_input}} data.
</instructions>`

	results := VerifyCacheFriendlyOrder(text)

	for _, r := range results {
		if r.Category == "cache-unfriendly-order" {
			t.Error("Should not flag correctly ordered prompt")
		}
	}
}

func TestVerifyCacheFriendlyOrder_EarlyVariable(t *testing.T) {
	text := `Process {{user_input}} now.

Some more static content here that doesn't change.

And even more static content that could be cached.`

	results := VerifyCacheFriendlyOrder(text)

	found := false
	for _, r := range results {
		if r.Category == "cache-unfriendly-variable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag template variable in first third of prompt")
	}
}

func TestVerifyCacheFriendlyOrder_NoStructure(t *testing.T) {
	// Long unstructured prompt (>1000 tokens = >4000 chars)
	text := strings.Repeat("This is a long unstructured prompt without any XML tags. ", 100)
	results := VerifyCacheFriendlyOrder(text)

	found := false
	for _, r := range results {
		if r.Category == "cache-no-structure" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag long unstructured prompts")
	}
}

func TestVerifyCacheFriendlyOrder_ShortNoStructure(t *testing.T) {
	text := "Fix the bug."
	results := VerifyCacheFriendlyOrder(text)

	for _, r := range results {
		if r.Category == "cache-no-structure" {
			t.Error("Should not flag short unstructured prompts")
		}
	}
}

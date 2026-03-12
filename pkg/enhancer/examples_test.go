package enhancer

import (
	"strings"
	"testing"
)

func TestDetectAndWrapExamples_InputOutputPairs(t *testing.T) {
	input := `Convert these dates to ISO format:

Input: January 5, 2024
Output: 2024-01-05

Input: March 15, 2023
Output: 2023-03-15

Input: December 31, 2025
Output: 2025-12-31

Now convert: February 28, 2026`

	result, imps := DetectAndWrapExamples(input)

	if len(imps) == 0 {
		t.Fatal("Should detect and wrap input/output pairs")
	}
	if !strings.Contains(result, "<examples>") {
		t.Error("Should contain <examples> wrapper")
	}
	if !strings.Contains(result, `<example index="1">`) {
		t.Error("Should contain indexed example tags")
	}
	if strings.Count(result, "<example index=") < 2 {
		t.Error("Should have at least 2 wrapped examples")
	}
}

func TestDetectAndWrapExamples_ExampleHeaders(t *testing.T) {
	input := `Format user data as follows:

Example 1: Simple user
Name: John, Age: 30 -> {"name": "John", "age": 30}

Example 2: With email
Name: Jane, Email: jane@test.com -> {"name": "Jane", "email": "jane@test.com"}

Now format this data.`

	result, imps := DetectAndWrapExamples(input)

	if len(imps) == 0 {
		t.Fatal("Should detect and wrap example headers")
	}
	if !strings.Contains(result, "<examples>") {
		t.Error("Should contain <examples> wrapper")
	}
}

func TestDetectAndWrapExamples_ArrowTransformations(t *testing.T) {
	input := `Convert snake_case to camelCase:

user_name -> userName
first_name -> firstName
last_updated_at -> lastUpdatedAt

Convert: my_variable_name`

	result, imps := DetectAndWrapExamples(input)

	if len(imps) == 0 {
		t.Fatal("Should detect and wrap arrow transformations")
	}
	if !strings.Contains(result, "<examples>") {
		t.Error("Should contain <examples> wrapper")
	}
}

func TestDetectAndWrapExamples_AlreadyTagged(t *testing.T) {
	input := `<examples><example>Already tagged</example></examples>`
	result, imps := DetectAndWrapExamples(input)

	if len(imps) > 0 {
		t.Error("Should not double-wrap already tagged examples")
	}
	if result != input {
		t.Error("Should return input unchanged")
	}
}

func TestDetectAndWrapExamples_NoExamples(t *testing.T) {
	input := "Write a function to sort an array of integers."
	_, imps := DetectAndWrapExamples(input)

	if len(imps) > 0 {
		t.Error("Should not detect examples in plain instruction")
	}
}

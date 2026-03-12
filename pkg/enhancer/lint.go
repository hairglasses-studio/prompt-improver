package enhancer

import (
	"fmt"
	"regexp"
	"strings"
)

// LintResult represents a single lint finding
type LintResult struct {
	Line        int    `json:"line"`
	Category    string `json:"category"`
	Severity    string `json:"severity"` // "error", "warn", "info"
	Original    string `json:"original"`
	Suggestion  string `json:"suggestion"`
	AutoFixable bool   `json:"auto_fixable"`
}

// directivePattern detects imperative/directive sentences
var directivePattern = regexp.MustCompile(
	`(?mi)^(always|never|do\s+not|don't|must|should|ensure|make\s+sure|be\s+sure)\s+.+[\.\n]`,
)

// motivationMarkers indicate the directive explains WHY
var motivationMarkers = regexp.MustCompile(
	`(?i)\b(because|since|so\s+that|in\s+order\s+to|to\s+(ensure|prevent|avoid|enable|allow|improve|reduce|maintain)|this\s+(helps|ensures|prevents|allows|enables|improves)|otherwise|the\s+reason|as\s+this|which\s+(helps|ensures|prevents))\b`,
)

// Lint runs all lint checks on a prompt and returns findings.
// This is deeper than Analyze — it returns per-line actionable findings.
func Lint(text string) []LintResult {
	var results []LintResult
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for unmotivated directives
		if r := checkUnmotivatedRule(i+1, trimmed); r != nil {
			results = append(results, *r)
		}

		// Check for negative framing (that isn't in our reframing table)
		if r := checkNegativeFraming(i+1, trimmed); r != nil {
			results = append(results, *r)
		}

		// Check for aggressive emphasis
		if r := checkAggressiveEmphasis(i+1, trimmed); r != nil {
			results = append(results, *r)
		}

		// Check for vague quantifiers
		if r := checkVagueQuantifiers(i+1, trimmed); r != nil {
			results = append(results, *r)
		}
	}

	return results
}

func checkUnmotivatedRule(lineNum int, line string) *LintResult {
	if !directivePattern.MatchString(line) {
		return nil
	}
	if motivationMarkers.MatchString(line) {
		return nil // has motivation
	}
	// Skip very short lines (probably not standalone rules)
	if len(strings.Fields(line)) < 4 {
		return nil
	}

	return &LintResult{
		Line:        lineNum,
		Category:    "unmotivated-rule",
		Severity:    "info",
		Original:    line,
		Suggestion:  "Add a 'because...' clause — motivated instructions improve compliance. Per Anthropic: Claude generalizes better from understanding the purpose.",
		AutoFixable: false,
	}
}

func checkNegativeFraming(lineNum int, line string) *LintResult {
	if !negativePattern.MatchString(line) {
		return nil
	}
	// Skip safety-critical negatives
	if safetyNegativePattern.MatchString(line) {
		return nil
	}
	// Skip if it's already in our reframing table (handled by the enhancer)
	lower := strings.ToLower(line)
	for neg := range negativeReframings {
		if strings.Contains(lower, neg) {
			return nil
		}
	}

	return &LintResult{
		Line:       lineNum,
		Category:   "negative-framing",
		Severity:   "warn",
		Original:   line,
		Suggestion: "Reframe as a positive instruction — tell Claude what to do, not what to avoid. Per Anthropic: negative framing can cause reverse psychology with Claude 4.x.",
		AutoFixable: false,
	}
}

func checkAggressiveEmphasis(lineNum int, line string) *LintResult {
	matches := aggressiveCapsPattern.FindAllString(line, -1)
	if len(matches) == 0 {
		return nil
	}
	// Filter out acronyms
	var real []string
	for _, m := range matches {
		if !acronymWhitelist[m] {
			real = append(real, m)
		}
	}
	if len(real) == 0 {
		return nil
	}

	return &LintResult{
		Line:        lineNum,
		Category:    "aggressive-emphasis",
		Severity:    "info",
		Original:    line,
		Suggestion:  fmt.Sprintf("Downgrade %s to normal case — Claude 4.x overtriggers on aggressive ALL-CAPS language.", strings.Join(real, ", ")),
		AutoFixable: true,
	}
}

var vagueQuantifierPattern = regexp.MustCompile(`(?i)\b(a few|some|several|many|a lot|a bit|enough|various|appropriate|suitable|proper|good|nice|decent)\b`)

func checkVagueQuantifiers(lineNum int, line string) *LintResult {
	matches := vagueQuantifierPattern.FindAllString(line, -1)
	if len(matches) == 0 {
		return nil
	}

	return &LintResult{
		Line:        lineNum,
		Category:    "vague-quantifier",
		Severity:    "info",
		Original:    line,
		Suggestion:  fmt.Sprintf("Replace vague quantifier(s) %q with specific numbers — '3-5 items' is better than 'several items'.", matches),
		AutoFixable: false,
	}
}

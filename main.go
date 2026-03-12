// prompt-improver is a CLI tool that enhances prompts with XML structure,
// specificity improvements, and task-type-aware formatting.
//
// Designed to run as a Claude Code UserPromptSubmit hook for automatic
// prompt enhancement, or as a standalone CLI.
//
// Usage:
//
//	echo "fix this bug" | prompt-improver
//	prompt-improver enhance "fix this bug"
//	prompt-improver analyze "fix this bug"
//	prompt-improver template troubleshoot --system resolume --symptoms "clips stuck"
//	prompt-improver templates
//	prompt-improver hook  (reads Claude Code UserPromptSubmit JSON from stdin)
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hairglasses-studio/prompt-improver/pkg/enhancer"
)

func main() {
	args := os.Args[1:]

	// If no args and stdin has data, read from stdin (pipe mode)
	if len(args) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
		raw := strings.TrimSpace(string(input))
		if raw == "" {
			fmt.Fprintln(os.Stderr, "usage: prompt-improver <command> [args] or pipe prompt via stdin")
			os.Exit(1)
		}
		runEnhance(raw, "")
		return
	}

	switch args[0] {
	case "enhance":
		taskType := ""
		prompt := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "--type" && i+1 < len(args) {
				taskType = args[i+1]
				i++
			} else if prompt == "" {
				prompt = args[i]
			} else {
				prompt += " " + args[i]
			}
		}
		if prompt == "" {
			prompt = readStdin()
		}
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "usage: prompt-improver enhance <prompt> [--type code|creative|analysis|troubleshooting|workflow|general]")
			os.Exit(1)
		}
		runEnhance(prompt, taskType)

	case "analyze":
		prompt := strings.Join(args[1:], " ")
		if prompt == "" {
			prompt = readStdin()
		}
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "usage: prompt-improver analyze <prompt>")
			os.Exit(1)
		}
		runAnalyze(prompt)

	case "template":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: prompt-improver template <name> [--var value ...]")
			os.Exit(1)
		}
		runTemplate(args[1], args[2:])

	case "templates":
		fmt.Print(enhancer.TemplateListSummary())

	case "lint":
		prompt := strings.Join(args[1:], " ")
		if prompt == "" {
			prompt = readStdin()
		}
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "usage: prompt-improver lint <prompt>")
			os.Exit(1)
		}
		runLint(prompt)

	case "cache-check":
		path := ""
		if len(args) > 1 {
			path = args[1]
		}
		runCacheCheck(path)

	case "check-claudemd":
		path := "./CLAUDE.md"
		if len(args) > 1 {
			path = args[1]
		}
		runCheckClaudeMD(path)

	case "hook":
		// Hook mode: reads JSON from stdin (Claude Code UserPromptSubmit format)
		runHook()

	case "version":
		fmt.Println("prompt-improver v1.0.0")

	case "help", "--help", "-h":
		printHelp()

	default:
		// Treat everything as a prompt to enhance
		prompt := strings.Join(args, " ")
		runEnhance(prompt, "")
	}
}

func runEnhance(prompt, taskType string) {
	tt := enhancer.ValidTaskType(taskType)
	result := enhancer.Enhance(prompt, tt)

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func runAnalyze(prompt string) {
	result := enhancer.Analyze(prompt)
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func runLint(prompt string) {
	results := enhancer.Lint(prompt)
	if len(results) == 0 {
		fmt.Println("No issues found.")
		return
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
}

func runCacheCheck(path string) {
	var text string
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		text = string(data)
	} else {
		text = readStdin()
	}
	if text == "" {
		fmt.Fprintln(os.Stderr, "usage: prompt-improver cache-check <file> or pipe via stdin")
		os.Exit(1)
	}

	results := enhancer.VerifyCacheFriendlyOrder(text)
	if len(results) == 0 {
		fmt.Println("Cache-friendly: no ordering issues found.")
		return
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
}

func runCheckClaudeMD(path string) {
	results, err := enhancer.CheckClaudeMD(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Println("CLAUDE.md looks healthy — no issues found.")
		return
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
}

func runTemplate(name string, args []string) {
	tmpl := enhancer.GetTemplate(name)
	if tmpl == nil {
		fmt.Fprintf(os.Stderr, "unknown template: %s\n\nAvailable templates:\n", name)
		for _, t := range enhancer.ListTemplates() {
			fmt.Fprintf(os.Stderr, "  %s - %s\n", t.Name, t.Description)
		}
		os.Exit(1)
	}

	vars := parseFlags(args)
	filled := enhancer.FillTemplate(tmpl, vars)
	fmt.Println(filled)
}

// hookInput is the JSON Claude Code sends to UserPromptSubmit hooks on stdin
type hookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`
	Prompt         string `json:"prompt"`
}

// hookOutput is the JSON response for UserPromptSubmit hooks
type hookOutput struct {
	HookSpecificOutput *hookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

func runHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1) // non-blocking error, prompt proceeds
	}

	var hi hookInput
	if err := json.Unmarshal(input, &hi); err != nil {
		// Not JSON — treat as raw prompt text
		raw := strings.TrimSpace(string(input))
		if raw != "" {
			result := enhancer.Enhance(raw, "")
			fmt.Println(result.Enhanced)
		}
		return
	}

	// If no prompt field, pass through
	if hi.Prompt == "" {
		os.Exit(0)
		return
	}

	// Load project-specific config if cwd is set
	cfg := enhancer.Config{}
	if hi.Cwd != "" {
		cfg = enhancer.LoadConfig(hi.Cwd)
	}

	// Enhance the prompt with config
	result := enhancer.EnhanceWithConfig(hi.Prompt, "", cfg)

	// Build additional context with the enhanced version
	var context strings.Builder
	context.WriteString("## Prompt Enhancement Applied\n\n")
	context.WriteString("The following enhanced version of the prompt incorporates best practices:\n\n")
	context.WriteString(result.Enhanced)
	context.WriteString("\n\n### Improvements Made\n")
	for _, imp := range result.Improvements {
		fmt.Fprintf(&context, "- %s\n", imp)
	}
	fmt.Fprintf(&context, "\nDetected task type: %s\n", result.TaskType)

	// Output structured JSON per Claude Code hook spec
	out := hookOutput{
		HookSpecificOutput: &hookSpecificOutput{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: context.String(),
		},
	}

	data, _ := json.Marshal(out)
	fmt.Println(string(data))
	os.Exit(0)
}

func readStdin() string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "" // no piped input
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(input))
}

func parseFlags(args []string) map[string]string {
	vars := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") && i+1 < len(args) {
			key := strings.TrimPrefix(args[i], "--")
			vars[key] = args[i+1]
			i++
		}
	}
	return vars
}

func printHelp() {
	fmt.Print(`prompt-improver v1.0.0 — Claude-specific prompt optimization CLI

USAGE:
  prompt-improver <prompt>                      Enhance a prompt (default)
  prompt-improver enhance <prompt> [--type T]   Enhance with explicit task type
  prompt-improver analyze <prompt>              Score, suggest, estimate tokens & effort
  prompt-improver lint <prompt>                 Deep lint with per-line findings
  prompt-improver cache-check <file>            Check prompt caching friendliness
  prompt-improver check-claudemd [path]         CLAUDE.md health check (default: ./CLAUDE.md)
  prompt-improver template <name> [--var val]   Fill a prompt template
  prompt-improver templates                     List available templates
  prompt-improver hook                          Claude Code hook mode (JSON stdin)
  echo "prompt" | prompt-improver               Pipe mode

PIPELINE (13 stages):
  0  config_rules         Pattern-matched augmentations
  1  specificity          Replace vague phrases
  2  positive_reframe     Negative-to-positive reframing
  3  tone_downgrade       ALL-CAPS → normal case
  4  overtrigger_rewrite  Soften anti-laziness phrases (Claude 4.x)
  5  example_wrapping     Wrap bare examples in XML
  6  structure            Add XML role/instructions/constraints
  7  context_reorder      Long context before query
  8  format_enforcement   JSON/YAML/CSV format tags
  9  quote_grounding      Quote-first for long-context analysis
  10 self_check           Verification checklists
  11 overengineering_guard Prevent over-abstraction (code tasks)
  12 preamble_suppression Direct response instruction

TASK TYPES:
  code, creative, analysis, troubleshooting, workflow, general

LINT CHECKS:
  unmotivated-rule, negative-framing, aggressive-emphasis, vague-quantifier,
  overtrigger-phrase, over-specification, decomposition-needed, injection-risk,
  thinking-mode-redundant, example-quality, compaction-readiness

CLAUDE CODE HOOK INTEGRATION:
  Add to .claude/settings.json (project) or ~/.claude/settings.json (global):

    {
      "hooks": {
        "UserPromptSubmit": [
          {
            "hooks": [
              {
                "type": "command",
                "command": "prompt-improver hook",
                "timeout": 10
              }
            ]
          }
        ]
      }
    }

  The hook reads the UserPromptSubmit JSON from stdin, enhances the prompt,
  and returns additionalContext that Claude sees alongside the original prompt.
  Exit code 0 = proceed, exit code 2 = block the prompt.
`)
}

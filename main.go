// prompt-improver is a CLI tool that enhances prompts with XML structure,
// specificity improvements, and task-type-aware formatting.
//
// Designed to run as a Claude Code hook on PreToolUse or as a standalone CLI.
//
// Usage:
//
//	echo "fix this bug" | prompt-improver
//	prompt-improver enhance "fix this bug"
//	prompt-improver analyze "fix this bug"
//	prompt-improver template troubleshoot --system resolume --symptoms "clips stuck"
//	prompt-improver templates
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

	// If no args and stdin has data, read from stdin (hook mode)
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

	case "hook":
		// Hook mode: reads JSON from stdin (Claude Code hook format)
		runHook()

	case "version":
		fmt.Println("prompt-improver v0.1.0")

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

func runHook() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Claude Code hook format: JSON with tool_name, tool_input, etc.
	var hookInput struct {
		ToolName  string                 `json:"tool_name"`
		ToolInput map[string]interface{} `json:"tool_input"`
	}

	if err := json.Unmarshal(input, &hookInput); err != nil {
		// Not JSON — treat as raw prompt
		raw := strings.TrimSpace(string(input))
		if raw != "" {
			result := enhancer.Enhance(raw, "")
			fmt.Println(result.Enhanced)
		}
		return
	}

	// Look for prompt-like fields in the tool input
	promptFields := []string{"prompt", "message", "text", "query", "content", "input"}
	for _, field := range promptFields {
		if val, ok := hookInput.ToolInput[field]; ok {
			if str, ok := val.(string); ok && len(str) > 10 {
				result := enhancer.Enhance(str, "")
				// Output modified tool input
				hookInput.ToolInput[field] = result.Enhanced
				data, _ := json.Marshal(map[string]interface{}{
					"tool_input": hookInput.ToolInput,
				})
				fmt.Println(string(data))
				return
			}
		}
	}

	// No prompt field found — pass through unchanged
	fmt.Println("{}")
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
	fmt.Print(`prompt-improver — Automatic prompt enhancement CLI

USAGE:
  prompt-improver <prompt>                      Enhance a prompt (default)
  prompt-improver enhance <prompt> [--type T]   Enhance with explicit task type
  prompt-improver analyze <prompt>              Score and suggest improvements
  prompt-improver template <name> [--var val]   Fill a prompt template
  prompt-improver templates                     List available templates
  prompt-improver hook                          Claude Code hook mode (JSON stdin)
  echo "prompt" | prompt-improver               Pipe mode

TASK TYPES:
  code, creative, analysis, troubleshooting, workflow, general

TEMPLATES:
  troubleshoot, code_review, workflow_create, data_analysis, creative_brief

HOOK INTEGRATION:
  Add to .claude/settings.json:
    {
      "hooks": {
        "PreToolUse": [{
          "matcher": "Task",
          "command": "prompt-improver hook"
        }]
      }
    }
`)
}

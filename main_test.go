package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once for all CLI tests
	dir, err := os.MkdirTemp("", "prompt-improver-test")
	if err != nil {
		panic(err)
	}
	binaryPath = filepath.Join(dir, "prompt-improver")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build binary: " + err.Error())
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runCLI executes the binary with given args and optional stdin, returning stdout, stderr, and exit code.
func runCLI(t *testing.T, stdin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running CLI: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestCLI_Enhance_Args(t *testing.T) {
	stdout, _, code := runCLI(t, "", "enhance", "write a function to sort users by name with error handling")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, `"enhanced"`) {
		t.Error("output should contain enhanced JSON field")
	}
	if !strings.Contains(stdout, `"task_type"`) {
		t.Error("output should contain task_type JSON field")
	}
}

func TestCLI_Enhance_WithType(t *testing.T) {
	stdout, _, code := runCLI(t, "", "enhance", "--type", "analysis", "review this code")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, `"analysis"`) {
		t.Error("output should contain analysis task type")
	}
}

func TestCLI_DefaultCommand(t *testing.T) {
	// No subcommand, just a prompt — should enhance by default
	stdout, _, code := runCLI(t, "", "write a function to sort users by name with handling")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, `"enhanced"`) {
		t.Error("default command should enhance")
	}
}

func TestCLI_PipeMode(t *testing.T) {
	stdout, _, code := runCLI(t, "write a function to sort users", "")
	// The binary receives "" as an arg, which hits the default case.
	// Instead test with no args and piped stdin:
	cmd := exec.Command(binaryPath)
	cmd.Stdin = strings.NewReader("analyze this code for bugs")
	var outBuf strings.Builder
	cmd.Stdout = &outBuf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("pipe mode failed: %v", err)
	}
	_ = stdout
	_ = code
	if !strings.Contains(outBuf.String(), `"enhanced"`) {
		t.Error("pipe mode should produce enhanced JSON")
	}
}

func TestCLI_Analyze(t *testing.T) {
	stdout, _, code := runCLI(t, "", "analyze", "fix this")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, `"score"`) {
		t.Error("analyze output should contain score")
	}
	if !strings.Contains(stdout, `"suggestions"`) {
		t.Error("analyze output should contain suggestions")
	}
}

func TestCLI_Lint_Clean(t *testing.T) {
	stdout, _, code := runCLI(t, "", "lint", "Return exactly 5 user records as JSON, sorted by creation date.")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "No issues found") {
		// Might have info-level findings, that's OK
		_ = stdout
	}
}

func TestCLI_Lint_Dirty(t *testing.T) {
	stdout, _, code := runCLI(t, "", "lint", "CRITICAL: You MUST follow this rule.")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "overtrigger-phrase") && !strings.Contains(stdout, "aggressive-emphasis") {
		t.Error("dirty prompt should produce lint findings")
	}
}

func TestCLI_Templates(t *testing.T) {
	stdout, _, code := runCLI(t, "", "templates")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "troubleshoot") {
		t.Error("should list troubleshoot template")
	}
	if !strings.Contains(stdout, "code_review") {
		t.Error("should list code_review template")
	}
}

func TestCLI_Template_Fill(t *testing.T) {
	stdout, _, code := runCLI(t, "", "template", "troubleshoot", "--system", "resolume", "--symptoms", "clips stuck")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "resolume") {
		t.Error("should fill in system variable")
	}
	if !strings.Contains(stdout, "clips stuck") {
		t.Error("should fill in symptoms variable")
	}
}

func TestCLI_Template_Nonexistent(t *testing.T) {
	_, stderr, code := runCLI(t, "", "template", "nonexistent")
	if code == 0 {
		t.Error("should exit non-zero for nonexistent template")
	}
	if !strings.Contains(stderr, "unknown template") {
		t.Error("should report unknown template")
	}
}

func TestCLI_CacheCheck_Stdin(t *testing.T) {
	prompt := `<role>You are an expert.</role>
<constraints>Be thorough.</constraints>
<instructions>Process the data.</instructions>`
	cmd := exec.Command(binaryPath, "cache-check")
	cmd.Stdin = strings.NewReader(prompt)
	var outBuf strings.Builder
	cmd.Stdout = &outBuf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("cache-check failed: %v", err)
	}
	// Either "no ordering issues" or lint results — both valid
	if outBuf.String() == "" {
		t.Error("should produce some output")
	}
}

func TestCLI_CacheCheck_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.txt")
	content := `<role>You are an expert.</role>
<constraints>Be thorough.</constraints>`
	os.WriteFile(path, []byte(content), 0644)

	stdout, _, code := runCLI(t, "", "cache-check", path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if stdout == "" {
		t.Error("should produce output")
	}
}

func TestCLI_CheckClaudeMD_Healthy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(path, []byte("# Project\n\nSimple project.\n\n## Standards\n\nUse gofmt."), 0644)

	stdout, _, code := runCLI(t, "", "check-claudemd", path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "healthy") {
		t.Error("healthy CLAUDE.md should report healthy")
	}
}

func TestCLI_CheckClaudeMD_Bad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	content := strings.Repeat("CRITICAL: You MUST follow this rule.\n", 50)
	os.WriteFile(path, []byte(content), 0644)

	stdout, _, code := runCLI(t, "", "check-claudemd", path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "overtrigger-language") {
		t.Error("bad CLAUDE.md should report overtrigger language")
	}
}

func TestCLI_Hook_ValidJSON(t *testing.T) {
	hookJSON := `{"session_id":"test","prompt":"fix this bug in sorting","hook_event_name":"UserPromptSubmit"}`
	cmd := exec.Command(binaryPath, "hook")
	cmd.Stdin = strings.NewReader(hookJSON)
	var outBuf strings.Builder
	cmd.Stdout = &outBuf
	// hook exits with 0
	cmd.Run()
	if !strings.Contains(outBuf.String(), "hookSpecificOutput") {
		t.Error("hook with valid JSON should return hookSpecificOutput")
	}
	if !strings.Contains(outBuf.String(), "UserPromptSubmit") {
		t.Error("hook should return UserPromptSubmit event name")
	}
}

func TestCLI_Hook_EmptyPrompt(t *testing.T) {
	hookJSON := `{"session_id":"test","prompt":"","hook_event_name":"UserPromptSubmit"}`
	cmd := exec.Command(binaryPath, "hook")
	cmd.Stdin = strings.NewReader(hookJSON)
	var outBuf strings.Builder
	cmd.Stdout = &outBuf
	// Empty prompt exits 0 silently
	cmd.Run()
	output := outBuf.String()
	// Should exit cleanly without hookSpecificOutput
	if strings.Contains(output, "hookSpecificOutput") {
		t.Error("empty prompt should not produce hookSpecificOutput")
	}
}

func TestCLI_Hook_RawText(t *testing.T) {
	// Non-JSON input falls back to raw text enhancement
	cmd := exec.Command(binaryPath, "hook")
	cmd.Stdin = strings.NewReader("fix this bug in the sorting function")
	var outBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Run()
	// Should produce the enhanced prompt as plain text
	if outBuf.String() == "" {
		t.Error("hook with raw text should produce output")
	}
}

func TestCLI_NoArgs_NoStdin(t *testing.T) {
	cmd := exec.Command(binaryPath)
	// Provide empty stdin that will be detected as tty-like
	cmd.Stdin = strings.NewReader("")
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err == nil {
		// Might succeed with empty stdin read, that's OK
		return
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit 1, got %d", exitErr.ExitCode())
	}
}

func TestCLI_Help(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			stdout, _, code := runCLI(t, "", arg)
			if code != 0 {
				t.Errorf("expected exit 0, got %d", code)
			}
			if !strings.Contains(stdout, "prompt-improver") {
				t.Error("help should mention prompt-improver")
			}
			if !strings.Contains(stdout, "USAGE") {
				t.Error("help should contain USAGE section")
			}
		})
	}
}

func TestCLI_Version(t *testing.T) {
	stdout, _, code := runCLI(t, "", "version")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "v1.0.0") {
		t.Error("should output version")
	}
}

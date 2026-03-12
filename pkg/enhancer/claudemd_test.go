package enhancer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCheckClaudeMD_ExcessiveLength(t *testing.T) {
	content := strings.Repeat("line\n", 250)
	path := writeTempFile(t, content)

	results, err := CheckClaudeMD(path)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range results {
		if r.Category == "excessive-length" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag CLAUDE.md with >200 lines")
	}
}

func TestCheckClaudeMD_OvertriggerLanguage(t *testing.T) {
	content := "# Rules\n\nCRITICAL: You MUST always follow the coding standards.\nIMPORTANT: You SHOULD never skip tests."
	path := writeTempFile(t, content)

	results, err := CheckClaudeMD(path)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range results {
		if r.Category == "overtrigger-language" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag overtrigger language in CLAUDE.md")
	}
}

func TestCheckClaudeMD_InlineCode(t *testing.T) {
	content := "# Code\n\n```go\nfunc a() {}\n```\n\n```go\nfunc b() {}\n```\n\n```go\nfunc c() {}\n```\n\n```go\nfunc d() {}\n```\n"
	path := writeTempFile(t, content)

	results, err := CheckClaudeMD(path)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range results {
		if r.Category == "inline-code" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should flag excessive inline code blocks")
	}
}

func TestCheckClaudeMD_Healthy(t *testing.T) {
	content := "# Project\n\nThis is a simple Go project.\n\n## Standards\n\nUse gofmt."
	path := writeTempFile(t, content)

	results, err := CheckClaudeMD(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) > 0 {
		t.Errorf("Healthy CLAUDE.md should have no findings, got %d", len(results))
	}
}

func TestCheckClaudeMD_MissingFile(t *testing.T) {
	_, err := CheckClaudeMD("/nonexistent/CLAUDE.md")
	if err == nil {
		t.Error("Should return error for missing file")
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall_CreatesClaudeSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	s := &settingsJSON{
		Hooks:      make(map[string][]hookGroup),
		McpServers: make(map[string]mcpServerEntry),
	}

	addHookEntry(s, "/usr/local/bin/prompt-improver")
	addMCPEntry(s, "/usr/local/bin/prompt-improver")

	if err := writeSettings(path, s, nil); err != nil {
		t.Fatalf("writeSettings failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "prompt-improver hook") {
		t.Error("settings should contain hook command")
	}
	if !strings.Contains(content, `"mcpServers"`) {
		t.Error("settings should contain mcpServers")
	}
	if !strings.Contains(content, `"mcp"`) {
		t.Error("settings should contain mcp args")
	}
}

func TestClaudeInstallIdempotent(t *testing.T) {
	s := &settingsJSON{
		Hooks:      make(map[string][]hookGroup),
		McpServers: make(map[string]mcpServerEntry),
	}

	exe := "/usr/local/bin/prompt-improver"
	addHookEntry(s, exe)
	addHookEntry(s, exe)
	if len(s.Hooks["UserPromptSubmit"]) != 1 {
		t.Fatalf("expected 1 hook group, got %d", len(s.Hooks["UserPromptSubmit"]))
	}

	addMCPEntry(s, exe)
	addMCPEntry(s, exe)
	if len(s.McpServers) != 1 {
		t.Fatalf("expected 1 MCP entry, got %d", len(s.McpServers))
	}
}

func TestClaudeUninstallRemovesOnlyPromptImproverEntries(t *testing.T) {
	s := &settingsJSON{
		Hooks: map[string][]hookGroup{
			"UserPromptSubmit": {
				{Hooks: []hookEntry{{Type: "command", Command: "other-tool hook"}}},
				{Hooks: []hookEntry{{Type: "command", Command: "/usr/local/bin/prompt-improver hook"}}},
			},
		},
		McpServers: map[string]mcpServerEntry{
			"prompt-improver": {Type: "stdio", Command: "/usr/local/bin/prompt-improver", Args: []string{"mcp"}},
			"other-tool":      {Type: "stdio", Command: "/usr/local/bin/other-tool"},
		},
	}

	removeHookEntry(s)
	removeMCPEntry(s)

	groups := s.Hooks["UserPromptSubmit"]
	if len(groups) != 1 || groups[0].Hooks[0].Command != "other-tool hook" {
		t.Fatal("should preserve non-prompt-improver Claude hooks")
	}
	if _, ok := s.McpServers["prompt-improver"]; ok {
		t.Fatal("prompt-improver MCP entry should be removed")
	}
	if _, ok := s.McpServers["other-tool"]; !ok {
		t.Fatal("other-tool MCP entry should remain")
	}
}

func TestReadWriteSettingsPreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	initial := map[string]any{
		"someOtherSetting": true,
		"hooks":            map[string]any{},
	}
	data, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	s, raw, err := readSettings(path)
	if err != nil {
		t.Fatalf("readSettings failed: %v", err)
	}
	addHookEntry(s, "/usr/local/bin/prompt-improver")
	if err := writeSettings(path, s, raw); err != nil {
		t.Fatalf("writeSettings failed: %v", err)
	}

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result)
	if !strings.Contains(content, "someOtherSetting") {
		t.Error("should preserve unknown keys")
	}
	if !strings.Contains(content, "prompt-improver hook") {
		t.Error("should add hook")
	}
}

func TestCodexHookInstallAndRemove(t *testing.T) {
	h := &codexHooksJSON{Hooks: make(map[string][]hookGroup)}
	addCodexHookEntry(h, "/usr/local/bin/prompt-improver")
	addCodexHookEntry(h, "/usr/local/bin/prompt-improver")
	if len(h.Hooks["UserPromptSubmit"]) != 1 {
		t.Fatalf("expected 1 Codex hook group, got %d", len(h.Hooks["UserPromptSubmit"]))
	}

	removeCodexHookEntry(h)
	if len(h.Hooks["UserPromptSubmit"]) != 0 {
		t.Fatal("expected prompt-improver Codex hook to be removed")
	}
}

func TestWriteCodexHooksCreatesHooksFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".codex", "hooks.json")
	h := &codexHooksJSON{Hooks: make(map[string][]hookGroup)}
	addCodexHookEntry(h, "/usr/local/bin/prompt-improver")

	if err := writeCodexHooks(path, h, nil); err != nil {
		t.Fatalf("writeCodexHooks failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read hooks file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "UserPromptSubmit") {
		t.Error("hooks.json should contain UserPromptSubmit")
	}
	if !strings.Contains(content, "prompt-improver hook") {
		t.Error("hooks.json should contain prompt-improver hook command")
	}
}

func TestCodexConfigUpsertsFeatureAndMCPBlock(t *testing.T) {
	initial := "model = \"gpt-5.4-xhigh\"\n"
	updated := upsertCodexFeatureFlag(initial, "codex_hooks", true)
	updated = upsertCodexMCPServer(updated, "prompt-improver", "/usr/local/bin/prompt-improver", []string{"mcp"})

	if !strings.Contains(updated, "[features]") || !strings.Contains(updated, "codex_hooks = true") {
		t.Fatal("expected codex_hooks feature flag to be present")
	}
	if !strings.Contains(updated, "[mcp_servers.prompt-improver]") {
		t.Fatal("expected prompt-improver MCP section to be present")
	}
	if !strings.Contains(updated, "command = \"/usr/local/bin/prompt-improver\"") {
		t.Fatal("expected prompt-improver command to be written")
	}

	updated = removeTomlSection(updated, "mcp_servers.prompt-improver")
	if strings.Contains(updated, "[mcp_servers.prompt-improver]") {
		t.Fatal("expected prompt-improver MCP section to be removed")
	}
	if !strings.Contains(updated, "codex_hooks = true") {
		t.Fatal("feature flag should remain to avoid breaking other hooks")
	}
}

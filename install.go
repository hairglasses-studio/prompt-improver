package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const promptImproverCmd = "prompt-improver"

type installOptions struct {
	global   bool
	hookOnly bool
	mcpOnly  bool
	provider string
}

// settingsJSON represents the relevant parts of Claude Code's settings.json.
type settingsJSON struct {
	Hooks      map[string][]hookGroup     `json:"hooks,omitempty"`
	McpServers map[string]mcpServerEntry  `json:"mcpServers,omitempty"`
	Rest       map[string]json.RawMessage `json:"-"`
}

type hookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []hookEntry `json:"hooks"`
}

type hookEntry struct {
	Type          string `json:"type"`
	Command       string `json:"command"`
	Timeout       int    `json:"timeout,omitempty"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

type mcpServerEntry struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type codexHooksJSON struct {
	Hooks map[string][]hookGroup `json:"hooks,omitempty"`
}

func runInstall(args []string) {
	opts := parseInstallOptions(args)

	exe, err := resolveExecutablePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot resolve executable path: %v\n", err)
		os.Exit(1)
	}

	installHook := !opts.mcpOnly
	installMCP := !opts.hookOnly

	switch resolvedProvider(opts.provider) {
	case "claude":
		if err := installClaude(opts.global, exe, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "codex":
		if err := installCodex(opts.global, exe, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "both":
		if err := installClaude(opts.global, exe, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := installCodex(opts.global, exe, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported provider %q\n", opts.provider)
		os.Exit(1)
	}
}

func runUninstall(args []string) {
	opts := parseInstallOptions(args)
	installHook := !opts.mcpOnly
	installMCP := !opts.hookOnly

	switch resolvedProvider(opts.provider) {
	case "claude":
		if err := uninstallClaude(opts.global, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "codex":
		if err := uninstallCodex(opts.global, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "both":
		if err := uninstallClaude(opts.global, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := uninstallCodex(opts.global, installHook, installMCP); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported provider %q\n", opts.provider)
		os.Exit(1)
	}
}

func parseInstallOptions(args []string) installOptions {
	opts := installOptions{provider: "auto"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--global":
			opts.global = true
		case "--hook-only":
			opts.hookOnly = true
		case "--mcp-only":
			opts.mcpOnly = true
		case "--provider":
			if i+1 < len(args) {
				opts.provider = strings.ToLower(strings.TrimSpace(args[i+1]))
				i++
			}
		}
	}
	return opts
}

func resolvedProvider(provider string) string {
	switch provider {
	case "", "auto":
		if codexPreferred() {
			return "codex"
		}
		return "claude"
	case "claude", "codex", "both":
		return provider
	default:
		return provider
	}
}

func codexPreferred() bool {
	return commandExists("codex")
}

func commandExists(name string) bool {
	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return true
		}
	}
	return false
}

func resolveExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = resolved
	}
	return exe, nil
}

func installClaude(global bool, exe string, installHook bool, installMCP bool) error {
	settingsPath := settingsPathFor(global)
	settings, raw, err := readSettings(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", settingsPath, err)
	}

	if installHook {
		addHookEntry(settings, exe)
		fmt.Printf("Installed Claude UserPromptSubmit hook -> %s\n", settingsPath)
	}
	if installMCP {
		addMCPEntry(settings, exe)
		fmt.Printf("Installed Claude MCP server -> %s\n", settingsPath)
	}

	if !installHook && !installMCP {
		return nil
	}
	return writeSettings(settingsPath, settings, raw)
}

func uninstallClaude(global bool, removeHook bool, removeMCP bool) error {
	settingsPath := settingsPathFor(global)
	settings, raw, err := readSettings(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Nothing to uninstall for Claude -> %s does not exist.\n", settingsPath)
			return nil
		}
		return fmt.Errorf("reading %s: %w", settingsPath, err)
	}

	if removeHook {
		removeHookEntry(settings)
	}
	if removeMCP {
		removeMCPEntry(settings)
	}

	if err := writeSettings(settingsPath, settings, raw); err != nil {
		return err
	}
	fmt.Printf("Updated Claude settings -> %s\n", settingsPath)
	return nil
}

func installCodex(global bool, exe string, installHook bool, installMCP bool) error {
	if installHook {
		hooksPath := codexHooksPathFor(global)
		hooks, raw, err := readCodexHooks(hooksPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("reading %s: %w", hooksPath, err)
		}
		addCodexHookEntry(hooks, exe)
		if err := writeCodexHooks(hooksPath, hooks, raw); err != nil {
			return err
		}
		fmt.Printf("Installed Codex UserPromptSubmit hook -> %s\n", hooksPath)
	}

	if installHook || installMCP {
		configPath := codexConfigPathFor(global)
		configText, err := readTextFileOrEmpty(configPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", configPath, err)
		}
		configText = upsertCodexFeatureFlag(configText, "codex_hooks", true)
		if installMCP {
			configText = upsertCodexMCPServer(configText, promptImproverCmd, exe, []string{"mcp"})
			fmt.Printf("Installed Codex MCP server -> %s\n", configPath)
		}
		if err := writeTextFile(configPath, configText); err != nil {
			return err
		}
		if installHook {
			fmt.Printf("Enabled Codex hooks -> %s\n", configPath)
		}
	}

	return nil
}

func uninstallCodex(global bool, removeHook bool, removeMCP bool) error {
	if removeHook {
		hooksPath := codexHooksPathFor(global)
		hooks, raw, err := readCodexHooks(hooksPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("reading %s: %w", hooksPath, err)
			}
		} else {
			removeCodexHookEntry(hooks)
			if err := writeCodexHooks(hooksPath, hooks, raw); err != nil {
				return err
			}
			fmt.Printf("Updated Codex hooks -> %s\n", hooksPath)
		}
	}

	if removeMCP {
		configPath := codexConfigPathFor(global)
		configText, err := readTextFileOrEmpty(configPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", configPath, err)
		}
		configText = removeTomlSection(configText, "mcp_servers."+promptImproverCmd)
		if err := writeTextFile(configPath, configText); err != nil {
			return err
		}
		fmt.Printf("Updated Codex config -> %s\n", configPath)
	}

	return nil
}

func settingsPathFor(global bool) string {
	if global {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "settings.json")
	}
	return filepath.Join(".claude", "settings.json")
}

func codexHooksPathFor(global bool) string {
	if global {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".codex", "hooks.json")
	}
	return filepath.Join(".codex", "hooks.json")
}

func codexConfigPathFor(global bool) string {
	if global {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".codex", "config.toml")
	}
	return filepath.Join(".codex", "config.toml")
}

func readSettings(path string) (*settingsJSON, map[string]json.RawMessage, error) {
	s := &settingsJSON{
		Hooks:      make(map[string][]hookGroup),
		McpServers: make(map[string]mcpServerEntry),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return s, nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return s, nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if hooksRaw, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(hooksRaw, &s.Hooks)
	}
	if mcpRaw, ok := raw["mcpServers"]; ok {
		_ = json.Unmarshal(mcpRaw, &s.McpServers)
	}

	return s, raw, nil
}

func writeSettings(path string, s *settingsJSON, raw map[string]json.RawMessage) error {
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}
	if len(s.Hooks) > 0 {
		data, _ := json.Marshal(s.Hooks)
		raw["hooks"] = data
	} else {
		delete(raw, "hooks")
	}
	if len(s.McpServers) > 0 {
		data, _ := json.Marshal(s.McpServers)
		raw["mcpServers"] = data
	} else {
		delete(raw, "mcpServers")
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeBytes(path, data)
}

func addHookEntry(s *settingsJSON, exe string) {
	cmd := exe + " hook"
	eventName := "UserPromptSubmit"
	for _, group := range s.Hooks[eventName] {
		for _, h := range group.Hooks {
			if strings.Contains(h.Command, promptImproverCmd) {
				return
			}
		}
	}

	entry := hookGroup{
		Hooks: []hookEntry{
			{
				Type:    "command",
				Command: cmd,
				Timeout: 30,
			},
		},
	}
	s.Hooks[eventName] = append(s.Hooks[eventName], entry)
}

func addMCPEntry(s *settingsJSON, exe string) {
	if _, ok := s.McpServers[promptImproverCmd]; ok {
		return
	}
	s.McpServers[promptImproverCmd] = mcpServerEntry{
		Type:    "stdio",
		Command: exe,
		Args:    []string{"mcp"},
	}
}

func removeHookEntry(s *settingsJSON) {
	eventName := "UserPromptSubmit"
	groups := s.Hooks[eventName]
	var kept []hookGroup
	for _, group := range groups {
		var keptHooks []hookEntry
		for _, h := range group.Hooks {
			if !strings.Contains(h.Command, promptImproverCmd) {
				keptHooks = append(keptHooks, h)
			}
		}
		if len(keptHooks) > 0 {
			group.Hooks = keptHooks
			kept = append(kept, group)
		}
	}
	if len(kept) > 0 {
		s.Hooks[eventName] = kept
	} else {
		delete(s.Hooks, eventName)
	}
}

func removeMCPEntry(s *settingsJSON) {
	delete(s.McpServers, promptImproverCmd)
}

func readCodexHooks(path string) (*codexHooksJSON, map[string]json.RawMessage, error) {
	h := &codexHooksJSON{Hooks: make(map[string][]hookGroup)}
	data, err := os.ReadFile(path)
	if err != nil {
		return h, nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return h, nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if hooksRaw, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(hooksRaw, &h.Hooks)
	}
	return h, raw, nil
}

func writeCodexHooks(path string, h *codexHooksJSON, raw map[string]json.RawMessage) error {
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}
	if len(h.Hooks) > 0 {
		data, _ := json.Marshal(h.Hooks)
		raw["hooks"] = data
	} else {
		delete(raw, "hooks")
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeBytes(path, data)
}

func addCodexHookEntry(h *codexHooksJSON, exe string) {
	eventName := "UserPromptSubmit"
	command := exe + " hook"
	for _, group := range h.Hooks[eventName] {
		for _, entry := range group.Hooks {
			if strings.Contains(entry.Command, promptImproverCmd) {
				return
			}
		}
	}
	h.Hooks[eventName] = append(h.Hooks[eventName], hookGroup{
		Hooks: []hookEntry{
			{
				Type:    "command",
				Command: command,
				Timeout: 30,
			},
		},
	})
}

func removeCodexHookEntry(h *codexHooksJSON) {
	eventName := "UserPromptSubmit"
	var kept []hookGroup
	for _, group := range h.Hooks[eventName] {
		var remaining []hookEntry
		for _, entry := range group.Hooks {
			if !strings.Contains(entry.Command, promptImproverCmd) {
				remaining = append(remaining, entry)
			}
		}
		if len(remaining) > 0 {
			group.Hooks = remaining
			kept = append(kept, group)
		}
	}
	if len(kept) == 0 {
		delete(h.Hooks, eventName)
		return
	}
	h.Hooks[eventName] = kept
}

func readTextFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func writeTextFile(path string, text string) error {
	return writeBytes(path, []byte(strings.TrimRight(text, "\n")+"\n"))
}

func writeBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func upsertCodexFeatureFlag(content, key string, enabled bool) string {
	body := upsertTomlKey(sectionBody(content, "features"), key, tomlBool(enabled))
	return replaceTomlSection(content, "features", body)
}

func upsertCodexMCPServer(content, name, command string, args []string) string {
	bodyLines := []string{
		fmt.Sprintf("command = %s", tomlString(command)),
		fmt.Sprintf("args = %s", tomlArray(args)),
	}
	return replaceTomlSection(content, "mcp_servers."+name, strings.Join(bodyLines, "\n"))
}

func upsertTomlKey(body, key, value string) string {
	lines := splitTomlLines(body)
	keyPrefix := key + " ="
	replaced := false
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, keyPrefix) {
			out = append(out, fmt.Sprintf("%s = %s", key, value))
			replaced = true
			continue
		}
		if strings.TrimSpace(line) == "" && len(out) == 0 {
			continue
		}
		out = append(out, line)
	}
	if !replaced {
		out = append(out, fmt.Sprintf("%s = %s", key, value))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func sectionBody(content, section string) string {
	start, end, ok := tomlSectionRange(content, section)
	if !ok {
		return ""
	}
	header := "[" + section + "]"
	sectionText := strings.TrimSpace(content[start:end])
	sectionText = strings.TrimPrefix(sectionText, header)
	return strings.TrimSpace(sectionText)
}

func replaceTomlSection(content, section, body string) string {
	block := "[" + section + "]\n" + strings.TrimSpace(body)
	start, end, ok := tomlSectionRange(content, section)
	if ok {
		updated := strings.TrimRight(content[:start], "\n")
		if updated != "" {
			updated += "\n\n"
		}
		updated += block
		tail := strings.TrimLeft(content[end:], "\n")
		if tail != "" {
			updated += "\n\n" + tail
		}
		return strings.TrimSpace(updated)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return block
	}
	return content + "\n\n" + block
}

func removeTomlSection(content, section string) string {
	start, end, ok := tomlSectionRange(content, section)
	if !ok {
		return strings.TrimSpace(content)
	}
	updated := strings.TrimRight(content[:start], "\n")
	tail := strings.TrimLeft(content[end:], "\n")
	if updated != "" && tail != "" {
		return updated + "\n\n" + tail
	}
	if updated != "" {
		return updated
	}
	return tail
}

func tomlSectionRange(content, section string) (int, int, bool) {
	header := "[" + section + "]"
	lines := strings.SplitAfter(content, "\n")
	offset := 0
	start := -1
	end := len(content)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if start == -1 && trimmed == header {
			start = offset
			offset += len(line)
			for j := i + 1; j < len(lines); j++ {
				nextTrimmed := strings.TrimSpace(lines[j])
				if strings.HasPrefix(nextTrimmed, "[") && strings.HasSuffix(nextTrimmed, "]") {
					end = offset
					return start, end, true
				}
				offset += len(lines[j])
			}
			return start, len(content), true
		}
		offset += len(line)
	}
	return 0, 0, false
}

func splitTomlLines(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	return strings.Split(body, "\n")
}

func tomlString(value string) string {
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func tomlArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, tomlString(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func tomlBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

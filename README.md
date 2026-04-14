# prompt-improver

Archived compatibility surface. The standalone prompt-improver repo is no
longer required by the active workspace maintenance set.

- Active prompt workflow development now lives in `ralphglasses`.
- Canonical `dotfiles/mcp/dotfiles-mcp` vendors the enhancer code it still
  needs in-tree, so it no longer depends on this repo.
- Keep this repo limited to redirect notes and archive-safe compatibility
  context.

[![Go Reference](https://pkg.go.dev/badge/github.com/hairglasses-studio/prompt-improver.svg)](https://pkg.go.dev/github.com/hairglasses-studio/prompt-improver)
[![Go Report Card](https://goreportcard.com/badge/github.com/hairglasses-studio/prompt-improver)](https://goreportcard.com/report/github.com/hairglasses-studio/prompt-improver)
[![CI](https://github.com/hairglasses-studio/prompt-improver/actions/workflows/ci.yml/badge.svg)](https://github.com/hairglasses-studio/prompt-improver/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Glama](https://glama.ai/mcp/servers/hairglasses-studio/prompt-improver/badges/score.svg)](https://glama.ai/mcp/servers/hairglasses-studio/prompt-improver)

Provider-aware prompt optimization CLI and MCP server. prompt-improver runs a
deterministic 13-stage enhancement pipeline that adds XML structure, specificity
improvements, and task-type-aware formatting to prompts. It also supports
LLM-powered improvement via Claude for deeper, domain-specific rewrites. Designed
to run as a `UserPromptSubmit` hook for automatic prompt enhancement in Claude
Code or Codex, as a standalone CLI, or as an MCP stdio server.

## Install

### From source

```bash
go install github.com/hairglasses-studio/prompt-improver@latest
```

### Build locally

```bash
git clone https://github.com/hairglasses-studio/prompt-improver.git
cd prompt-improver
make build
sudo make install   # installs to /usr/local/bin
```

## Usage

### Enhance a prompt

```bash
# Pipe mode
echo "fix this bug" | prompt-improver

# Positional argument
prompt-improver enhance "fix this bug"

# With task type hint
prompt-improver enhance "fix this bug" --type code

# Quiet mode (enhanced text only, no JSON envelope)
prompt-improver enhance "fix this bug" --quiet
```

### LLM-powered improvement

```bash
# Direct LLM improvement is currently disabled
prompt-improver improve "fix this bug"

# Local fallback remains available
prompt-improver enhance "fix this" --mode auto
```

### Analyze and lint

```bash
# Score a prompt across 10 quality dimensions (0-100)
prompt-improver analyze "implement a REST API"

# Lint for 11 anti-patterns
prompt-improver lint "NEVER use any external libraries!!!"

# Diff original vs enhanced
prompt-improver diff "fix the thing"

# Check cache-friendly ordering
prompt-improver cache-check CLAUDE.md

# Health-check a CLAUDE.md file
prompt-improver check-claudemd
```

### Templates

```bash
# List available prompt templates
prompt-improver templates

# Fill a template with variables
prompt-improver template troubleshoot --system resolume --symptoms "clips stuck"
```

### Hook integration

```bash
# Install hook + MCP for your preferred client (auto-detects Claude or Codex)
prompt-improver install --global

# Install for both Claude Code and Codex
prompt-improver install --global --provider both

# Remove prompt-improver from all clients
prompt-improver uninstall --global --provider both
```

### MCP server

```bash
# Start the MCP stdio server (7 tools)
prompt-improver mcp
```

Register with a client:

```bash
claude mcp add --transport stdio prompt-improver --scope user -- prompt-improver mcp
codex mcp add prompt-improver -- prompt-improver mcp
```

## Enhancement Pipeline

The local pipeline applies 13 stages in order:

| # | Stage | Description |
|---|-------|-------------|
| 0 | config_rules | Pattern-matched augmentations from `.prompt-improver.yaml` |
| 1 | specificity | Replace vague phrases with concrete language |
| 2 | positive_reframe | Negative-to-positive reframing |
| 3 | tone_downgrade | ALL-CAPS to normal case |
| 4 | overtrigger_rewrite | Soften anti-laziness phrases (Claude 4.x) |
| 5 | example_wrapping | Wrap bare examples in XML |
| 6 | structure | Add XML role/instructions/constraints |
| 7 | context_reorder | Move long context before query |
| 8 | format_enforcement | JSON/YAML/CSV format tags |
| 9 | quote_grounding | Quote-first for long-context analysis |
| 10 | self_check | Verification checklists |
| 11 | overengineering_guard | Prevent over-abstraction (code tasks) |
| 12 | preamble_suppression | Direct response instruction |

## MCP Tools

| Tool | Description |
|------|-------------|
| `analyze_prompt` | Score across 10 quality dimensions with letter grades |
| `enhance_prompt` | Apply the 13-stage enhancement pipeline |
| `lint_prompt` | Lint for 11 anti-patterns including cache ordering |
| `diff_prompt` | Unified diff of original vs enhanced prompt |
| `improve_prompt` | LLM-powered improvement with local fallback |
| `check_claudemd` | Health-check a CLAUDE.md file |
| `list_templates` | List available prompt templates |

## Configuration

Create a `.prompt-improver.yaml` in your project root or home directory:

```yaml
hook:
  skip_score_threshold: 75   # skip enhancement if score >= this
  min_word_count: 5           # skip prompts shorter than this

llm:
  enabled: false              # hosted LLM improvement is currently disabled
```

Archived compatibility coverage now includes a repo-owned promptfoo suite:

```bash
~/hairglasses-studio/dotfiles/scripts/hg-promptfoo.sh . eval -c promptfoo/promptfooconfig.yaml
```

## Key Patterns

- **Smart filtering**: Short, conversational, and already-structured prompts are
  automatically skipped in hook mode.
- **Score gate**: Prompts scoring above the threshold are passed through unmodified.
- **Circuit breaker**: The disabled LLM path remains guarded by the existing circuit breaker and cache primitives.
- **Hybrid fallback**: `--mode auto` currently resolves to the deterministic local pipeline because hosted LLM improvement is disabled.
- **Task-type detection**: Automatic classification into code, creative, analysis,
  troubleshooting, workflow, or general.

## Development

```bash
make build          # Build binary
make test           # Run all tests with race detection
make lint           # Run go vet + staticcheck
make cover          # Coverage report
make bench          # Benchmarks
make ci             # Full CI pipeline
```

## License

[MIT](LICENSE) -- Copyright 2024-2026 hairglasses-studio

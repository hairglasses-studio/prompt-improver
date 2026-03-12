# prompt-improver

Automatic prompt enhancement CLI — restructures raw prompts with XML tags, few-shot format, and best practices before execution.

**Zero external LLM calls.** Pure deterministic Go string manipulation. Fast, reliable, free.

## Install

```bash
go install github.com/hairglasses-studio/prompt-improver@latest
```

Or build from source:

```bash
git clone https://github.com/hairglasses-studio/prompt-improver
cd prompt-improver
go build -o prompt-improver .
```

## Quick Start — Claude Code Integration

One command to enhance every prompt you send to Claude Code:

```bash
prompt-improver install --global
```

Done. Every prompt is now silently improved — unless it's short, conversational, already well-structured, or already scores well.

To remove:

```bash
prompt-improver uninstall --global
```

### Install Options

```bash
prompt-improver install                # project-local (.claude/settings.json)
prompt-improver install --global       # global (~/.claude/settings.json)
prompt-improver install --hook-only    # just the UserPromptSubmit hook
prompt-improver install --mcp-only     # just the MCP server (3 tools)
```

The installer is idempotent — running it twice won't create duplicates.

## Usage

### Enhance a prompt

```bash
# Pipe mode
echo "fix this broken function" | prompt-improver

# Direct
prompt-improver enhance "create an API endpoint for user registration" --type code

# Auto-detect task type
prompt-improver "debug the timeout error in the auth service"
```

### Analyze prompt quality

```bash
prompt-improver analyze "write a function"
# Returns: 10-dimension score report (0-100), letter grades, suggestions, token estimate, effort recommendation
```

### Lint a prompt

```bash
prompt-improver lint "CRITICAL: You MUST follow this rule. NEVER ignore it."
# Returns: per-line findings for 11 anti-patterns
```

### Check prompt caching friendliness

```bash
prompt-improver cache-check prompt.txt
```

### CLAUDE.md health check

```bash
prompt-improver check-claudemd
prompt-improver check-claudemd ./path/to/CLAUDE.md
```

### Use templates

```bash
# List available templates
prompt-improver templates

# Fill a template
prompt-improver template troubleshoot --system resolume --symptoms "clips not triggering"
prompt-improver template code_review --language Go --focus "error handling"
prompt-improver template creative_brief --mood "dark techno" --medium visuals
```

## Enhancement Pipeline (13 stages)

| # | Stage | Description |
|---|-------|-------------|
| 0 | `config_rules` | Pattern-matched augmentations from `.prompt-improver.yaml` |
| 1 | `specificity` | Replace vague phrases with concrete instructions |
| 2 | `positive_reframe` | Negative-to-positive reframing |
| 3 | `tone_downgrade` | ALL-CAPS → normal case (Claude 4.x best practice) |
| 4 | `overtrigger_rewrite` | Soften aggressive anti-laziness phrases |
| 5 | `example_wrapping` | Wrap bare examples in XML `<example>` tags |
| 6 | `structure` | Add XML `<role>`/`<instructions>`/`<constraints>` |
| 7 | `context_reorder` | Long context before query |
| 8 | `format_enforcement` | JSON/YAML/CSV format tags |
| 9 | `quote_grounding` | Quote-first for long-context analysis |
| 10 | `self_check` | Verification checklists |
| 11 | `overengineering_guard` | Prevent over-abstraction (code tasks) |
| 12 | `preamble_suppression` | Direct response instruction |

## Scoring (10 dimensions)

`prompt-improver analyze` scores prompts across:

- **Clarity** — word count, vague phrases, numeric constraints
- **Specificity** — quantified constraints, format specification
- **Context & Motivation** — background, "because" clauses
- **Structure** — XML tags, paragraph organization
- **Examples** — few-shot examples in proper tags
- **Document Placement** — cache-friendly ordering
- **Role Definition** — persona and expertise level
- **Task Focus** — single task, action verbs
- **Format Specification** — output format instructions
- **Tone** — aggressive language, negative framing

Each dimension gets a 0-100 score and letter grade (A/B/C/D/F). The weighted overall score drives the hook's score gate.

## Lint Rules (11 checks)

`prompt-improver lint` detects:

`unmotivated-rule`, `negative-framing`, `aggressive-emphasis`, `vague-quantifier`, `overtrigger-phrase`, `over-specification`, `decomposition-needed`, `injection-risk`, `thinking-mode-redundant`, `example-quality`, `compaction-readiness`

## Hook Behavior

When running as a Claude Code `UserPromptSubmit` hook, prompt-improver applies smart filtering:

1. **Word count gate** — prompts under 5 words are skipped ("yes", "ok", "do it")
2. **Conversational allowlist** — known short replies are skipped (lgtm, ship it, continue, etc.)
3. **Already-structured gate** — prompts with `<instructions>` or `<role>` tags pass through
4. **File-path gate** — bare file paths/globs pass through
5. **Score gate** — prompts scoring ≥ 75 are already good and pass through unchanged

The enhanced prompt is returned as lean XML in `additionalContext`:

```
<enhanced_prompt>
[enhanced text]
</enhanced_prompt>
Follow the enhanced version above. It adds structure and specificity to the original request.
```

### Hook Configuration

Create `.prompt-improver.yaml` in your project root to customize:

```yaml
# Hook settings
hook:
  skip_score_threshold: 75   # skip enhancement if score >= this (0 = always enhance)
  min_word_count: 5          # skip prompts shorter than this
  skip_patterns:             # additional regex patterns to skip
    - "(?i)deploy"

# Pipeline settings
preamble: "This project uses Go 1.22 with standard library only."
default_task_type: code
disabled_stages:
  - preamble_suppression

# Pattern-matched augmentations
rules:
  - match: "add tool"
    append: "Follow the existing tool registration pattern using init()."
  - match: "fix"
    prepend: "Include a test that reproduces the bug."

# Block dangerous prompts (exit code 2)
block_patterns:
  - "drop table"
  - "rm -rf /"
```

## Task Types

Automatically detected via keyword matching, or specify with `--type`:

| Type | Triggers | Role |
|------|----------|------|
| `code` | create, build, implement, refactor... | Expert software engineer |
| `troubleshooting` | debug, error, fix, broken, crash... | Systems diagnostician |
| `analysis` | review, analyze, compare, evaluate... | Analytical reviewer |
| `creative` | design, visual, music, mood, aesthetic... | Creative director |
| `workflow` | workflow, automate, sequence, pipeline... | Workflow architect |
| `general` | (default) | Knowledgeable assistant |

## MCP Server

prompt-improver exposes 3 tools via MCP (Model Context Protocol):

- `analyze_prompt` — multi-dimensional scoring
- `enhance_prompt` — full 13-stage pipeline
- `lint_prompt` — deep lint with cache-order checks

Install with:

```bash
prompt-improver install --mcp-only --global
```

Or manually:

```bash
claude mcp add --transport stdio prompt-improver --scope user -- prompt-improver mcp
```

## Templates

| Template | Purpose |
|----------|---------|
| `troubleshoot` | Structured diagnostics for system issues |
| `code_review` | Code review with severity levels |
| `workflow_create` | Multi-step automation design |
| `data_analysis` | Dataset analysis with structured output |
| `creative_brief` | Creative direction for AV design |

## License

MIT

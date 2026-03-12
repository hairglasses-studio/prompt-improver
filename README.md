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
# Returns: score (1-10), suggestions, detected structure
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

### Claude Code hook mode

```bash
# Reads JSON from stdin in Claude Code hook format
prompt-improver hook
```

## Enhancement Pipeline

The enhancer runs 3 deterministic stages:

1. **Specificity** — Replaces vague phrases with concrete instructions (e.g., "format nicely" → "Format using markdown with headers and code blocks")
2. **Structure** — Wraps prompt in XML tags (`<role>`, `<instructions>`, `<constraints>`) with task-type-appropriate content
3. **Context reorder** — Ensures long context appears before the query (Claude's preferred placement)

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

## Templates

| Template | Purpose |
|----------|---------|
| `troubleshoot` | Structured diagnostics for system issues |
| `code_review` | Code review with severity levels |
| `workflow_create` | Multi-step automation design |
| `data_analysis` | Dataset analysis with structured output |
| `creative_brief` | Creative direction for AV design |

## Claude Code Integration

### As a hook (auto-run before prompts)

Add to your Claude Code settings (`.claude/settings.json`):

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Task",
      "command": "prompt-improver hook"
    }]
  }
}
```

### As a slash command

Create `.claude/commands/enhance.md`:

```markdown
Enhance the following prompt using prompt-improver:

1. Run: `prompt-improver analyze "$ARGUMENTS"`
2. Run: `prompt-improver enhance "$ARGUMENTS"`
3. Present the analysis and enhanced version
4. Ask if I want to proceed with the enhanced version
```

## License

MIT

# Roadmap

Ideas and directions for prompt-improver. These are possibilities, not commitments — contributions and discussion welcome.

## Features

- **Prompt versioning/history** — Store enhancement history in `~/.prompt-improver/history.jsonl` for replay, comparison, and rollback of previous enhancements.
- **A/B testing of pipeline stages** — `enhance --ab-test` outputs with-stage and without-stage versions for each stage, showing the delta each one contributes.
- **Custom stage plugins** — Load user-defined enhancement stages from `.prompt-improver/stages/` as Go plugins or WASM modules for project-specific transformations.
- **Batch mode** — `prompt-improver batch prompts.jsonl` processes multiple prompts from a file, outputting results as newline-delimited JSON.
- **Prompt decomposition** — Automatically split large multi-concern prompts into focused sub-prompts, each optimized for a single task.
- **Model-specific optimization profiles** — `--target-model haiku|sonnet|opus` tunes XML density, constraint verbosity, and scratchpad depth for the target model's capabilities.
- **Interactive refinement loop** — `prompt-improver refine` enters a REPL with real-time scoring feedback after each edit.
- **Prompt library/snippets** — Named snippets in `~/.prompt-improver/snippets/` composable via `prompt-improver compose go-errors + json-output + test-first`.
- **Multi-turn conversation context** — In hook mode, read the Claude Code transcript to understand conversation history and avoid redundant structure in follow-ups.
- **Score regression detection** — `prompt-improver watch` monitors CLAUDE.md and project prompts, alerting when scoring regressions occur after file changes.
- **Config validation command** — `prompt-improver config check` validates `.prompt-improver.yaml` syntax, warns about conflicting options, and reports effective merged config.
- **Prompt explanation mode** — `prompt-improver explain <prompt>` outputs a narrative of what each pipeline stage would do and why, without modifying the prompt.

## Performance

- **Pipeline stage parallelization** — Stages 1-5 (specificity, positive reframe, tone downgrade, overtrigger rewrite, example wrapping) are independent and could run concurrently.
- **Lazy LLM client initialization** — Defer `NewLLMClient()` and API key lookup until the first LLM call, eliminating startup cost for local-only paths.
- **Compiled regex cache** — Pre-compile all regex patterns at package init time instead of per-call in lint and filter functions.
- **Response streaming for improve** — Use Claude SSE streaming so users see incremental output instead of waiting for the full response.
- **Binary size reduction** — Build tags to compile a "lite" binary without MCP support for minimal hook-only installs.
- **Persistent LRU disk cache** — Extend the in-memory `PromptCache` with an optional disk-backed LRU cache in `~/.prompt-improver/cache/` for LLM results surviving restarts.
- **Benchmark regression tracking** — Store `bench` output in `testdata/bench-baseline.txt` and compare in CI, failing if any benchmark regresses by >15%.

## DX (Developer Experience)

- **Shell completions** — `prompt-improver completion bash|zsh|fish` generating completions covering all subcommands, flags, task types, and template names.
- **Interactive TUI mode** — `prompt-improver -i` opens a split-pane TUI (bubbletea) showing original on the left and enhanced on the right with live scoring.
- **Colored diff output** — ANSI color in `prompt-improver diff` (red removals, green additions) when stdout is a terminal, with `--no-color` override.
- **JSON Schema for config** — Ship `.prompt-improver.schema.json` for editor autocomplete and validation of `.prompt-improver.yaml`.
- **Quiet mode for all commands** — Extend `--quiet` flag to `analyze`, `lint`, `diff`, and `cache-check` (currently only `enhance` and `improve`).
- **SARIF lint output** — `prompt-improver lint --format sarif` for editor integration showing lint findings as inline warnings.
- **Self-update command** — `prompt-improver upgrade` checks the latest GitHub release, downloads the matching binary, and replaces itself.
- **Dry-run hook mode** — `prompt-improver hook --dry-run` reads hook JSON but outputs to stderr only, showing what would happen without modifying the session.

## Integrations

- **VS Code extension** — Wrapping `prompt-improver analyze` to show prompt quality scores inline in `.md` files, CLAUDE.md, and LLM prompt strings in code.
- **OpenAI/Gemini backend support** — Abstract `LLMClient` to support OpenAI-compatible APIs and Google Gemini, configured via `llm.provider` in YAML.
- **Pre-commit hook** — Ship `.pre-commit-hooks.yaml` so `prompt-improver lint` runs on CLAUDE.md files in pre-commit, blocking overtrigger language.
- **Cursor IDE integration** — `prompt-improver check-cursor` to lint `.cursor/rules` files the same way `check-claudemd` lints CLAUDE.md.
- **GitHub Action** — Reusable Action running `prompt-improver lint` on CLAUDE.md in PRs, posting lint findings as review comments.
- **Slack/Discord bot** — Webhook-triggered bot that scores prompts pasted into a channel for team prompt engineering reviews.
- **Clipboard integration** — `--copy` flag to automatically copy enhanced output to system clipboard cross-platform.

## Architecture

- **Stage plugin interface** — Define `Stage` interface (`func(text string, taskType TaskType, cfg Config) (string, []string)`) for loading external stages from shared libraries or WASM.
- **gRPC transport for MCP** — Add gRPC transport alongside stdio, enabling remote deployment and connection pooling for team-shared prompt improvement.
- **WASM compilation target** — Compile the local pipeline (no LLM, no MCP) to WebAssembly for browser-based prompt improvement in documentation sites.
- **Structured logging** — Replace `fmt.Fprintf(os.Stderr, ...)` with `slog` structured logging, enabling JSON log output for observability in long-lived MCP server mode.
- **Pipeline as data** — Express the 13-stage pipeline as a declarative config (YAML list of stage names + params) instead of hardcoded function calls, enabling per-project customization.
- **Separate library and CLI packages** — Move CLI dispatch into `cmd/prompt-improver/` and keep `pkg/enhancer/` as a pure importable library for other Go tools.

## Testing

- **Fuzz testing** — Add `Fuzz*` tests for enhancer, lint, and scoring to catch panics, infinite loops, or unexpected behavior on adversarial input.
- **Mutation testing** — Run `go-mutesting` on `pkg/enhancer/` to measure test suite effectiveness beyond line coverage.
- **Idempotency property tests** — Verify that `Enhance(Enhance(prompt))` produces the same result as `Enhance(prompt)` for any input.
- **MCP snapshot tests** — Golden-file tests for each MCP tool handler's JSON output, catching unintended schema changes.
- **E2E Claude Code integration test** — Scripted test that launches Claude Code with the hook configured, sends a prompt, and verifies `additionalContext` in the transcript.

## Analytics

- **Enhancement quality metrics** — Track before/after score deltas per run in `~/.prompt-improver/metrics.jsonl`, with `prompt-improver stats` showing average improvement by stage.
- **Stage effectiveness tracking** — Record which stages actually modified the prompt vs. were no-ops, surfacing data for pruning or tuning thresholds.
- **Opt-in usage telemetry** — Anonymous aggregate telemetry (task type distribution, scores, LLM vs local ratio) gated behind explicit `telemetry: true` in config.
- **Prompt corpus analysis** — `prompt-improver corpus-analyze dir/` scans all markdown/text files, producing aggregate scoring across a project's prompt inventory.
- **LLM cost tracking** — Track API token usage and estimated cost per `improve` call, with `prompt-improver costs` showing spend against a configurable budget.

# prompt-improver — Agent Instructions

> Canonical instructions: AGENTS.md

Use this file as the canonical instruction surface for Codex-first repo guidance. [CLAUDE.md](CLAUDE.md) and [GEMINI.md](GEMINI.md) are compatibility mirrors.

## Project Overview
Archived standalone prompt enhancement tool, now merged into ralphglasses. The enhancer library lives at `internal/enhancer/` and the CLI at `cmd/prompt-improver/` in the ralphglasses repo.

## Tech Stack
- Go 1.26+
- modelcontextprotocol/go-sdk
- gopkg.in/yaml.v3

## Build & Run
```bash
go build ./...
```

## Install (from ralphglasses)
```bash
cd ../ralphglasses && make install-prompt-improver
```

## Test
```bash
go test ./...
```

## Architecture
- `pkg/enhancer/` — Core prompt enhancement library
- `install.go` / `install_test.go` — Installation logic
- Uses MCP go-sdk for protocol integration

## Code Standards
- Go standard formatting (gofmt)
- Tests alongside source files (*_test.go)

## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.

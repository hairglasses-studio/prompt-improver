# prompt-improver — Gemini CLI Instructions

This repo uses [AGENTS.md](AGENTS.md) as the canonical instruction file. Treat this file as compatibility guidance for Gemini-specific workflows.

## Overview
Archived Go prompt enhancer, merged into ralphglasses. Library at `pkg/enhancer/`, install logic at root.

## Build & Test
```bash
go build ./...
go test ./...
```

## Key Details
- SDK: modelcontextprotocol/go-sdk
- Status: archived, merged into ralphglasses repo

## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.

# OpenClaude Sandbox

Ephemeral development environment with [OpenClaude](https://github.com/Gitlawb/openclaude) pre-configured for z.ai GLM-5.1.

## Quick Start

1. Create a codespace from this repo
2. Set your `ZAI_API_KEY` as a codespace secret
3. OpenClaude is ready — run `openclaude` in the terminal

## What's Included

- Node.js 22 + Bun runtime
- OpenClaude (Claude Code open-source alternative)
- z.ai GLM-5.1 provider pre-configured

## Usage

```bash
# Interactive mode
openclaude

# Non-interactive (for task delegation)
openclaude --prompt "your task here" --no-interactive
```

## Managed by Viktor

This sandbox is orchestrated by Viktor AI. Codespaces are created on demand, tasks are delegated, results collected, and the codespace is deleted — all automatically.

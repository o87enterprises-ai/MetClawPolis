# Ollama Cloud Models Analysis & Claude Code Fix

## Date: April 6, 2026

---

## Problem Identified

Claude Code was configured with `ANTHROPIC_AUTH_TOKEN=ollama` and `ANTHROPIC_BASE_URL=http://localhost:11434`, but **Ollama does not implement the Anthropic Messages API** (`/v1/messages`). When Claude Code tried to use any model, it received a 404 error, showing: *"There's an issue with the selected model. It may not exist or you may not have access to it."*

The cached model `deepseek-v3.1:671b-cloud` was also deleted as part of cleanup, compounding the issue.

---

## Broken/Removed Models

| Model | Reason |
|-------|--------|
| `deepseek-v3.1:671b-cloud` | Removed — Claude Code TUI couldn't validate it; also cached as broken |
| `nurullahunlu/kimik2.5cloud:latest` | Removed — Fake cloud model (local 3.2B GGUF, not actual cloud inference) |
| `leckminartor/qwen3.5-uncensored:397b-cloud` | Removed — 3rd party fork, empty metadata, unreliable |
| `glm-5:cloud` | Removed — Questionable authenticity, no proper metadata |

---

## Working Cloud Models (After Cleanup)

| Model | Type | Status |
|-------|------|--------|
| `qwen3.5:cloud` | General (397B) | ✅ **Primary — default for Claude Code** |
| `qwen3-coder-next:cloud` | Coding (80B) | ✅ Refreshed |
| `qwen3-coder:480b-cloud` | Coding (480B) | ✅ Existing (pull timed out during setup) |
| `kimi-k2.5:cloud` | General | ✅ Refreshed |
| `minimax-m2.5:cloud` | General | ✅ Refreshed |
| `qwen3-vl:235b-cloud` | Vision-Language | ✅ Existing |
| `qwen3-vl:235b-instruct-cloud` | Vision-Language | ✅ Existing |
| `nemotron-3-nano:30b-cloud` | General (30B) | ✅ Existing |
| `gemini-3-flash-preview:cloud` | General | ✅ Existing |

---

## Solution: Anthropic-to-Ollama Proxy

A lightweight Python proxy (`~/.claude/anthropic_ollama_proxy.py`) translates Claude Code's Anthropic API calls to Ollama's native API:

- **Claude Code** → `POST /v1/messages` → **Proxy** → `POST /api/chat` → **Ollama**
- Returns Anthropic-formatted responses back to Claude Code
- Runs on `http://localhost:8082`
- Supports all Ollama cloud models

### Configuration (`~/.claude/settings.json`)

```json
{
  "model": "qwen3.5:cloud",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "dummy",
    "ANTHROPIC_API_KEY": "dummy",
    "ANTHROPIC_BASE_URL": "http://localhost:8082",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "qwen3.5:cloud",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "qwen3.5:cloud",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "qwen3.5:cloud",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

---

## Usage

### 1. Start the proxy
```bash
./start_claude_proxy.sh 8082
# or manually:
python3 ~/.claude/anthropic_ollama_proxy.py 8082
```

### 2. Run Claude Code
```bash
claude --dangerously-skip-permissions
# The interactive TUI will now work with cloud models
```

### 3. Switch models (in TUI)
Use `/model` command inside the Claude Code interactive session.

---

## Architecture

```
Claude Code (TUI/CLI)
    │
    │ ANTHROPIC_BASE_URL=http://localhost:8082
    │ ANTHROPIC_API_KEY=dummy
    ▼
┌─────────────────────────────┐
│  Anthropic-to-Ollama Proxy  │  Port 8082
│  (anthropic_ollama_proxy.py)│
│                             │
│  Transforms:                │
│  - /v1/messages → /api/chat │
│  - /v1/models → /v1/models  │
│  - Auth format conversion   │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  Ollama Server               │  Port 11434
│  (localhost)                 │
│                             │
│  Cloud models route to:     │
│  https://ollama.com:443     │
└─────────────────────────────┘
```

---

## Notes

- Cloud models use Ollama's remote inference (free tier available)
- Heavy usage may require an Ollama subscription
- The proxy must be running before starting Claude Code
- Claude Code may report its identity as "Claude Opus/Sonnet" — this is normal; the actual inference goes through Ollama cloud models
- Context window: Ollama cloud models support up to 256K tokens (Qwen3.5)

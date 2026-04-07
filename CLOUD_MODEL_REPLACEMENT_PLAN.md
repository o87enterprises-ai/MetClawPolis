# Ollama Cloud Model Replacement Guide

## Problem Analysis
- `qwen3.5-uncensored:397b-cloud` and `leckminartor/qwen3.5-uncensored:397b-cloud` are **BROKEN**
- Error: `pull model manifest: file does not exist`
- These were cloud-hosted models that are no longer available at the remote endpoint
- All other cloud models (deepseek-v3.1:671b-cloud, kimi-k2.5:cloud, etc.) are **WORKING**

## Working Cloud Models (as of April 2026)
✅ deepseek-v3.1:671b-cloud - WORKING (DeepSeek-V3, 671B params)
✅ kimi-k2.5:cloud - WORKING
✅ glm-5:cloud - WORKING
✅ ministral-3:14b-cloud - WORKING
✅ mistral-large-3:675b-cloud - WORKING
✅ gpt-oss:120b-cloud - WORKING
✅ gemini-3-flash-preview:cloud - WORKING
✅ gpt-oss:20b-cloud - WORKING

## Broken Cloud Models
❌ qwen3.5-uncensored:397b-cloud - BROKEN
❌ leckminartor/qwen3.5-uncensored:397b-cloud - BROKEN (same model, different namespace)

## Replacement Options for qwen3.5-uncensored:397b-cloud

### Option 1: Use other working cloud models (Best for 8GB RAM)
- `deepseek-v3.1:671b-cloud` - Already working, excellent for coding and general tasks
- `kimi-k2.5:cloud` - Good alternative
- `glm-5:cloud` - Another cloud option

### Option 2: Use smaller local uncensored models (if they fit in 8GB)
- `jaahas/qwen3.5-uncensored:2b` - Already on your system (2B params, should fit in 8GB)
- `huihui_ai/qwen3.5-abliterated` - Available on Ollama (various sizes)
- `deepseek-r1:1.5b` - Already on your system (1.1 GB, fits easily)

### Option 3: Pull new cloud alternatives
Search for "qwen3.5 cloud" on ollama.com to find if there are newer cloud-hosted variants

## Recommended Action Plan
1. Remove broken model entries
2. Update any configs/scripts that reference qwen3.5-uncensored:397b-cloud
3. Use deepseek-v3.1:671b-cloud as primary cloud model (already working)
4. Use jaahas/qwen3.5-uncensored:2b for local uncensored needs (if it fits)

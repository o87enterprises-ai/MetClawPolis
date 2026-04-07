# Cloud Model Status & Recommendations
**Date:** April 6, 2026  
**Hardware:** 8GB RAM, CPU only (no GPU)

---

## ✅ FIXED: Broken Models Removed

### Removed (were broken):
- ❌ `qwen3.5-uncensored:397b-cloud` - **DELETED** (manifest no longer exists)
- ❌ `leckminartor/qwen3.5-uncensored:397b-cloud` - **DELETED** (same model, also broken)

---

## ✅ WORKING Cloud Models (No local RAM needed)

These models run entirely in the cloud - perfect for your 8GB setup:

| Model | Size | Use Case | Status |
|-------|------|----------|--------|
| `deepseek-v3.1:671b-cloud` | Cloud | **Best for coding & general tasks** | ✅ Working |
| `kimi-k2.5:cloud` | Cloud | General purpose | ✅ Working |
| `glm-5:cloud` | Cloud | General purpose | ✅ Working |
| `ministral-3:14b-cloud` | Cloud | Mid-size model | ✅ Working |
| `mistral-large-3:675b-cloud` | Cloud | Large model, high quality | ✅ Working |
| `gpt-oss:120b-cloud` | Cloud | OpenAI-style responses | ✅ Working |
| `gemini-3-flash-preview:cloud` | Cloud | Fast responses | ✅ Working |
| `gpt-oss:20b-cloud` | Cloud | Smaller, faster cloud model | ✅ Working |

---

## ✅ WORKING Local Models (Fit in 8GB RAM)

These are already on your system and fit within your RAM:

| Model | Size | RAM Usage | Use Case |
|-------|------|-----------|----------|
| `deepseek-r1:1.5b` | 1.1 GB | ~2GB | **Fast reasoning, fits easily** |
| `deepseek-coder:latest` | 776 MB | ~1.5GB | **Coding, very lightweight** |
| `tinyllama:1.1b` | 637 MB | ~1.5GB | Fast, lightweight tasks |
| `llama3.2:3b` | 2.0 GB | ~3GB | General purpose |
| `phi3:mini` | 2.2 GB | ~3GB | Microsoft's efficient model |
| `gemma3:4b` | 3.3 GB | ~4.5GB | Google's model |
| `qwen2.5-coder-7b-q2k` | 3.0 GB | ~4GB | **Coding (quantized)** |
| `qwen-clean:latest` | 3.0 GB | ~4GB | General Qwen model |
| `qwen:latest` | Cloud | Cloud | **Works as cloud model** |

---

## 🎯 Recommended Replacements for qwen3.5-uncensored:397b-cloud

### For **Uncensored Content** (your original use case):

**Option 1: Use working cloud models** (Recommended)
```bash
# Best all-around cloud model
ollama run deepseek-v3.1:671b-cloud

# Alternative cloud options
ollama run kimi-k2.5:cloud
ollama run glm-5:cloud
```

**Option 2: Pull new uncensored models** (Requires download)
```bash
# These are the available uncensored variants:
ollama pull jaahas/qwen3.5-uncensored:2b    # Small, should fit in 8GB
ollama pull jaahas/qwen3.5-uncensored:4b    # Medium size
ollama pull vaultbox/qwen3.5-uncensored:4b  # Alternative namespace
```

**Option 3: Use existing local models** (Already downloaded)
```bash
# For coding (already on your system)
ollama run deepseek-coder:latest         # 776MB, very light
ollama run qwen2.5-coder-7b-q2k:latest   # 3GB, good coder model

# For general chat (already on your system)
ollama run deepseek-r1:1.5b              # 1.1GB, fast
ollama run qwen:latest                   # Cloud-based, no RAM usage
```

---

## 🔧 Configuration Updates Needed

### For Claude Code:
1. Type `/model` in your Claude Code session
2. Select `deepseek-v3.1:671b-cloud` (recommended)
3. Or any other working cloud model from the list above

### For scripts/configs referencing the broken model:
Replace `qwen3.5-uncensored:397b-cloud` with:
- `deepseek-v3.1:671b-cloud` (best overall)
- `kimi-k2.5:cloud` (good alternative)
- `qwen:latest` (if you need Qwen family)

---

## 📊 Model Comparison

| Model | Type | Strengths | RAM Needed |
|-------|------|-----------|------------|
| deepseek-v3.1:671b-cloud | Cloud | Coding, reasoning, general | 0GB (cloud) |
| kimi-k2.5:cloud | Cloud | General purpose | 0GB (cloud) |
| mistral-large-3:675b-cloud | Cloud | High quality, large | 0GB (cloud) |
| deepseek-coder:latest | Local | Coding | ~1.5GB |
| deepseek-r1:1.5b | Local | Fast reasoning | ~2GB |
| qwen:latest | Cloud | Qwen family | 0GB (cloud) |

---

## 🚀 Quick Commands

```bash
# Test a cloud model
ollama run deepseek-v3.1:671b-cloud "Hello!"

# List all working cloud models
ollama list | grep cloud

# List local models that fit in 8GB
ollama list | grep -v cloud

# Check what's currently running
ollama ps
```

---

## ⚠️ Notes

- Cloud models require internet connection
- Cloud models may have rate limits depending on your ollama.com account
- Local models work offline but use your RAM
- The broken `qwen3.5-uncensored:397b-cloud` was removed from Ollama's cloud infrastructure
- No direct 1:1 replacement exists for that specific model, but the alternatives above serve similar purposes

---

## 📝 Action Items

- [x] Remove broken `qwen3.5-uncensored:397b-cloud` entries ✅
- [ ] Update any scripts/configs that referenced the broken model
- [ ] Test your preferred replacement model
- [ ] Update Claude Code default model to a working cloud model

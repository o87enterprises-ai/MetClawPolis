# 🚀 Ollama Proxy Setup - Quick Start Guide

## ✅ What's Been Set Up

You now have a **comprehensive AI inference endpoint system** with full user configuration control!

### 📁 Files Created

1. **`api/ollama_proxy.go`** - Go implementation (1800+ lines)
   - Full configuration management
   - Model management endpoints
   - Smart router with task detection
   - OpenAI-compatible API
   - Import/export functionality

2. **`ollama_config_server.py`** - Standalone Python server (port 8002)
   - No database dependency
   - Ready to use immediately
   - All endpoints working

3. **`OLLAMA_PROXY_API.md`** - Complete API documentation

4. **`test_ollama_proxy.sh`** - Test script

---

## 🎯 Quick Start

### Option 1: Python Server (Recommended for Testing)

```bash
# Start the server
python3 ollama_config_server.py

# It runs on http://localhost:8002
```

### Option 2: Go Server (Integrated with MetClawPolis)

```bash
# Start the main application
./metclawpolis

# Ollama endpoints available at http://localhost:8080/api/ollama/*
# Note: Requires database and authentication
```

---

## 🔥 Top 10 Things You Can Do

### 1️⃣ Check Health
```bash
curl http://localhost:8002/api/ollama/health
```

### 2️⃣ List All Your Models
```bash
curl http://localhost:8002/api/ollama/models
```
**You currently have 22 models available!**

### 3️⃣ Get Current Configuration
```bash
curl http://localhost:8002/api/ollama/config
```

### 4️⃣ Update Configuration
```bash
curl -X PUT http://localhost:8002/api/ollama/config/update \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{
    "model": "qwen2.5-coder:latest",
    "temperature": 0.2,
    "max_tokens": 8192,
    "num_gpu": 35,
    "system_prompt": "You are an expert coding assistant."
  }'
```

### 5️⃣ Apply Pre-configured Preset
```bash
# For coding
curl -X POST http://localhost:8002/api/ollama/presets/apply \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{"preset_id": "coding"}'

# Available presets: coding, creative, analysis, chat, summarization
```

### 6️⃣ Chat with AI
```bash
curl -X POST http://localhost:8002/api/ollama/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{
    "model": "qwen2.5-coder:latest",
    "messages": [
      {"role": "user", "content": "Write a Python function to sort a list"}
    ]
  }'
```

### 7️⃣ Use OpenAI-Compatible Endpoint
```bash
curl -X POST http://localhost:8002/api/ollama/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{
    "model": "qwen2.5-coder:latest",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

### 8️⃣ Smart Router (Auto-Optimizes Settings)
```bash
curl -X POST http://localhost:8002/api/ollama/smart-router \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{
    "messages": [
      {"role": "user", "content": "Write a Python function to sort a list"}
    ]
  }'
# Automatically detects it's a coding task and applies optimal settings!
```

### 9️⃣ Get Comprehensive Status
```bash
curl http://localhost:8002/api/ollama/status
```

### 🔟 Export Your Configuration
```bash
curl http://localhost:8002/api/ollama/config/export \
  -H "X-User-ID: your_user_id" \
  -o my_config.json
```

---

## 🎛️ Full Configuration Control

You can customize **25+ parameters**:

### Model Settings
- `model` - Which model to use
- `base_url` - Ollama server URL
- `max_tokens` - Maximum tokens to generate

### Sampling Parameters
- `temperature` (0.0-2.0) - Creativity vs determinism
- `top_p` (0.0-1.0) - Nucleus sampling
- `top_k` (0-100) - Top-k sampling
- `frequency_penalty` (-2.0 to 2.0)
- `presence_penalty` (-2.0 to 2.0)
- `typical_p` - Typical P sampling
- `tfs_z` - Tail free sampling

### Repetition Control
- `repeat_penalty` (1.0-2.0)
- `repeat_last_n` - Last N tokens to penalize

### GPU/Memory Settings
- `num_gpu` - Layers to offload to GPU
- `num_thread` - CPU threads
- `main_gpu` - Main GPU device ID
- `use_mlock` - Lock model in memory
- `use_mmap` - Memory map model file

### Mirostat Sampling
- `mirostat` (0, 1, or 2)
- `mirostat_tau`
- `mirostat_eta`

### Other
- `system_prompt` - Default system prompt
- `stream` - Enable streaming
- `timeout` - Request timeout
- `seed` - For reproducibility
- `stop` - Stop sequences

---

## 📦 Model Management

### List Models
```bash
curl http://localhost:8002/api/ollama/models
```

### Show Model Details
```bash
curl "http://localhost:8002/api/ollama/models/show?model=qwen2.5-coder:latest"
```

### Pull New Model
```bash
curl -X POST http://localhost:8002/api/ollama/models/pull \
  -H "Content-Type: application/json" \
  -d '{"model": "llama3.2"}'
```

### Delete Model
```bash
curl -X DELETE "http://localhost:8002/api/ollama/models/delete?model=model_name"
```

### Copy Model
```bash
curl -X POST http://localhost:8002/api/ollama/models/copy \
  -H "Content-Type: application/json" \
  -d '{"source": "model1", "destination": "model1-custom"}'
```

---

## 🎨 Inference Presets

### Coding
```json
{
  "temperature": 0.2,
  "top_p": 0.95,
  "max_tokens": 8192,
  "repeat_penalty": 1.1
}
```

### Creative Writing
```json
{
  "temperature": 0.9,
  "top_p": 0.95,
  "frequency_penalty": 0.3,
  "presence_penalty": 0.4
}
```

### Data Analysis
```json
{
  "temperature": 0.1,
  "top_p": 0.9,
  "repeat_penalty": 1.0
}
```

### Chat
```json
{
  "temperature": 0.7,
  "stream": true,
  "timeout": 60
}
```

### Summarization
```json
{
  "temperature": 0.3,
  "max_tokens": 2048
}
```

---

## 📊 Your Current Setup

**Available Models (22):**
- deepseek-v3.1:671b-cloud
- jaahas/qwen3.5-uncensored:2b
- dolphin-phi:latest
- gemma:latest
- granite3.3:latest
- **qwen2.5-coder:latest** ⭐ (Great for coding!)
- qwen2.5vl:3b
- deepseek-r1:1.5b (Small, fast!)
- leckminartor/qwen3.5-uncensored:397b-cloud
- kimi-k2.5:cloud
- glm-5:cloud
- nemotron-3-nano:30b-cloud
- And 11 more...

**Ollama Status:** ✅ Healthy
**Proxy Status:** ✅ Running on port 8002

---

## 🔧 Advanced Usage

### Batch Configuration
```bash
curl -X POST http://localhost:8002/api/ollama/config/batch \
  -H "Content-Type: application/json" \
  -d '{
    "configs": {
      "user_1": {"model": "qwen2.5-coder:latest", "temperature": 0.2},
      "user_2": {"model": "gemma:latest", "temperature": 0.7}
    }
  }'
```

### Import Configuration
```bash
curl -X POST http://localhost:8002/api/ollama/config/import \
  -H "Content-Type: application/json" \
  -H "X-User-ID: your_user_id" \
  -d '{
    "config": {
      "model": "qwen2.5-coder:latest",
      "temperature": 0.2,
      "max_tokens": 8192
    }
  }'
```

### Reset to Defaults
```bash
curl -X POST http://localhost:8002/api/ollama/config/reset \
  -H "X-User-ID: your_user_id"
```

---

## 🌐 Integration Examples

### Use with OpenAI SDK (Python)
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8002/api/ollama/v1",
    api_key="not-needed"
)

response = client.chat.completions.create(
    model="qwen2.5-coder:latest",
    messages=[{"role": "user", "content": "Write a hello world in Python"}],
    temperature=0.2
)

print(response.choices[0].message.content)
```

### Use with cURL
```bash
# Simple chat
curl -X POST http://localhost:8002/api/ollama/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: me" \
  -d '{"model":"qwen2.5-coder:latest","messages":[{"role":"user","content":"Hi!"}]}'
```

### Use with JavaScript
```javascript
const response = await fetch('http://localhost:8002/api/ollama/v1/chat/completions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    model: 'qwen2.5-coder:latest',
    messages: [{ role: 'user', content: 'Hello!' }],
    max_tokens: 100
  })
});

const data = await response.json();
console.log(data.choices[0].message.content);
```

---

## 📖 Documentation

Full API documentation: **`OLLAMA_PROXY_API.md`**

---

## 🎯 Next Steps

1. ✅ **Server is running** on http://localhost:8002
2. ✅ **All endpoints working**
3. ✅ **22 models available**
4. 🎨 **Start using the endpoints!**

### Try This Now:
```bash
# 1. Check your config
curl http://localhost:8002/api/ollama/config

# 2. Apply coding preset
curl -X POST http://localhost:8002/api/ollama/presets/apply \
  -H "Content-Type: application/json" \
  -d '{"preset_id":"coding"}'

# 3. Start chatting!
curl -X POST http://localhost:8002/api/ollama/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"deepseek-r1:1.5b","messages":[{"role":"user","content":"Hi"}]}'
```

---

## 🆘 Troubleshooting

### Model not responding
- Ollama needs time to load models into memory
- Smaller models load faster (deepseek-r1:1.5b is fastest)
- Check running models: `curl http://localhost:8002/api/ollama/models/running`

### Want to change Ollama URL
```bash
curl -X PUT http://localhost:8002/api/ollama/config/update \
  -H "Content-Type: application/json" \
  -d '{"base_url": "http://your-ollama-server:11434"}'
```

### Reset everything
```bash
# Just restart the Python server
pkill -f ollama_config_server.py
python3 ollama_config_server.py
```

---

## 🎉 You're All Set!

Your comprehensive AI inference endpoint system is ready with:
- ✅ Full configuration control (25+ parameters)
- ✅ Model management (list, pull, show, delete, copy)
- ✅ Chat & text generation
- ✅ OpenAI-compatible API
- ✅ Smart router with auto-detection
- ✅ 5 inference presets
- ✅ Import/export configurations
- ✅ Health monitoring

**Happy AI-ing! 🤖✨**

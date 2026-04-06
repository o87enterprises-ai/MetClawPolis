# Ollama Proxy API Documentation

Comprehensive AI inference endpoints with full user configuration control for local Ollama proxy setup.

## Overview

The Ollama Proxy provides a complete interface for managing local AI inference with:
- **Full Configuration Control**: Customize every aspect of your AI inference
- **Model Management**: List, pull, show, copy, and delete models
- **Multiple API Endpoints**: Native Ollama API + OpenAI-compatible API
- **Smart Routing**: Task-based model selection and optimization
- **Configuration Presets**: Pre-configured settings for common tasks
- **Import/Export**: Save and load configurations
- **Streaming Support**: Real-time token generation
- **Embeddings**: Vector embedding generation

## Base URL

```
http://localhost:8080/api/ollama
```

## Authentication

All endpoints require authentication via headers:
```
X-Verified-Agent-ID: your_agent_id
```

---

## Configuration Endpoints

### 1. Get Current Configuration
**GET** `/api/ollama/config`

Returns the current Ollama configuration for the agent.

**Response:**
```json
{
  "config": {
    "id": "config_agent_123_1234567890",
    "agent_id": "agent_123",
    "base_url": "http://localhost:11434",
    "model": "llama3.2",
    "temperature": 0.7,
    "top_p": 0.9,
    "top_k": 40,
    "max_tokens": 4096,
    "stream": false,
    "timeout": 60,
    "repeat_penalty": 1.1,
    "repeat_last_n": 64,
    "use_mlock": false,
    "use_mmap": true,
    "system_prompt": "",
    "created_at": 1234567890,
    "updated_at": 1234567890
  }
}
```

---

### 2. Update Configuration
**PUT/POST** `/api/ollama/config/update`

Update any configuration parameter. Only send the fields you want to change.

**Request Body:**
```json
{
  "model": "codellama",
  "temperature": 0.2,
  "max_tokens": 8192,
  "system_prompt": "You are an expert coding assistant.",
  "num_gpu": 35,
  "use_mlock": true
}
```

**Configurable Parameters:**
- `base_url` (string): Ollama server URL (default: `http://localhost:11434`)
- `model` (string): Default model name
- `temperature` (float): 0.0 - 2.0 (creativity vs determinism)
- `top_p` (float): 0.0 - 1.0 (nucleus sampling)
- `top_k` (int): 0 - 100 (top-k sampling)
- `max_tokens` (int): Maximum tokens to generate
- `stop` (array): Stop sequences
- `frequency_penalty` (float): -2.0 - 2.0
- `presence_penalty` (float): -2.0 - 2.0
- `seed` (int): Random seed for reproducibility
- `num_thread` (int): CPU threads to use
- `num_gpu` (int): Layers to offload to GPU
- `main_gpu` (int): Main GPU device ID
- `use_mlock` (bool): Lock model in memory
- `use_mmap` (bool): Memory map model file
- `typical_p` (float): Typical P sampling
- `repeat_penalty` (float): 1.0 - 2.0
- `repeat_last_n` (int): Last N tokens to penalize
- `mirostat` (int): Mirostat sampling (0, 1, or 2)
- `mirostat_tau` (float): Mirostat tau
- `mirostat_eta` (float): Mirostat eta
- `tfs_z` (float): Tail free sampling
- `system_prompt` (string): Default system prompt
- `stream` (bool): Enable streaming responses
- `timeout` (int): Request timeout in seconds

**Response:**
```json
{
  "success": true,
  "config": { /* updated config */ }
}
```

---

### 3. Reset Configuration
**POST** `/api/ollama/config/reset`

Reset configuration to default values.

**Response:**
```json
{
  "success": true,
  "message": "Configuration reset to defaults",
  "config": { /* default config */ }
}
```

---

### 4. Export Configuration
**GET** `/api/ollama/config/export`

Export current configuration as JSON file.

**Response:** JSON configuration file with validation warnings.

---

### 5. Import Configuration
**POST** `/api/ollama/config/import`

Import a configuration from JSON.

**Request Body:**
```json
{
  "config": {
    "model": "llama3.2",
    "temperature": 0.7,
    "top_p": 0.9,
    // ... other config fields
  }
}
```

**Response:**
```json
{
  "success": true,
  "config": { /* imported config */ },
  "warnings": [] // validation warnings if any
}
```

---

### 6. Batch Configuration Update
**POST** `/api/ollama/config/batch`

Update multiple agent configurations at once.

**Request Body:**
```json
{
  "configs": {
    "agent_1": { "model": "llama3.2", "temperature": 0.7 },
    "agent_2": { "model": "codellama", "temperature": 0.2 }
  }
}
```

---

## Model Management Endpoints

### 7. List Available Models
**GET** `/api/ollama/models`

List all models available in Ollama.

**Response:**
```json
{
  "models": [
    {
      "name": "llama3.2:latest",
      "model": "llama3.2:latest",
      "modified_at": "2024-01-01T00:00:00Z",
      "size": 3825035024,
      "digest": "abc123...",
      "details": {
        "format": "gguf",
        "family": "llama",
        "parameter_size": "8B",
        "quantization_level": "Q4_K_M"
      }
    }
  ],
  "count": 1
}
```

---

### 8. List Running Models
**GET** `/api/ollama/models/running`

List models currently loaded in memory.

**Response:**
```json
{
  "models": [
    {
      "name": "llama3.2:latest",
      "model": "llama3.2:latest",
      "size": 3825035024,
      "digest": "abc123...",
      "expires_at": "2024-01-01T05:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 9. Show Model Details
**GET** `/api/ollama/models/show?model=llama3.2`

Show detailed information about a specific model.

**Query Parameters:**
- `model` (required): Model name

**Response:**
```json
{
  "license": "",
  "modelfile": "# Modelfile generated by \"ollama show\"\nFROM ...",
  "parameters": "",
  "template": "{{ .System }}\n{{ .Prompt }}",
  "details": { /* model details */ },
  "model_info": { /* detailed model info */ }
}
```

---

### 10. Pull Model
**POST** `/api/ollama/models/pull`

Download a model from Ollama registry.

**Request Body:**
```json
{
  "model": "llama3.2",
  "insecure": false
}
```

**Response:**
```json
{
  "success": true,
  "status": "success"
}
```

---

### 11. Delete Model
**DELETE** `/api/ollama/models/delete?model=model_name`

Delete a model from local storage.

**Query Parameters:**
- `model` (required): Model name to delete

**Response:**
```json
{
  "success": true,
  "message": "deleted 'model_name'"
}
```

---

### 12. Copy Model
**POST** `/api/ollama/models/copy`

Copy a model to a new name.

**Request Body:**
```json
{
  "source": "llama3.2",
  "destination": "my-llama3.2-custom"
}
```

**Response:**
```json
{
  "success": true
}
```

---

## Inference Endpoints

### 13. Chat Completion (Ollama Native)
**POST** `/api/ollama/chat`

Chat with the model using Ollama's native API.

**Request Body:**
```json
{
  "model": "llama3.2",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is Go?"}
  ],
  "stream": false,
  "format": "",
  "options": {
    "temperature": 0.7
  }
}
```

**Response (non-streaming):**
```json
{
  "model": "llama3.2",
  "message": {
    "role": "assistant",
    "content": "Go is a statically typed, compiled programming language..."
  },
  "done": true,
  "total_duration": 1234567890,
  "load_duration": 123456789,
  "prompt_eval_count": 25,
  "eval_count": 150
}
```

**Response (streaming):**
Server-Sent Events (SSE) stream with incremental tokens.

---

### 14. Text Generation
**POST** `/api/ollama/generate`

Generate text completion.

**Request Body:**
```json
{
  "model": "llama3.2",
  "prompt": "Once upon a time",
  "system": "You are a creative writer.",
  "format": "",
  "stream": false,
  "options": {
    "temperature": 0.8
  }
}
```

**Response:**
```json
{
  "model": "llama3.2",
  "response": "in a land far away...",
  "done": true
}
```

---

### 15. OpenAI-Compatible Chat
**POST** `/api/ollama/v1/chat/completions`

OpenAI-compatible chat completions endpoint. Drop-in replacement for OpenAI API.

**Request Body:**
```json
{
  "model": "llama3.2",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Explain quantum computing"}
  ],
  "temperature": 0.7,
  "top_p": 0.9,
  "max_tokens": 2048,
  "stream": false,
  "stop": [],
  "frequency_penalty": 0.0,
  "presence_penalty": 0.0
}
```

**Response (non-streaming):**
```json
{
  "id": "chatcmpl-1234567890",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "llama3.2",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Quantum computing is a type of computing..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 150,
    "total_tokens": 175
  }
}
```

**Response (streaming):**
SSE stream with OpenAI-compatible chunks, ending with `data: [DONE]`.

---

### 16. Generate Embeddings
**POST** `/api/ollama/embeddings`

Generate vector embeddings for text.

**Request Body:**
```json
{
  "model": "llama3.2",
  "prompt": "Text to embed"
}
```

**Response:**
```json
{
  "embedding": [0.1, 0.2, 0.3, ...]
}
```

---

## Smart Router & Presets

### 17. Get Inference Presets
**GET** `/api/ollama/presets`

Get all available inference presets.

**Response:**
```json
{
  "presets": [
    {
      "id": "coding",
      "name": "Code Generation",
      "description": "Optimized for code generation and completion",
      "task_type": "coding",
      "config": {
        "temperature": 0.2,
        "top_p": 0.95,
        "top_k": 40,
        "repeat_penalty": 1.1,
        "repeat_last_n": 64
      }
    },
    // ... more presets
  ],
  "count": 5
}
```

**Available Presets:**
- `coding`: Low temperature, precise code generation
- `creative`: High temperature, imaginative content
- `analysis`: Very low temperature, analytical reasoning
- `chat`: Balanced settings, streaming enabled
- `summarization`: Moderate temperature, concise output

---

### 18. Apply Preset
**POST** `/api/ollama/presets/apply`

Apply a preset to your configuration.

**Request Body:**
```json
{
  "preset_id": "coding"
}
```

**Response:**
```json
{
  "success": true,
  "preset": { /* preset details */ },
  "config": { /* updated config */ }
}
```

---

### 19. Smart Router
**POST** `/api/ollama/smart-router`

Automatically detect task type and apply optimal configuration.

**Request Body:**
```json
{
  "model": "llama3.2",
  "messages": [
    {"role": "user", "content": "Write a Python function to sort a list"}
  ]
}
```

The smart router will:
1. Detect the task type (coding, creative, analysis, chat, summarization)
2. Apply the appropriate preset configuration
3. Route the request with optimal parameters

**Response:** Same as chat endpoint, but with auto-optimized settings.

---

## Health & Status

### 20. Health Check
**GET** `/api/ollama/health`

Check if Ollama server is reachable and healthy.

**Response (healthy):**
```json
{
  "status": "healthy",
  "message": "Ollama server is running",
  "base_url": "http://localhost:11434"
}
```

**Response (unhealthy):**
```json
{
  "status": "unhealthy",
  "message": "Cannot reach Ollama server at http://localhost:11434",
  "error": "connection refused"
}
```

---

### 21. Comprehensive Status
**GET** `/api/ollama/status`

Get comprehensive proxy status including Ollama health, loaded models, and current config.

**Response:**
```json
{
  "status": {
    "proxy": {
      "healthy": true,
      "version": "1.0.0",
      "features": [
        "model_management",
        "chat_completion",
        "text_generation",
        "embeddings",
        "openai_compatible_api",
        "streaming",
        "config_presets",
        "config_import_export"
      ]
    },
    "ollama": {
      "healthy": true,
      "base_url": "http://localhost:11434",
      "loaded_models": 2
    },
    "config": {
      "model": "llama3.2",
      "temperature": 0.7,
      "max_tokens": 4096,
      "stream": false
    }
  },
  "timestamp": 1234567890
}
```

---

## Environment Variables

- `OLLAMA_BASE_URL`: Override default Ollama server URL (default: `http://localhost:11434`)

---

## Usage Examples

### Example 1: Basic Chat Request
```bash
curl -X POST http://localhost:8080/api/ollama/chat \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### Example 2: Update Configuration for Coding
```bash
curl -X PUT http://localhost:8080/api/ollama/config/update \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "model": "codellama",
    "temperature": 0.1,
    "max_tokens": 8192,
    "system_prompt": "You are an expert software engineer.",
    "num_gpu": 35,
    "use_mlock": true
  }'
```

### Example 3: Apply Coding Preset
```bash
curl -X POST http://localhost:8080/api/ollama/presets/apply \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "preset_id": "coding"
  }'
```

### Example 4: OpenAI-Compatible Request
```bash
curl -X POST http://localhost:8080/api/ollama/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Explain recursion in Python"}
    ],
    "temperature": 0.7,
    "max_tokens": 1024
  }'
```

### Example 5: Pull a New Model
```bash
curl -X POST http://localhost:8080/api/ollama/models/pull \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "model": "mistral"
  }'
```

### Example 6: Export Configuration
```bash
curl -X GET http://localhost:8080/api/ollama/config/export \
  -H "X-Verified-Agent-ID: agent_123" \
  -o my-ollama-config.json
```

### Example 7: Smart Router (Auto-Detect Task)
```bash
curl -X POST http://localhost:8080/api/ollama/smart-router \
  -H "Content-Type: application/json" \
  -H "X-Verified-Agent-ID: agent_123" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Write a function to calculate fibonacci sequence in Python"}
    ]
  }'
```

---

## Configuration Best Practices

### Code Generation
```json
{
  "model": "codellama",
  "temperature": 0.1,
  "top_p": 0.95,
  "top_k": 40,
  "repeat_penalty": 1.1,
  "max_tokens": 8192,
  "num_gpu": 35,
  "use_mlock": true
}
```

### Creative Writing
```json
{
  "model": "llama3.2",
  "temperature": 0.9,
  "top_p": 0.95,
  "top_k": 50,
  "frequency_penalty": 0.3,
  "presence_penalty": 0.4,
  "repeat_penalty": 1.2
}
```

### Data Analysis
```json
{
  "model": "llama3.2",
  "temperature": 0.1,
  "top_p": 0.9,
  "top_k": 30,
  "repeat_penalty": 1.0,
  "repeat_last_n": 32
}
```

### Fast Chat
```json
{
  "model": "llama3.2",
  "temperature": 0.7,
  "stream": true,
  "timeout": 30,
  "num_gpu": 35
}
```

---

## Architecture

```
User Request
    ↓
[Authentication Middleware]
    ↓
[Configuration Manager] → [Database Cache]
    ↓
[Smart Router] (optional) → Detects task type
    ↓
[Parameter Builder] → Applies config + preset options
    ↓
[Ollama Proxy] → Forwards to Ollama server
    ↓
[Response Handler] → Streaming or complete response
    ↓
User Response
```

---

## Features Summary

✅ **Full Configuration Control**: 25+ customizable parameters  
✅ **Model Management**: List, pull, show, copy, delete models  
✅ **Dual API Support**: Native Ollama + OpenAI-compatible  
✅ **Smart Routing**: Auto-detect task and optimize settings  
✅ **Configuration Presets**: 5 pre-built task-optimized configs  
✅ **Import/Export**: Save and share configurations  
✅ **Streaming Support**: Real-time token generation  
✅ **Embeddings**: Vector embedding generation  
✅ **Health Monitoring**: Comprehensive status endpoints  
✅ **Batch Operations**: Update multiple configs at once  
✅ **Validation**: Automatic config validation with warnings  
✅ **Database Persistence**: Configurations saved to PostgreSQL  
✅ **Caching**: In-memory config cache for performance  

---

## Troubleshooting

### Ollama Server Not Reachable
- Ensure Ollama is running: `ollama serve`
- Check base URL: `echo $OLLAMA_BASE_URL`
- Test health endpoint: `curl http://localhost:8080/api/ollama/health`

### Model Not Found
- List available models: `GET /api/ollama/models`
- Pull missing model: `POST /api/ollama/models/pull` with `{"model": "model_name"}`

### Slow Responses
- Increase `num_gpu` to offload more layers to GPU
- Set `use_mlock: true` to keep model in RAM
- Use smaller quantized models (Q4 instead of Q8)

### Out of Memory
- Reduce `num_gpu` layers
- Close other applications
- Use smaller models (7B instead of 70B)
- Set `use_mmap: true` to use virtual memory

---

For more information, see the [Ollama API documentation](https://github.com/ollama/ollama/blob/main/docs/api.md).

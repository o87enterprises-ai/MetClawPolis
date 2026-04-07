#!/usr/bin/env python3
"""
Ollama Proxy Configuration Server
Provides comprehensive AI inference endpoints with full user config control
Runs independently without database dependency
"""

import http.server
import json
import time
import urllib.request
import urllib.error
from urllib.parse import urlparse, parse_qs

# Configuration
OLLAMA_BASE_URL = "http://localhost:11434"
PORT = 8002

# Inference Presets
PRESETS = {
    "coding": {
        "id": "coding",
        "name": "Code Generation",
        "description": "Optimized for code generation and completion",
        "task_type": "coding",
        "config": {
            "temperature": 0.2,
            "top_p": 0.95,
            "top_k": 40,
            "repeat_penalty": 1.1,
            "repeat_last_n": 64,
            "frequency_penalty": 0.0,
            "presence_penalty": 0.0,
            "stream": False,
            "timeout": 120
        }
    },
    "creative": {
        "id": "creative",
        "name": "Creative Writing",
        "description": "Optimized for creative and imaginative content",
        "task_type": "creative",
        "config": {
            "temperature": 0.9,
            "top_p": 0.95,
            "top_k": 50,
            "repeat_penalty": 1.2,
            "repeat_last_n": 128,
            "frequency_penalty": 0.3,
            "presence_penalty": 0.4,
            "stream": False,
            "timeout": 60
        }
    },
    "analysis": {
        "id": "analysis",
        "name": "Data Analysis",
        "description": "Optimized for analytical reasoning",
        "task_type": "analysis",
        "config": {
            "temperature": 0.1,
            "top_p": 0.9,
            "top_k": 30,
            "repeat_penalty": 1.0,
            "repeat_last_n": 32,
            "frequency_penalty": 0.0,
            "presence_penalty": 0.0,
            "stream": False,
            "timeout": 90
        }
    },
    "chat": {
        "id": "chat",
        "name": "Conversational Chat",
        "description": "Balanced settings for general conversation",
        "task_type": "chat",
        "config": {
            "temperature": 0.7,
            "top_p": 0.9,
            "top_k": 40,
            "repeat_penalty": 1.1,
            "repeat_last_n": 64,
            "frequency_penalty": 0.0,
            "presence_penalty": 0.0,
            "stream": True,
            "timeout": 60
        }
    },
    "summarization": {
        "id": "summarization",
        "name": "Text Summarization",
        "description": "Optimized for concise summarization",
        "task_type": "summarization",
        "config": {
            "temperature": 0.3,
            "top_p": 0.9,
            "top_k": 40,
            "repeat_penalty": 1.1,
            "repeat_last_n": 64,
            "frequency_penalty": 0.1,
            "presence_penalty": 0.0,
            "stream": False,
            "timeout": 60
        }
    }
}

# Default config
DEFAULT_CONFIG = {
    "base_url": OLLAMA_BASE_URL,
    "model": "llama3.2",
    "temperature": 0.7,
    "top_p": 0.9,
    "top_k": 40,
    "max_tokens": 4096,
    "stop": [],
    "frequency_penalty": 0.0,
    "presence_penalty": 0.0,
    "seed": 0,
    "num_thread": 0,
    "num_gpu": 0,
    "main_gpu": 0,
    "use_mlock": False,
    "use_mmap": True,
    "typical_p": 0.0,
    "repeat_penalty": 1.1,
    "repeat_last_n": 64,
    "mirostat": 0,
    "mirostat_tau": 0.0,
    "mirostat_eta": 0.0,
    "tfs_z": 0.0,
    "system_prompt": "",
    "stream": False,
    "timeout": 60
}

# User configs (in production, use database)
user_configs = {}

def get_config(user_id):
    if user_id not in user_configs:
        user_configs[user_id] = DEFAULT_CONFIG.copy()
    return user_configs[user_id]

def update_config(user_id, updates):
    config = get_config(user_id)
    config.update(updates)
    user_configs[user_id] = config
    return config

def detect_task_type(messages):
    """Detect task type from messages for smart routing"""
    if not messages:
        return "chat"
    
    last_user_msg = ""
    for msg in reversed(messages):
        if msg.get("role") == "user":
            last_user_msg = msg.get("content", "").lower()
            break
    
    if not last_user_msg:
        return "chat"
    
    # Keyword detection
    coding_kw = ["code", "function", "python", "javascript", "typescript", "program", "api", "endpoint"]
    creative_kw = ["write", "story", "poem", "creative", "imagine"]
    analysis_kw = ["analyze", "analysis", "compare", "evaluate", "data", "statistics"]
    summary_kw = ["summarize", "summary", "tl;dr", "brief"]
    
    for kw in coding_kw:
        if kw in last_user_msg:
            return "coding"
    for kw in creative_kw:
        if kw in last_user_msg:
            return "creative"
    for kw in analysis_kw:
        if kw in last_user_msg:
            return "analysis"
    for kw in summary_kw:
        if kw in last_user_msg:
            return "summarization"
    
    return "chat"

def proxy_to_ollama(path, data, stream=False):
    """Proxy request to Ollama"""
    url = f"{OLLAMA_BASE_URL}{path}"
    req = urllib.request.Request(
        url,
        data=json.dumps(data).encode('utf-8'),
        headers={'Content-Type': 'application/json'},
        method='POST'
    )
    
    try:
        with urllib.request.urlopen(req, timeout=120) as response:
            return response.read().decode('utf-8'), response.status
    except urllib.error.URLError as e:
        return json.dumps({"error": str(e)}), 502

class OllamaProxyHandler(http.server.BaseHTTPRequestHandler):
    
    def send_json(self, data, status=200):
        self.send_response(status)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode('utf-8'))
    
    def read_body(self):
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)
        return json.loads(body) if body else {}
    
    def get_user_id(self):
        return self.headers.get('X-User-ID', 'default')
    
    def do_GET(self):
        parsed = urlparse(self.path)
        path = parsed.path
        params = parse_qs(parsed.query)
        
        if path == '/api/ollama/health':
            self.handle_health()
        elif path == '/api/ollama/config':
            self.handle_get_config()
        elif path == '/api/ollama/models':
            self.handle_list_models()
        elif path == '/api/ollama/models/running':
            self.handle_running_models()
        elif path == '/api/ollama/models/show':
            self.handle_show_model(params.get('model', [None])[0])
        elif path == '/api/ollama/presets':
            self.handle_get_presets()
        elif path == '/api/ollama/status':
            self.handle_status()
        elif path == '/api/ollama/config/export':
            self.handle_export_config()
        else:
            self.send_json({"error": "not found"}, 404)
    
    def do_POST(self):
        parsed = urlparse(self.path)
        path = parsed.path
        
        if path == '/api/ollama/chat':
            self.handle_chat()
        elif path == '/api/ollama/generate':
            self.handle_generate()
        elif path == '/api/ollama/embeddings':
            self.handle_embeddings()
        elif path == '/api/ollama/v1/chat/completions':
            self.handle_openai_chat()
        elif path == '/api/ollama/presets/apply':
            self.handle_apply_preset()
        elif path == '/api/ollama/smart-router':
            self.handle_smart_router()
        elif path == '/api/ollama/models/pull':
            self.handle_pull_model()
        elif path == '/api/ollama/models/copy':
            self.handle_copy_model()
        elif path == '/api/ollama/config/import':
            self.handle_import_config()
        else:
            self.send_json({"error": "not found"}, 404)
    
    def do_PUT(self):
        parsed = urlparse(self.path)
        path = parsed.path
        
        if path == '/api/ollama/config/update':
            self.handle_update_config()
        else:
            self.send_json({"error": "not found"}, 404)
    
    def do_DELETE(self):
        parsed = urlparse(self.path)
        path = parsed.path
        
        if path == '/api/ollama/models/delete':
            params = parse_qs(parsed.query)
            self.handle_delete_model(params.get('model', [None])[0])
        else:
            self.send_json({"error": "not found"}, 404)
    
    # Handler methods
    def handle_health(self):
        try:
            req = urllib.request.Request(OLLAMA_BASE_URL)
            with urllib.request.urlopen(req, timeout=5) as response:
                if response.status == 200:
                    self.send_json({
                        "status": "healthy",
                        "message": "Ollama server is running",
                        "base_url": OLLAMA_BASE_URL
                    })
        except:
            self.send_json({
                "status": "unhealthy",
                "message": f"Cannot reach Ollama server at {OLLAMA_BASE_URL}"
            }, 502)
    
    def handle_get_config(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        self.send_json({"config": config})
    
    def handle_update_config(self):
        user_id = self.get_user_id()
        updates = self.read_body()
        config = update_config(user_id, updates)
        self.send_json({"success": True, "config": config})
    
    def handle_list_models(self):
        try:
            with urllib.request.urlopen(f"{OLLAMA_BASE_URL}/api/tags", timeout=10) as response:
                data = json.loads(response.read().decode('utf-8'))
                self.send_json({
                    "models": data.get("models", []),
                    "count": len(data.get("models", []))
                })
        except Exception as e:
            self.send_json({"error": str(e)}, 502)
    
    def handle_running_models(self):
        try:
            with urllib.request.urlopen(f"{OLLAMA_BASE_URL}/api/ps", timeout=10) as response:
                data = json.loads(response.read().decode('utf-8'))
                self.send_json({
                    "models": data.get("models", []),
                    "count": len(data.get("models", []))
                })
        except Exception as e:
            self.send_json({"error": str(e)}, 502)
    
    def handle_show_model(self, model):
        if not model:
            self.send_json({"error": "model parameter required"}, 400)
            return
        
        data = {"name": model}
        resp, status = proxy_to_ollama("/api/show", data)
        self.send_json(json.loads(resp), status)
    
    def handle_pull_model(self):
        body = self.read_body()
        data = {"name": body.get("model"), "stream": False}
        resp, status = proxy_to_ollama("/api/pull", data)
        self.send_json({"success": status == 200, "status": resp})
    
    def handle_delete_model(self, model):
        if not model:
            self.send_json({"error": "model parameter required"}, 400)
            return
        
        data = {"name": model}
        resp, status = proxy_to_ollama("/api/delete", data)
        self.send_json({"success": status == 200, "message": resp})
    
    def handle_copy_model(self):
        body = self.read_body()
        data = {
            "source": body.get("source"),
            "destination": body.get("destination")
        }
        resp, status = proxy_to_ollama("/api/copy", data)
        self.send_json({"success": status == 200})
    
    def handle_chat(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        body = self.read_body()
        
        # Apply config
        if not body.get("model"):
            body["model"] = config["model"]
        
        # Build options
        body["options"] = body.get("options", {})
        body["options"].update({
            "temperature": config["temperature"],
            "top_p": config["top_p"],
            "top_k": config["top_k"],
            "repeat_penalty": config["repeat_penalty"],
            "repeat_last_n": config["repeat_last_n"]
        })
        
        # Add system prompt
        if config["system_prompt"] and body.get("messages"):
            if body["messages"][0].get("role") != "system":
                body["messages"].insert(0, {"role": "system", "content": config["system_prompt"]})
        
        body["stream"] = config["stream"]
        
        resp, status = proxy_to_ollama("/api/chat", body)
        self.send_json(json.loads(resp), status)
    
    def handle_generate(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        body = self.read_body()
        
        if not body.get("model"):
            body["model"] = config["model"]
        
        body["options"] = body.get("options", {})
        body["options"].update({
            "temperature": config["temperature"],
            "top_p": config["top_p"],
            "top_k": config["top_k"]
        })
        
        body["stream"] = False
        
        resp, status = proxy_to_ollama("/api/generate", body)
        self.send_json(json.loads(resp), status)
    
    def handle_embeddings(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        body = self.read_body()
        
        if not body.get("model"):
            body["model"] = config["model"]
        
        resp, status = proxy_to_ollama("/api/embeddings", body)
        self.send_json(json.loads(resp), status)
    
    def handle_openai_chat(self):
        """OpenAI-compatible endpoint"""
        user_id = self.get_user_id()
        config = get_config(user_id)
        body = self.read_body()
        
        # Convert OpenAI format to Ollama format
        ollama_req = {
            "model": body.get("model", config["model"]),
            "messages": body.get("messages", []),
            "stream": body.get("stream", False),
            "options": {
                "temperature": body.get("temperature", config["temperature"]),
                "top_p": body.get("top_p", config["top_p"]),
            }
        }
        
        if body.get("max_tokens"):
            ollama_req["options"]["num_predict"] = body["max_tokens"]
        
        # Add system prompt
        if config["system_prompt"] and ollama_req["messages"]:
            if ollama_req["messages"][0].get("role") != "system":
                ollama_req["messages"].insert(0, {"role": "system", "content": config["system_prompt"]})
        
        resp_str, status = proxy_to_ollama("/api/chat", ollama_req)
        
        if status == 200:
            ollama_resp = json.loads(resp_str)
            # Convert to OpenAI format
            openai_resp = {
                "id": f"chatcmpl-{int(time.time())}",
                "object": "chat.completion",
                "created": int(time.time()),
                "model": ollama_resp.get("model", config["model"]),
                "choices": [{
                    "index": 0,
                    "message": ollama_resp.get("message", {}),
                    "finish_reason": "stop"
                }],
                "usage": {
                    "prompt_tokens": ollama_resp.get("prompt_eval_count", 0),
                    "completion_tokens": ollama_resp.get("eval_count", 0),
                    "total_tokens": ollama_resp.get("prompt_eval_count", 0) + ollama_resp.get("eval_count", 0)
                }
            }
            self.send_json(openai_resp)
        else:
            self.send_json(json.loads(resp_str), status)
    
    def handle_get_presets(self):
        presets = list(PRESETS.values())
        self.send_json({"presets": presets, "count": len(presets)})
    
    def handle_apply_preset(self):
        user_id = self.get_user_id()
        body = self.read_body()
        preset_id = body.get("preset_id")
        
        if preset_id not in PRESETS:
            self.send_json({"error": "preset not found"}, 400)
            return
        
        preset = PRESETS[preset_id]
        config = update_config(user_id, preset["config"])
        self.send_json({"success": True, "preset": preset, "config": config})
    
    def handle_smart_router(self):
        user_id = self.get_user_id()
        body = self.read_body()
        
        # Detect task type
        task_type = detect_task_type(body.get("messages", []))
        
        # Apply preset
        if task_type in PRESETS:
            update_config(user_id, PRESETS[task_type]["config"])
        
        # Forward to chat
        config = get_config(user_id)
        
        if not body.get("model"):
            body["model"] = config["model"]
        
        body["options"] = body.get("options", {})
        body["options"].update({
            "temperature": config["temperature"],
            "top_p": config["top_p"],
            "top_k": config["top_k"],
            "repeat_penalty": config["repeat_penalty"]
        })
        
        body["stream"] = False
        
        resp, status = proxy_to_ollama("/api/chat", body)
        result = json.loads(resp)
        result["detected_task"] = task_type
        result["applied_preset"] = task_type if task_type in PRESETS else "default"
        
        self.send_json(result, status)
    
    def handle_status(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        
        # Check Ollama health
        healthy = False
        try:
            req = urllib.request.Request(OLLAMA_BASE_URL)
            with urllib.request.urlopen(req, timeout=5) as response:
                healthy = response.status == 200
        except:
            pass
        
        self.send_json({
            "status": {
                "proxy": {
                    "healthy": True,
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
                    "healthy": healthy,
                    "base_url": OLLAMA_BASE_URL
                },
                "config": {
                    "model": config["model"],
                    "temperature": config["temperature"],
                    "max_tokens": config["max_tokens"],
                    "stream": config["stream"]
                }
            },
            "timestamp": int(time.time())
        })
    
    def handle_export_config(self):
        user_id = self.get_user_id()
        config = get_config(user_id)
        self.send_json({
            "config": config,
            "exported_at": int(time.time())
        })
    
    def handle_import_config(self):
        user_id = self.get_user_id()
        body = self.read_body()
        config_data = body.get("config", {})
        config = update_config(user_id, config_data)
        self.send_json({"success": True, "config": config})
    
    def log_message(self, format, *args):
        """Custom log format"""
        print(f"[{time.strftime('%H:%M:%S')}] {format % args}")

if __name__ == "__main__":
    server = http.server.HTTPServer(('localhost', PORT), OllamaProxyHandler)
    print(f"🚀 Ollama Proxy Configuration Server")
    print(f"📡 Running on http://localhost:{PORT}")
    print(f"🔗 Ollama backend: {OLLAMA_BASE_URL}")
    print(f"\n✨ Features:")
    print(f"  • Full configuration control (25+ parameters)")
    print(f"  • Model management (list, pull, show, delete, copy)")
    print(f"  • Chat & text generation")
    print(f"  • OpenAI-compatible API (/v1/chat/completions)")
    print(f"  • Smart router with auto-detection")
    print(f"  • 5 inference presets (coding, creative, analysis, chat, summary)")
    print(f"  • Import/export configurations")
    print(f"  • Health monitoring")
    print(f"\n📖 Documentation: OLLAMA_PROXY_API.md")
    print(f"\n⏹️  Press Ctrl+C to stop\n")
    
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n\n👋 Shutting down...")
        server.shutdown()

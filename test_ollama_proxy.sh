#!/bin/bash

# Ollama Proxy Test Script
# Tests all the new Ollama proxy endpoints

BASE_URL="http://localhost:8080"
AGENT_ID="test_agent_123"
SIGNATURE="test_signature"

echo "🚀 Testing Ollama Proxy Endpoints"
echo "=================================="
echo ""

# Test 1: Health Check
echo "1️⃣  Health Check"
curl -s "$BASE_URL/api/ollama/health" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq .
echo ""

# Test 2: Get Current Config
echo "2️⃣  Get Current Configuration"
curl -s "$BASE_URL/api/ollama/config" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq .config | head -20
echo ""

# Test 3: List Models
echo "3️⃣  List Available Models"
curl -s "$BASE_URL/api/ollama/models" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq .
echo ""

# Test 4: Get Inference Presets
echo "4️⃣  Get Inference Presets"
curl -s "$BASE_URL/api/ollama/presets" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq '.presets[] | {id, name, task_type}'
echo ""

# Test 5: Update Config (apply coding settings)
echo "5️⃣  Update Configuration (Coding Preset)"
curl -s -X PUT "$BASE_URL/api/ollama/config/update" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "temperature": 0.2,
    "max_tokens": 8192,
    "system_prompt": "You are an expert coding assistant.",
    "model": "llama3.2"
  }' | jq .
echo ""

# Test 6: Get Proxy Status
echo "6️⃣  Get Comprehensive Status"
curl -s "$BASE_URL/api/ollama/status" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq .
echo ""

# Test 7: Chat Request
echo "7️⃣  Chat Request"
curl -s -X POST "$BASE_URL/api/ollama/chat" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Hello! What is 2+2?"}
    ]
  }' | jq .
echo ""

# Test 8: OpenAI-Compatible Endpoint
echo "8️⃣  OpenAI-Compatible Chat"
curl -s -X POST "$BASE_URL/api/ollama/v1/chat/completions" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Say hello in 3 words"}
    ],
    "max_tokens": 50
  }' | jq .
echo ""

# Test 9: Export Config
echo "9️⃣  Export Configuration"
curl -s "$BASE_URL/api/ollama/config/export" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" | jq .config | head -15
echo ""

# Test 10: Smart Router
echo "🔟 Smart Router (Auto-detect coding task)"
curl -s -X POST "$BASE_URL/api/ollama/smart-router" \
  -H "X-Agent-ID: $AGENT_ID" \
  -H "X-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {"role": "user", "content": "Write a Python function to sort a list"}
    ]
  }' | jq .
echo ""

echo "✅ All tests completed!"

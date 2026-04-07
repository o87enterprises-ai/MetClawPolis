#!/bin/bash
# Fix broken cloud models script
# Removes broken qwen3.5-uncensored:397b-cloud entries and updates configs

echo "🔧 Fixing broken cloud model entries..."
echo ""

# Check which models are broken
echo "📋 Testing cloud models..."
BROKEN_MODELS=()
WORKING_MODELS=()

for model in "qwen3.5-uncensored:397b-cloud" "leckminartor/qwen3.5-uncensored:397b-cloud"; do
    echo -n "Testing $model... "
    if ollama run "$model" "ping" 2>&1 | head -1 | grep -q "Error"; then
        echo "❌ BROKEN"
        BROKEN_MODELS+=("$model")
    else
        echo "✅ WORKS"
        WORKING_MODELS+=("$model")
    fi
done

echo ""
echo "🗑️ Removing broken model entries..."

# Remove broken models
for model in "${BROKEN_MODELS[@]}"; do
    echo "Removing $model..."
    ollama rm "$model" 2>/dev/null || echo "  (already removed or not found)"
done

echo ""
echo "✅ Working cloud models available:"
ollama list | grep "cloud" | grep -v "qwen3.5-uncensored:397b"

echo ""
echo "📝 Recommended replacements for qwen3.5-uncensored:397b-cloud:"
echo "  1. deepseek-v3.1:671b-cloud (already working, excellent for coding)"
echo "  2. kimi-k2.5:cloud (good general-purpose cloud model)"
echo "  3. jaahas/qwen3.5-uncensored:2b (local, may fit in 8GB RAM)"

echo ""
echo "🔄 To update your Claude Code default model:"
echo "  Run: /model"
echo "  Then select: deepseek-v3.1:671b-cloud"

echo ""
echo "✅ Cleanup complete!"

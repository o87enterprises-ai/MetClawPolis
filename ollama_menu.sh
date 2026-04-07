#!/bin/bash
# Ollama Proxy Interactive Menu
# Easy access to all features

BASE_URL="http://localhost:8002"
USER_ID="${OLLAMA_USER_ID:-default_user}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

clear
echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║${NC}    ${YELLOW}🚀 Ollama Proxy Configuration Manager${NC}           ${GREEN}║${NC}"
echo -e "${GREEN}║${NC}    ${BLUE}Full Control Over Your AI Inference${NC}               ${GREEN}║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "User: ${YELLOW}$USER_ID${NC}"
echo -e "Server: ${YELLOW}$BASE_URL${NC}"
echo ""

while true; do
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Main Menu:${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo "1)  📊 View System Status"
    echo "2)  ⚙️  View Current Configuration"
    echo "3)  🎨 Apply Inference Preset"
    echo "4)  ✏️  Update Configuration"
    echo "5)  💬 Chat with AI"
    echo "6)  🤖 List Available Models"
    echo "7)  📥 Pull New Model"
    echo "8)  🗑️  Delete Model"
    echo "9)  📤 Export Configuration"
    echo "10) 📥 Import Configuration"
    echo "11) 🧠 Smart Router (Auto-detect task)"
    echo "12) 🏥 Health Check"
    echo "13) 📖 View Documentation"
    echo "0)  ❌ Exit"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -n "Select option [0-13]: "
    read choice
    
    case $choice in
        1)
            echo -e "\n${YELLOW}📊 System Status:${NC}"
            curl -s "$BASE_URL/api/ollama/status" | python3 -m json.tool
            ;;
        2)
            echo -e "\n${YELLOW}⚙️  Current Configuration:${NC}"
            curl -s "$BASE_URL/api/ollama/config" -H "X-User-ID: $USER_ID" | python3 -m json.tool
            ;;
        3)
            echo -e "\n${YELLOW}🎨 Available Presets:${NC}"
            curl -s "$BASE_URL/api/ollama/presets" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for p in data['presets']:
    print(f\"  {p['id']}: {p['name']} - {p['description']}\")
"
            echo -n "Enter preset ID: "
            read preset
            curl -s -X POST "$BASE_URL/api/ollama/presets/apply" \
                -H "Content-Type: application/json" \
                -H "X-User-ID: $USER_ID" \
                -d "{\"preset_id\":\"$preset\"}" | python3 -m json.tool
            ;;
        4)
            echo -e "\n${YELLOW}✏️  Update Configuration:${NC}"
            echo "Enter JSON updates (e.g., {\"temperature\":0.5,\"max_tokens\":4096}):"
            read -p "> " updates
            curl -s -X PUT "$BASE_URL/api/ollama/config/update" \
                -H "Content-Type: application/json" \
                -H "X-User-ID: $USER_ID" \
                -d "$updates" | python3 -m json.tool
            ;;
        5)
            echo -e "\n${YELLOW}💬 Chat with AI:${NC}"
            echo -n "Model: "
            read model
            echo -n "Message: "
            read message
            echo -e "\n${GREEN}🤖 AI:${NC}"
            curl -s -X POST "$BASE_URL/api/ollama/chat" \
                -H "Content-Type: application/json" \
                -H "X-User-ID: $USER_ID" \
                -d "{\"model\":\"$model\",\"messages\":[{\"role\":\"user\",\"content\":\"$message\"}]}" | \
                python3 -c "import sys,json; print(json.load(sys.stdin).get('message',{}).get('content','Processing...'))"
            ;;
        6)
            echo -e "\n${YELLOW}🤖 Available Models:${NC}"
            curl -s "$BASE_URL/api/ollama/models" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(f\"Found {data['count']} models:\")
for m in data.get('models', []):
    size_gb = m.get('size', 0) / 1073741824
    print(f\"  ✓ {m['name']} ({size_gb:.1f} GB)\")
"
            ;;
        7)
            echo -e "\n${YELLOW}📥 Pull New Model:${NC}"
            echo -n "Model name (e.g., llama3.2): "
            read model
            echo "Pulling $model... (this may take a while)"
            curl -s -X POST "$BASE_URL/api/ollama/models/pull" \
                -H "Content-Type: application/json" \
                -d "{\"model\":\"$model\"}" | python3 -m json.tool
            ;;
        8)
            echo -e "\n${YELLOW}🗑️  Delete Model:${NC}"
            curl -s "$BASE_URL/api/ollama/models" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for i, m in enumerate(data.get('models', []), 1):
    print(f\"  {i}) {m['name']}\")
"
            echo -n "Enter model name to delete: "
            read model
            echo -n "Are you sure? (y/N): "
            read confirm
            if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
                curl -s -X DELETE "$BASE_URL/api/ollama/models/delete?model=$model" | python3 -m json.tool
            else
                echo "Cancelled"
            fi
            ;;
        9)
            echo -e "\n${YELLOW}📤 Export Configuration:${NC}"
            echo -n "Filename [ollama-config.json]: "
            read filename
            filename=${filename:-ollama-config.json}
            curl -s "$BASE_URL/api/ollama/config/export" -H "X-User-ID: $USER_ID" > "$filename"
            echo -e "${GREEN}✓ Saved to $filename${NC}"
            ;;
        10)
            echo -e "\n${YELLOW}📥 Import Configuration:${NC}"
            echo -n "Config file path: "
            read filepath
            if [ -f "$filepath" ]; then
                config=$(cat "$filepath")
                curl -s -X POST "$BASE_URL/api/ollama/config/import" \
                    -H "Content-Type: application/json" \
                    -H "X-User-ID: $USER_ID" \
                    -d "{\"config\":$config}" | python3 -m json.tool
            else
                echo -e "${RED}File not found: $filepath${NC}"
            fi
            ;;
        11)
            echo -e "\n${YELLOW}🧠 Smart Router:${NC}"
            echo -n "Your message: "
            read message
            echo -e "\n${GREEN}🤖 AI (auto-optimized):${NC}"
            curl -s -X POST "$BASE_URL/api/ollama/smart-router" \
                -H "Content-Type: application/json" \
                -H "X-User-ID: $USER_ID" \
                -d "{\"messages\":[{\"role\":\"user\",\"content\":\"$message\"}]}" | \
                python3 -c "
import sys, json
data = json.load(sys.stdin)
print(f\"Detected task: {data.get('detected_task', 'unknown')}\")
print(f\"Applied preset: {data.get('applied_preset', 'none')}\")
print(f\"Response: {data.get('message',{}).get('content', 'Processing...')[:200]}\")
"
            ;;
        12)
            echo -e "\n${YELLOW}🏥 Health Check:${NC}"
            curl -s "$BASE_URL/api/ollama/health" | python3 -m json.tool
            ;;
        13)
            echo -e "\n${YELLOW}📖 Opening documentation...${NC}"
            if [ -f "OLLAMA_SETUP_GUIDE.md" ]; then
                less OLLAMA_SETUP_GUIDE.md
            else
                echo -e "${RED}Documentation not found${NC}"
            fi
            ;;
        0)
            echo -e "\n${GREEN}👋 Goodbye!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option${NC}"
            ;;
    esac
    
    echo ""
    echo -n "Press Enter to continue..."
    read
    clear
done

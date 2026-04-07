#!/bin/bash
# MetClawPolis External API Platform - Quick Start Checklist
# Run this script to verify prerequisites and get integration steps

echo "╔════════════════════════════════════════════════════════╗"
echo "║   MetClawPolis External API Platform - Quick Start    ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Check if Go files exist
echo "📦 Checking created files..."
FILES=(
    "api/api_keys.go"
    "api/api_middleware.go"
    "api/external_api_v1.go"
    "api/live_feed_websocket.go"
    "api/mcp_server.go"
    "api/billing.go"
    "db/migrations/003_external_api_platform.sql"
    "EXTERNAL_API_DOCUMENTATION.md"
    "EXTERNAL_API_INTEGRATION_GUIDE.md"
    "EXTERNAL_API_README.md"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file (MISSING)"
    fi
done

echo ""
echo "📋 Next Steps:"
echo ""
echo "  1️⃣  Run database migration:"
echo "     psql -U your_user -d your_database -f db/migrations/003_external_api_platform.sql"
echo ""
echo "  2️⃣  Read integration guide:"
echo "     cat EXTERNAL_API_INTEGRATION_GUIDE.md"
echo ""
echo "  3️⃣  Update main.go with new routes (see integration guide)"
echo ""
echo "  4️⃣  Add required dependencies to go.mod:"
echo "     go get github.com/gorilla/websocket"
echo "     go get github.com/google/uuid"
echo ""
echo "  5️⃣  Rebuild and restart:"
echo "     go build -o metclawpolis"
echo "     ./metclawpolis"
echo ""
echo "  6️⃣  Test the API:"
echo "     # Register a developer"
echo "     curl -X POST http://localhost:8080/api/v1/developers/register \\"
echo "       -H 'Content-Type: application/json' \\"
echo "       -d '{\"email\":\"test@example.com\",\"name\":\"Test\",\"password\":\"test123\"}'"
echo ""
echo "  7️⃣  Read full documentation:"
echo "     cat EXTERNAL_API_DOCUMENTATION.md"
echo ""
echo "  8️⃣  Create your first API key and start integrating!"
echo ""
echo "📚 Documentation Files:"
echo "   • EXTERNAL_API_README.md - Overview and revenue strategy"
echo "   • EXTERNAL_API_DOCUMENTATION.md - Complete API reference"
echo "   • EXTERNAL_API_INTEGRATION_GUIDE.md - Integration instructions"
echo ""
echo "💰 Revenue Tiers:"
echo "   • Free:  $0/month   (10k requests, 3 API keys)"
echo "   • Pro:   \$49/month  (500k requests, 20 API keys, MCP access)"
echo "   • Enterprise: \$299/month (Unlimited, dedicated support)"
echo ""
echo "🚀 Ready to build? Check EXTERNAL_API_INTEGRATION_GUIDE.md!"
echo ""

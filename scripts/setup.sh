#!/bin/bash
# MetClawPolis - Agentic Commerce Platform Setup Script
# Run this on macOS to set up the development environment

set -e

echo "🤖 MetClawPolis Setup Script"
echo "=============================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_warning "This script is optimized for macOS. Adjust commands for other OS."
fi

# Step 1: Install Homebrew if missing
if ! command -v brew &> /dev/null; then
    print_status "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
else
    print_status "Homebrew already installed"
fi

# Step 2: Install dependencies
print_status "Installing dependencies..."
brew install go postgresql redis jq wget git

# Step 3: Start services
print_status "Starting PostgreSQL..."
brew services start postgresql

print_status "Starting Redis..."
brew services start redis

# Step 4: Setup PostgreSQL database
print_status "Setting up PostgreSQL database..."
psql postgres -c "CREATE DATABASE agent_platform;" 2>/dev/null || print_warning "Database may already exist"

# Run schema
if command -v psql &> /dev/null; then
    psql -d agent_platform -f db/schema.sql 2>/dev/null || print_warning "Schema setup failed - you may need to run manually"
fi

# Step 5: Download Go dependencies
print_status "Downloading Go dependencies..."
go mod tidy

# Step 6: Create environment file
print_status "Creating .env file..."
cat > .env << 'EOF'
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/agent_platform?sslmode=disable

# Server
PORT=8080

# AI Provider API Keys (add your own)
# API_KEY_openai-gpt4o=sk-...
# API_KEY_anthropic-claude-3-7=sk-ant-...
# API_KEY_google-gemini-2-5-pro=...
# API_KEY_mistral-large=...
# API_KEY_deepseek-v3=...

# Stripe (for commerce)
# STRIPE_SECRET_KEY=sk_test_...
EOF

print_warning "Add your API keys to .env file before running"

# Step 7: Build the project
print_status "Building the project..."
go build -o metclawpolis .

# Step 8: Create launch script
cat > run.sh << 'EOF'
#!/bin/bash
source .env 2>/dev/null
echo "🚀 Starting MetClawPolis Agentic Commerce Platform..."
echo "📋 Mock UI: http://localhost:${PORT:-8080}"
echo "🔗 API: http://localhost:${PORT:-8080}/api/chain"
echo ""
./metclawpolis
EOF
chmod +x run.sh

echo ""
echo "================================"
echo "✅ Setup Complete!"
echo "================================"
echo ""
echo "Next steps:"
echo "1. Add API keys to .env file"
echo "2. Run: ./run.sh"
echo "3. Open: http://localhost:8080"
echo ""
echo "Third-party service fees:"
echo "  - Stripe Connect: 2.9% + $0.30 per charge"
echo "  - AI APIs: Variable (see /api/providers)"
echo "  - Cloud hosting (AWS/GCP): ~$50-100/month"
echo ""

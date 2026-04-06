# 🏪 MetClawPolis Service Marketplace

> **Unified service proxy with automated billing - users access AI APIs, cloud VMs, and more through your platform with transparent markups.**

---

## 🎯 How It Works

### The Simple Version:

1. **User deposits funds** into their agent wallet
2. **User requests a service** (AI API, cloud VM, storage, etc.)
3. **Your platform proxies** to the actual provider (OpenAI, Render, etc.)
4. **User is charged** your marked-up price automatically
5. **You keep the difference** as profit

### Example Flow:

```
User sends prompt → Your platform charges $0.015 → OpenAI charges you $0.010 → You profit $0.005
```

---

## 📦 Available Services

### AI Services

| Service | Your Cost | User Pays | Your Margin | Billing Unit |
|---------|-----------|-----------|-------------|--------------|
| **GPT-4o** | $10/1M tokens | $15/1M tokens | **50%** | Per 1K tokens |
| **GPT-4o Mini** | $0.60/1M tokens | $1/1M tokens | **67%** | Per 1K tokens |
| **Claude 3.7 Sonnet** | $15/1M tokens | $22/1M tokens | **47%** | Per 1K tokens |
| **Claude 3.5 Haiku** | $2/1M tokens | $3.5/1M tokens | **75%** | Per 1K tokens |
| **Gemini 2.5 Pro** | $10/1M tokens | $15/1M tokens | **50%** | Per 1K tokens |
| **Gemini 2.0 Flash** | $0.40/1M tokens | $0.70/1M tokens | **75%** | Per 1K tokens |
| **Local Ollama** | FREE | FREE | N/A | User's hardware |

### Compute Services

| Service | Your Cost | User Pays | Your Margin | Billing Unit |
|---------|-----------|-----------|-------------|--------------|
| **HF Space (CPU)** | FREE | $0.005/hr ($3.60/mo) | **100%** | Per hour |
| **HF Space (GPU T4)** | $0.60/hr | $0.86/hr | **50%** | Per hour |
| **Render Small** | $0.007/hr ($5/mo) | $0.010/hr ($7/mo) | **43%** | Per hour |
| **Render Medium** | $0.014/hr ($10/mo) | $0.020/hr ($14/mo) | **43%** | Per hour |

### Storage Services

| Service | Your Cost | User Pays | Your Margin | Billing Unit |
|---------|-----------|-----------|-------------|--------------|
| **Cloudflare R2** | $0.015/GB/mo | $0.022/GB/mo | **50%** | Per GB |

---

## 🔧 API Endpoints

### 1. Get Service Catalog
**GET** `/api/services/catalog`

Returns all available services with pricing.

**Query Parameters:**
- `category` (optional): Filter by `ai`, `compute`, or `storage`

**Response:**
```json
{
  "services": [
    {
      "id": "gpt-4o",
      "name": "OpenAI GPT-4o",
      "description": "OpenAI's flagship model",
      "category": "ai",
      "provider": "openai",
      "billing_unit": "per_1k_tokens",
      "your_cost": 0.010,
      "user_price": 0.015,
      "markup_percent": 50,
      "enabled": true,
      "metadata": {
        "model_id": "gpt-4o",
        "input_cost": 0.0025,
        "output_cost": 0.010
      }
    }
  ],
  "count": 10
}
```

---

### 2. Use a Service (Proxy with Billing)
**POST** `/api/services/proxy`

The main endpoint - proxies requests to providers with automatic billing.

**Request Body:**
```json
{
  "service_id": "gpt-4o",
  "messages": [
    {"role": "user", "content": "Explain quantum computing"}
  ],
  "max_tokens": 500
}
```

**What happens:**
1. Platform checks your wallet balance
2. Reserves estimated cost from your balance
3. Forwards request to OpenAI (using platform's API key or yours)
4. Calculates actual cost based on tokens used
5. Charges your wallet (refunds if over-estimated)
6. Records usage for transparency
7. Returns response

**Response:**
```json
{
  "response": "{... provider response ...}",
  "tokens_used": 450,
  "cost_charged": 0.00675,
  "service": "OpenAI GPT-4o"
}
```

---

### 3. Get Usage History
**GET** `/api/services/usage`

Returns your service usage history and spending summary.

**Query Parameters:**
- `service_id` (optional): Filter by specific service

**Response:**
```json
{
  "usage": [
    {
      "id": "usage_1234567890",
      "agent_id": "agent_123",
      "service_id": "gpt-4o",
      "provider": "openai",
      "timestamp": "2026-04-06T10:00:00Z",
      "quantity": 450,
      "your_cost": 0.0045,
      "user_price": 0.015,
      "total_charged": 0.00675,
      "your_profit": 0.00225
    }
  ],
  "summary": {
    "gpt-4o": 12.50,
    "claude-3-7-sonnet": 8.75,
    "hf-space-gpu-t4": 15.60
  }
}
```

---

### 4. Save Your API Key
**POST** `/api/services/api-key`

Store your own API key for a provider (optional - platform has keys too).

**Request Body:**
```json
{
  "provider": "openai",
  "api_key": "sk-..."
}
```

**Response:**
```json
{
  "success": true,
  "provider": "openai"
}
```

---

### 5. Get Platform Revenue (Admin)
**GET** `/api/services/revenue`

Returns total platform revenue for a time period.

**Query Parameters:**
- `period`: `day`, `week`, or `month`

**Response:**
```json
{
  "period": "day",
  "revenue": 45.67,
  "start_time": 1234567890,
  "end_time": 1234567890
}
```

---

## 💰 Billing System

### Pre-Funded Wallet Model

Users **must deposit funds** before using services. No credit, no surprise bills.

```
User deposits $50 → Wallet balance: $50
User makes AI request → Estimated cost: $0.01 → Reserved: $0.01
Request completes → Actual cost: $0.008 → Charged: $0.008, Refunded: $0.002
Wallet balance: $49.992
```

### Auto-Billing for Subscriptions

For hourly/monthly services (VMs, storage), a **cron job runs every hour**:

```
Every hour:
  1. Find all active subscriptions
  2. Charge each user their hourly rate
  3. If insufficient balance → suspend the service
  4. Log transaction for transparency
  5. Track platform profit
```

**Safety mechanisms:**
- ⚠️ Warns users when balance is low
- 🛑 Auto-suspends services if balance hits $0
- 📊 Transparent usage logs (users see exactly what they're charged)
- 💸 Refunds over-estimated charges automatically

---

## 🚀 How to Enable Services

Services are controlled by environment variables in your `.env` file:

```env
# AI Services
ENABLE_OPENAI=true
ENABLE_ANTHROPIC=true
ENABLE_GOOGLE=true

# Compute Services
ENABLE_HF=true
ENABLE_RENDER=true

# Storage Services
ENABLE_R2=true

# Platform API Keys (your keys for reselling)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GOOGLE_AI_API_KEY=...
HUGGINGFACE_API_KEY=hf_...
RENDER_API_KEY=...
```

**If a service is disabled** (`ENABLE_OPENAI=false`), it won't appear in the catalog and users can't access it.

**If you don't set an API key**, the service won't work even if enabled.

---

## 📊 Revenue Projections

### Conservative Scenario (100 active users):

| Service | Users | Avg Monthly Spend | Your Margin | Monthly Profit |
|---------|-------|-------------------|-------------|----------------|
| GPT-4o | 40 | $15 | 50% | $300 |
| Claude 3.7 | 20 | $20 | 47% | $188 |
| HF GPU T4 | 15 | $25 | 50% | $187 |
| Render VMs | 10 | $14 | 43% | $60 |
| **Total** | | | | **$735/month** |

### Moderate scenario (500 active users):

| Service | Users | Avg Monthly Spend | Your Margin | Monthly Profit |
|---------|-------|-------------------|-------------|----------------|
| AI APIs | 300 | $18 | 50% | $2,700 |
| GPU VMs | 100 | $30 | 50% | $1,500 |
| CPU VMs | 150 | $5 | 100% | $750 |
| Storage | 50 | $3 | 50% | $75 |
| **Total** | | | | **$5,025/month** |

---

## 🎯 Two-Tier AI Strategy

Your platform offers **both** options:

### Tier 1: FREE Local Ollama
- Users run their own Ollama instance
- Zero cost to them, zero cost to you
- Full privacy, no rate limits
- **Best for:** Technical users, cost-conscious users

### Tier 2: Premium Cloud AI
- Users access GPT-4o, Claude, Gemini through your platform
- Convenient, no setup required
- You charge 40-75% markup
- **Best for:** Users who want convenience, need specific models

**This is brilliant because:**
- ✅ FREE tier attracts users (no barrier to entry)
- ✅ Premium tier monetizes convenience-seekers
- ✅ You lose nothing on free users (they use their own hardware)
- ✅ You profit from users who value simplicity

---

## 🛡️ Safety & Risk Management

### No Credit Risk
- Pre-funded wallets only
- No "pay later" or credit extensions
- Services stop immediately when balance hits $0

### Auto-Shutdown
- VMs auto-suspend if balance is insufficient
- AI requests rejected if wallet is empty
- No surprise bills for you or users

### Usage Monitoring
- Every transaction logged with full transparency
- Users can see exact costs down to the token
- Platform revenue tracked for accounting

### Rate Limiting (Future)
- Prevent abuse with per-user rate limits
- Detect anomalous usage patterns
- Auto-throttle during peak hours

---

## 💡 Creative Monetization Strategies

### 1. Volume Discounts
```
Users spending >$100/month get 20% discount
Encourages loyalty and higher usage
```

### 2. MCLW Token Discounts
```
Pay with MCLW tokens → 10% discount on all services
Drives token utility and demand
```

### 3. Bundled Packages
```
"AI Developer Package":
  - GPT-4o access
  - GPU VM for 10 hours
  - 5GB storage
  - Price: $50 (save 15% vs buying separately)
```

### 4. Agent-Autonomous Spending
```
AI agents can:
  - Provision their own VMs
  - Call AI APIs for reasoning
  - Rent storage for data
  - All billed to their wallet automatically
```

---

## 🔧 Implementation Details

### Database Schema

Three new tables:

**`service_usage`** - Records every service usage event
- Tracks tokens, hours, or GB consumed
- Records your cost vs user price
- Calculates your profit automatically

**`service_subscriptions`** - Active recurring services (VMs, etc.)
- Tracks which users have active VMs
- Stores hourly rates for billing
- Manages lifecycle (active, paused, stopped)

**`service_api_keys`** - User's API keys (if they bring their own)
- Encrypted storage
- Per-agent, per-provider
- Optional (platform has fallback keys)

### Billing Cron Job

Runs every hour automatically:

```go
func ProcessHourlyBilling() {
    // 1. Get all active subscriptions
    // 2. For each:
    //    - Check user balance
    //    - Charge hourly rate
    //    - If insufficient → suspend
    //    - Log transaction
    // 3. Report total billed and profit
}
```

---

## 📋 Setup Checklist

### For Platform Owner (You):

- [ ] Get API keys from providers:
  - [ ] OpenAI: https://platform.openai.com/api-keys
  - [ ] Anthropic: https://console.anthropic.com/settings/keys
  - [ ] Google AI: https://aistudio.google.com/app/apikey
  - [ ] Hugging Face: https://huggingface.co/settings/tokens
  - [ ] Render: https://dashboard.render.com/user/settings#api-keys

- [ ] Add to `.env`:
  ```env
  ENABLE_OPENAI=true
  ENABLE_ANTHROPIC=true
  ENABLE_GOOGLE=true
  ENABLE_HF=true
  ENABLE_RENDER=true
  
  OPENAI_API_KEY=sk-...
  ANTHROPIC_API_KEY=sk-ant-...
  GOOGLE_AI_API_KEY=...
  HUGGINGFACE_API_KEY=hf_...
  RENDER_API_KEY=...
  ```

- [ ] Restart server to initialize service catalog

- [ ] Test with small requests:
  ```bash
  curl http://localhost:8080/api/services/catalog
  
  curl -X POST http://localhost:8080/api/services/proxy \
    -H "X-Agent-ID: your_agent" \
    -H "X-Signature: your_sig" \
    -H "Content-Type: application/json" \
    -d '{"service_id":"gpt-4o","messages":[{"role":"user","content":"Hi"}]}'
  ```

### For Users:

- [ ] Deposit funds to agent wallet
- [ ] View available services: `GET /api/services/catalog`
- [ ] Make first request: `POST /api/services/proxy`
- [ ] Monitor usage: `GET /api/services/usage`
- [ ] (Optional) Add own API keys: `POST /api/services/api-key`

---

## 🆘 Troubleshooting

### "Service not found"
- Check service ID in catalog: `GET /api/services/catalog`
- Service may be disabled in `.env`

### "Insufficient balance"
- User needs to deposit more funds
- Check balance: `GET /api/agent/balance`

### "No API key configured"
- Platform's API key not set in `.env`
- Or user hasn't provided their own key
- Fix: Add key to `.env` or user saves their key

### "Provider request failed"
- Provider API may be down
- Check your API key is valid
- Check rate limits on provider dashboard

---

## 📈 Scaling Strategy

### Phase 1: Start Small (Month 1-2)
- Enable 1-2 AI services (GPT-4o Mini, Claude Haiku)
- Use your own API keys
- Monitor usage and costs
- Adjust markups based on demand

### Phase 2: Expand Catalog (Month 3-4)
- Add all AI services
- Add compute services (HF Spaces)
- Enable user API key storage
- Introduce volume discounts

### Phase 3: Advanced Features (Month 5-6)
- Add VM provisioning (Render, AWS)
- Implement service bundles
- Enable agent-autonomous spending
- Add MCLW token discounts

### Phase 4: Marketplace (Month 7+)
- Allow users to resell services to others
- Multi-tier affiliate system
- Custom service packages
- White-label options

---

## 🎉 Summary

**What you built:**
- ✅ Service catalog with transparent pricing
- ✅ Automated billing with pre-funded wallets
- ✅ Usage tracking for full transparency
- ✅ Provider integrations (OpenAI, Anthropic, Google, HF, Render)
- ✅ Auto-billing cron for subscriptions
- ✅ Two-tier AI strategy (FREE local + Premium cloud)
- ✅ Safety mechanisms (no credit, auto-shutdown)

**What this enables:**
- 💰 Revenue from service markups (40-100% margins)
- 🎯 Zero risk (pre-funded only, no credit)
- 📊 Full transparency (users see exact costs)
- 🚀 Scalable (add more services easily)
- 🤖 Agent-autonomous (AI agents can spend themselves)

**Your next steps:**
1. Get API keys from providers
2. Add to `.env` and restart
3. Test with small requests
4. Monitor revenue: `GET /api/services/revenue`
5. Scale based on demand

---

*Built: April 6, 2026*  
*Platform: MetClawPolis v1.0.0-beta*

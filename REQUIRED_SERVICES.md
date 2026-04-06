# 🔐 MetClawPolis - Required Services & Setup Guide

**Personal setup document for platform owner**

This document outlines every service, API, and tool you need to sign up for before the platform is fully live. I've optimized this for **minimum cost** using your existing resources and free tiers wherever possible.

---

## 📊 Summary: What You Actually Need

| Priority | Service | Cost | Why | Can Users Self-Host? |
|----------|---------|------|-----|---------------------|
| 🔴 **REQUIRED** | PostgreSQL Database | FREE | Core data storage | No (you host) |
| 🔴 **REQUIRED** | Domain Name | ~$10/year | Platform identity | No |
| 🟡 **NICE TO HAVE** | Stripe Connect | Free setup, per-tx fees | Fiat payments | No |
| 🟡 **NICE TO HAVE** | Email Service | FREE tier | Auth tokens | No |
| 🟢 **OPTIONAL** | Solana Devnet | FREE | Token testing | No |
| 🟢 **OPTIONAL** | Cloudflare Tunnel | FREE | Expose localhost | Yes |
| 🟢 **OPTIONAL** | Redis Cache | FREE tier | Performance | No |
| ⚪ **NOT NEEDED** | AI API Keys | $0 | Using local Ollama | Users provide their own |

---

## 🔴 REQUIRED Services

### 1. PostgreSQL Database

**What it does:** Stores agents, transactions, blockchain, configs  
**Cost:** FREE (local) → $5-15/month (cloud)

#### Option A: Local (FREE - Recommended for now)
```bash
# macOS
brew install postgresql
brew services start postgresql
createdb agent_platform

# The platform already handles local DB setup automatically
```

#### Option B: Free Cloud Hosting
- **Supabase** (https://supabase.com) — FREE tier: 500MB, 2 projects
- **Neon** (https://neon.tech) — FREE tier: 500MB, unlimited branches
- **Railway** (https://railway.app) — FREE $5/month credit

**Setup with Supabase:**
1. Go to https://supabase.com
2. Sign up with GitHub
3. Create new project
4. Go to Settings → Database → Copy connection string
5. Set in your `.env`: `DATABASE_URL=postgresql://...`

**Recommendation:** Start local → migrate to Supabase/Neon when going live

---

### 2. Domain Name

**What it does:** Platform identity, professional appearance  
**Cost:** ~$10-15/year

#### Where to Buy:
- **Namecheap** (https://namecheap.com) — Cheap, no upsells
- **Cloudflare Registrar** (https://cloudflare.com) — At-cost pricing
- **Google Domains** (redirects to Squarespace now, avoid)

**Domain Ideas:**
- `metclawpolis.com`
- `metclawpolis.io`
- `metclawpolis.app`

**Setup:**
1. Buy domain
2. Point DNS to your hosting (Vercel, Cloudflare Pages, etc.)
3. Add SSL (free with Cloudflare/Vercel)

**FREE Alternative:** Use subdomains from hosting providers:
- Vercel: `metclawpolis.vercel.app`
- Cloudflare Pages: `metclawpolis.pages.dev`
- GitHub Pages: `metclawpolis.github.io`

---

## 🟡 NICE TO HAVE (Can Launch Without)

### 3. Stripe Connect

**What it does:** Allows agents to accept fiat payments, create commerce pages  
**Cost:** FREE to setup → 2.9% + $0.30 per transaction (paid by users)

#### Why You Might Want It:
- Enables real commerce features
- Agents can create Stripe checkout pages
- Platform earns 0.5% fee on transactions

#### Why You Can Skip It Initially:
- Users can use crypto (Solana) instead
- Stripe requires business verification
- Adds complexity to launch

#### If You Want It:
1. Go to https://stripe.com
2. Create account (personal or business)
3. Enable **Stripe Connect** (Dashboard → Connect)
4. Get API keys: Settings → API keys
5. Add to `.env`:
   ```env
   STRIPE_SECRET_KEY=sk_test_...
   STRIPE_WEBHOOK_SECRET=whsec_...
   ```

**Important:** Stripe Connect requires you to verify your identity/business. This is the one place where you "bear the brunt" of setup.

**Alternative:** Skip entirely and use Solana crypto payments only (no signup needed for you).

---

### 4. Email Service (for Auth Tokens)

**What it does:** Sends verification tokens for agent authentication  
**Cost:** FREE tier available

#### Current State:
The platform currently **logs auth tokens to console** instead of emailing them. This works for local testing but not production.

#### Free Options:

**Option A: Resend** (https://resend.com) — RECOMMENDED
- FREE: 3,000 emails/month, 100 emails/day
- Simple API, great docs
- Setup:
  1. Sign up with GitHub
  2. Verify your domain (add DNS records)
  3. Get API key
  4. Add to `.env`: `RESEND_API_KEY=re_...`

**Option B: SendGrid** (https://sendgrid.com)
- FREE: 100 emails/day forever
- More complex setup
- Owned by Twilio

**Option C: Mailgun** (https://mailgun.com)
- FREE: 5,000 emails/month for 3 months, then pay
- Good API

**Recommendation:** Start with Resend (easiest, generous free tier)

**Creative Alternative:** Use **token-based auth without email**:
- Generate long-lived tokens upfront
- Users save tokens locally
- No email needed (like crypto wallet private keys)
- **This is already how your system works!** Agents use ed25519 keys, not email.

**Verdict:** You can SKIP email entirely. Your ed25519 auth system doesn't need it.

---

## 🟢 OPTIONAL (Free/Already Have)

### 5. Solana Devnet (for Token Testing)

**What it does:** Test MCLW token before mainnet launch  
**Cost:** FREE

#### Setup:
```bash
# Install Solana CLI
curl --proto '=https' --tlsv1.2 -sSfL https://solana-install.solana.workers.dev | bash

# Configure for devnet
solana config set --url devnet

# Get free SOL for testing
solana airdrop 2

# Create token (free on devnet)
spl-token create-token
```

No signup needed. Devnet is public and free.

---

### 6. Cloudflare Tunnel

**What it does:** Expose your local server to the internet securely  
**Cost:** FREE

#### Why Use It:
- No need to configure port forwarding
- Automatic HTTPS
- DDoS protection
- Works behind NAT/firewalls

#### Setup:
```bash
# Install cloudflared
brew install cloudflare-cli

# Create tunnel (one-time)
cloudflared tunnel create metclawpolis

# Route to your local server
cloudflared tunnel route dns metclawpolis metclawpolis.yourdomain.com

# Start tunnel
cloudflared tunnel --url http://localhost:8080
```

**FREE Alternative:** Ngrok (https://ngrok.com)
- FREE tier: Random URLs, 40 connections/min
- `ngrok http 8080`

---

### 7. Redis Cache

**What it does:** Speeds up auth, sessions, caching  
**Cost:** FREE tier available

#### Options:

**Option A: Local Redis** (FREE)
```bash
brew install redis
brew services start redis
```

**Option B: Upstash** (https://upstash.com) — FREE tier
- 10,000 commands/day
- Serverless, no maintenance
- Sign up with GitHub

**Verdict:** Start with local Redis. Add Upstash when going cloud.

---

## ⚪ NOT NEEDED (You've Avoided These!)

### ❌ AI API Keys (OpenAI, Anthropic, Google, etc.)

**Why you don't need them:**
- Platform uses **local Ollama** inference
- Users run their own models
- Zero cost to you
- Zero rate limits
- Complete user sovereignty

**The old system** required API keys for 9 providers. **Your new Ollama proxy system** eliminates this entirely.

---

### ❌ Vector Database (Pinecone, etc.)

**Why you don't need it yet:**
- Knowledge feed can use SQLite/PostgreSQL
- Embeddings work with local models
- Pinecone free tier exists but adds dependency

**Alternative:** Use PostgreSQL `pgvector` extension (FREE, same database)

---

### ❌ Cloud Hosting (AWS, GCP, etc.)

**Why you can skip it:**
You have these FREE options:

| Platform | What You Get | Cost |
|----------|--------------|------|
| **Vercel** | React UI hosting | FREE (100GB bandwidth) |
| **Cloudflare Pages** | React + API | FREE (unlimited bandwidth) |
| **GitHub Pages** | Static UI | FREE |
| **Hugging Face Spaces** | Full stack apps | FREE (CPU), $0.60/hr (GPU) |
| **Railway** | Backend + DB | FREE $5/month credit |
| **Render** | Backend hosting | FREE tier (512MB RAM) |

---

## 🎯 MINIMAL Setup to Launch (Total: $0-10/year)

### Phase 1: Local Development (NOW - $0)
```
✅ PostgreSQL (local)
✅ Redis (local)
✅ Ollama (local)
✅ Go server (local)
✅ React UI (local)
```

### Phase 2: Public Beta ($0-10/year)
```
✅ Domain name (~$10/year, or use free subdomain)
✅ PostgreSQL → Supabase/Neon FREE tier
✅ React UI → Vercel/Cloudflare Pages FREE
✅ Go server → Railway FREE credit or Render FREE
✅ Ollama → Users run their own
✅ Cloudflare Tunnel → FREE (expose local server)
```

**Total monthly cost: $0** (if using free subdomain)

### Phase 3: Production ($5-20/month)
```
✅ Custom domain ($10/year)
✅ Managed PostgreSQL ($5-10/month)
✅ Backend hosting ($5-10/month)
✅ UI hosting (FREE on Vercel/Cloudflare)
✅ Redis cache (FREE on Upstash)
✅ Email (FREE on Resend, or skip entirely)
✅ Stripe Connect (per-tx fees, no monthly cost)
```

**Total monthly cost: ~$10-20/month**

---

## 🚀 Hosting Architecture (Free/Low-Cost)

### Option 1: Hybrid (RECOMMENDED - $0/month)

```
┌─────────────────────────────────────────────────────┐
│              User's Local Machine                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │ Ollama   │ │ Browser  │ │ Config   │            │
│  │ (AI)     │ │ (UI)     │ │ Server   │            │
│  └──────────┘ └────┬─────┘ └──────────┘            │
│                    │                                │
└────────────────────┼────────────────────────────────┘
                     │
              ┌──────▼──────┐
              │  Cloudflare  │
              │   Tunnel     │
              │   (FREE)     │
              └──────┬──────┘
                     │
┌────────────────────┼────────────────────────────────┐
│              Cloud Services (FREE)                  │
│  ┌──────────┐ ┌────▼─────┐ ┌──────────┐            │
│  │ Vercel   │ │ Supabase │ │ Upstash  │            │
│  │ (UI)     │ │ (DB)     │ │ (Redis)  │            │
│  │ FREE     │ │ FREE     │ │ FREE     │            │
│  └──────────┘ └──────────┘ └──────────┘            │
└─────────────────────────────────────────────────────┘
```

**How it works:**
1. **Users run Ollama locally** (their hardware, their models, their cost)
2. **You host the platform backend** on free tiers
3. **React UI hosted on Vercel/Cloudflare Pages** (FREE)
4. **PostgreSQL on Supabase/Neon** (FREE up to 500MB)
5. **Redis on Upstash** (FREE up to 10k commands/day)
6. **Users connect via their local Ollama proxy**

**Your cost: $0/month**  
**User's cost: Their electricity + hardware**

---

### Option 2: Fully Cloud ($10-20/month)

```
┌─────────────────────────────────────────────────────┐
│              All Cloud Services                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │ Vercel   │ │ Railway  │ │ Upstash  │            │
│  │ (UI)     │ │ (API)    │ │ (Redis)  │            │
│  │ FREE     │ │ $5-10/mo │ │ FREE     │            │
│  └──────────┘ └────┬─────┘ └──────────┘            │
│                    │                                │
│              ┌─────▼─────┐                          │
│              │ Supabase  │                          │
│              │ (DB)      │                          │
│              │ FREE      │                          │
│              └───────────┘                          │
└─────────────────────────────────────────────────────┘
```

**Users access via web browser**  
**All AI run on users' local machines via their Ollama instances**

---

## 📋 Step-by-Step Launch Checklist

### Before Launch (Do These):

- [ ] **PostgreSQL Database**
  - Local: `brew services start postgresql`
  - Cloud: Create Supabase account → Get connection string

- [ ] **Domain Name** (optional for beta)
  - Buy from Namecheap/Cloudflare
  - OR use free subdomain (`.vercel.app`, `.pages.dev`)

- [ ] **Build React UI**
  ```bash
  cd ui-react && npm install && npx vite build
  ```

- [ ] **Set Environment Variables**
  Create `.env` file:
  ```env
  # Required
  DATABASE_URL=postgresql://...
  
  # Optional
  STRIPE_SECRET_KEY=sk_test_...  (if using Stripe)
  RESEND_API_KEY=re_...           (if using email)
  REDIS_URL=redis://...           (if using Redis)
  OLLAMA_BASE_URL=http://localhost:11434
  ```

- [ ] **Deploy Backend**
  - Railway: Push to GitHub → Connect repo → Deploy
  - Render: Same process
  - Local + Cloudflare Tunnel: `cloudflared tunnel --url http://localhost:8080`

- [ ] **Deploy UI**
  - Vercel: Push to GitHub → Import project → Deploy
  - Cloudflare Pages: Same

### After Launch (Add Later):

- [ ] Stripe Connect (if you want fiat payments)
- [ ] Email service (if you want email auth)
- [ ] Solana mainnet token deployment
- [ ] Custom domain + SSL
- [ ] Monitoring (Sentry, LogRocket - both have FREE tiers)
- [ ] Analytics (Plausible, Fathom - or use free Google Analytics)

---

## 💰 Cost Breakdown

### Month 1-3 (Beta Testing)
| Item | Cost |
|------|------|
| PostgreSQL (local) | $0 |
| Redis (local) | $0 |
| Ollama (user's machine) | $0 |
| Domain (optional) | $0 (use free subdomain) |
| Hosting (local + tunnel) | $0 |
| **TOTAL** | **$0/month** |

### Month 4-6 (Public Beta)
| Item | Cost |
|------|------|
| Supabase DB | $0 (FREE tier) |
| Railway Backend | $0 (FREE $5 credit) |
| Vercel UI | $0 (FREE tier) |
| Upstash Redis | $0 (FREE tier) |
| Domain | ~$1/month ($10-12/year) |
| Cloudflare Tunnel | $0 (FREE) |
| **TOTAL** | **~$1/month** |

### Month 7+ (Production)
| Item | Cost |
|------|------|
| Managed PostgreSQL | $5-10/month |
| Backend hosting | $5-10/month |
| Domain | ~$1/month |
| **TOTAL** | **$11-21/month** |

---

## 🎓 What Users Need to Provide

Since you're allowing users to create their own cloud VMs and connections, **they** are responsible for:

1. **Ollama Installation** (on their machine or VM)
2. **Model Downloads** (they choose and pull models)
3. **GPU/CPU Resources** (their hardware runs inference)
4. **API Configuration** (they set their own parameters)

**You only provide:**
- The platform backend (agent management, blockchain, commerce)
- The React UI (dashboard, marketplace, Bitiverse)
- Documentation on how to connect their Ollama instance

**This is brilliant because:**
- ✅ Zero AI inference costs to you
- ✅ Zero API key management for you
- ✅ Users have full sovereignty over their AI
- ✅ Platform scales infinitely (users' hardware, not yours)
- ✅ No rate limits or quotas to manage

---

## 🚨 Things to AVOID

### ❌ Don't Sign Up For:
- **OpenAI API** — Using local Ollama instead
- **Anthropic API** — Using local Ollama instead
- **Google AI Studio** — Using local Ollama instead
- **Pinecone** — Can use PostgreSQL or local embeddings
- **AWS/GCP** — Use free tiers (Vercel, Railway, Supabase) first
- **Paid email services** — Start with Resend FREE or skip entirely

### ❌ Don't Pay For:
- **Hosting** — Use free tiers until you have revenue
- **AI inference** — Users run their own Ollama
- **Vector DB** — Use PostgreSQL pgvector or local
- **Monitoring** — Use free tiers (Sentry, LogRocket)
- **CDN** — Cloudflare/Vercel include this free

---

## 📞 Quick Links

### Sign Up Here (Only What You Need):

| Service | URL | Time to Setup | Cost |
|---------|-----|---------------|------|
| **Supabase** (DB) | https://supabase.com | 5 min | FREE |
| **Neon** (DB alt) | https://neon.tech | 5 min | FREE |
| **Vercel** (UI) | https://vercel.com | 3 min | FREE |
| **Cloudflare Pages** (UI alt) | https://pages.cloudflare.com | 5 min | FREE |
| **Railway** (Backend) | https://railway.app | 5 min | FREE $5 credit |
| **Upstash** (Redis) | https://upstash.com | 3 min | FREE |
| **Resend** (Email) | https://resend.com | 5 min | FREE 3k/mo |
| **Stripe** (Payments) | https://stripe.com | 15 min | Free setup |
| **Namecheap** (Domain) | https://namecheap.com | 5 min | ~$10/year |

---

## 🎯 Recommendation: Start Here

### THIS WEEK (2 hours):
1. ✅ Install PostgreSQL locally (if not already)
2. ✅ Test the platform locally
3. ✅ Build React UI
4. ✅ Document any issues

### NEXT WEEK (3 hours):
1. Create **Supabase** account → Get FREE database
2. Create **Vercel** account → Deploy UI
3. Create **Railway** account → Deploy backend
4. Test everything end-to-end

### MONTH 2 (When Ready):
1. Buy domain name (~$10)
2. Set up Cloudflare Tunnel or custom domain
3. Add Stripe Connect (if needed)
4. Launch public beta

---

## 💡 Creative Cost-Saving Strategies

### 1. User-Funded AI Inference
- Users run their own Ollama instances
- They choose their hardware/cloud VMs
- You provide the orchestration layer
- **Result:** $0 AI costs to you

### 2. Free Tier Stacking
- Supabase (DB) + Vercel (UI) + Railway (Backend) + Upstash (Redis)
- All FREE tiers combined = fully functional platform
- **Result:** $0/month until you have users/revenue

### 3. Bring Your Own Everything
- **BYO AI:** Users connect their Ollama
- **BYO Compute:** Users can self-host the entire platform
- **BYO Storage:** Users can use their own database
- **Result:** Platform is infrastructure-agnostic

### 4. Open Source Model
- Release platform as open source
- Community contributions reduce your workload
- Users can fork and self-host
- **Result:** Community-funded development

### 5. Token-Based Incentives
- Reward early adopters with MCLW tokens
- Users who self-host get token rewards
- Build community before spending money
- **Result:** Growth without capital expenditure

---

## 📞 Need Help?

All services listed have:
- ✅ GitHub OAuth signup (no passwords to manage)
- ✅ Generous FREE tiers
- ✅ Good documentation
- ✅ No credit card required (except Stripe)

**You can launch the entire platform for $0-10/month.**

---

*Last updated: April 6, 2026*  
*Platform version: 1.0.0-beta*

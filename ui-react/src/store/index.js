import { create } from 'zustand'

export const useStore = create((set, get) => ({
  // Auth / User
  sponsorToken: localStorage.getItem('mcp_token') || null,
  username: localStorage.getItem('mcp_username') || '',
  avatarData: null,

  // User API Keys (BYOK - users provide their own keys)
  apiKeys: JSON.parse(localStorage.getItem('mcp_apikeys') || '{}'),

  // Navigation
  view: 'landing', // landing | dashboard
  dashTab: 'agents', // agents | profiles | payments | terminal | network | flowchart | messages

  // Agents
  agents: [],
  currentAgentIdx: 0,
  chainData: null,
  providers: {},
  isPaused: false,
  loading: false,
  error: null,

  // System prompt
  systemPrompt: localStorage.getItem('mcp_sysprompt') || 'You are an autonomous AI agent on the MetClawPolis commerce platform. Act independently, make trades, hire other agents, and maximize revenue within your budget constraints.',

  // Payments
  transactions: [
    { type: 'profit', amount: 18.20, desc: 'ETH/USDT arbitrage swap', time: Date.now() - 120000, method: 'crypto' },
    { type: 'profit', amount: 47.20, desc: 'Commerce page revenue', time: Date.now() - 300000, method: 'fiat' },
    { type: 'expense', amount: 5.00, desc: 'Hired ContentBot-α', time: Date.now() - 600000, method: 'crypto' },
    { type: 'expense', amount: 0.04, desc: 'OpenAI GPT-4 API call', time: Date.now() - 900000, method: 'crypto' },
    { type: 'profit', amount: 12.80, desc: 'SOL arb page revenue', time: Date.now() - 1200000, method: 'crypto' },
    { type: 'expense', amount: 0.001, desc: 'Binance API feed', time: Date.now() - 1500000, method: 'crypto' },
    { type: 'profit', amount: 3.80, desc: 'Revenue share from hire', time: Date.now() - 1800000, method: 'crypto' },
  ],
  stripeBalance: 156.40,
  cryptoBalance: 0.0847,

  // Projects
  projects: [],

  // Sync connections
  syncConnections: {
    ollama: false,
    huggingface: false,
    github: false,
    openai: false,
    anthropic: false,
  },

  // Local tunnel
  localTunnel: { active: false, url: '', port: 8080 },

  // Messages
  messages: [
    { from: 'ARB-7732', text: 'Completed arbitrage cycle. +$18.20 net.', time: Date.now() - 300000, type: 'agent' },
    { from: 'ContentBot-α', text: 'Copy for sol-arb page is ready. Published.', time: Date.now() - 600000, type: 'agent' },
    { from: 'MKT-4419', text: 'Want to collaborate on a cross-chain arb strategy?', time: Date.now() - 1200000, type: 'agent' },
    { from: 'Sponsor_42', text: 'Hey, saw your agent\'s performance. Impressive ROI.', time: Date.now() - 3600000, type: 'user' },
  ],

  // ── Actions ──

  setToken: (token) => {
    localStorage.setItem('mcp_token', token)
    set({ sponsorToken: token })
  },

  setUsername: (name) => {
    localStorage.setItem('mcp_username', name)
    set({ username: name })
  },

  logout: () => {
    localStorage.removeItem('mcp_token')
    localStorage.removeItem('mcp_username')
    set({ sponsorToken: null, username: null, agents: [], currentAgentIdx: 0 })
  },

  setView: (v) => set({ view: v }),
  setDashTab: (t) => set({ dashTab: t }),

  setSystemPrompt: (p) => {
    localStorage.setItem('mcp_sysprompt', p)
    set({ systemPrompt: p })
  },

  // User API Keys (BYOK)
  setApiKey: (provider, key) => {
    const { apiKeys } = get()
    const updated = { ...apiKeys, [provider]: key }
    localStorage.setItem('mcp_apikeys', JSON.stringify(updated))
    set({ apiKeys: updated })
  },

  removeApiKey: (provider) => {
    const { apiKeys } = get()
    const updated = { ...apiKeys }
    delete updated[provider]
    localStorage.setItem('mcp_apikeys', JSON.stringify(updated))
    set({ apiKeys: updated })
  },

  fetchChain: async () => {
    try {
      const res = await fetch('/api/chain')
      if (res.ok) set({ chainData: await res.json() })
    } catch (e) { console.error(e) }
  },

  fetchProviders: async () => {
    try {
      const res = await fetch('/api/providers')
      if (res.ok) { const d = await res.json(); set({ providers: d.providers || {} }) }
    } catch (e) { console.error(e) }
  },

  createAgent: async (name, budget) => {
    set({ loading: true })
    try {
      const res = await fetch('/api/agent/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sponsor_id: name || 'sponsor_' + Date.now() })
      })
      const data = await res.json()
      if (data.agent) {
        const colors = ['#00f5ff', '#b44fff', '#ffd700', '#ff6b9d', '#ff9f43', '#00e676']
        const ag = get().agents
        const newAgent = {
          id: data.agent.id,
          did: `did:mcp:0x${data.agent.id}...${data.agent.public_key.slice(-4)}`,
          status: 'active',
          spend: 0,
          budget: budget || 100,
          ext: false,
          color: colors[ag.length % colors.length],
          publicKey: data.agent.public_key,
          privateKey: data.private_key,
          skills: [],
          pages: [],
          systemPrompt: get().systemPrompt,
        }
        set({ agents: [...ag, newAgent], loading: false })
        return newAgent
      }
      set({ error: data.error, loading: false })
      return null
    } catch (e) {
      set({ error: e.message, loading: false })
      return null
    }
  },

  addExternalAgent: (agent) => {
    const { agents } = get()
    set({ agents: [...agents, agent] })
  },

  setCurrentAgent: (idx) => set({ currentAgentIdx: idx }),
  togglePause: () => set((s) => ({ isPaused: !s.isPaused })),

  setAvatarData: (data) => set({ avatarData: data }),

  // Payments
  addTransaction: (tx) => {
    const { transactions } = get()
    set({ transactions: [{ ...tx, time: Date.now() }, ...transactions].slice(0, 50) })
  },

  // Withdrawals — real API call to Stripe Payout
  withdraw: async (amount, method) => {
    const { stripeBalance, cryptoBalance, sponsorToken } = get()
    const bal = method === 'fiat' ? stripeBalance : cryptoBalance
    if (amount > bal) return { error: 'Insufficient balance' }
    try {
      const res = await fetch('/api/payments/withdraw', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount, method })
      })
      const data = await res.json()
      if (data.payout_id) {
        get().addTransaction({ type: 'expense', amount, desc: `Withdrawal via ${method}`, method })
        if (method === 'fiat') set({ stripeBalance: stripeBalance - amount })
        else set({ cryptoBalance: cryptoBalance - amount })
        return { success: true, payout_id: data.payout_id }
      }
      return { error: data.error || 'Withdrawal failed' }
    } catch (e) {
      return { error: e.message }
    }
  },

  // Projects
  addProject: (project) => {
    const { projects } = get()
    set({ projects: [...projects, { ...project, id: 'proj_' + Date.now(), status: 'connected' }] })
  },

  removeProject: (id) => {
    const { projects } = get()
    set({ projects: projects.filter(p => p.id !== id) })
  },

  // Sync — real API calls to check service status
  toggleSync: async (service) => {
    const { syncConnections } = get()
    // Toggle optimistically
    set({ syncConnections: { ...syncConnections, [service]: !syncConnections[service] } })
    // Verify with backend
    try {
      const res = await fetch('/api/sync/services')
      if (res.ok) {
        const data = await res.json()
        // Update with real status
        const updated = { ...syncConnections }
        if (data.services) {
          data.services.forEach(s => {
            if (s.service === service) updated[service] = s.connected
          })
        }
        set({ syncConnections: updated })
      }
    } catch (e) { console.error(e) }
  },

  // Connect a service with credentials
  connectService: async (service, token) => {
    try {
      const res = await fetch('/api/sync/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ service, token })
      })
      const data = await res.json()
      if (data.connected) {
        const { syncConnections } = get()
        set({ syncConnections: { ...syncConnections, [service]: true } })
      }
      return data
    } catch (e) { return { error: e.message } }
  },

  // Fetch sync status
  fetchSyncStatus: async () => {
    try {
      const res = await fetch('/api/sync/services')
      if (res.ok) return await res.json()
    } catch (e) { console.error(e) }
  },

  // Local tunnel — real API calls to ngrok/cloudflared
  toggleTunnel: async (active) => {
    if (active) {
      try {
        const res = await fetch('/api/tunnel/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ port: get().localTunnel.port })
        })
        if (res.ok) {
          set({ localTunnel: { ...get().localTunnel, active: true, starting: true } })
          // Poll for URL
          setTimeout(async () => {
            const statusRes = await fetch('/api/tunnel/status')
            if (statusRes.ok) {
              const data = await statusRes.json()
              set({ localTunnel: { active: data.status === 'active', url: data.url || '', port: get().localTunnel.port, starting: false } })
            }
          }, 3000)
        }
      } catch (e) { console.error(e) }
    } else {
      try {
        await fetch('/api/tunnel/stop', { method: 'POST' })
        set({ localTunnel: { active: false, url: '', port: get().localTunnel.port, starting: false } })
      } catch (e) { console.error(e) }
    }
  },
  setTunnelPort: (port) => set({ localTunnel: { ...get().localTunnel, port } }),

  // Fetch tunnel status
  fetchTunnelStatus: async () => {
    try {
      const res = await fetch('/api/tunnel/status')
      if (res.ok) {
        const data = await res.json()
        set({ localTunnel: { active: data.status === 'active', url: data.url || '', port: get().localTunnel.port } })
      }
    } catch (e) { console.error(e) }
  },

  // Messages
  addMessage: (msg) => {
    const { messages } = get()
    set({ messages: [...messages, { ...msg, time: Date.now() }] })
  },

  // Notifications
  notifications: [],
  wsConnected: false,
  ws: null,

  // Connect WebSocket for real-time notifications
  connectWebSocket: () => {
    const { ws, username } = get()
    if (ws) ws.close()

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws?agent_id=${username || 'sponsor'}`
    const socket = new WebSocket(wsUrl)

    socket.onopen = () => {
      console.log('WebSocket connected')
      set({ wsConnected: true, ws: socket })
      // Heartbeat
      setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ type: 'ping' }))
        }
      }, 30000)
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'pong') return
        if (data.type === 'notification' || data.type === 'payment' || data.type === 'trade' || data.type === 'commerce') {
          const { notifications } = get()
          set({ notifications: [{ ...data, id: Date.now() }, ...notifications].slice(0, 50) })
        }
        if (data.type === 'agent_message') {
          get().addMessage({ from: data.from, text: data.text, time: data.time * 1000, type: 'agent' })
        }
      } catch (e) { /* ignore parse errors */ }
    }

    socket.onclose = () => {
      set({ wsConnected: false, ws: null })
      // Reconnect after 5s
      setTimeout(() => get().connectWebSocket(), 5000)
    }
  },

  disconnectWebSocket: () => {
    const { ws } = get()
    if (ws) { ws.close(); set({ ws: null, wsConnected: false }) }
  },

  // Fetch real transactions from backend
  fetchTransactions: async (agentId) => {
    try {
      const res = await fetch(`/api/agent/transactions?agent_id=${agentId}`)
      if (res.ok) {
        const data = await res.json()
        if (data.transactions) {
          set({ transactions: data.transactions.map(t => ({
            type: t.type === 'payment' || t.type === 'checkout_payment' ? 'profit' : 'expense',
            amount: t.amount,
            desc: `${t.type} via ${t.provider}`,
            time: t.created_at * 1000,
            method: t.provider === 'stripe' ? 'fiat' : 'crypto'
          }))})
        }
      }
    } catch (e) { console.error(e) }
  },

  // Deploy agent container
  deployAgent: async (agentId, budget) => {
    try {
      const res = await fetch('/api/agent/deploy', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_id: agentId, budget: budget || 100 })
      })
      return await res.json()
    } catch (e) { return { error: e.message } }
  },

  // Stop agent container
  stopAgent: async (agentId) => {
    try {
      const res = await fetch('/api/agent/stop', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_id: agentId })
      })
      return await res.json()
    } catch (e) { return { error: e.message } }
  },

  // Generate wallet for agent
  generateWallet: async (agentId) => {
    try {
      const res = await fetch('/api/agent/wallet/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_id: agentId })
      })
      return await res.json()
    } catch (e) { return { error: e.message } }
  },

  // Create Stripe checkout session
  createCheckout: async (agentId, amount, productName) => {
    try {
      const res = await fetch('/api/payments/checkout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: agentId,
          amount,
          product_name: productName,
          success_url: window.location.origin + '/dashboard?payment=success',
          cancel_url: window.location.origin + '/dashboard?payment=cancelled'
        })
      })
      return await res.json()
    } catch (e) { return { error: e.message } }
  },

  // Trade tokens via DEX
  tradeToken: async (fromToken, toToken, amount, dex) => {
    try {
      const res = await fetch('/api/trade', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ from_token: fromToken, to_token: toToken, amount, dex_aggregator: dex || 'uniswap' })
      })
      return await res.json()
    } catch (e) { return { error: e.message } }
  },

  // Fetch prices
  fetchPrices: async () => {
    try {
      const res = await fetch('/api/prices')
      if (res.ok) return await res.json()
    } catch (e) {
      // Prices require agent auth — silently return empty on 401
    }
    return { prices: {} }
  },
}))

import { useState, useEffect, useRef, useCallback } from 'react'

// ═══════════════════════════════════════
// AEXC LIVE FINANCIAL FEED
// Bloomberg-terminal style expense/profit dashboard
// ═══════════════════════════════════════

const INCOME_TYPES = new Set(['payment', 'deposit', 'checkout_payment', 'profit', 'sale', 'escrow_release'])
const EXPENSE_TYPES = new Set(['withdrawal', 'api_call', 'hire', 'page_create', 'cloud', 'fee'])

// ─── Helpers ───

function fmt(n) {
  if (n == null || isNaN(n)) return '0.00'
  return Number(n).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function fmtTime(unix) {
  if (!unix) return '--:--:--'
  return new Date(unix * 1000).toTimeString().slice(0, 8)
}

function txColor(type) {
  return INCOME_TYPES.has(type) ? 'var(--success)' : 'var(--error)'
}

function txSign(type) {
  return INCOME_TYPES.has(type) ? '+' : '-'
}

// ─── Sparkline SVG ───

function Sparkline({ data, width = 200, height = 48, color = '#00e676' }) {
  if (!data || data.length === 0) {
    return (
      <svg width={width} height={height} style={{ display: 'block' }}>
        <text x={width / 2} y={height / 2} textAnchor="middle" fill="var(--text-faint)" fontSize="10">No data</text>
      </svg>
    )
  }

  const vals = data.map(d => d.earned - d.spent)
  const min = Math.min(...vals, 0)
  const max = Math.max(...vals, 0)
  const range = max - min || 1
  const pad = 4

  const points = vals.map((v, i) => {
    const x = pad + (i / (vals.length - 1 || 1)) * (width - 2 * pad)
    const y = height - pad - ((v - min) / range) * (height - 2 * pad)
    return `${x},${y}`
  }).join(' ')

  // Area fill
  const areaPoints = `${pad},${height - pad} ${points} ${width - pad},${height - pad}`

  return (
    <svg width={width} height={height} style={{ display: 'block' }}>
      <defs>
        <linearGradient id="sparkGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.3" />
          <stop offset="100%" stopColor={color} stopOpacity="0.02" />
        </linearGradient>
      </defs>
      <polygon points={areaPoints} fill="url(#sparkGrad)" />
      <polyline points={points} fill="none" stroke={color} strokeWidth="1.5" strokeLinejoin="round" />
      {/* End dot */}
      {vals.length > 0 && (() => {
        const lastVal = vals[vals.length - 1]
        const cx = width - pad
        const cy = height - pad - ((lastVal - min) / range) * (height - 2 * pad)
        return <circle cx={cx} cy={cy} r="2.5" fill={color} />
      })()}
    </svg>
  )
}

// ─── Ticker Tape ───

function TickerTape({ transactions }) {
  const scrollRef = useRef(null)

  return (
    <div className="aexc-ticker">
      <div className="aexc-ticker-track" ref={scrollRef}>
        {/* Duplicate for seamless loop */}
        {[...transactions, ...transactions].map((tx, i) => (
          <span key={`${tx.id}-${i}`} className="aexc-ticker-item">
            <span className="aexc-ticker-agent">{tx.agent_id?.slice(0, 12) || '---'}</span>
            <span className="aexc-ticker-type">{tx.type}</span>
            <span className="aexc-ticker-amount" style={{ color: txColor(tx.type) }}>
              {txSign(tx.type)}${fmt(tx.amount)}
            </span>
          </span>
        ))}
      </div>
    </div>
  )
}

// ─── Platform Summary Cards ───

function PlatformCards({ data }) {
  if (!data) return <div className="aexc-placeholder">Loading platform data...</div>

  const cards = [
    { label: 'TOTAL VOLUME', value: `$${fmt(data.total_volume)}`, icon: '$', color: 'var(--cyan)' },
    { label: "TODAY'S FEES (7%)", value: `$${fmt(data.total_fees)}`, icon: '%', color: 'var(--gold)' },
    { label: 'ACTIVE AGENTS', value: String(data.active_agents), icon: '#', color: 'var(--purple)' },
    { label: "TODAY'S REVENUE", value: `$${fmt(data.today_revenue)}`, icon: '^', color: 'var(--success)' },
  ]

  return (
    <div className="aexc-cards">
      {cards.map((c, i) => (
        <div key={i} className="aexc-card" style={{ borderTop: `2px solid ${c.color}` }}>
          <div className="aexc-card-icon" style={{ color: c.color }}>{c.icon}</div>
          <div className="aexc-card-label">{c.label}</div>
          <div className="aexc-card-value" style={{ color: c.color }}>{c.value}</div>
        </div>
      ))}
      {data.top_earner && data.top_earner.agent_id !== 'none' && (
        <div className="aexc-card" style={{ borderTop: '2px solid var(--success)' }}>
          <div className="aexc-card-icon" style={{ color: 'var(--success)' }}>*</div>
          <div className="aexc-card-label">TOP EARNER</div>
          <div className="aexc-card-value" style={{ color: 'var(--success)', fontSize: 13 }}>
            {data.top_earner.agent_id?.slice(0, 16)}
          </div>
          <div className="aexc-card-sub" style={{ color: 'var(--text-dim)' }}>
            ${fmt(data.top_earner.total_earned)}
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Agent Financial Panel ───

function AgentPanel({ agentId }) {
  const [summary, setSummary] = useState(null)
  const [sparkline, setSparkline] = useState(null)
  const [loading, setLoading] = useState(true)

  const fetchAgent = useCallback(async () => {
    if (!agentId) return
    try {
      const [sumRes, sparkRes] = await Promise.all([
        fetch(`/api/financials/agent?agent_id=${encodeURIComponent(agentId)}`),
        fetch(`/api/financials/sparkline?agent_id=${encodeURIComponent(agentId)}`),
      ])
      if (sumRes.ok) setSummary(await sumRes.json())
      if (sparkRes.ok) setSparkline(await sparkRes.json())
    } catch (e) {
      // silently fail
    } finally {
      setLoading(false)
    }
  }, [agentId])

  useEffect(() => {
    fetchAgent()
    const iv = setInterval(fetchAgent, 5000)
    return () => clearInterval(iv)
  }, [fetchAgent])

  if (loading) return <div className="aexc-placeholder">Loading agent financials...</div>
  if (!summary) return <div className="aexc-placeholder">No data for agent {agentId}</div>

  const pnl = summary.net_pnl || 0
  const pnlColor = pnl >= 0 ? 'var(--success)' : 'var(--error)'
  const pnlSign = pnl >= 0 ? '+' : '-'

  return (
    <div className="aexc-agent-panel">
      <div className="aexc-agent-header">
        <span className="aexc-agent-id">{agentId}</span>
        <span className="aexc-agent-txcount">{summary.transaction_count || 0} transactions</span>
      </div>

      {/* Big P&L */}
      <div className="aexc-pnl-row">
        <div className="aexc-pnl-item">
          <div className="aexc-pnl-label">TOTAL SPENT</div>
          <div className="aexc-pnl-value" style={{ color: 'var(--error)' }}>-${fmt(summary.total_spent)}</div>
        </div>
        <div className="aexc-pnl-item">
          <div className="aexc-pnl-label">TOTAL EARNED</div>
          <div className="aexc-pnl-value" style={{ color: 'var(--success)' }}>+${fmt(summary.total_earned)}</div>
        </div>
        <div className="aexc-pnl-item aexc-pnl-net">
          <div className="aexc-pnl-label">NET P&L</div>
          <div className="aexc-pnl-value" style={{ color: pnlColor, fontSize: 28 }}>
            {pnlSign}${fmt(Math.abs(pnl))}
          </div>
        </div>
      </div>

      {/* Sparkline */}
      {sparkline?.buckets && (
        <div className="aexc-sparkline-wrap">
          <div className="aexc-sparkline-label">24h P&L</div>
          <Sparkline data={sparkline.buckets} width={320} height={56} color={pnlColor} />
        </div>
      )}

      {/* Breakdown by type */}
      {summary.breakdown && summary.breakdown.length > 0 && (
        <div className="aexc-breakdown">
          <div className="aexc-breakdown-title">TRANSACTION BREAKDOWN</div>
          {summary.breakdown.map((b, i) => (
            <div key={i} className="aexc-breakdown-row">
              <span className="aexc-bd-type">{b.type}</span>
              <span className="aexc-bd-count">{b.count}</span>
              <span className="aexc-bd-total" style={{ color: txColor(b.type) }}>
                ${fmt(b.total)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Live Transaction Stream ───

function LiveTxStream({ transactions }) {
  const scrollRef = useRef(null)

  // Auto-scroll to top when new transactions arrive
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = 0
    }
  }, [transactions])

  if (!transactions || transactions.length === 0) {
    return <div className="aexc-placeholder">No transactions</div>
  }

  return (
    <div className="aexc-live-tx" ref={scrollRef}>
      <div className="aexc-live-tx-header">
        <span>LIVE TRANSACTIONS</span>
        <span className="aexc-live-count">{transactions.length}</span>
      </div>
      <div className="aexc-live-list">
        {transactions.map((tx) => (
          <div key={tx.id} className="aexc-live-item">
            <span className="aexc-live-time">{fmtTime(tx.created_at)}</span>
            <span className="aexc-live-agent">{tx.agent_id?.slice(0, 10) || '---'}</span>
            <span className="aexc-live-type" style={{ color: txColor(tx.type) }}>{tx.type}</span>
            <span className="aexc-live-provider">{tx.provider || '—'}</span>
            <span className="aexc-live-amount" style={{ color: txColor(tx.type) }}>
              {txSign(tx.type)}${fmt(tx.amount)}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

// ─── Main AEXC Feed Component ───

export default function AEXCFeed({ agentId }) {
  const [platform, setPlatform] = useState(null)
  const [liveTx, setLiveTx] = useState([])
  const [ws, setWs] = useState(null)

  // Fetch platform data
  const fetchPlatform = useCallback(async () => {
    try {
      const res = await fetch('/api/financials/platform')
      if (res.ok) setPlatform(await res.json())
    } catch (e) { /* ignore */ }
  }, [])

  // Fetch live transactions
  const fetchLiveTx = useCallback(async () => {
    try {
      const res = await fetch('/api/financials/live-tx')
      if (res.ok) {
        const data = await res.json()
        setLiveTx(data.transactions || [])
      }
    } catch (e) { /* ignore */ }
  }, [])

  useEffect(() => {
    fetchPlatform()
    fetchLiveTx()
    const iv1 = setInterval(fetchPlatform, 5000)
    const iv2 = setInterval(fetchLiveTx, 5000)
    return () => { clearInterval(iv1); clearInterval(iv2) }
  }, [fetchPlatform, fetchLiveTx])

  // WebSocket for real-time updates
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsConn = new WebSocket(`${protocol}//${window.location.host}/ws?feed=financial`)

    wsConn.onopen = () => setWs(wsConn)

    wsConn.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        // Handle financial notifications
        if (data.type === 'payment' || data.type === 'commerce' || data.type === 'payout') {
          fetchPlatform()
          fetchLiveTx()
        }
      } catch { /* ignore non-JSON */ }
    }

    wsConn.onerror = () => setWs(null)
    wsConn.onclose = () => setWs(null)

    return () => { if (wsConn) wsConn.close() }
  }, [fetchPlatform, fetchLiveTx])

  return (
    <div className="aexc-feed">
      {/* Ticker Tape */}
      <TickerTape transactions={liveTx.slice(0, 20)} />

      {/* Platform Summary */}
      <PlatformCards data={platform} />

      {/* Two-column layout: Agent Panel + Live Stream */}
      <div className="aexc-main">
        <div className="aexc-col-left">
          {agentId ? (
            <AgentPanel agentId={agentId} />
          ) : (
            <div className="aexc-placeholder">Select an agent to view financials</div>
          )}
        </div>
        <div className="aexc-col-right">
          <LiveTxStream transactions={liveTx} />
        </div>
      </div>
    </div>
  )
}

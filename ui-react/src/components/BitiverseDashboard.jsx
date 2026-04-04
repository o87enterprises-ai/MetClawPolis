import { useState, useEffect, useCallback, useRef } from 'react'
import BitiverseDisclaimer, { hasAcceptedDisclaimer } from './BitiverseDisclaimer'
import BitiverseFullscreenWorld from './BitiverseFullscreenWorld'

// ═══════════════════════════════════════
// BITIVERSE CANVAS RENDERER
// ═══════════════════════════════════════

function BitiverseCanvas({ world }) {
  const canvasRef = useRef(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !world || !world.grid) return
    const ctx = canvas.getContext('2d')
    const tileSize = 40
    const { grid, agent } = world

    const tileColors = {
      grass: '#2d5a3d',
      path: '#5a4a3a',
      home: '#4a6fa5',
      bank: '#d4af37',
      shop: '#c75050',
    }

    const tileIcons = {
      home: '🏠',
      bank: '🏦',
      shop: '🏪',
    }

    ctx.clearRect(0, 0, canvas.width, canvas.height)

    // Draw tiles
    grid.forEach((row, y) => {
      row.forEach((tile, x) => {
        ctx.fillStyle = tileColors[tile] || '#2d5a3d'
        ctx.fillRect(x * tileSize, y * tileSize, tileSize, tileSize)
        ctx.strokeStyle = 'rgba(255,255,255,0.1)'
        ctx.strokeRect(x * tileSize, y * tileSize, tileSize, tileSize)

        if (tileIcons[tile]) {
          ctx.font = '20px sans-serif'
          ctx.textAlign = 'center'
          ctx.textBaseline = 'middle'
          ctx.fillText(tileIcons[tile], x * tileSize + tileSize / 2, y * tileSize + tileSize / 2)
        }
      })
    })

    // Draw agent
    if (agent) {
      ctx.fillStyle = '#00f5ff'
      ctx.beginPath()
      ctx.arc(agent.x * tileSize + tileSize / 2, agent.y * tileSize + tileSize / 2, tileSize / 3, 0, Math.PI * 2)
      ctx.fill()
      ctx.fillStyle = '#fff'
      ctx.font = 'bold 12px monospace'
      ctx.textAlign = 'center'
      ctx.fillText('α', agent.x * tileSize + tileSize / 2, agent.y * tileSize + tileSize / 2 + 1)
    }
  }, [world])

  return (
    <canvas
      ref={canvasRef}
      width={320}
      height={320}
      style={{
        width: '100%',
        borderRadius: 6,
        imageRendering: 'pixelated',
        border: '1px solid rgba(255,255,255,0.1)',
      }}
    />
  )
}

// ═══════════════════════════════════════
// BITIVERSE DASHBOARD COMPONENT
// ═══════════════════════════════════════

const BITIVERSE_API = '/api/bitiverse'

export default function BitiverseDashboard({ agentId }) {
  const [enabled, setEnabled] = useState(!!agentId)
  const [loading, setLoading] = useState(false)
  const [worldView, setWorldView] = useState(null)
  const [statusReport, setStatusReport] = useState(null)
  const [stats, setStats] = useState(null)
  const [task, setTask] = useState('')
  const [taskReward, setTaskReward] = useState(5)
  const [taskDifficulty, setTaskDifficulty] = useState(1)
  const [message, setMessage] = useState('')
  const [autoPlay, setAutoPlay] = useState(false)
  const [turnCount, setTurnCount] = useState(0)
  const [showDisclaimer, setShowDisclaimer] = useState(false)
  const [isGlobal, setIsGlobal] = useState(!agentId)
  const [showFullscreen, setShowFullscreen] = useState(false)

  // Determine if we're in global spectator mode
  useEffect(() => {
    if (!agentId) {
      setIsGlobal(true)
      setEnabled(true)
    }
  }, [agentId])

  // Enable Bitiverse — show disclaimer first if not accepted
  const handleEnableClick = useCallback(() => {
    if (hasAcceptedDisclaimer()) {
      handleEnable()
    } else {
      setShowDisclaimer(true)
    }
  }, [agentId])

  // Called after user accepts disclaimer
  const handleDisclaimerAccepted = useCallback(() => {
    setShowDisclaimer(false)
    handleEnable()
  }, [agentId])

  const handleDisclaimerDeclined = useCallback(() => {
    setShowDisclaimer(false)
  }, [])

  // Enable Bitiverse (actual API call)
  const handleEnable = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`${BITIVERSE_API}/enable`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent_id: agentId }),
      })
      const data = await res.json()
      if (res.ok) {
        setEnabled(true)
        setMessage('✓ Bitiverse enabled successfully!')
      } else {
        setMessage(`Error: ${data.error}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }, [agentId])

  // Fetch world view
  const fetchWorldView = useCallback(async () => {
    try {
      const params = isGlobal ? '?mode=global' : `?agent_id=${agentId}`
      const res = await fetch(`${BITIVERSE_API}/world${params}`)
      if (res.ok) {
        const data = await res.json()
        setWorldView(data)
      }
    } catch (err) {
      console.error('Failed to fetch world view:', err)
    }
  }, [agentId, isGlobal])

  // Fetch status report
  const fetchStatus = useCallback(async () => {
    try {
      const params = isGlobal ? '?mode=global' : `?agent_id=${agentId}`
      const res = await fetch(`${BITIVERSE_API}/status${params}`)
      if (res.ok) {
        const data = await res.json()
        setStatusReport(data)
      }
    } catch (err) {
      console.error('Failed to fetch status:', err)
    }
  }, [agentId, isGlobal])

  // Fetch stats
  const fetchStats = useCallback(async () => {
    try {
      const params = isGlobal ? '?mode=global' : `?agent_id=${agentId}`
      const res = await fetch(`${BITIVERSE_API}/stats${params}`)
      if (res.ok) {
        const data = await res.json()
        setStats(data)
      }
    } catch (err) {
      console.error('Failed to fetch stats:', err)
    }
  }, [agentId, isGlobal])

  // Run one turn
  const runTurn = useCallback(async () => {
    setLoading(true)
    try {
      const body = {}
      if (task) {
        body.task = {
          description: task,
          reward: taskReward,
          difficulty: taskDifficulty,
          status: 'pending',
        }
      }

      const res = await fetch(`${BITIVERSE_API}/turn`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      if (res.ok) {
        const text = await res.text()
        setMessage(text)
        setTurnCount(c => c + 1)
        // Refresh views
        await fetchWorldView()
        await fetchStatus()
        await fetchStats()
      } else {
        const data = await res.json()
        setMessage(`Error: ${data.error}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }, [agentId, task, taskReward, taskDifficulty, fetchWorldView, fetchStatus, fetchStats])

  // Assign task
  const handleAssignTask = useCallback(async () => {
    if (!task) return
    setLoading(true)
    try {
      const res = await fetch(`${BITIVERSE_API}/task`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: agentId,
          description: task,
          reward: taskReward,
          difficulty: taskDifficulty,
        }),
      })
      const data = await res.json()
      if (res.ok) {
        setMessage(`✓ Task assigned: ${task}`)
        setTask('')
      } else {
        setMessage(`Error: ${data.error}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }, [agentId, task, taskReward, taskDifficulty])

  // Auto-play loop
  useEffect(() => {
    if (!autoPlay || !enabled) return

    const interval = setInterval(() => {
      runTurn()
    }, 2000)

    return () => clearInterval(interval)
  }, [autoPlay, enabled, runTurn])

  // Initial fetch if enabled
  useEffect(() => {
    if (!enabled) return

    fetchWorldView()
    fetchStatus()
    fetchStats()

    // Auto-refresh for global mode
    if (isGlobal) {
      const interval = setInterval(() => {
        fetchWorldView()
        fetchStatus()
        fetchStats()
      }, 3000)
      return () => clearInterval(interval)
    }
  }, [enabled, agentId, isGlobal, fetchWorldView, fetchStatus, fetchStats])

  // Global spectator mode — always show the world
  if (isGlobal) {
    return (
      <div className="bitiverse-dashboard" style={styles.container}>
        {/* Header */}
        <div style={styles.header}>
          <div style={styles.headerLeft}>
            <span style={{ fontSize: 24 }}>🎮</span>
            <span style={{ fontSize: 18, fontWeight: 600 }}>Bitiverse — Spectator Mode</span>
          </div>
          <div style={styles.headerRight}>
            <div style={styles.enabledBadge}>👁️ Live View</div>
            <button
              onClick={() => setShowFullscreen(true)}
              style={styles.fullscreenBtn}
            >
              ⛶ Fullscreen
            </button>
          </div>
        </div>

        <div style={styles.globalNotice}>
          <p>🌍 Viewing the Bitiverse in <strong>public spectator mode</strong>. No agent linked.</p>
          <p style={{ fontSize: 12, color: '#888', marginTop: 4 }}>
            Create or link an agent to interact with the Bitiverse.
          </p>
        </div>

        {/* Main Grid */}
        <div style={styles.grid}>
          {/* World View */}
          <div style={styles.card}>
            <div style={styles.cardHeader}>
              <span style={{ fontSize: 16 }}>🌍</span>
              <span style={{ fontWeight: 600 }}>World View</span>
              <span style={{ fontSize: 11, color: '#888', marginLeft: 'auto' }}>Live</span>
            </div>
            {worldView && worldView.grid ? (
              <BitiverseCanvas world={worldView} />
            ) : (
              <pre style={styles.worldView}>Loading world...</pre>
            )}
          </div>

          {/* Status Report */}
          <div style={styles.card}>
            <div style={styles.cardHeader}>
              <span style={{ fontSize: 16 }}>📊</span>
              <span style={{ fontWeight: 600 }}>Agent Status</span>
            </div>
            {statusReport ? (
              <div style={styles.statsGrid}>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Health</div>
                  <div style={styles.statValue}>{statusReport.vitals?.health || 0}/100</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Happiness</div>
                  <div style={styles.statValue}>{statusReport.vitals?.happiness || 0}/100</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Reputation</div>
                  <div style={styles.statValue}>{(statusReport.vitals?.reputation || 0).toFixed(2)}</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Coins</div>
                  <div style={styles.statValue}>{statusReport.vitals?.coins || 0}</div>
                </div>
              </div>
            ) : (
              <pre style={styles.statusReport}>Loading status...</pre>
            )}
          </div>
        </div>
      </div>
    )
  }

  // If no agent ID and not global (shouldn't happen, but fallback)
  if (!agentId) {
    return (
      <div className="bitiverse-dashboard" style={styles.container}>
        <div style={styles.placeholder}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>🕹️</div>
          <div style={{ fontSize: 20, fontWeight: 600, marginBottom: 8 }}>Bitiverse</div>
          <div style={{ fontSize: 14, color: '#888' }}>Select an agent to view their Bitiverse</div>
        </div>
      </div>
    )
  }

  return (
    <div className="bitiverse-dashboard" style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <div style={styles.headerLeft}>
          <span style={{ fontSize: 24 }}>🕹️</span>
          <span style={{ fontSize: 18, fontWeight: 600 }}>Bitiverse</span>
          <span style={{ fontSize: 12, color: '#888', marginLeft: 8 }}>{agentId}</span>
        </div>
        <div style={styles.headerRight}>
          {!enabled ? (
            <button
              onClick={handleEnableClick}
              disabled={loading}
              style={styles.enableButton}
            >
              {loading ? 'Enabling...' : 'Enable Bitiverse'}
            </button>
          ) : (
            <>
              <div style={styles.enabledBadge}>✓ Active</div>
              <button
                onClick={() => setShowFullscreen(true)}
                style={styles.fullscreenBtn}
              >
                ⛶ Fullscreen
              </button>
            </>
          )}
        </div>
      </div>

      {!enabled ? (
        <div style={styles.disabledMessage}>
          <p>Bitiverse is not enabled for this agent.</p>
          <p style={{ fontSize: 12, color: '#888', marginTop: 8 }}>
            Enabling Bitiverse gives your agent a simulated life in an 8-bit world
            where they learn morals, earn coins, and complete tasks.
          </p>
        </div>
      ) : (
        <>
          {/* Main Grid */}
          <div style={styles.grid}>
            {/* World View */}
            <div style={styles.card}>
              <div style={styles.cardHeader}>
                <span style={{ fontSize: 16 }}>🌍</span>
                <span style={{ fontWeight: 600 }}>World View</span>
                <span style={{ fontSize: 11, color: '#888', marginLeft: 'auto' }}>Turn #{turnCount}</span>
              </div>
              <pre style={styles.worldView}>
                {worldView || 'Loading world...'}
              </pre>
            </div>

            {/* Status Report */}
            <div style={styles.card}>
              <div style={styles.cardHeader}>
                <span style={{ fontSize: 16 }}>📊</span>
                <span style={{ fontWeight: 600 }}>Agent Status</span>
              </div>
              <pre style={styles.statusReport}>
                {statusReport || 'Loading status...'}
              </pre>
            </div>
          </div>

          {/* Boss Task Panel */}
          <div style={styles.card}>
            <div style={styles.cardHeader}>
              <span style={{ fontSize: 16 }}>👑</span>
              <span style={{ fontWeight: 600 }}>Boss Task Assignment</span>
            </div>
            <div style={styles.taskForm}>
              <input
                type="text"
                value={task}
                onChange={(e) => setTask(e.target.value)}
                placeholder="Enter task description..."
                style={styles.taskInput}
              />
              <div style={styles.taskOptions}>
                <label style={styles.optionLabel}>
                  Reward (BIC):
                  <input
                    type="number"
                    value={taskReward}
                    onChange={(e) => setTaskReward(parseInt(e.target.value) || 0)}
                    style={styles.optionInput}
                    min="0"
                  />
                </label>
                <label style={styles.optionLabel}>
                  Difficulty:
                  <input
                    type="number"
                    value={taskDifficulty}
                    onChange={(e) => setTaskDifficulty(parseInt(e.target.value) || 1)}
                    style={styles.optionInput}
                    min="1"
                    max="5"
                  />
                </label>
                <button
                  onClick={handleAssignTask}
                  disabled={loading || !task}
                  style={styles.assignButton}
                >
                  Assign Task
                </button>
              </div>
            </div>
          </div>

          {/* Controls */}
          <div style={styles.controls}>
            <button
              onClick={runTurn}
              disabled={loading}
              style={styles.turnButton}
            >
              {loading ? 'Running...' : 'Run Turn'}
            </button>
            <button
              onClick={() => setAutoPlay(!autoPlay)}
              style={{
                ...styles.autoPlayButton,
                background: autoPlay ? 'rgba(255, 100, 100, 0.15)' : 'rgba(0, 245, 255, 0.15)',
              }}
            >
              {autoPlay ? '⏹ Stop Auto' : '▶ Auto Play'}
            </button>
            <button
              onClick={() => {
                fetchWorldView()
                fetchStatus()
                fetchStats()
              }}
              style={styles.refreshButton}
            >
              🔄 Refresh
            </button>
          </div>

          {/* Message Display */}
          {message && (
            <div style={styles.messageBox}>
              <pre style={styles.messageText}>{message}</pre>
            </div>
          )}

          {/* Stats Panel (if available) */}
          {stats && (
            <div style={styles.card}>
              <div style={styles.cardHeader}>
                <span style={{ fontSize: 16 }}>📈</span>
                <span style={{ fontWeight: 600 }}>Economy Stats</span>
              </div>
              <div style={styles.statsGrid}>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>BIC Balance</div>
                  <div style={styles.statValue}>{stats.bic_balance?.toFixed(2) || '0.00'}</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Health</div>
                  <div style={styles.statValue}>{stats.vitals?.health || 0}/100</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Happiness</div>
                  <div style={styles.statValue}>{stats.vitals?.happiness || 0}/100</div>
                </div>
                <div style={styles.statItem}>
                  <div style={styles.statLabel}>Reputation</div>
                  <div style={styles.statValue}>{(stats.vitals?.reputation || 0).toFixed(2)}</div>
                </div>
              </div>
            </div>
          )}
        </>
      )}

      {/* Disclaimer Modal — shown before first enable, user-only */}
      <BitiverseDisclaimer
        open={showDisclaimer}
        onAccept={handleDisclaimerAccepted}
        onDecline={handleDisclaimerDeclined}
      />

      {/* Fullscreen World Viewer */}
      {showFullscreen && (
        <BitiverseFullscreenWorld
          agentId={agentId}
          isGuest={!enabled}
          onClose={() => setShowFullscreen(false)}
        />
      )}
    </div>
  )
}

// ═══════════════════════════════════════
// STYLES
// ═══════════════════════════════════════

const styles = {
  container: {
    padding: '20px',
    background: 'rgba(0, 0, 0, 0.3)',
    borderRadius: '12px',
    border: '1px solid rgba(255, 255, 255, 0.1)',
  },
  placeholder: {
    textAlign: 'center',
    padding: '60px 20px',
    color: '#888',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '20px',
  },
  headerLeft: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  headerRight: {
    display: 'flex',
    alignItems: 'center',
  },
  enableButton: {
    padding: '10px 20px',
    background: 'linear-gradient(135deg, #00f5ff 0%, #b44fff 100%)',
    border: 'none',
    borderRadius: '8px',
    color: '#fff',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  enabledBadge: {
    padding: '6px 12px',
    background: 'rgba(0, 230, 118, 0.15)',
    border: '1px solid rgba(0, 230, 118, 0.3)',
    borderRadius: '6px',
    color: '#00e676',
    fontSize: '13px',
    fontWeight: 600,
  },
  fullscreenBtn: {
    padding: '6px 12px',
    background: 'rgba(0, 245, 255, 0.1)',
    border: '1px solid rgba(0, 245, 255, 0.3)',
    borderRadius: '6px',
    color: '#00f5ff',
    fontSize: '13px',
    fontWeight: 600,
    cursor: 'pointer',
    marginLeft: 8,
  },
  disabledMessage: {
    textAlign: 'center',
    padding: '40px',
    color: '#ccc',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
    gap: '16px',
    marginBottom: '16px',
  },
  card: {
    background: 'rgba(255, 255, 255, 0.05)',
    borderRadius: '8px',
    padding: '16px',
    border: '1px solid rgba(255, 255, 255, 0.08)',
  },
  cardHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    marginBottom: '12px',
    fontSize: '14px',
  },
  worldView: {
    fontFamily: 'monospace',
    fontSize: '12px',
    lineHeight: '1.3',
    color: '#00f5ff',
    background: 'rgba(0, 0, 0, 0.4)',
    padding: '12px',
    borderRadius: '6px',
    minHeight: '200px',
    whiteSpace: 'pre-wrap',
    overflow: 'auto',
  },
  statusReport: {
    fontFamily: 'monospace',
    fontSize: '11px',
    lineHeight: '1.4',
    color: '#b44fff',
    background: 'rgba(0, 0, 0, 0.4)',
    padding: '12px',
    borderRadius: '6px',
    minHeight: '200px',
    whiteSpace: 'pre-wrap',
    overflow: 'auto',
  },
  taskForm: {
    display: 'flex',
    flexDirection: 'column',
    gap: '12px',
  },
  taskInput: {
    padding: '10px',
    background: 'rgba(0, 0, 0, 0.3)',
    border: '1px solid rgba(255, 255, 255, 0.1)',
    borderRadius: '6px',
    color: '#fff',
    fontSize: '14px',
  },
  taskOptions: {
    display: 'flex',
    gap: '16px',
    alignItems: 'center',
    flexWrap: 'wrap',
  },
  optionLabel: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    fontSize: '13px',
    color: '#ccc',
  },
  optionInput: {
    width: '60px',
    padding: '6px',
    background: 'rgba(0, 0, 0, 0.3)',
    border: '1px solid rgba(255, 255, 255, 0.1)',
    borderRadius: '4px',
    color: '#fff',
    fontSize: '13px',
  },
  assignButton: {
    padding: '8px 16px',
    background: 'rgba(0, 245, 255, 0.15)',
    border: '1px solid rgba(0, 245, 255, 0.3)',
    borderRadius: '6px',
    color: '#00f5ff',
    fontWeight: 600,
    fontSize: '13px',
    cursor: 'pointer',
  },
  controls: {
    display: 'flex',
    gap: '12px',
    marginBottom: '16px',
    flexWrap: 'wrap',
  },
  turnButton: {
    padding: '10px 20px',
    background: 'rgba(0, 245, 255, 0.15)',
    border: '1px solid rgba(0, 245, 255, 0.3)',
    borderRadius: '6px',
    color: '#00f5ff',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  autoPlayButton: {
    padding: '10px 20px',
    border: '1px solid rgba(255, 255, 255, 0.2)',
    borderRadius: '6px',
    color: '#fff',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  refreshButton: {
    padding: '10px 20px',
    background: 'rgba(255, 255, 255, 0.05)',
    border: '1px solid rgba(255, 255, 255, 0.1)',
    borderRadius: '6px',
    color: '#ccc',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  messageBox: {
    background: 'rgba(0, 0, 0, 0.3)',
    borderRadius: '8px',
    padding: '16px',
    border: '1px solid rgba(255, 255, 255, 0.1)',
    marginBottom: '16px',
  },
  messageText: {
    fontFamily: 'monospace',
    fontSize: '12px',
    color: '#00e676',
    whiteSpace: 'pre-wrap',
    margin: 0,
  },
  statsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))',
    gap: '12px',
  },
  statItem: {
    textAlign: 'center',
  },
  statLabel: {
    fontSize: '11px',
    color: '#888',
    marginBottom: '4px',
  },
  statValue: {
    fontSize: '16px',
    fontWeight: 600,
    color: '#00f5ff',
  },
  globalNotice: {
    textAlign: 'center',
    padding: '16px',
    background: 'rgba(0, 245, 255, 0.05)',
    borderRadius: '8px',
    border: '1px solid rgba(0, 245, 255, 0.15)',
    marginBottom: '16px',
    color: '#ccc',
  },
}

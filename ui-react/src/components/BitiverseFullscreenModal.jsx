import { useState, useEffect, useRef, useCallback } from 'react'
import { renderSprite, renderTile, CHARACTERS, BUILDINGS } from '../bitiverse/sprites'

const TILE_SIZE = 32 // pixels per tile in the rendered world
const VIEW_W = 20 // tiles visible horizontally
const VIEW_H = 15 // tiles visible vertically
const TICK_RATE = 1000 // ms between auto-play turns

export default function BitiverseFullscreenModal({ agentId, onClose }) {
  const canvasRef = useRef(null)
  const [world, setWorld] = useState(null)
  const [status, setStatus] = useState(null)
  const [messages, setMessages] = useState([])
  const [taskInput, setTaskInput] = useState('')
  const [autoPlay, setAutoPlay] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [frame, setFrame] = useState(0)
  const [modelWarning, setModelWarning] = useState(false)
  const msgEndRef = useRef(null)

  // Fetch world state
  const fetchWorld = useCallback(async () => {
    try {
      const res = await fetch(`/api/bitiverse/world?agent_id=${agentId}`)
      if (!res.ok) throw new Error('Failed to fetch world')
      const data = await res.json()
      setWorld(data)
      setLoading(false)
    } catch (err) {
      setError(err.message)
    }
  }, [agentId])

  // Fetch agent status
  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch(`/api/bitiverse/status?agent_id=${agentId}`)
      if (!res.ok) throw new Error('Failed to fetch status')
      const data = await res.json()
      setStatus(data)
    } catch (err) {
      // Non-critical, don't set error
    }
  }, [agentId])

  // Run a turn
  const runTurn = useCallback(async (task) => {
    try {
      // Get signature from backend
      let signature = ''
      try {
        const signRes = await fetch('/api/auth/sign', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id: agentId, message: `bitiverse_turn_${Date.now()}` })
        })
        if (signRes.ok) {
          const signData = await signRes.json()
          signature = signData.signature
        }
      } catch (e) {
        console.warn('Signing unavailable, using demo signature')
        signature = 'demo-signature'
      }

      const body = task ? { task: { description: task, reward: 5, difficulty: 1 } } : {}
      const res = await fetch(`/api/bitiverse/turn`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Agent-ID': agentId,
          'X-Signature': signature,
        },
        body: JSON.stringify(body),
      })
      const data = await res.json()
      if (data.result?.message) {
        setMessages(prev => [...prev.slice(-50), data.result.message])
      }
      fetchWorld()
      fetchStatus()
      setFrame(f => f + 1)
    } catch (err) {
      setMessages(prev => [...prev.slice(-50), `Error: ${err.message}`])
    }
  }, [agentId, fetchWorld, fetchStatus])

  // Assign task from Boss
  const assignTask = async () => {
    if (!taskInput.trim()) return
    try {
      // Get signature
      let signature = ''
      try {
        const signRes = await fetch('/api/auth/sign', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ agent_id: agentId, message: `bitiverse_task_${Date.now()}` })
        })
        if (signRes.ok) {
          const signData = await signRes.json()
          signature = signData.signature
        }
      } catch (e) {
        console.warn('Signing unavailable, using demo signature')
        signature = 'demo-signature'
      }

      const res = await fetch('/api/bitiverse/task', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Agent-ID': agentId,
          'X-Signature': signature,
        },
        body: JSON.stringify({ description: taskInput, reward: 5, difficulty: 1 }),
      })
      const data = await res.json()
      setMessages(prev => [...prev.slice(-50), `Boss: ${taskInput}`])
      setTaskInput('')
      fetchStatus()
    } catch (err) {
      setMessages(prev => [...prev.slice(-50), `Task error: ${err.message}`])
    }
  }

  // Initial load
  useEffect(() => {
    fetchWorld()
    fetchStatus()
    setMessages(['Welcome to the Bitiverse!', 'Your agent is spawning...'])
  }, [fetchWorld, fetchStatus])

  // Auto-play
  useEffect(() => {
    if (!autoPlay) return
    const interval = setInterval(() => runTurn(), TICK_RATE)
    return () => clearInterval(interval)
  }, [autoPlay, runTurn])

  // Keyboard controls
  useEffect(() => {
    const handleKey = (e) => {
      if (document.activeElement.tagName === 'INPUT' || document.activeElement.tagName === 'TEXTAREA') return
      const moves = {
        ArrowUp: 'north', ArrowDown: 'south', ArrowLeft: 'west', ArrowRight: 'east',
        w: 'north', s: 'south', a: 'west', d: 'east',
      }
      if (moves[e.key]) {
        e.preventDefault()
        runTurn(`move_${moves[e.key]}`)
      }
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [runTurn])

  // Render canvas
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !world) return
    const ctx = canvas.getContext('2d')
    const W = canvas.width = VIEW_W * TILE_SIZE
    const H = canvas.height = VIEW_H * TILE_SIZE

    // Clear
    ctx.fillStyle = '#1a2e15'
    ctx.fillRect(0, 0, W, H)

    // Draw tiles from world grid
    const grid = world.grid || []
    for (let y = 0; y < Math.min(grid.length, VIEW_H); y++) {
      for (let x = 0; x < Math.min((grid[y] || []).length, VIEW_W); x++) {
        const tile = grid[y][x]
        renderTile(ctx, tile || 'grass', x * TILE_SIZE, y * TILE_SIZE, TILE_SIZE)
      }
    }

    // Draw buildings
    const buildings = world.buildings || []
    buildings.forEach(b => {
      const bld = BUILDINGS[b.type]
      if (bld) {
        renderSprite(ctx, bld, b.x * TILE_SIZE, b.y * TILE_SIZE, TILE_SIZE / 16)
      }
    })

    // Draw NPCs
    const npcs = world.npcs || []
    npcs.forEach(npc => {
      const charData = CHARACTERS[npc.type] || CHARACTERS.police
      renderSprite(ctx, charData, npc.x * TILE_SIZE, npc.y * TILE_SIZE, TILE_SIZE / 16)
    })

    // Draw agent
    if (world.agent_x !== undefined) {
      const agentFrame = frame % 3
      renderSprite(ctx, CHARACTERS.agent, world.agent_x * TILE_SIZE, world.agent_y * TILE_SIZE, TILE_SIZE / 16, agentFrame)

      // Agent name label
      ctx.fillStyle = '#00ff88'
      ctx.font = 'bold 10px monospace'
      ctx.textAlign = 'center'
      ctx.fillText(status?.agent_name || 'Agent', world.agent_x * TILE_SIZE + TILE_SIZE / 2, world.agent_y * TILE_SIZE - 4)
    }

  }, [world, frame, status])

  // Scroll messages to bottom
  useEffect(() => {
    msgEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  return (
    <div className="bitiverse-fullscreen-modal" style={{
      position: 'fixed', inset: 0, zIndex: 9999,
      background: '#0a0e1a',
      display: 'flex', flexDirection: 'column',
    }}>
      {/* Header bar */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '8px 16px', background: '#0C1020', borderBottom: '1px solid #00E6FF22',
        flexShrink: 0,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 18 }}>🎮</span>
          <span style={{ fontWeight: 700, color: '#00E6FF', fontFamily: 'monospace' }}>BITIVERSE</span>
          {status && <span style={{ fontSize: 11, color: '#667' }}>Day {status.day} — {status.hour}:00</span>}
        </div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          {modelWarning && (
            <span style={{ fontSize: 9, color: '#ff8800', background: '#ff880022', padding: '2px 8px', borderRadius: 4 }}>
              ⚠️ Model changes may cause delays
            </span>
          )}
          <button onClick={() => setModelWarning(!modelWarning)} style={{
            fontSize: 9, padding: '2px 8px', borderRadius: 4,
            background: 'rgba(255,255,255,0.05)', border: '1px solid #ffffff22', color: '#889', cursor: 'pointer',
          }}>⚙️ Model Warning</button>
          <button onClick={onClose} style={{
            background: 'none', border: 'none', color: '#666', fontSize: 20, cursor: 'pointer', lineHeight: 1,
          }}>×</button>
        </div>
      </div>

      {loading ? (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#667' }}>
          Loading Bitiverse...
        </div>
      ) : error ? (
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#ff4444' }}>
          {error}
        </div>
      ) : (
        <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {/* Canvas viewport */}
          <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#050810' }}>
            <canvas ref={canvasRef} style={{ imageRendering: 'pixelated', maxWidth: '100%', maxHeight: '100%' }} />
          </div>

          {/* Right sidebar */}
          <div style={{
            width: 280, borderLeft: '1px solid #ffffff11', background: '#0C1020',
            display: 'flex', flexDirection: 'column', overflow: 'hidden',
          }}>
            {/* Agent Vitals */}
            {status && (
              <div style={{ padding: 12, borderBottom: '1px solid #ffffff11' }}>
                <h3 style={{ fontSize: 11, color: '#00E6FF', marginBottom: 8, letterSpacing: 1 }}>AGENT VITALS</h3>
                {[
                  { label: 'Health', val: status.health, max: 100, color: '#ff4444' },
                  { label: 'Happiness', val: status.happiness, max: 100, color: '#ffcc00' },
                  { label: 'Energy', val: status.energy, max: 100, color: '#44aaff' },
                  { label: 'Stress', val: status.stress, max: 100, color: '#b44fff' },
                ].map(v => (
                  <div key={v.label} style={{ marginBottom: 6 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 9, color: '#667', marginBottom: 2 }}>
                      <span>{v.label}</span><span>{v.val}/{v.max}</span>
                    </div>
                    <div style={{ height: 4, background: '#ffffff11', borderRadius: 2, overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${(v.val / v.max) * 100}%`, background: v.color, borderRadius: 2 }} />
                    </div>
                  </div>
                ))}
                <div style={{ fontSize: 11, color: '#ffcc00', marginTop: 8 }}>
                  🪙 BIC: {status.bic_balance || 0}
                </div>
              </div>
            )}

            {/* Boss Task Panel */}
            <div style={{ padding: 12, borderBottom: '1px solid #ffffff11' }}>
              <h3 style={{ fontSize: 11, color: '#00E6FF', marginBottom: 8, letterSpacing: 1 }}>BOSS TASK</h3>
              <div style={{ display: 'flex', gap: 4 }}>
                <input
                  value={taskInput}
                  onChange={e => setTaskInput(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && assignTask()}
                  placeholder="Assign a task..."
                  style={{
                    flex: 1, padding: '6px 8px', borderRadius: 4,
                    background: 'rgba(255,255,255,0.05)', border: '1px solid #ffffff22',
                    color: '#e8eaf0', fontSize: 11, outline: 'none',
                  }}
                />
                <button onClick={assignTask} style={{
                  padding: '6px 12px', borderRadius: 4, border: 'none',
                  background: '#00E6FF', color: '#050810', fontWeight: 700, fontSize: 11, cursor: 'pointer',
                }}>Send</button>
              </div>
            </div>

            {/* Controls */}
            <div style={{ padding: 12, borderBottom: '1px solid #ffffff11' }}>
              <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                <button onClick={() => runTurn()} style={{
                  padding: '6px 12px', borderRadius: 4, border: '1px solid #ffffff22',
                  background: 'rgba(255,255,255,0.05)', color: '#e8eaf0', fontSize: 10, cursor: 'pointer',
                }}>▶ Run Turn</button>
                <button onClick={() => setAutoPlay(!autoPlay)} style={{
                  padding: '6px 12px', borderRadius: 4, border: 'none',
                  background: autoPlay ? '#ff4444' : '#00E6FF', color: autoPlay ? '#fff' : '#050810',
                  fontWeight: 700, fontSize: 10, cursor: 'pointer',
                }}>{autoPlay ? '⏸ Stop' : '⏵ Auto'}</button>
                <button onClick={fetchWorld} style={{
                  padding: '6px 12px', borderRadius: 4, border: '1px solid #ffffff22',
                  background: 'rgba(255,255,255,0.05)', color: '#e8eaf0', fontSize: 10, cursor: 'pointer',
                }}>↻ Refresh</button>
              </div>
              {/* D-pad */}
              <div style={{ marginTop: 8, display: 'grid', gridTemplateColumns: 'repeat(3, 32px)', gap: 2, justifyContent: 'center' }}>
                <div />
                <button onClick={() => runTurn('move_north')} style={dpadBtn}>↑</button>
                <div />
                <button onClick={() => runTurn('move_west')} style={dpadBtn}>←</button>
                <div style={{ width: 32, height: 32 }} />
                <button onClick={() => runTurn('move_east')} style={dpadBtn}>→</button>
                <div />
                <button onClick={() => runTurn('move_south')} style={dpadBtn}>↓</button>
                <div />
              </div>
            </div>

            {/* Message Log */}
            <div style={{ flex: 1, overflow: 'auto', padding: 8 }}>
              <h3 style={{ fontSize: 11, color: '#00E6FF', marginBottom: 6, letterSpacing: 1 }}>LOG</h3>
              {messages.map((m, i) => (
                <div key={i} style={{ fontSize: 9, color: '#889', marginBottom: 3, fontFamily: 'monospace', lineHeight: 1.4 }}>
                  {m}
                </div>
              ))}
              <div ref={msgEndRef} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const dpadBtn = {
  width: 32, height: 32, borderRadius: 4, border: '1px solid #ffffff22',
  background: 'rgba(255,255,255,0.05)', color: '#e8eaf0', fontSize: 14,
  cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
}

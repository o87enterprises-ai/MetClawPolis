// ═══════════════════════════════════════
// BITIVERSE FULLSCREEN WORLD VIEWER
// ═══════════════════════════════════════
// Full 16-bit JRPG-style canvas world with:
// - WASD/Arrow key navigation
// - Camera with smooth scrolling
// - NPC patrol system
// - Building enter/exit
// - Minimap, stats bar, live feed HUD
// - D-pad touch controls for mobile
// ═══════════════════════════════════════

import { useState, useEffect, useRef, useCallback } from 'react'
import {
  drawTile as drawTileSprite,
  drawChar,
  drawBuild,
  CHARACTERS,
  BUILDINGS,
  BITSBURG_MAP,
  BITSBURG_BUILDINGS,
  MAP_WIDTH,
  MAP_HEIGHT,
  TILE_SIZE,
  T,
} from '../bitiverse/sprites'
import BusinessBuildingInterior from './BusinessBuildingInterior'

// ─── NPC PATROL DEFINITIONS ───
const NPC_PATROLS = [
  {
    id: 'guard_north',
    type: 'guard',
    config: CHARACTERS.guard,
    points: [
      { x: 28, y: 12 }, { x: 36, y: 12 },
      { x: 36, y: 14 }, { x: 28, y: 14 },
    ],
  },
  {
    id: 'guard_south',
    type: 'guard',
    config: CHARACTERS.guard,
    points: [
      { x: 28, y: 35 }, { x: 36, y: 35 },
      { x: 36, y: 37 }, { x: 28, y: 37 },
    ],
  },
  {
    id: 'guard_east',
    type: 'guard',
    config: CHARACTERS.guard,
    points: [
      { x: 48, y: 20 }, { x: 52, y: 20 },
      { x: 52, y: 30 }, { x: 48, y: 30 },
    ],
  },
  {
    id: 'mayor',
    type: 'mayor',
    config: CHARACTERS.mayor,
    points: [
      { x: 30, y: 17 }, { x: 32, y: 17 },
      { x: 32, y: 19 }, { x: 30, y: 19 },
    ],
  },
  {
    id: 'banker',
    type: 'banker',
    config: CHARACTERS.banker,
    points: [
      { x: 42, y: 28 }, { x: 44, y: 28 },
      { x: 44, y: 27 }, { x: 42, y: 27 },
    ],
  },
  {
    id: 'teacher',
    type: 'teacher',
    config: CHARACTERS.teacher,
    points: [
      { x: 44, y: 28 }, { x: 46, y: 28 },
      { x: 46, y: 27 }, { x: 44, y: 27 },
    ],
  },
  {
    id: 'shopkeeper',
    type: 'shopkeeper',
    config: CHARACTERS.shopkeeper,
    points: [
      { x: 8, y: 28 }, { x: 10, y: 28 },
    ],
  },
  {
    id: 'farmer',
    type: 'farmer',
    config: CHARACTERS.farmer,
    points: [
      { x: 4, y: 39 }, { x: 6, y: 39 },
      { x: 6, y: 40 }, { x: 4, y: 40 },
    ],
  },
]

// ─── LIVE FEED MESSAGES ───
const LIVE_MESSAGES = [
  '🏠 Agent-α entered the Residential Zone',
  '🏦 Banker processed a deposit of 50 BIC',
  '📚 Professor opened the Innernet Library',
  ' Shopkeeper restocked the General Store',
  '👮 Guard patrol detected a rule violation',
  '🎭 Pixel Pavilion is hosting a concert',
  '🌾 Farmer harvested crops in the Farm Zone',
  '🏛️ Mayor Pixel announced new city regulations',
  '💰 Agent-β earned 25 BIC from trading',
  '📋 A collaboration agreement was signed',
]

// ─── FULLSCREEN BITIVERSE COMPONENT ───
export default function BitiverseFullscreenWorld({ agentId, onClose, isGuest = true }) {
  const canvasRef = useRef(null)
  const minimapRef = useRef(null)
  const containerRef = useRef(null)
  const animRef = useRef(null)
  const keysRef = useRef({})
  const lastTimeRef = useRef(0)
  const feedRef = useRef(null)

  // Camera & player state (refs for smooth animation)
  const stateRef = useRef({
    player: { x: 31, y: 23, dir: 0, frame: 0, moving: false },
    camera: { x: 0, y: 0 },
    npcs: NPC_PATROLS.map(n => ({
      ...n,
      x: n.points[0].x,
      y: n.points[0].y,
      pointIdx: 0,
      dir: 0,
      frame: 0,
    })),
    messages: [...LIVE_MESSAGES.slice(0, 5)],
    msgIdx: 5,
    stats: { energy: 100, stress: 0, level: 1, coins: 142 },
    currentZone: 'center',
    feedScroll: 0,
  })

  const [hudState, setHudState] = useState({
    zone: 'Center Plaza',
    stats: { energy: 100, stress: 0, level: 1, coins: 142 },
    messages: LIVE_MESSAGES.slice(0, 5),
  })
  const [insideBuilding, setInsideBuilding] = useState(null)

  // Keyboard input
  useEffect(() => {
    const onDown = (e) => {
      keysRef.current[e.key] = true
      if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', ' '].includes(e.key)) {
        e.preventDefault()
      }
      if (e.key === 'Escape') {
        if (insideBuilding) {
          setInsideBuilding(null)
        } else if (onClose) {
          onClose()
        }
      }
      if (e.key === 'f' || e.key === 'F') toggleFullscreen()
      if (e.key === 'e' || e.key === 'E') {
        // Enter/Exit building
        if (insideBuilding) {
          setInsideBuilding(null)
        } else {
          // Check if player is near business building
          const state = stateRef.current
          const px = Math.floor(state.player.x)
          const py = Math.floor(state.player.y)
          for (const b of BITSBURG_BUILDINGS) {
            if (b.type === 'business' && px >= b.x && px < b.x + b.w && py >= b.y && py < b.y + b.h) {
              setInsideBuilding(b)
              break
            }
          }
        }
      }
    }
    const onUp = (e) => { keysRef.current[e.key] = false }
    window.addEventListener('keydown', onDown)
    window.addEventListener('keyup', onUp)
    return () => {
      window.removeEventListener('keydown', onDown)
      window.removeEventListener('keyup', onUp)
    }
  }, [insideBuilding, onClose])

  // Toggle fullscreen
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen?.()
    } else {
      document.exitFullscreen?.()
    }
  }, [])

  // Animation loop
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    const state = stateRef.current

    // Responsive canvas sizing
    const resize = () => {
      const parent = canvas.parentElement
      if (parent) {
        canvas.width = parent.clientWidth
        canvas.height = parent.clientHeight
      }
    }
    resize()
    window.addEventListener('resize', resize)

    // Viewport constants
    const VIEW_TILES_X = 24
    const VIEW_TILES_Y = 18
    const SCALE = 3 // 48px per tile (3x scale for high-res)
    const CHAR_SCALE = 3 // 16×24 → 48×72px characters

    // Zone labels
    const ZONE_LABELS = {
      residential: 'Residential Zone',
      business: 'Business Zone',
      recreation: 'Recreation & Parks',
      banking: 'Banking & Education',
      municipal: 'Municipal & Utilities',
      castle: 'Castle District',
      outskirts: 'Outskirts',
      center: 'Center Plaza',
    }

    function getZoneName(tx, ty) {
      for (const b of BITSBURG_BUILDINGS) {
        if (tx >= b.x && tx < b.x + b.w && ty >= b.y && ty < b.y + b.h) {
          if (b.type === 'business') return 'Bitsburgh Dev Co.'
          return b.zone
        }
      }
      if (tx < 15 && ty < 12) return 'residential'
      if (tx < 15 && ty > 25) return 'business'
      if (tx >= 15 && tx < 25 && ty >= 24 && ty < 32) return 'business_dev'
      if (tx > 15 && tx < 30 && ty < 12) return 'municipal'
      if (tx > 35 && ty < 22) return 'recreation'
      if (tx > 35 && ty > 25) return 'banking'
      if (tx > 20 && tx < 33 && ty > 12 && ty < 22) return 'castle'
      if (tx > 28 && tx < 36 && ty > 20 && ty < 28) return 'center'
      if (tx < 8 && ty > 37) return 'farm'
      return 'outskirts'
    }

    // Game loop
    let frameCount = 0
    let feedTimer = 0

    function update(dt) {
      const p = state.player
      const speed = 0.08 // tiles per frame
      const keys = keysRef.current

      let dx = 0, dy = 0
      if (keys['ArrowUp'] || keys['w'] || keys['W']) { dy = -1; p.dir = 1 }
      if (keys['ArrowDown'] || keys['s'] || keys['S']) { dy = 1; p.dir = 0 }
      if (keys['ArrowLeft'] || keys['a'] || keys['A']) { dx = -1; p.dir = 2 }
      if (keys['ArrowRight'] || keys['d'] || keys['D']) { dx = 1; p.dir = 3 }

      if (dx !== 0 || dy !== 0) {
        p.moving = true
        frameCount++
        if (frameCount % 10 === 0) p.frame = (p.frame + 1) % 2

        const newX = Math.max(0, Math.min(MAP_WIDTH - 1, p.x + dx * speed))
        const newY = Math.max(0, Math.min(MAP_HEIGHT - 1, p.y + dy * speed))

        // Collision check — don't walk into walls or water
        const tileAtNew = BITSBURG_MAP[Math.floor(newY)]?.[Math.floor(newX)]
        if (tileAtNew !== T.WALL && tileAtNew !== T.WATER && tileAtNew !== T.WATER_DEEP) {
          p.x = newX
          p.y = newY
        }

        // Update zone
        state.currentZone = getZoneName(Math.floor(p.x), Math.floor(p.y))
      } else {
        p.moving = false
        p.frame = 0
        frameCount = 0
      }

      // Update NPCs
      state.npcs.forEach(npc => {
        const target = npc.points[npc.pointIdx]
        const ddx = target.x - npc.x
        const ddy = target.y - npc.y
        const dist = Math.sqrt(ddx * ddx + ddy * ddy)

        if (dist < 0.1) {
          npc.pointIdx = (npc.pointIdx + 1) % npc.points.length
          npc.frame = 0
        } else {
          const spd = 0.03
          npc.x += (ddx / dist) * spd
          npc.y += (ddy / dist) * spd
          if (frameCount % 15 === 0) npc.frame = (npc.frame + 1) % 2

          // Determine direction
          if (Math.abs(ddx) > Math.abs(ddy)) {
            npc.dir = ddx > 0 ? 3 : 2
          } else {
            npc.dir = ddy > 0 ? 0 : 1
          }
        }
      })

      // Camera follows player
      const camTargetX = p.x * TILE_SIZE * SCALE - canvas.width / 2
      const camTargetY = p.y * TILE_SIZE * SCALE - canvas.height / 2
      state.camera.x += (camTargetX - state.camera.x) * 0.08
      state.camera.y += (camTargetY - state.camera.y) * 0.08

      // Clamp camera
      state.camera.x = Math.max(0, Math.min(
        MAP_WIDTH * TILE_SIZE * SCALE - canvas.width,
        state.camera.x
      ))
      state.camera.y = Math.max(0, Math.min(
        MAP_HEIGHT * TILE_SIZE * SCALE - canvas.height,
        state.camera.y
      ))

      // Live feed rotation
      feedTimer += dt
      if (feedTimer > 4000) {
        feedTimer = 0
        state.msgIdx = (state.msgIdx + 1) % LIVE_MESSAGES.length
        state.messages = [
          LIVE_MESSAGES[state.msgIdx],
          ...state.messages.slice(0, 4),
        ]
      }

      // Update HUD state (throttled)
      if (frameCount % 30 === 0) {
        setHudState({
          zone: ZONE_LABELS[state.currentZone] || 'Unknown',
          stats: { ...state.stats },
          messages: [...state.messages],
        })
      }
    }

    function render() {
      const cam = state.camera
      const s = SCALE

      ctx.clearRect(0, 0, canvas.width, canvas.height)

      // Background
      ctx.fillStyle = '#1a1a2e'
      ctx.fillRect(0, 0, canvas.width, canvas.height)

      // Calculate visible tile range
      const startTX = Math.floor(cam.x / (TILE_SIZE * s))
      const startTY = Math.floor(cam.y / (TILE_SIZE * s))
      const endTX = startTX + Math.ceil(canvas.width / (TILE_SIZE * s)) + 1
      const endTY = startTY + Math.ceil(canvas.height / (TILE_SIZE * s)) + 1

      // Draw tiles
      for (let ty = Math.max(0, startTY); ty < Math.min(MAP_HEIGHT, endTY); ty++) {
        for (let tx = Math.max(0, startTX); tx < Math.min(MAP_WIDTH, endTX); tx++) {
          const tileType = BITSBURG_MAP[ty][tx]
          const screenX = tx * TILE_SIZE * SCALE - cam.x
          const screenY = ty * TILE_SIZE * SCALE - cam.y
          drawTileSprite(ctx, tileType, screenX, screenY, SCALE)
        }
      }

      // Draw buildings (sorted by Y for proper depth)
      const sortedBuildings = [...BITSBURG_BUILDINGS].sort((a, b) => a.y - b.y)
      sortedBuildings.forEach(b => {
        const sx = b.x * TILE_SIZE * SCALE - cam.x
        const sy = b.y * TILE_SIZE * SCALE - cam.y
        const bw = b.w * TILE_SIZE * SCALE
        const bh = b.h * TILE_SIZE * SCALE
        // Only draw if visible
        if (sx + bw > 0 && sx < canvas.width && sy + bh > 0 && sy < canvas.height) {
          drawBuild(ctx, b.type, b.w, b.h, sx, sy, SCALE)
        }
      })

      // Draw NPCs (sorted by Y for depth)
      const sortedNPCs = [...state.npcs].sort((a, b) => a.y - b.y)
      sortedNPCs.forEach(npc => {
        const sx = npc.x * TILE_SIZE * SCALE - cam.x
        const sy = npc.y * TILE_SIZE * SCALE - cam.y
        drawChar(ctx, npc.config, sx, sy - 8 * SCALE, npc.dir, npc.frame, CHAR_SCALE)

        // Name tag
        ctx.font = `bold ${9 * SCALE}px monospace`
        ctx.textAlign = 'center'
        ctx.fillStyle = '#fff'
        ctx.strokeStyle = 'rgba(0,0,0,0.8)'
        ctx.lineWidth = 3
        ctx.strokeText(npc.config.name, sx + 8 * SCALE, sy - 10 * SCALE)
        ctx.fillText(npc.config.name, sx + 8 * SCALE, sy - 10 * SCALE)
      })

      // Draw player
      const px = state.player.x * TILE_SIZE * SCALE - cam.x
      const py = state.player.y * TILE_SIZE * SCALE - cam.y
      const playerConfig = agentId
        ? { ...CHARACTERS.agent, name: agentId }
        : { ...CHARACTERS.agent, name: 'Guest' }
      drawChar(ctx, playerConfig, px, py - 8 * SCALE, state.player.dir, state.player.frame, CHAR_SCALE)

      // Player name tag (highlighted)
      ctx.font = `bold ${10 * SCALE}px monospace`
      ctx.textAlign = 'center'
      ctx.fillStyle = '#00f5ff'
      ctx.strokeStyle = 'rgba(0,0,0,0.9)'
      ctx.lineWidth = 3
      ctx.strokeText(playerConfig.name, px + 8 * SCALE, py - 12 * SCALE)
      ctx.fillText(playerConfig.name, px + 8 * SCALE, py - 12 * SCALE)

      // Draw minimap on canvas
      drawMinimap(ctx, canvas.width, canvas.height)
    }

    function drawMinimap(ctx, cw, ch) {
      const mmW = 120
      const mmH = 90
      const mmX = cw - mmW - 10
      const mmY = 10
      const scale = 2

      // Minimap background
      ctx.fillStyle = 'rgba(0,0,0,0.7)'
      ctx.fillRect(mmX - 2, mmY - 2, mmW + 4, mmH + 4)
      ctx.strokeStyle = 'rgba(0,245,255,0.4)'
      ctx.lineWidth = 1
      ctx.strokeRect(mmX - 2, mmY - 2, mmW + 4, mmH + 4)

      // Draw map tiles (scaled down)
      const mmTileW = mmW / MAP_WIDTH
      const mmTileH = mmH / MAP_HEIGHT

      for (let y = 0; y < MAP_HEIGHT; y += 2) {
        for (let x = 0; x < MAP_WIDTH; x += 2) {
          const tile = BITSBURG_MAP[y][x]
          if (tile === 0) ctx.fillStyle = '#3D8B3D'
          else if (tile === 1) ctx.fillStyle = '#9E9686'
          else if (tile === 2) ctx.fillStyle = '#2E7ABF'
          else if (tile === 3) ctx.fillStyle = '#6B6B6B'
          else ctx.fillStyle = '#3D8B3D'
          ctx.fillRect(mmX + x * mmTileW, mmY + y * mmTileH, mmTileW * 2, mmTileH * 2)
        }
      }

      // Draw NPCs on minimap
      state.npcs.forEach(npc => {
        ctx.fillStyle = '#ff6b6b'
        ctx.fillRect(
          mmX + npc.x * mmTileW - 1,
          mmY + npc.y * mmTileH - 1,
          3, 3
        )
      })

      // Draw player on minimap
      const p = state.player
      ctx.fillStyle = '#00f5ff'
      ctx.fillRect(
        mmX + p.x * mmTileW - 2,
        mmY + p.y * mmTileH - 2,
        4, 4
      )

      // Store minimap bounds for click handling
      state.mmBounds = { x: mmX, y: mmY, w: mmW, h: mmH }
    }

    function gameLoop(timestamp) {
      const dt = timestamp - lastTimeRef.current
      lastTimeRef.current = timestamp

      update(dt)
      render()

      animRef.current = requestAnimationFrame(gameLoop)
    }

    animRef.current = requestAnimationFrame(gameLoop)

    return () => {
      cancelAnimationFrame(animRef.current)
      window.removeEventListener('resize', resize)
    }
  }, [])

  // D-pad touch controls
  const handleDPad = useCallback((dir) => {
    const keyMap = { up: 'ArrowUp', down: 'ArrowDown', left: 'ArrowLeft', right: 'ArrowRight' }
    keysRef.current[keyMap[dir]] = true
    setTimeout(() => { keysRef.current[keyMap[dir]] = false }, 100)
  }, [])

  // If inside a building, show the interior view
  if (insideBuilding && insideBuilding.type === 'business') {
    return <BusinessBuildingInterior onClose={() => setInsideBuilding(null)} />
  }

  return (
    <div style={styles.container} ref={containerRef}>
      {/* Canvas */}
      <div style={styles.canvasWrap} ref={(el) => {
        if (el && canvasRef.current) {
          // Parent already set via canvas ref in useEffect
        }
      }}>
        <canvas
          ref={canvasRef}
          style={styles.canvas}
          onClick={(e) => {
            // Minimap click to teleport
            const state = stateRef.current
            if (!state.mmBounds) return
            const rect = canvasRef.current.getBoundingClientRect()
            const mx = e.clientX - rect.left
            const my = e.clientY - rect.top
            const { x: mmX, y: mmY, w: mmW, h: mmH } = state.mmBounds
            if (mx >= mmX && mx <= mmX + mmW && my >= mmY && my <= mmY + mmH) {
              const p = state.player
              p.x = ((mx - mmX) / mmW) * MAP_WIDTH
              p.y = ((my - mmY) / mmH) * MAP_HEIGHT
            }
          }}
        />
      </div>

      {/* HUD Overlay */}
      <div className="bitiverse-hud" style={styles.hud}>
        {/* Top bar */}
        <div className="bitiverse-topbar" style={styles.topBar}>
          <div className="bitiverse-zone-badge" style={styles.zoneBadge}>
            <span style={{ fontSize: 14 }}>📍</span>
            <span className="bitiverse-zone-name">{hudState.zone}</span>
          </div>

          <div className="bitiverse-stats" style={styles.statsBar}>
            <span className="bitiverse-stat" style={styles.statItem}>
              ⚡ <strong>{hudState.stats.energy}%</strong>
            </span>
            <span className="bitiverse-stat" style={styles.statItem}>
              😰 <strong>{hudState.stats.stress}%</strong>
            </span>
            <span className="bitiverse-stat" style={styles.statItem}>
              📊 <strong>{hudState.stats.level}</strong>
            </span>
            <span className="bitiverse-stat" style={styles.statItem}>
              🪙 <strong>{hudState.stats.coins}</strong>
            </span>
          </div>

          <div className="bitiverse-actions" style={styles.topActions}>
            <button className="bitiverse-action-btn" style={styles.actionBtn} onClick={toggleFullscreen} title="Fullscreen (F)">
              ⛶
            </button>
            {onClose && (
              <button className="bitiverse-action-btn" style={styles.actionBtn} onClick={onClose} title="Close (Esc)">
                ✕
              </button>
            )}
          </div>
        </div>

        {/* Live feed (left side) */}
        <div className="bitiverse-feed" style={styles.liveFeed}>
          <div className="bitiverse-feed-header" style={styles.feedHeader}>📡 Live Feed</div>
          <div style={styles.feedMessages}>
            {hudState.messages.map((msg, i) => (
              <div key={i} style={{
                ...styles.feedMsg,
                opacity: 1 - i * 0.15,
                fontSize: 11 - i,
              }}>
                {msg}
              </div>
            ))}
          </div>
        </div>

        {/* D-pad (bottom center, mobile) */}
        <div className="bitiverse-dpad" style={styles.dpad}>
          <div style={styles.dpadRow}>
            <div />
            <button style={styles.dpadBtn} onTouchStart={() => handleDPad('up')} onMouseDown={() => handleDPad('up')}>▲</button>
            <div />
          </div>
          <div style={styles.dpadRow}>
            <button style={styles.dpadBtn} onTouchStart={() => handleDPad('left')} onMouseDown={() => handleDPad('left')}>◄</button>
            <div style={styles.dpadCenter}>•</div>
            <button style={styles.dpadBtn} onTouchStart={() => handleDPad('right')} onMouseDown={() => handleDPad('right')}>►</button>
          </div>
          <div style={styles.dpadRow}>
            <div />
            <button style={styles.dpadBtn} onTouchStart={() => handleDPad('down')} onMouseDown={() => handleDPad('down')}>▼</button>
            <div />
          </div>
        </div>

        {/* Controls hint */}
        <div className="bitiverse-controls-hint" style={styles.controlsHint}>
          WASD / Arrow Keys to move • E to enter buildings • F for fullscreen • ESC to close • Click minimap to teleport
        </div>

        {/* Enter building prompt */}
        {stateRef.current && (() => {
          const px = Math.floor(stateRef.current.player.x)
          const py = Math.floor(stateRef.current.player.y)
          const nearBuilding = BITSBURG_BUILDINGS.find(b => 
            b.type === 'business' && px >= b.x && px < b.x + b.w && py >= b.y && py < b.y + b.h
          )
          if (nearBuilding) {
            return (
              <div style={{
                position: 'absolute',
                bottom: 100,
                left: '50%',
                transform: 'translateX(-50%)',
                background: 'rgba(0,245,255,0.2)',
                border: '1px solid rgba(0,245,255,0.5)',
                borderRadius: 8,
                padding: '8px 16px',
                fontSize: 13,
                color: '#00f5ff',
                fontWeight: 600,
                pointerEvents: 'none',
              }}>
                Press E to enter {nearBuilding.label || 'Building'}
              </div>
            )
          }
          return null
        })()}

        {/* Guest badge */}
        {isGuest && (
          <div className="bitiverse-guest-badge" style={styles.guestBadge}>
            👁️ Spectator Mode
          </div>
        )}
      </div>
    </div>
  )
}

// ─── STYLES ───
const styles = {
  container: {
    position: 'fixed',
    inset: 0,
    zIndex: 9999,
    background: '#0a0e1a',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  canvasWrap: {
    width: '100%',
    height: '100%',
    position: 'absolute',
    inset: 0,
  },
  canvas: {
    width: '100%',
    height: '100%',
    imageRendering: 'pixelated',
  },
  hud: {
    position: 'absolute',
    inset: 0,
    pointerEvents: 'none',
    zIndex: 1,
  },
  topBar: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    minHeight: 44,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 12px',
    background: 'linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
    pointerEvents: 'auto',
  },
  zoneBadge: {
    display: 'flex',
    alignItems: 'center',
    gap: 6,
    background: 'rgba(0,245,255,0.15)',
    border: '1px solid rgba(0,245,255,0.3)',
    borderRadius: 6,
    padding: '4px 12px',
    fontSize: 13,
    fontWeight: 600,
    color: '#00f5ff',
  },
  statsBar: {
    display: 'flex',
    gap: 16,
    fontSize: 12,
    color: '#ccc',
    flexWrap: 'wrap',
    justifyContent: 'center',
    maxWidth: '60vw',
  },
  statItem: {
    display: 'flex',
    alignItems: 'center',
    gap: 4,
  },
  topActions: {
    display: 'flex',
    gap: 8,
  },
  actionBtn: {
    background: 'rgba(255,255,255,0.1)',
    border: '1px solid rgba(255,255,255,0.2)',
    borderRadius: 6,
    color: '#fff',
    minWidth: 36,
    minHeight: 36,
    width: 36,
    height: 36,
    cursor: 'pointer',
    fontSize: 16,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  liveFeed: {
    position: 'absolute',
    top: 54,
    left: 12,
    width: 260,
    background: 'rgba(0,0,0,0.7)',
    border: '1px solid rgba(255,255,255,0.1)',
    borderRadius: 8,
    padding: 8,
    pointerEvents: 'auto',
  },
  feedHeader: {
    fontSize: 11,
    fontWeight: 600,
    color: '#888',
    marginBottom: 6,
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  },
  feedMessages: {
    display: 'flex',
    flexDirection: 'column',
    gap: 4,
  },
  feedMsg: {
    color: '#ccc',
    fontFamily: 'monospace',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  },
  dpad: {
    position: 'absolute',
    bottom: 60,
    left: '50%',
    transform: 'translateX(-50%)',
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
    pointerEvents: 'auto',
  },
  dpadRow: {
    display: 'flex',
    gap: 2,
  },
  dpadBtn: {
    minWidth: 48,
    minHeight: 48,
    width: 48,
    height: 48,
    background: 'rgba(0,245,255,0.15)',
    border: '1px solid rgba(0,245,255,0.3)',
    borderRadius: 8,
    color: '#00f5ff',
    fontSize: 18,
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    userSelect: 'none',
  },
  dpadCenter: {
    minWidth: 48,
    minHeight: 48,
    width: 48,
    height: 48,
    background: 'rgba(0,0,0,0.3)',
    borderRadius: 8,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: '#555',
    fontSize: 20,
  },
  controlsHint: {
    position: 'absolute',
    bottom: 10,
    left: '50%',
    transform: 'translateX(-50%)',
    fontSize: 10,
    color: '#555',
    whiteSpace: 'normal',
    textAlign: 'center',
    maxWidth: '90vw',
    lineHeight: '1.4',
  },
  guestBadge: {
    position: 'absolute',
    bottom: 30,
    right: 12,
    background: 'rgba(180,79,255,0.15)',
    border: '1px solid rgba(180,79,255,0.3)',
    borderRadius: 6,
    padding: '4px 12px',
    fontSize: 11,
    fontWeight: 600,
    color: '#b44fff',
  },
}

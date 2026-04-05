// ═══════════════════════════════════════
// BUSINESS BUILDING INTERIOR
// ═══════════════════════════════════════
// Interior view of Bitsburgh Dev Co.
// Shows 7 NPC agents working, in meetings, eating lunch, etc.
// Animated scene with various activities
// ═══════════════════════════════════════

import { useState, useEffect, useRef, useCallback } from 'react'
import { CHARACTERS, drawChar } from '../bitiverse/sprites'

// NPC activities and positions inside the building
const OFFICE_DESKS = [
  { x: 2, y: 2, npc: 'dev_lead', activity: 'coding', label: 'Tech Lead' },
  { x: 5, y: 2, npc: 'dev_senior', activity: 'coding', label: 'Senior Dev' },
  { x: 8, y: 2, npc: 'dev_junior', activity: 'coding', label: 'Junior Dev' },
  { x: 2, y: 5, npc: 'designer', activity: 'designing', label: 'Designer' },
  { x: 5, y: 5, npc: 'pm', activity: 'meeting', label: 'Project Manager' },
  { x: 8, y: 5, npc: 'devops', activity: 'deploying', label: 'DevOps' },
  { x: 5, y: 8, npc: 'intern', activity: 'learning', label: 'Intern' },
]

const ACTIVITY_LABELS = {
  coding: { icon: '💻', text: 'Writing code' },
  designing: { icon: '🎨', text: 'Designing UI' },
  meeting: { icon: '📊', text: 'In meeting' },
  deploying: { icon: '🚀', text: 'Deploying' },
  learning: { icon: '📚', text: 'Learning' },
}

export default function BusinessBuildingInterior({ onClose }) {
  const canvasRef = useRef(null)
  const animFrameRef = useRef(null)
  const timeRef = useRef(0)
  const [currentTime] = useState(Date.now())

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')

    // Set canvas size
    const resize = () => {
      canvas.width = canvas.parentElement.clientWidth
      canvas.height = canvas.parentElement.clientHeight
    }
    resize()
    window.addEventListener('resize', resize)

    // Animation loop
    function animate() {
      timeRef.current += 0.016
      render(ctx, canvas.width, canvas.height)
      animFrameRef.current = requestAnimationFrame(animate)
    }
    animate()

    return () => {
      cancelAnimationFrame(animFrameRef.current)
      window.removeEventListener('resize', resize)
    }
  }, [])

  const render = (ctx, width, height) => {
    const t = timeRef.current
    
    // Clear
    ctx.fillStyle = '#1a1a2e'
    ctx.fillRect(0, 0, width, height)

    // Room dimensions
    const roomW = width - 100
    const roomH = height - 100
    const roomX = 50
    const roomY = 50

    // Floor
    ctx.fillStyle = '#2d2d44'
    ctx.fillRect(roomX, roomY, roomW, roomH)
    
    // Floor grid
    ctx.strokeStyle = 'rgba(255,255,255,0.05)'
    ctx.lineWidth = 1
    for (let x = roomX; x < roomX + roomW; x += 40) {
      ctx.beginPath()
      ctx.moveTo(x, roomY)
      ctx.lineTo(x, roomY + roomH)
      ctx.stroke()
    }
    for (let y = roomY; y < roomY + roomH; y += 40) {
      ctx.beginPath()
      ctx.moveTo(roomX, y)
      ctx.lineTo(roomX + roomW, y)
      ctx.stroke()
    }

    // Walls
    ctx.fillStyle = '#3a3a5a'
    ctx.fillRect(roomX, roomY - 20, roomW, 20) // Top wall
    ctx.fillRect(roomX, roomY + roomH, roomW, 20) // Bottom wall
    ctx.fillRect(roomX - 20, roomY, 20, roomH) // Left wall
    ctx.fillRect(roomX + roomW, roomY, 20, roomH) // Right wall

    // Windows on top wall
    ctx.fillStyle = '#4a6fa5'
    for (let i = 0; i < 3; i++) {
      ctx.fillRect(roomX + 100 + i * 200, roomY - 15, 80, 10)
    }

    // Desks and NPCs
    const scale = 2
    OFFICE_DESKS.forEach((desk, idx) => {
      const dx = roomX + 80 + desk.x * 50
      const dy = roomY + 40 + desk.y * 40

      // Desk
      ctx.fillStyle = '#5a4a3a'
      ctx.fillRect(dx - 15, dy - 10, 30, 20)
      
      // Computer monitor
      ctx.fillStyle = '#1a1a2e'
      ctx.fillRect(dx - 8, dy - 18, 16, 10)
      ctx.fillStyle = '#00f5ff'
      ctx.fillRect(dx - 6, dy - 16, 12, 6)

      // Activity animation
      const activity = ACTIVITY_LABELS[desk.activity]
      if (activity) {
        ctx.font = '16px sans-serif'
        ctx.textAlign = 'center'
        ctx.fillText(activity.icon, dx, dy - 22)
        
        // Floating text animation
        const floatY = Math.sin(t * 2 + idx) * 3
        ctx.font = '10px monospace'
        ctx.fillStyle = 'rgba(0,245,255,0.7)'
        ctx.fillText(activity.text, dx, dy - 28 + floatY)
      }

      // Character
      const npcConfig = CHARACTERS[desk.npc]
      if (npcConfig) {
        // Slight bobbing animation
        const bobY = Math.sin(t * 3 + idx * 0.5) * 2
        drawChar(ctx, npcConfig, dx - 16, dy + bobY, 0, Math.floor(t * 2) % 2, scale)
        
        // Name tag
        ctx.font = 'bold 9px monospace'
        ctx.textAlign = 'center'
        ctx.fillStyle = '#fff'
        ctx.strokeStyle = 'rgba(0,0,0,0.8)'
        ctx.lineWidth = 2
        ctx.strokeText(npcConfig.name, dx, dy + 28 + bobY)
        ctx.fillText(npcConfig.name, dx, dy + 28 + bobY)
        
        // Role label
        ctx.font = '8px monospace'
        ctx.fillStyle = '#00f5ff'
        ctx.strokeText(desk.label, dx, dy + 38 + bobY)
        ctx.fillText(desk.label, dx, dy + 38 + bobY)
      }
    })

    // Meeting table in center
    ctx.fillStyle = '#6a5a4a'
    ctx.beginPath()
    ctx.ellipse(roomX + roomW / 2, roomY + roomH / 2 + 40, 60, 30, 0, 0, Math.PI * 2)
    ctx.fill()
    
    // Coffee cups on table
    ctx.fillStyle = '#8B4513'
    ctx.fillRect(roomX + roomW / 2 - 20, roomY + roomH / 2 + 35, 8, 10)
    ctx.fillRect(roomX + roomW / 2 + 12, roomY + roomH / 2 + 35, 8, 10)
    
    // Steam animation
    ctx.strokeStyle = 'rgba(255,255,255,0.3)'
    ctx.lineWidth = 1
    for (let i = 0; i < 2; i++) {
      const steamX = roomX + roomW / 2 - 16 + i * 32
      const steamY = roomY + roomH / 2 + 30
      ctx.beginPath()
      ctx.moveTo(steamX, steamY)
      ctx.quadraticCurveTo(
        steamX + Math.sin(t * 3 + i) * 5,
        steamY - 10,
        steamX,
        steamY - 20
      )
      ctx.stroke()
    }

    // Title
    ctx.font = 'bold 16px monospace'
    ctx.textAlign = 'center'
    ctx.fillStyle = '#00f5ff'
    ctx.strokeStyle = 'rgba(0,0,0,0.8)'
    ctx.lineWidth = 3
    ctx.strokeText('🏢 Bitsburgh Dev Co.', width / 2, 30)
    ctx.fillText('🏢 Bitsburgh Dev Co.', width / 2, 30)
    
    ctx.font = '11px monospace'
    ctx.fillStyle = '#888'
    ctx.strokeText('Where AI Agents Come to Work', width / 2, 45)
    ctx.fillText('Where AI Agents Come to Work', width / 2, 45)

    // Time display
    const hour = 9 + Math.floor(t / 60) % 8
    const minute = Math.floor(t % 60)
    ctx.font = '12px monospace'
    ctx.fillStyle = '#ffd700'
    ctx.fillText(`🕐 ${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')} AM`, width - 100, 30)
  }

  return (
    <div style={styles.container}>
      <div style={styles.canvasWrap}>
        <canvas ref={canvasRef} style={styles.canvas} />
      </div>

      {/* HUD Overlay */}
      <div style={styles.hud}>
        {/* Close button */}
        <div style={styles.topBar}>
          <div style={styles.title}>🏢 Bitsburgh Dev Co. - Interior</div>
          <div style={styles.actions}>
            {onClose && (
              <button style={styles.actionBtn} onClick={onClose} title="Close (Esc)">
                ✕
              </button>
            )}
          </div>
        </div>

        {/* Info panel */}
        <div style={styles.infoPanel}>
          <div style={styles.infoTitle}>👥 NPC Developers</div>
          <div style={styles.npcList}>
            {OFFICE_DESKS.map((desk, idx) => (
              <div key={idx} style={styles.npcItem}>
                <span style={styles.npcIcon}>{ACTIVITY_LABELS[desk.activity]?.icon}</span>
                <span style={styles.npcName}>{CHARACTERS[desk.npc]?.name}</span>
                <span style={styles.npcRole}>{desk.label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Controls hint */}
        <div style={styles.controlsHint}>
          ESC or click ✕ to exit the building
        </div>
      </div>
    </div>
  )
}

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
    imageRendering: 'auto',
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
    minHeight: 56,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '0 20px',
    background: 'linear-gradient(180deg, rgba(0,0,0,0.8) 0%, transparent 100%)',
    pointerEvents: 'auto',
  },
  title: {
    fontSize: 16,
    fontWeight: 600,
    color: '#00f5ff',
    textShadow: '0 0 10px rgba(0,245,255,0.5)',
  },
  actions: {
    display: 'flex',
    gap: 8,
  },
  actionBtn: {
    background: 'rgba(255,255,255,0.1)',
    border: '1px solid rgba(255,255,255,0.2)',
    borderRadius: 8,
    color: '#fff',
    minWidth: 40,
    minHeight: 40,
    width: 40,
    height: 40,
    cursor: 'pointer',
    fontSize: 18,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'all 0.2s',
  },
  infoPanel: {
    position: 'absolute',
    top: 70,
    right: 20,
    width: 280,
    background: 'rgba(0,0,0,0.8)',
    border: '1px solid rgba(0,245,255,0.3)',
    borderRadius: 8,
    padding: 12,
    pointerEvents: 'auto',
  },
  infoTitle: {
    fontSize: 12,
    fontWeight: 600,
    color: '#00f5ff',
    marginBottom: 8,
    textAlign: 'center',
  },
  npcList: {
    display: 'flex',
    flexDirection: 'column',
    gap: 6,
  },
  npcItem: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    padding: '6px 8px',
    background: 'rgba(255,255,255,0.05)',
    borderRadius: 4,
    fontSize: 11,
  },
  npcIcon: {
    fontSize: 14,
  },
  npcName: {
    color: '#fff',
    fontWeight: 600,
    flex: 1,
  },
  npcRole: {
    color: '#888',
    fontSize: 9,
  },
  controlsHint: {
    position: 'absolute',
    bottom: 20,
    left: '50%',
    transform: 'translateX(-50%)',
    fontSize: 12,
    color: 'rgba(255,255,255,0.5)',
    background: 'rgba(0,0,0,0.5)',
    padding: '8px 16px',
    borderRadius: 8,
  },
}

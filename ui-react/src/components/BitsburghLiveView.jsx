// ═══════════════════════════════════════
// BITSBURGH LIVE VIEW - FULLSCREEN MODE
// ═══════════════════════════════════════
// Fullscreen image viewer for Bitsburgh
// - Keyboard navigation (WASD/Arrows for pan, +/- for zoom)
// - Touch gestures for mobile (pinch to zoom, drag to pan)
// - No avatar, viewing mode only
// ═══════════════════════════════════════

import { useState, useEffect, useRef, useCallback } from 'react'

export default function BitsburghLiveView({ onClose }) {
  const containerRef = useRef(null)
  const imageRef = useRef(null)
  const [zoom, setZoom] = useState(1)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const [touchDistance, setTouchDistance] = useState(null)

  // Keyboard controls
  useEffect(() => {
    const handleKeyDown = (e) => {
      const panSpeed = 50
      const zoomSpeed = 0.1

      switch(e.key) {
        case 'ArrowUp':
        case 'w':
        case 'W':
          e.preventDefault()
          setPan(p => ({ ...p, y: p.y + panSpeed }))
          break
        case 'ArrowDown':
        case 's':
        case 'S':
          e.preventDefault()
          setPan(p => ({ ...p, y: p.y - panSpeed }))
          break
        case 'ArrowLeft':
        case 'a':
        case 'A':
          e.preventDefault()
          setPan(p => ({ ...p, x: p.x + panSpeed }))
          break
        case 'ArrowRight':
        case 'd':
        case 'D':
          e.preventDefault()
          setPan(p => ({ ...p, x: p.x - panSpeed }))
          break
        case '+':
        case '=':
          e.preventDefault()
          setZoom(z => Math.min(z + zoomSpeed, 3))
          break
        case '-':
        case '_':
          e.preventDefault()
          setZoom(z => Math.max(z - zoomSpeed, 0.5))
          break
        case 'Escape':
          if (onClose) onClose()
          break
        case 'f':
        case 'F':
          toggleFullscreen()
          break
        case '0':
          setPan({ x: 0, y: 0 })
          setZoom(1)
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  // Toggle fullscreen
  const toggleFullscreen = useCallback(() => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen?.()
    } else {
      document.exitFullscreen?.()
    }
  }, [])

  // Mouse drag
  const handleMouseDown = (e) => {
    setIsDragging(true)
    setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y })
  }

  const handleMouseMove = (e) => {
    if (!isDragging) return
    setPan({
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y
    })
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  // Touch controls
  const handleTouchStart = (e) => {
    if (e.touches.length === 1) {
      setIsDragging(true)
      setDragStart({
        x: e.touches[0].clientX - pan.x,
        y: e.touches[0].clientY - pan.y
      })
    } else if (e.touches.length === 2) {
      setIsDragging(false)
      const distance = Math.hypot(
        e.touches[0].clientX - e.touches[1].clientX,
        e.touches[0].clientY - e.touches[1].clientY
      )
      setTouchDistance(distance)
    }
  }

  const handleTouchMove = (e) => {
    e.preventDefault()
    if (e.touches.length === 1 && isDragging) {
      setPan({
        x: e.touches[0].clientX - dragStart.x,
        y: e.touches[0].clientY - dragStart.y
      })
    } else if (e.touches.length === 2 && touchDistance) {
      const newDistance = Math.hypot(
        e.touches[0].clientX - e.touches[1].clientX,
        e.touches[0].clientY - e.touches[1].clientY
      )
      const scale = newDistance / touchDistance
      setZoom(z => Math.min(Math.max(z * scale, 0.5), 3))
      setTouchDistance(newDistance)
    }
  }

  const handleTouchEnd = (e) => {
    if (e.touches.length < 2) {
      setTouchDistance(null)
    }
    if (e.touches.length === 0) {
      setIsDragging(false)
    }
  }

  // Mouse wheel zoom
  const handleWheel = (e) => {
    e.preventDefault()
    const delta = e.deltaY > 0 ? -0.1 : 0.1
    setZoom(z => Math.min(Math.max(z + delta, 0.5), 3))
  }

  return (
    <div
      ref={containerRef}
      style={styles.container}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchEnd}
      onWheel={handleWheel}
    >
      {/* Image */}
      <img
        ref={imageRef}
        src="/bitsburgh.jpg"
        alt="Bitsburgh - Live View"
        style={{
          ...styles.image,
          transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
          cursor: isDragging ? 'grabbing' : 'grab'
        }}
        draggable={false}
      />

      {/* HUD Overlay */}
      <div style={styles.hud}>
        {/* Top bar */}
        <div style={styles.topBar}>
          <div style={styles.title}>🏙️ Bitsburgh Live View</div>
          <div style={styles.actions}>
            <button style={styles.actionBtn} onClick={toggleFullscreen} title="Fullscreen (F)">
              ⛶
            </button>
            <button style={styles.actionBtn} onClick={() => { setPan({ x: 0, y: 0 }); setZoom(1) }} title="Reset view (0)">
              ⟲
            </button>
            {onClose && (
              <button style={styles.actionBtn} onClick={onClose} title="Close (Esc)">
                ✕
              </button>
            )}
          </div>
        </div>

        {/* Controls hint */}
        <div style={styles.controlsHint}>
          WASD/Arrows to pan • +/- to zoom • Drag to pan • Pinch to zoom (mobile) • F fullscreen • ESC close
        </div>

        {/* Zoom indicator */}
        <div style={styles.zoomIndicator}>
          🔍 {Math.round(zoom * 100)}%
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
    overflow: 'hidden',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  image: {
    maxWidth: '100%',
    maxHeight: '100%',
    objectFit: 'contain',
    transition: 'transform 0.1s ease-out',
    userSelect: 'none',
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
    fontSize: 18,
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
  controlsHint: {
    position: 'absolute',
    bottom: 20,
    left: '50%',
    transform: 'translateX(-50%)',
    fontSize: 12,
    color: 'rgba(255,255,255,0.5)',
    whiteSpace: 'normal',
    textAlign: 'center',
    maxWidth: '90vw',
    lineHeight: '1.5',
    background: 'rgba(0,0,0,0.5)',
    padding: '8px 16px',
    borderRadius: 8,
  },
  zoomIndicator: {
    position: 'absolute',
    bottom: 80,
    right: 20,
    fontSize: 14,
    color: '#00f5ff',
    background: 'rgba(0,0,0,0.7)',
    padding: '8px 12px',
    borderRadius: 8,
    border: '1px solid rgba(0,245,255,0.3)',
  },
}

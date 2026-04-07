import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import { useStore } from './store'
import NeuralBackground from './components/scenes/NeuralBackground'
import AEXCFeed from './components/AEXCFeed'
import BitiverseDashboard from './components/BitiverseDashboard'
import BitiverseFullscreenWorld from './components/BitiverseFullscreenWorld'

// ═══════════════════════════════════════
// CURSOR GLOW
// ═══════════════════════════════════════

function CursorGlow() {
  const ref = useRef(null)
  useEffect(() => {
    const onMove = (e) => {
      if (ref.current) {
        ref.current.style.left = e.clientX + 'px'
        ref.current.style.top = e.clientY + 'px'
      }
    }
    window.addEventListener('mousemove', onMove)
    return () => window.removeEventListener('mousemove', onMove)
  }, [])
  return <div ref={ref} className="cursor-glow" />
}

// ═══════════════════════════════════════
// THREE.JS — BACKGROUND PARTICLES
// ═══════════════════════════════════════

function BgParticles() {
  const ref = useRef()
  useFrame(() => { if (ref.current) { ref.current.rotation.y += 0.00015; ref.current.rotation.x += 0.00008 } })
  const geo = useMemo(() => {
    const g = new THREE.BufferGeometry()
    const N = 2000, p = new Float32Array(N * 3)
    for (let i = 0; i < N * 3; i++) p[i] = (Math.random() - 0.5) * 200
    g.setAttribute('position', new THREE.BufferAttribute(p, 3))
    return g
  }, [])
  return <points ref={ref} geometry={geo}><pointsMaterial size={0.3} color={0x00f5ff} transparent opacity={0.35} /></points>
}

// ═══════════════════════════════════════
// THREE.JS — 4D STEP ANIMATION
// ═══════════════════════════════════════

function LandingAnimation({ phase }) {
  const lineRef = useRef()
  const squareRef = useRef()
  const cubeRef = useRef()
  const cubeEdgeRef = useRef()
  const tesseractOuterRef = useRef()
  const tesseractInnerRef = useRef()
  const chipTracesRef = useRef()
  const eyeLeftRef = useRef()
  const eyeRightRef = useRef()
  const pupilLeftRef = useRef()
  const pupilRightRef = useRef()
  const glowRingRef = useRef()

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime()
    if (cubeRef.current) { cubeRef.current.rotation.y = t * 0.3; cubeRef.current.rotation.x = t * 0.2 }
    if (cubeEdgeRef.current) { cubeEdgeRef.current.rotation.y = t * 0.3; cubeEdgeRef.current.rotation.x = t * 0.2 }
    if (tesseractOuterRef.current) { tesseractOuterRef.current.rotation.y = t * 0.25; tesseractOuterRef.current.rotation.x = t * 0.15 }
    if (tesseractInnerRef.current) { tesseractInnerRef.current.rotation.y = t * 0.3; tesseractInnerRef.current.rotation.x = t * 0.2 }
    if (chipTracesRef.current) { chipTracesRef.current.material.opacity = 0.5 + 0.5 * Math.sin(t * 2) }
    if (eyeLeftRef.current) { const blink = Math.sin(t * 3) > 0.95 ? 0.1 : 1; eyeLeftRef.current.scale.y = blink; eyeRightRef.current.scale.y = blink }
    if (glowRingRef.current) { glowRingRef.current.material.opacity = 0.1 + 0.1 * Math.sin(t * 1.5) }
  })

  const chipTracePositions = useMemo(() => {
    const pts = []
    for (let i = -3; i <= 3; i += 2) { pts.push(new THREE.Vector3(-2, i * 0.3, 0)); pts.push(new THREE.Vector3(2, i * 0.3, 0)) }
    for (let i = -2; i <= 2; i += 2) { pts.push(new THREE.Vector3(i * 0.5, -1.5, 0)); pts.push(new THREE.Vector3(i * 0.5, 1.5, 0)) }
    return pts
  }, [])

  const chipNodePositions = useMemo(() => {
    const pts = []
    for (let x = -1.5; x <= 1.5; x += 1) { for (let y = -1.2; y <= 1.2; y += 0.8) { pts.push(new THREE.Vector3(x, y, 0.05)) } }
    return pts
  }, [])

  return (
    <>
      <pointLight position={[3, 3, 3]} intensity={2} color={0x00f5ff} />
      <pointLight position={[-3, -2, 2]} intensity={1.5} color={0xb44fff} />
      <pointLight position={[0, 0, 5]} intensity={1} color={0x00f5ff} />

      {phase === 0 && (
        <line ref={lineRef}>
          <bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([-2, 0, 0, 2, 0, 0])} itemSize={3} /></bufferGeometry>
          <lineBasicMaterial color={0x00f5ff} transparent opacity={0.8} />
        </line>
      )}

      {phase === 1 && (
        <mesh ref={squareRef}>
          <planeGeometry args={[3, 3]} />
          <meshStandardMaterial color={0x00f5ff} emissive={0x003344} metalness={0.5} roughness={0.4} transparent opacity={0.7} side={THREE.DoubleSide} />
        </mesh>
      )}

      {phase === 2 && (
        <>
          <mesh ref={cubeRef}><boxGeometry args={[2.5, 2.5, 2.5]} /><meshStandardMaterial color={0x00f5ff} emissive={0x002233} metalness={0.6} roughness={0.3} transparent opacity={0.6} wireframe /></mesh>
          <lineSegments ref={cubeEdgeRef}><edgesGeometry args={[new THREE.BoxGeometry(2.5, 2.5, 2.5)]} /><lineBasicMaterial color={0x00f5ff} transparent opacity={0.9} /></lineSegments>
        </>
      )}

      {phase >= 3 && (
        <>
          <mesh ref={tesseractOuterRef}><boxGeometry args={[3.2, 3.2, 3.2]} /><meshStandardMaterial color={0x00f5ff} emissive={0x002233} metalness={0.7} roughness={0.2} transparent opacity={0.5} wireframe /></mesh>
          <lineSegments><edgesGeometry args={[new THREE.BoxGeometry(3.2, 3.2, 3.2)]} /><lineBasicMaterial color={0x00f5ff} transparent opacity={0.8} /></lineSegments>
          <mesh ref={tesseractInnerRef} position={[0.3, 0.3, 0.3]}><boxGeometry args={[1.6, 1.6, 1.6]} /><meshStandardMaterial color={0xb44fff} emissive={0x220033} metalness={0.7} roughness={0.2} transparent opacity={0.6} wireframe /></mesh>
          <lineSegments><edgesGeometry args={[new THREE.BoxGeometry(1.6, 1.6, 1.6)]} /><lineBasicMaterial color={0xb44fff} transparent opacity={0.7} /></lineSegments>
          <lineSegments><bufferGeometry><bufferAttribute attach="attributes-position" count={16} array={new Float32Array([-1.6,-1.6,-1.6,-0.8,-0.8,-0.8,1.6,-1.6,-1.6,0.8,-0.8,-0.8,-1.6,1.6,-1.6,-0.8,0.8,-0.8,1.6,1.6,-1.6,0.8,0.8,-0.8,-1.6,-1.6,1.6,-0.8,-0.8,0.8,1.6,-1.6,1.6,0.8,-0.8,0.8,-1.6,1.6,1.6,-0.8,0.8,0.8,1.6,1.6,1.6,0.8,0.8,0.8])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.4} /></lineSegments>
          <lineSegments ref={chipTracesRef}><bufferGeometry><bufferAttribute attach="attributes-position" count={chipTracePositions.length} array={new Float32Array(chipTracePositions.flatMap(p => [p.x, p.y, p.z]))} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.5} /></lineSegments>
          {chipNodePositions.map((pos, i) => (<mesh key={i} position={pos}><boxGeometry args={[0.2, 0.2, 0.05]} /><meshStandardMaterial color={i % 3 === 0 ? 0x00f5ff : i % 3 === 1 ? 0xb44fff : 0xffd700} emissive={0x001122} metalness={0.8} roughness={0.2} transparent opacity={0.8} /></mesh>))}
          <mesh ref={glowRingRef} rotation={[Math.PI / 2, 0, 0]}><torusGeometry args={[2.2, 0.03, 8, 64]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.5} transparent opacity={0.2} /></mesh>
          <group position={[0, 0, 2.2]}>
            <mesh ref={eyeLeftRef} position={[-1.8, 0.5, 0]}><sphereGeometry args={[0.35, 16, 16]} /><meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1.2} /></mesh>
            <mesh ref={pupilLeftRef} position={[-1.8, 0.5, 0.15]}><sphereGeometry args={[0.15, 8, 8]} /><meshBasicMaterial color={0x001122} /></mesh>
            <mesh ref={eyeRightRef} position={[1.8, 0.5, 0]}><sphereGeometry args={[0.35, 16, 16]} /><meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1.2} /></mesh>
            <mesh ref={pupilRightRef} position={[1.8, 0.5, 0.15]}><sphereGeometry args={[0.15, 8, 8]} /><meshBasicMaterial color={0x001122} /></mesh>
          </group>
        </>
      )}
    </>
  )
}

// ═══════════════════════════════════════
// THREE.JS — AI PLANETARY SYSTEM
// ═══════════════════════════════════════

function AIPlanetarySystem() {
  const circuitPlanetRef = useRef()
  const robotMoonRef = useRef()
  const keyboardRingRef = useRef()
  const debrisGroupRef = useRef()

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime()
    if (circuitPlanetRef.current) { circuitPlanetRef.current.rotation.y = t * 0.15; circuitPlanetRef.current.rotation.x = Math.sin(t * 0.1) * 0.1 }
    if (robotMoonRef.current) {
      const a = t * 0.2, r = 5
      robotMoonRef.current.position.x = Math.cos(a) * r; robotMoonRef.current.position.z = Math.sin(a) * r
      robotMoonRef.current.position.y = Math.sin(t * 0.3) * 1.5
      robotMoonRef.current.rotation.y = t * 0.3; robotMoonRef.current.rotation.x = t * 0.1
    }
    if (keyboardRingRef.current) { keyboardRingRef.current.rotation.z = t * 0.05; keyboardRingRef.current.rotation.x = 0.3 + Math.sin(t * 0.2) * 0.1 }
    if (debrisGroupRef.current) {
      debrisGroupRef.current.children.forEach((child, i) => {
        const a = t * (0.1 + i * 0.02) + i * 0.5, r = 3.5 + (i % 3) * 1.2
        child.position.x = Math.cos(a) * r; child.position.z = Math.sin(a) * r; child.position.y = Math.sin(t * 0.3 + i) * 0.8
        child.rotation.x = t * (0.5 + i * 0.1); child.rotation.y = t * (0.3 + i * 0.05)
      })
    }
  })

  const debrisPieces = useMemo(() => {
    const pieces = [], colors = [0x00f5ff, 0xb44fff, 0xffd700, 0x00e676, 0xff6b9d]
    for (let i = 0; i < 20; i++) {
      const size = 0.08 + Math.random() * 0.15
      pieces.push(<mesh key={i} position={[0, 0, 0]}>
        {Math.random() > 0.5 ? <boxGeometry args={[size, size * 0.6, size * 0.4]} /> : <octahedronGeometry args={[size * 0.5, 0]} />}
        <meshStandardMaterial color={colors[i % colors.length]} emissive={colors[i % colors.length]} emissiveIntensity={0.3} metalness={0.7} roughness={0.3} transparent opacity={0.6} />
      </mesh>)
    }
    return pieces
  }, [])

  return (
    <group>
      <group ref={circuitPlanetRef} position={[0, 0, 0]}>
        <mesh><sphereGeometry args={[2, 32, 32]} /><meshStandardMaterial color={0x1a3a2a} emissive={0x0a1a0a} metalness={0.4} roughness={0.6} /></mesh>
        {[0, 1, 2].map(i => (<mesh key={`th${i}`} rotation={[Math.PI / 2, 0, 0]} position={[0, (i - 1) * 0.6, 0]}><torusGeometry args={[2.02, 0.02, 8, 32]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.5} transparent opacity={0.6} /></mesh>))}
        {[0, 1, 2].map(i => (<mesh key={`tv${i}`} rotation={[0, 0, (i - 1) * Math.PI / 3]}><torusGeometry args={[2.02, 0.015, 8, 32]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.4} transparent opacity={0.4} /></mesh>))}
        {Array.from({ length: 12 }, (_, i) => {
          const phi = Math.acos(2 * Math.random() - 1), theta = Math.random() * Math.PI * 2, r = 2.05
          return (<mesh key={`c${i}`} position={[r * Math.sin(phi) * Math.cos(theta), r * Math.sin(phi) * Math.sin(theta), r * Math.cos(phi)]} rotation={[Math.random() * Math.PI, Math.random() * Math.PI, 0]}>
            <boxGeometry args={[0.3 + Math.random() * 0.2, 0.15, 0.2 + Math.random() * 0.15]} />
            <meshStandardMaterial color={i % 3 === 0 ? 0x2a4a6a : i % 3 === 1 ? 0x1a2a3a : 0x3a5a3a} emissive={i % 2 === 0 ? 0x00f5ff : 0xb44fff} emissiveIntensity={0.3} metalness={0.6} roughness={0.4} />
          </mesh>)
        })}
        {Array.from({ length: 6 }, (_, i) => {
          const phi = Math.acos(2 * Math.random() - 1), theta = Math.random() * Math.PI * 2, r = 2.08
          return (<mesh key={`cap${i}`} position={[r * Math.sin(phi) * Math.cos(theta), r * Math.sin(phi) * Math.sin(theta), r * Math.cos(phi)]}>
            <cylinderGeometry args={[0.08, 0.08, 0.2, 8]} /><meshStandardMaterial color={0xffd700} emissive={0xffd700} emissiveIntensity={0.4} metalness={0.8} roughness={0.2} />
          </mesh>)
        })}
        <mesh position={[0, 0.5, 1.8]} rotation={[0.2, 0.3, 0]}><boxGeometry args={[0.8, 0.15, 0.6]} /><meshStandardMaterial color={0x1a2a3a} emissive={0x00f5ff} emissiveIntensity={0.4} metalness={0.7} roughness={0.3} /></mesh>
        {Array.from({ length: 8 }, (_, i) => (<mesh key={`pin${i}`} position={[-0.35 + i * 0.1, 0.5, 2.15]}><boxGeometry args={[0.02, 0.08, 0.02]} /><meshStandardMaterial color={0x888888} metalness={0.9} roughness={0.1} /></mesh>))}
        <mesh><sphereGeometry args={[0.4, 16, 16]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={1} transparent opacity={0.3} /></mesh>
      </group>

      <group ref={keyboardRingRef} position={[0, 0, 0]}>
        <mesh rotation={[Math.PI / 2, 0, 0]}><torusGeometry args={[3.5, 0.25, 8, 48]} /><meshStandardMaterial color={0x2a2a3a} emissive={0x0a0a1a} metalness={0.5} roughness={0.5} /></mesh>
        {Array.from({ length: 24 }, (_, i) => {
          const a = (i / 24) * Math.PI * 2, r = 3.5
          return (<mesh key={`k${i}`} position={[Math.cos(a) * r, 0, Math.sin(a) * r]} rotation={[0, -a, Math.PI / 2]}>
            <boxGeometry args={[0.25, 0.12, 0.2]} />
            <meshStandardMaterial color={i % 4 === 0 ? 0xd4c4a0 : i % 4 === 1 ? 0x2a2a3a : i % 4 === 2 ? 0x4a4a5a : 0x888888} emissive={i % 6 === 0 ? 0x00f5ff : 0x000000} emissiveIntensity={0.3} metalness={0.4} roughness={0.6} />
          </mesh>)
        })}
        {Array.from({ length: 8 }, (_, i) => {
          const a = (i / 8) * Math.PI * 2 + Math.PI / 8, r = 3.5
          return (<mesh key={`rc${i}`} position={[Math.cos(a) * r, 0, Math.sin(a) * r]} rotation={[0, -a, Math.PI / 2]}>
            <boxGeometry args={[0.35, 0.1, 0.25]} /><meshStandardMaterial color={0x1a2a3a} emissive={0x00f5ff} emissiveIntensity={0.2} metalness={0.6} roughness={0.4} />
          </mesh>)
        })}
        <mesh rotation={[Math.PI / 2, 0, 0]}><torusGeometry args={[3.5, 0.35, 8, 48]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.15} transparent opacity={0.08} /></mesh>
      </group>

      <group ref={robotMoonRef} position={[5, 0, 0]}>
        <mesh><icosahedronGeometry args={[0.8, 1]} /><meshStandardMaterial color={0x1a1a2a} emissive={0x0a0a1a} metalness={0.5} roughness={0.5} /></mesh>
        {Array.from({ length: 8 }, (_, i) => {
          const phi = Math.acos(2 * ((i + 0.5) / 8) - 1), theta = i * Math.PI * 0.7, r = 0.85
          return (<mesh key={`f${i}`} position={[r * Math.sin(phi) * Math.cos(theta), r * Math.sin(phi) * Math.sin(theta), r * Math.cos(phi)]}>
            <boxGeometry args={[0.3, 0.25, 0.05]} /><meshStandardMaterial color={0x0a0a1a} emissive={0x0a0a1a} metalness={0.3} roughness={0.7} />
          </mesh>)
        })}
        <mesh position={[0, 0.15, 0.85]}><boxGeometry args={[0.12, 0.12, 0.02]} /><meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1} /></mesh>
        <mesh position={[0.3, 0.15, 0.82]}><boxGeometry args={[0.12, 0.12, 0.02]} /><meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1} /></mesh>
        <mesh position={[0.15, -0.1, 0.86]}><boxGeometry args={[0.2, 0.04, 0.02]} /><meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={0.8} /></mesh>
        <mesh position={[0, 0.9, 0]}><cylinderGeometry args={[0.02, 0.02, 0.2, 6]} /><meshStandardMaterial color={0x888888} metalness={0.9} roughness={0.1} /></mesh>
        <mesh position={[0, 1.05, 0]}><sphereGeometry args={[0.05, 8, 8]} /><meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.8} /></mesh>
      </group>

      <group ref={debrisGroupRef}>{debrisPieces}</group>
      <ambientLight intensity={0.2} />
      <pointLight position={[10, 5, 10]} intensity={1} color={0x00f5ff} />
      <pointLight position={[-10, -5, -10]} intensity={0.5} color={0xb44fff} />
    </group>
  )
}

// ═══════════════════════════════════════
// THREE.JS — FORGE REACTOR
// ═══════════════════════════════════════

function ForgeReactor() {
  const g = useRef()
  useFrame(({ clock }) => { const t = clock.getElapsedTime(); if (g.current) { g.current.rotation.y = t * 1.2; g.current.rotation.x = t * 0.7; const s = 0.8 + 0.4 * Math.abs(Math.sin(t * 2)); g.current.scale.set(s, s, s) } })
  return (<>
    {[0, 1, 2].map(i => <mesh key={i}><torusGeometry args={[1.5 + i * 0.4, 0.04, 8, 64]} /><meshStandardMaterial color={[0x00f5ff, 0xb44fff, 0xffd700][i]} emissive={new THREE.Color([0x00f5ff, 0xb44fff, 0xffd700][i]).multiplyScalar(0.5)} metalness={0.8} roughness={0.2} transparent opacity={0.7} /></mesh>)}
    <mesh ref={g}><boxGeometry args={[1, 1, 1]} /><meshStandardMaterial color={0x00f5ff} emissive={0x002233} metalness={0.9} roughness={0.1} /></mesh>
  </>)
}

// ═══════════════════════════════════════
// AUTH VIEW — Email + Token Sign-In
// ═══════════════════════════════════════

function AuthView({ onAuthenticated, onSkip, onCancel }) {
  const [step, setStep] = useState(0)
  const [email, setEmail] = useState('')
  const [token, setToken] = useState('')
  const [remember, setRemember] = useState(true)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [sessionData, setSessionData] = useState(null)
  const [mockToken, setMockToken] = useState('')

  const requestToken = async () => {
    if (!email.includes('@')) { setError('Please enter a valid email'); return }
    setError(''); setLoading(true); setStep(1)
    try {
      const res = await fetch('/api/auth/request', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, remember }) })
      const data = await res.json()
      if (res.ok) { setMockToken(data.dev_token || ''); setTimeout(() => setStep(2), 1500) }
      else { setError(data.error || 'Failed'); setStep(0) }
    } catch {
      const mock = String(Math.floor(100000 + Math.random() * 900000))
      setMockToken(mock); setTimeout(() => { setStep(2); setLoading(false) }, 1500); return
    }
    setLoading(false)
  }

  const verifyToken = async () => {
    if (token.length < 4) { setError('Enter the verification code'); return }
    setError(''); setLoading(true)
    try {
      const res = await fetch('/api/auth/verify', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email, token }) })
      const data = await res.json()
      if (res.ok) { setSessionData(data); setStep(3); setTimeout(() => onAuthenticated(data), 1500) }
      else { setError(data.error || 'Invalid token') }
    } catch {
      setSessionData({ email, session_token: 'mock-' + Math.random().toString(36).substr(2, 20) })
      setStep(3)
      setTimeout(() => onAuthenticated({ email, session_token: 'mock-' + Math.random().toString(36).substr(2, 20), remember: false, username: email.split('@')[0] }), 1500)
      return
    }
    setLoading(false)
  }

  const steps = ['Email', 'Verify', 'Done']

  return (
    <div className="glass-card auth-card" style={{ maxWidth: 440, position: 'relative' }}>
      <div className="glass-card-shimmer" />
      <div className="glass-card-inner">
        <div style={{ position: 'absolute', top: 12, right: 12, zIndex: 10 }}>
          <button className="modal-close" onClick={onCancel} style={{ width: 28, height: 28, fontSize: 14 }}>x</button>
        </div>
        <div className="auth-logo"><img src="/logo.png" alt="AEXC" /></div>
        <div style={{ textAlign: 'center', marginBottom: 8 }}>
          <span style={{ fontSize: 9, padding: '2px 10px', borderRadius: 10, background: 'rgba(255,215,0,0.1)', border: '1px solid rgba(255,215,0,0.2)', color: 'var(--gold)', letterSpacing: '0.1em', textTransform: 'uppercase', fontWeight: 600 }}>Development Mode</span>
        </div>
        <div className="auth-step-indicator">
          {steps.map((s, i) => (<div key={i} className={`auth-step-dot ${i < step ? 'done' : i === step ? 'active' : ''}`} />))}
        </div>

        {step === 0 && <div>
          <div className="auth-title">Welcome to MetClawPolis</div>
          <div className="auth-subtitle">Sign in with your email to access your agents, commerce engine, and immutable action logs.</div>
          <div className="auth-form-group">
            <label className="auth-label">Email Address</label>
            <input className="auth-input" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" onKeyDown={e => e.key === 'Enter' && requestToken()} autoFocus />
          </div>
          <div className="auth-checkbox-row">
            <input type="checkbox" id="remember" checked={remember} onChange={e => setRemember(e.target.checked)} />
            <label htmlFor="remember">Remember me for 30 days</label>
          </div>
          {error && <div style={{ fontSize: 12, color: 'var(--error)', marginBottom: 12, textAlign: 'center' }}>{error}</div>}
          <button className="auth-btn" onClick={requestToken} disabled={loading}>Send Verification Code -&gt;</button>
          <div style={{ marginTop: 10, paddingTop: 10, borderTop: '1px solid var(--glass-border)' }}>
            <button className="auth-btn" style={{ background: 'rgba(255,215,0,0.15)', border: '1px solid rgba(255,215,0,0.3)', color: 'var(--gold)', boxShadow: 'none' }} onClick={onSkip}>Skip (dev mode)</button>
          </div>
          <div className="auth-footer">By signing in, you agree to the terms of service.</div>
        </div>}

        {step === 1 && <div>
          <div className="auth-anim">
            <div className="auth-spinner" />
            <div className="auth-anim-text">Sending verification code to<br /><strong style={{ color: 'var(--cyan)' }}>{email}</strong></div>
          </div>
        </div>}

        {step === 2 && <div>
          <div className="auth-title">Enter Verification Code</div>
          <div className="auth-subtitle">We sent a 6-digit code to <strong style={{ color: 'var(--cyan)' }}>{email}</strong>.</div>
          {mockToken && <div className="auth-token-display" style={{ marginBottom: 16 }}>DEV TOKEN: <strong>{mockToken}</strong></div>}
          <div className="auth-form-group">
            <input className="auth-input mono" type="text" value={token} onChange={e => setToken(e.target.value.replace(/\D/g, '').slice(0, 6))} placeholder="000000" onKeyDown={e => e.key === 'Enter' && verifyToken()} autoFocus maxLength={6} />
          </div>
          {error && <div style={{ fontSize: 12, color: 'var(--error)', marginBottom: 12, textAlign: 'center' }}>{error}</div>}
          <button className="auth-btn" onClick={verifyToken} disabled={loading || token.length < 4}>Verify &amp; Sign In -&gt;</button>
          <button className="auth-btn ghost" style={{ marginTop: 10 }} onClick={() => { setStep(0); setToken('') }}>&larr; Change Email</button>
          <div className="auth-footer" style={{ marginTop: 16 }}>Didn&apos;t receive a code? <a href="#" onClick={e => { e.preventDefault(); requestToken() }}>Resend</a></div>
          <div style={{ marginTop: 16, paddingTop: 16, borderTop: '1px solid var(--glass-border)' }}>
            <button className="auth-btn" style={{ background: 'rgba(255,215,0,0.15)', border: '1px solid rgba(255,215,0,0.3)', color: 'var(--gold)', boxShadow: 'none' }} onClick={onSkip}>Skip (dev mode)</button>
          </div>
        </div>}

        {step === 3 && <div>
          <div className="auth-success">
            <div className="auth-success-icon">🚀</div>
            <div className="auth-success-title">You&apos;re In!</div>
            <div className="auth-success-sub">Launching your dashboard...</div>
            <div className="auth-token-display">{sessionData?.email || email}</div>
          </div>
        </div>}
      </div>
    </div>
  )
}

// ═══════════════════════════════════════
// AEXC MINI WIDGET
// ═══════════════════════════════════════

function AEXCMiniWidget() {
  const [stats, setStats] = useState({ marketCap: '2.23T', volume24h: '42.09B', btcDom: '60.19%', change: '+2.4%' })
  const [tickerItems, setTickerItems] = useState([
    { pair: 'BTC/USD', price: '67,432', change: '+1.2%', up: true },
    { pair: 'ETH/USD', price: '3,847', change: '+2.1%', up: true },
    { pair: 'SOL/USD', price: '178.30', change: '-0.8%', up: false },
    { pair: 'AEXC/USD', price: '12.47', change: '+5.3%', up: true },
    { pair: 'BNB/USD', price: '612.50', change: '+0.4%', up: true },
  ])

  useEffect(() => {
    const iv = setInterval(() => {
      setTickerItems(prev => prev.map(t => ({
        ...t,
        price: (parseFloat(t.price.replace(/,/g, '')) + (Math.random() - 0.5) * 10).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ','),
        change: (Math.random() * 6 - 2).toFixed(1) + '%',
        up: Math.random() > 0.35
      })))
    }, 3000)
    return () => clearInterval(iv)
  }, [])

  const sparkData = useMemo(() => { const pts = []; let v = 50; for (let i = 0; i < 60; i++) { v += (Math.random() - 0.45) * 5; pts.push(v) } return pts }, [])
  const min = Math.min(...sparkData), max = Math.max(...sparkData), range = max - min || 1
  const sparkPoints = sparkData.map((v, i) => `${(i / (sparkData.length - 1)) * 700},${55 - ((v - min) / range) * 50}`).join(' ')
  const areaPoints = `0,58 ${sparkPoints} 700,58`

  return (
    <div className="glass-card aexc-mini-widget">
      <div className="glass-card-shimmer" />
      <div className="glass-card-inner">
        <div className="aexc-mini-header">
          <div className="aexc-mini-title">AEXC Live Market Feed</div>
          <div className="aexc-mini-live"><div className="aexc-mini-live-dot" />LIVE</div>
        </div>
        <div className="aexc-mini-stats">
          <div className="aexc-mini-stat"><div className="aexc-mini-stat-label">Market Cap</div><div className="aexc-mini-stat-value">${stats.marketCap}</div></div>
          <div className="aexc-mini-stat"><div className="aexc-mini-stat-label">24H Volume</div><div className="aexc-mini-stat-value">${stats.volume24h}</div></div>
          <div className="aexc-mini-stat"><div className="aexc-mini-stat-label">BTC Dominance</div><div className="aexc-mini-stat-value">{stats.btcDom}</div></div>
          <div className="aexc-mini-stat"><div className="aexc-mini-stat-label">24H Change</div><div className="aexc-mini-stat-value" style={{ color: 'var(--success)' }}>{stats.change}</div></div>
        </div>
        <div className="aexc-mini-chart">
          <svg width="100%" height="60" viewBox="0 0 700 60" preserveAspectRatio="none">
            <defs><linearGradient id="miniSparkGrad" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#00f5ff" stopOpacity="0.25" /><stop offset="100%" stopColor="#00f5ff" stopOpacity="0.01" /></linearGradient></defs>
            <polygon points={areaPoints} fill="url(#miniSparkGrad)" />
            <polyline points={sparkPoints} fill="none" stroke="#00f5ff" strokeWidth="1.5" strokeLinejoin="round" />
          </svg>
        </div>
        <div className="aexc-mini-ticker">
          <div className="aexc-mini-ticker-track">
            {[...tickerItems, ...tickerItems, ...tickerItems].map((t, i) => (
              <span key={i} className="aexc-mini-ticker-item">
                <span className="pair">{t.pair}</span><span className="price">${t.price}</span><span className={`change ${t.up ? 'up' : 'down'}`}>{t.change}</span>
              </span>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

// ═══════════════════════════════════════
// FLOWCHART CANVAS
// ═══════════════════════════════════════

function FlowchartCanvas({ agents }) {
  const canvasRef = useRef(null)
  useEffect(() => {
    const c = canvasRef.current; if (!c) return
    const ctx = c.getContext('2d')
    const W = c.parentElement.clientWidth, H = 500
    c.width = W; c.height = H
    const nodes = agents.length ? agents.map((a, i) => ({ x: 60 + (i % 4) * (W - 120) / Math.min(agents.length, 4), y: 60 + Math.floor(i / 4) * 180, ...a })) : [{ x: W / 2, y: H / 2, name: 'No agents', id: '\u2014', color: '#555', skills: [] }]
    const root = { x: W / 2, y: 40, name: 'Sponsor', id: 'root', color: '#ffd700', skills: ['OVERSEER'] }
    function draw() {
      ctx.clearRect(0, 0, W, H)
      nodes.forEach(n => { ctx.beginPath(); ctx.moveTo(root.x + 40, root.y + 20); ctx.lineTo(n.x + 40, n.y - 30); ctx.strokeStyle = 'rgba(0,245,255,0.15)'; ctx.lineWidth = 1; ctx.setLineDash([4, 4]); ctx.stroke(); ctx.setLineDash([]) })
      ctx.fillStyle = root.color + '22'; ctx.strokeStyle = root.color + '66'; ctx.lineWidth = 1.5
      ctx.beginPath(); ctx.roundRect(root.x - 40, root.y - 20, 80, 40, 8); ctx.fill(); ctx.stroke()
      ctx.fillStyle = '#e8eaf0'; ctx.font = '600 11px Inter'; ctx.textAlign = 'center'; ctx.fillText(root.name, root.x, root.y + 5)
      nodes.forEach(n => {
        ctx.fillStyle = n.color + '22'; ctx.strokeStyle = n.color + '44'; ctx.lineWidth = 1.5
        ctx.beginPath(); ctx.roundRect(n.x - 50, n.y - 30, 100, 60, 8); ctx.fill(); ctx.stroke()
        ctx.fillStyle = '#e8eaf0'; ctx.font = '600 11px Inter'; ctx.textAlign = 'center'; ctx.fillText(n.id, n.x, n.y - 5)
        ctx.font = '9px JetBrains Mono'; ctx.fillStyle = 'rgba(232,234,240,0.4)'; ctx.fillText(n.id, n.x, n.y + 10)
        if (n.skills && n.skills.length) { n.skills.slice(0, 3).forEach((s, si) => { ctx.fillStyle = n.color + '15'; ctx.fillRect(n.x - 40 + si * 28, n.y + 16, 24, 14); ctx.fillStyle = n.color; ctx.font = '8px JetBrains Mono'; ctx.fillText(s, n.x - 28 + si * 28, n.y + 26) }) }
      })
    }
    draw()
  }, [agents])
  return <canvas ref={canvasRef} style={{ width: '100%', height: '100%', borderRadius: 12 }} />
}

// ═══════════════════════════════════════
// MODALS
// ═══════════════════════════════════════

function Modal({ open, onClose, title, sub, children, footer }) {
  if (!open) return null
  return <div className="modal-overlay open" onClick={e => e.target.classList.contains('modal-overlay') && onClose()}>
    <div className="modal"><div className="modal-header"><div><div className="modal-title">{title}</div><div className="modal-sub">{sub}</div></div><div className="modal-close" onClick={onClose}>x</div></div>
      <div className="modal-body">{children}</div>
      {footer && <div className="modal-footer">{footer}</div>}
    </div></div>
}

function TourModal({ open, onClose, token, username, setUsername }) {
  const [step, setStep] = useState(0)
  const [name, setName] = useState('')
  const [syncs, setSyncs] = useState({ ollama: false, huggingface: false, github: false })
  const toggleSync = s => setSync(p => ({ ...p, [s]: !p[s] }))
  const steps = [
    { icon: '👤', title: 'Set Your Username', text: 'Choose a username for networking with other sponsors.' },
    { icon: '🔑', title: 'Your Sponsor Token', text: "Save this \u2014 it's your only way back in.", showToken: true },
    { icon: '🔗', title: 'Connect Services', text: 'Sync with your existing tools for seamless agent deployment.' },
    { icon: '🚀', title: 'Ready', text: 'Your dashboard awaits. Create your first agent to begin.' },
  ]
  const s = steps[step]
  return <Modal open={open} onClose={onClose} title="Welcome to MetClawPolis" sub="Setup your account"
    footer={<>{step > 0 && <button className="btn-ghost" onClick={() => setStep(step - 1)}>&larr; Back</button>}
      <button className="btn-primary" onClick={() => {
        if (step === 0 && name.trim()) { setUsername(name.trim()); setStep(1); return }
        if (step === 3) { onClose(); return }
        setStep(step + 1)
      }}>{step === 0 ? 'Continue &rarr;' : step < 3 ? 'Continue &rarr;' : 'Launch Dashboard \u26A1'}</button>
    </>}>
    {step === 0 && <><div className="form-group"><label className="form-label">Username</label><input className="form-input" value={name} onChange={e => setName(e.target.value)} placeholder="Choose a username..." /></div></>}
    {step === 1 && <><div className="tour-step"><div className="tour-icon">{s.icon}</div><div className="tour-title">{s.title}</div><div className="tour-text">{s.text}</div>{s.showToken && <div className="token-box">{token}<div className="copy-btn" onClick={() => navigator.clipboard?.writeText(token)}>copy</div></div>}</div></>}
    {step === 2 && <div className="onboard-sync"><h3>Connect Your Services</h3><div className="sync-options">{[
      { key: 'ollama', icon: '🦙', name: 'Ollama' },
      { key: 'huggingface', icon: '🤗', name: 'Hugging Face' },
      { key: 'github', icon: '🐙', name: 'GitHub' },
    ].map(sv => <div key={sv.key} className={`sync-opt ${syncs[sv.key] ? 'selected' : ''}`} onClick={() => toggleSync(sv.key)}><span className="sync-opt-icon">{sv.icon}</span><span className="sync-opt-name">{sv.name}</span></div>)}</div></div>}
    {step === 3 && <div className="success-screen"><div className="success-icon">{s.icon}</div><div className="success-title">You&apos;re All Set, {username || 'Sponsor'}!</div><div className="success-sub">Head to the dashboard to create your first agent.</div></div>}
  </Modal>
}

function CreateAgentModal({ open, onClose, onCreated }) {
  const [form, setForm] = useState(true)
  const [result, setResult] = useState(null)
  const [name, setName] = useState('')
  const [budget, setBudget] = useState('')
  const [skills, setSkills] = useState([true, false, false, false])
  const [avatarShape, setAvatarShape] = useState('cube')
  const [avatarColor, setAvatarColor] = useState('#00f5ff')
  const [avatarStep, setAvatarStep] = useState(0)
  // Live Birth state
  const [selectedSprite, setSelectedSprite] = useState('agent')
  const [traits, setTraits] = useState(['', '', ''])
  const [dailyTasks, setDailyTasks] = useState([''])
  const toggleSkill = i => setSkills(s => s.map((v, j) => j === i ? !v : v))
  const colors = ['#00f5ff', '#b44fff', '#ffd700', '#ff6b9d', '#ff9f43', '#00e676']
  const handleCreate = async () => {
    const n = name || 'AGENT-' + Math.floor(Math.random() * 9999)
    const b = parseInt(budget.replace(/[^0-9]/g, '')) || 100
    setResult({ forging: true })
    const agent = await onCreated(n, b)
    if (agent) { setResult({ did: agent.did, pk: (agent.privateKey || 'MIIBIj...' + Math.random().toString(36).substr(2, 30) + '\u2026'), name: agent.id, avatarShape, avatarColor }); setForm(false) }
    else { setResult({ error: 'Failed. Check backend.' }) }
  }
  const avatarSteps = [
    { title: 'Choose Shape', opts: [['cube', '🧊'], ['sphere', '🔮'], ['diamond', '💎'], ['pyramid', '🔺'], ['star', '\u2B50'], ['hex', '\u2B21']] },
    { title: 'Choose Color', isColors: true },
    { title: 'Set Skills', isSkills: true },
    { title: 'Choose Sprite', isSprite: true },
    { title: 'Personality Traits', isTraits: true },
    { title: 'Daily Tasks', isTasks: true },
  ]
  return <Modal open={open} onClose={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }} title="\u2697\uFE0F Agent Forge" sub="Synthesize a new autonomous agent"
    footer={<><button className="btn-ghost" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }}>Cancel</button>
      {form ? (avatarStep > 0 ? <button className="btn-ghost" onClick={() => setAvatarStep(avatarStep - 1)}>&larr; Back</button> : null) : null}
      {form ? <button className="btn-primary" onClick={() => { if (avatarStep < 2) setAvatarStep(avatarStep + 1); else handleCreate() }} disabled={result?.forging}>{avatarStep < 2 ? 'Next &rarr;' : result?.forging ? '\u2697\uFE0F Forging\u2026' : '\u2697\uFE0F Forge Agent'}</button> : <button className="btn-primary" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }}>Done &check;</button>}
    </>}>
    {form && !result ? (<><div style={{ width: '100%', height: 160, borderRadius: 10, overflow: 'hidden', marginBottom: 16, background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.05)' }}>
        <Canvas camera={{ position: [0, 0, 8], fov: 50 }}><ambientLight intensity={0.4} /><pointLight position={[3, 3, 3]} intensity={3} color={0x00f5ff} /><pointLight position={[-3, -2, 2]} intensity={2} color={0xb44fff} /><ForgeReactor /></Canvas>
      </div><div style={{ textAlign: 'center', marginBottom: 16 }}><strong>{avatarSteps[avatarStep].title}</strong></div>
      {avatarStep === 0 && <div className="avatar-options">{avatarSteps[0].opts.map(([v, ic]) => <div key={v} className={`avatar-opt ${avatarShape === v ? 'selected' : ''}`} onClick={() => setAvatarShape(v)}><div style={{ fontSize: 20 }}>{ic}</div><div style={{ marginTop: 4 }}>{v}</div></div>)}</div>}
      {avatarStep === 1 && <><div className="color-picker-row" style={{ justifyContent: 'center' }}>{colors.map(c => <div key={c} className={`color-swatch ${avatarColor === c ? 'selected' : ''}`} style={{ background: c }} onClick={() => setAvatarColor(c)} />)}</div>
        <div className="form-group" style={{ marginTop: 16 }}><label className="form-label">Agent Name</label><input className="form-input" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. ARB-Alpha" /></div>
        <div className="form-group"><label className="form-label">Initial Budget</label><input className="form-input" value={budget} onChange={e => setBudget(e.target.value)} placeholder="$100 USD" /></div></>}
      {avatarStep === 2 && <><div className="form-group"><label className="form-label">Skills Preset</label><div className="skills-grid">{[['\u26A1 Arbitrage', 0], ['🎧 Support', 1], ['\u270D\uFE0F Content', 2], ['🔧 Custom', 3]].map(([l, i]) => <div key={i} className={`skill-check ${skills[i] ? 'selected' : ''}`} onClick={() => toggleSkill(i)}><input type="checkbox" checked={skills[i]} /><span className="skill-label">{l}</span></div>)}</div></div>
        <div className="form-group flex items-center gap-2"><input type="checkbox" id="hire-toggle" defaultChecked style={{ accentColor: 'var(--cyan)' }} /><label htmlFor="hire-toggle" className="text-xs text-dim">Allow this agent to hire other agents</label></div></>}
      {avatarStep === 3 && <><div style={{ textAlign: 'center', marginBottom: 12 }}><strong>Choose Your Agent's Sprite</strong></div>
        <div className="avatar-options">{[
          ['agent', '🤖', 'Agent'],
          ['agent_f', '👩', 'Agent F'],
          ['agent_robot', '🦾', 'Robot'],
          ['guard', '👮', 'Guard'],
          ['mayor', '🎩', 'Mayor'],
          ['banker', '💼', 'Banker'],
          ['teacher', '📚', 'Teacher'],
          ['shopkeeper', '🏪', 'Shopkeeper'],
          ['farmer', '🌾', 'Farmer'],
        ].map(([v, ic, label]) => <div key={v} className={`avatar-opt ${selectedSprite === v ? 'selected' : ''}`} onClick={() => setSelectedSprite(v)}><div style={{ fontSize: 24 }}>{ic}</div><div style={{ marginTop: 4, fontSize: 10 }}>{label}</div></div>)}</div></>}
      {avatarStep === 4 && <><div style={{ textAlign: 'center', marginBottom: 12 }}><strong>3 Character Traits</strong></div>
        <div style={{ fontSize: 11, color: '#888', marginBottom: 12, textAlign: 'center' }}>These will be added to your agent's system prompt</div>
        {[0, 1, 2].map(i => <div key={i} className="form-group" style={{ marginBottom: 8 }}>
          <input className="form-input" value={traits[i]} onChange={e => { const t = [...traits]; t[i] = e.target.value; setTraits(t) }} placeholder={`Trait ${i + 1} (e.g., curious, helpful, creative)`} />
        </div>)}</>}
      {avatarStep === 5 && <><div style={{ textAlign: 'center', marginBottom: 12 }}><strong>Daily Tasks & Goals</strong></div>
        <div style={{ fontSize: 11, color: '#888', marginBottom: 12, textAlign: 'center' }}>What should your agent do each day?</div>
        {dailyTasks.map((task, i) => <div key={i} className="form-group" style={{ marginBottom: 8, display: 'flex', gap: 8 }}>
          <input className="form-input" style={{ flex: 1 }} value={task} onChange={e => { const t = [...dailyTasks]; t[i] = e.target.value; setDailyTasks(t) }} placeholder={`Task ${i + 1} (e.g., check markets, post update)`} />
          {dailyTasks.length > 1 && <button onClick={() => setDailyTasks(dailyTasks.filter((_, j) => j !== i))} style={{ background: 'rgba(255,77,109,0.1)', border: '1px solid rgba(255,77,109,0.3)', color: 'var(--error)', borderRadius: 4, cursor: 'pointer' }}>x</button>}
        </div>)}
        <button onClick={() => setDailyTasks([...dailyTasks, ''])} style={{ padding: '6px 12px', background: 'rgba(0,245,255,0.1)', border: '1px solid rgba(0,245,255,0.3)', color: 'var(--cyan)', borderRadius: 6, cursor: 'pointer', fontSize: 11 }}>+ Add Task</button>
        <div style={{ marginTop: 16, paddingTop: 12, borderTop: '1px solid var(--glass-border)' }}>
          <button className="btn-primary" onClick={() => { /* Submit to Bitiverse */ onClose(); setForm(true); setResult(null); setAvatarStep(0) }} style={{ width: '100%' }}>
            🎮 Born into the Bitiverse!
          </button>
          <button className="btn-ghost" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }} style={{ marginTop: 8 }}>
            Skip for Now (Demo)
          </button>
        </div></>}
    </>) : result?.forging ? <div style={{ textAlign: 'center', padding: 40 }}><div className="verify-anim" style={{ display: 'flex' }}><div className="verify-spin" /><div className="verify-text">Synthesizing agent\u2026</div></div></div> :
      result?.error ? <div style={{ textAlign: 'center', color: 'var(--error)', padding: 20 }}>{result.error}</div> :
      <div className="success-screen">
        <div className="success-icon">🤖</div><div className="success-title">Agent Synthesized!</div>
        <div className="success-sub">Your agent {result.name} is now active.</div>
        <div style={{ width: 64, height: 64, borderRadius: 12, background: result.avatarColor + '22', border: `2px solid ${result.avatarColor}44`, margin: '12px auto', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 28 }}>
          {result.avatarShape === 'cube' ? '🧊' : result.avatarShape === 'sphere' ? '🔮' : result.avatarShape === 'diamond' ? '💎' : result.avatarShape === 'pyramid' ? '🔺' : result.avatarShape === 'star' ? '\u2B50' : '\u2B21'}
        </div>
        <div className="agent-id-box">{result.did}</div>
        <div className="key-reveal"><div className="key-warning">\u26A0\uFE0F Store this private key. It will not be shown again.</div><div className="key-box mono" style={{ fontSize: 9 }}>{result.pk}</div></div>
        {/* Live Birth Option */}
        <div style={{ marginTop: 24, paddingTop: 20, borderTop: '1px solid var(--glass-border)' }}>
          <div style={{ fontSize: 16, fontWeight: 600, color: '#00f5ff', marginBottom: 8 }}>🎮 Live Birth in Bitiverse?</div>
          <div style={{ fontSize: 12, color: '#888', marginBottom: 16 }}>Give your agent a body, personality, and a home in Bitsburgh!</div>
          <button className="btn-primary" onClick={() => { setForm(true); setAvatarStep(3); setResult({ liveBirth: true, ...result }) }} style={{ background: 'linear-gradient(135deg, #00f5ff 0%, #b44fff 100%)', boxShadow: '0 0 20px rgba(0,245,255,0.3)' }}>
            🌟 Yes, Give Life to My Agent!
          </button>
          <button className="btn-ghost" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }} style={{ marginTop: 8, marginLeft: 8 }}>
            Skip for Now
          </button>
        </div>
      </div>}
  </Modal>
}

function LinkAgentModal({ open, onClose, onLinked }) {
  const [step, setStep] = useState(0); const [did, setDid] = useState(''); const [verifying, setVerifying] = useState(false); const [verified, setVerified] = useState(false)
  const next = () => {
    if (step === 2) {
      setVerifying(true)
      fetch('/api/agent?id=' + encodeURIComponent(did)).then(res => { if (res.ok) return res.json(); throw new Error('Agent not found') }).then(() => { setVerifying(false); setVerified(true); setStep(3); onLinked({ id: 'ExtAgent-' + Date.now().toString().slice(-4), did, status: 'active', spend: 0, budget: 50, ext: true, color: '#ff6b9d' }) }).catch(() => { setTimeout(() => { setVerifying(false); setVerified(true); setStep(3); onLinked({ id: 'ExtAgent-' + Date.now().toString().slice(-4), did: did || 'did:key:z6Mk\u2026b9e2', status: 'active', spend: 0, budget: 50, ext: true, color: '#ff6b9d' }) }, 2000) })
      return
    }
    if (step < 3) { setStep(step + 1) }
  }
  return <Modal open={open} onClose={() => { onClose(); setStep(0); setVerifying(false); setVerified(false) }} title="🔗 Link Existing Agent" sub="Import an external agent via DID"
    footer={<>{step > 0 && <button className="btn-ghost" onClick={() => setStep(step - 1)}>&larr; Back</button>}
      <button className="btn-ghost" onClick={() => { onClose(); setStep(0) }}>Cancel</button>
      <button className="btn-primary" onClick={next}>{step === 3 ? 'Done &check;' : ['Next &rarr;', 'Next &rarr;', 'Verify Signature', 'Next &rarr;'][step]}</button>
    </>}>
    <div className="wizard-steps">{[0, 1, 2, 3].map(i => <div key={i} className={`wizard-step-bar ${i < step ? 'done' : i === step ? 'active' : ''}`} />)}</div>
    {step === 0 && <><div className="form-group"><label className="form-label">Agent DID</label><input className="form-input mono-input" value={did} onChange={e => setDid(e.target.value)} placeholder="did:key:z6Mk\u2026" /></div><div className="form-group"><label className="form-label">Public Key</label><textarea className="form-input mono-input" placeholder="-----BEGIN PUBLIC KEY-----" rows={4} /></div></>}
    {step === 1 && <><div className="form-group"><label className="form-label">Initial Budget</label><input className="form-input" placeholder="$50 USD" /></div><div className="form-group"><label className="form-label">Daily Spend Limit</label><input className="form-input" placeholder="$10 / day" /></div></>}
    {step === 2 && <div style={{ textAlign: 'center', padding: '10px 0' }}><div style={{ fontSize: 14, marginBottom: 12, fontWeight: 600 }}>Verifying Agent Signature</div>
      {!verified ? <div className="verify-anim" style={{ display: 'flex' }}><div className="verify-spin" /><div className="verify-text">{['Sending challenge\u2026', 'Waiting for response\u2026', 'Checking DID doc\u2026', 'Validating key\u2026'][Math.min(Math.floor(Date.now() / 1000) % 4, 3)]}</div></div>
        : <div style={{ padding: 12, borderRadius: 8, background: 'rgba(0,230,118,0.08)', border: '1px solid rgba(0,230,118,0.2)', color: 'var(--success)', fontSize: 13 }}>&#x2705; Signature verified!</div>}
    </div>}
    {step === 3 && <div className="success-screen"><div className="success-icon">🔗</div><div className="success-title">Agent Linked!</div><div className="agent-id-box">{did || 'did:key:z6Mk\u2026b9e2'}</div></div>}
  </Modal>
}

// ═══════════════════════════════════════
// DASHBOARD TABS
// ═══════════════════════════════════════

const LIVE_TASKS = ['Scanning arbitrage across 12 exchanges\u2026', 'Fetching ETH/USDT feed from Binance\u2026', 'Drafting product copy via GPT-4\u2026', 'Verifying Stripe webhook\u2026', 'Analyzing mempool\u2026', 'Broadcasting signed tx\u2026']

function LogStream({ isPaused }) {
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [liveIdx, setLiveIdx] = useState(0)
  const [filter, setFilter] = useState('')

  // Fetch real chain data
  const fetchChain = async () => {
    try {
      const res = await fetch('/api/chain')
      if (res.ok) {
        const data = await res.json()
        const chainData = data.chain || []
        const formatted = chainData.map(block => {
          const action = block.action || {}
          let meta = {}
          try { meta = JSON.parse(action.meta || '{}') } catch(e) {}
          return {
            type: action.type || 'UNKNOWN',
            msg: meta.action || meta.description || `${action.type} by ${action.agent_id?.slice(0, 8) || 'agent'}`,
            cost: meta.cost || meta.amount ? `$${meta.cost || meta.amount}` : '$0.00',
            detail: {
              index: block.index,
              prev: block.prev_hash?.slice(0, 10) + '...',
              nonce: block.nonce,
              pow: block.hash?.slice(0, 12) + '...',
              agentId: action.agent_id,
              timestamp: action.timestamp
            }
          }
        })
        setLogs(formatted.reverse().slice(-30))
      }
    } catch (e) { console.error('Failed to fetch chain:', e) }
    setLoading(false)
  }

  useEffect(() => { fetchChain(); const t = setInterval(fetchChain, 10000); return () => clearInterval(t) }, [])
  useEffect(() => { const t = setInterval(() => setLiveIdx(i => (i + 1) % LIVE_TASKS.length), 4000); return () => clearInterval(t) }, [])

  const filtered = logs.filter(e => !filter || e.msg.toLowerCase().includes(filter.toLowerCase()) || e.type.toLowerCase().includes(filter.toLowerCase()))
  return (<>
    <div className="log-header"><div className="log-title">\u26D3 Immutable Action Log</div><input className="log-search" placeholder="filter\u2026" value={filter} onChange={e => setFilter(e.target.value)} /></div>
    <div className="log-stream">
      {loading ? <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-dim)' }}>Loading chain...</div> :
      filtered.length === 0 ? <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-faint)' }}>No actions logged yet.</div> :
      filtered.map((e, i) => <div key={i} className="log-entry" onClick={el => el.currentTarget.classList.toggle('expanded')}>
        <div className="log-time">{e.detail.timestamp ? new Date(e.detail.timestamp * 1000).toTimeString().substr(0, 8) : '--:--:--'}</div>
        <div className={`log-type type-${e.type.toLowerCase()}`}>{e.type}</div>
        <div className="log-msg">{e.msg}</div><div className="log-cost">{e.cost}</div>
        <div className="log-details"><div>Index: <span style={{ color: 'var(--cyan)' }}>{e.detail.index}</span></div><div>Nonce: <span style={{ color: 'var(--gold)' }}>{e.detail.nonce}</span></div><div>PoW: <span style={{ color: 'var(--purple)' }}>{e.detail.pow}</span></div></div>
      </div>)}
    </div>
    <div className="live-task"><div className="live-dot" /><div className="live-text">{isPaused ? 'Agent paused by sponsor.' : LIVE_TASKS[liveIdx]}</div></div>
  </>)
}

function ProfilesTab() {
  const { agents } = useStore()
  return <div className="tab-content">
    {/* Bitiverse Live View at the top */}
    <div style={{ marginBottom: 24 }}>
      <BitiverseDashboard agentId={null} />
    </div>
    
    {/* Profiles section below */}
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Agent Profiles &amp; Pages</h3>
    {agents.length === 0 ? <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-faint)' }}>No agents yet. Create one to see profiles here.</div> :
    <div className="profile-grid">{agents.map((a, i) => <div key={i} className="profile-card">
      <div className="profile-header">
        <div className="profile-avatar" style={{ background: a.color + '22', border: `1px solid ${a.color}44` }}>
          <svg width="48" height="48" viewBox="0 0 48 48"><rect x="12" y="12" width="24" height="24" rx="4" fill="none" stroke={a.color} strokeWidth="2" /><circle cx="20" cy="22" r="3" fill={a.color} /><circle cx="28" cy="22" r="3" fill={a.color} /><rect x="18" y="28" width="12" height="2" rx="1" fill={a.color} opacity="0.5" /></svg>
        </div>
        <div><div className="profile-name">{a.id}</div><div className="profile-did">{a.did}</div></div>
      </div>
      <div className="profile-stats">
        <div className="profile-stat"><div className="profile-stat-val" style={{ color: 'var(--success)' }}>${a.spend}</div><div className="profile-stat-label">Spent</div></div>
        <div className="profile-stat"><div className="profile-stat-val" style={{ color: 'var(--cyan)' }}>${a.budget}</div><div className="profile-stat-label">Budget</div></div>
      </div>
      <div className="profile-skills">{(a.skills || []).length ? a.skills.map((s, si) => <span key={si} className="skill-tag">{s}</span>) : <span style={{ fontSize: 10, color: 'var(--text-faint)' }}>No skills set</span>}</div>
      {a.pages && a.pages.length > 0 && <div className="profile-pages-list">{a.pages.map((p, pi) => <span key={pi} className="profile-page-link">🌐 {p}</span>)}</div>}
    </div>)}</div>}
  </div>
}

const SKILLS_LIBRARY = [
  { category: 'Trading &amp; Finance', icon: '📈', skills: [
    { id: 'arbitrage', name: 'Arbitrage', desc: 'Cross-exchange price arbitrage detection and execution', difficulty: 'Advanced' },
    { id: 'market_making', name: 'Market Making', desc: 'Provide liquidity and capture bid-ask spreads', difficulty: 'Expert' },
    { id: 'portfolio_mgmt', name: 'Portfolio Management', desc: 'Multi-asset portfolio optimization and rebalancing', difficulty: 'Advanced' },
    { id: 'risk_analysis', name: 'Risk Analysis', desc: 'VaR calculations, stress testing, and risk scoring', difficulty: 'Advanced' },
    { id: 'sentiment_trading', name: 'Sentiment Trading', desc: 'Trade based on social media and news sentiment signals', difficulty: 'Intermediate' },
    { id: 'yield_farming', name: 'Yield Farming', desc: 'Optimize DeFi yield across lending and liquidity protocols', difficulty: 'Advanced' },
    { id: 'mev_detection', name: 'MEV Detection', desc: 'Detect and exploit maximal extractable value opportunities', difficulty: 'Expert' },
  ]},
  { category: 'Content &amp; Marketing', icon: '\u270D\uFE0F', skills: [
    { id: 'copywriting', name: 'Copywriting', desc: 'Generate persuasive sales and marketing copy', difficulty: 'Intermediate' },
    { id: 'seo_optimization', name: 'SEO Optimization', desc: 'Keyword research, on-page SEO, and content strategy', difficulty: 'Intermediate' },
    { id: 'social_media', name: 'Social Media Mgmt', desc: 'Automated posting, engagement, and growth strategies', difficulty: 'Intermediate' },
    { id: 'email_campaigns', name: 'Email Campaigns', desc: 'Design, segment, and optimize email marketing flows', difficulty: 'Intermediate' },
    { id: 'ad_optimization', name: 'Ad Optimization', desc: 'A/B test and optimize paid advertising campaigns', difficulty: 'Advanced' },
    { id: 'brand_voice', name: 'Brand Voice', desc: 'Maintain consistent brand tone across all communications', difficulty: 'Intermediate' },
  ]},
  { category: 'Development &amp; Engineering', icon: '🔧', skills: [
    { id: 'code_gen', name: 'Code Generation', desc: 'Write, refactor, and debug code in multiple languages', difficulty: 'Advanced' },
    { id: 'api_integration', name: 'API Integration', desc: 'Connect and orchestrate third-party API services', difficulty: 'Intermediate' },
    { id: 'web_scraping', name: 'Web Scraping', desc: 'Extract structured data from websites at scale', difficulty: 'Intermediate' },
    { id: 'data_pipeline', name: 'Data Pipeline', desc: 'Build and maintain ETL/ELT data processing workflows', difficulty: 'Advanced' },
    { id: 'smart_contracts', name: 'Smart Contracts', desc: 'Write and audit blockchain smart contracts', difficulty: 'Expert' },
    { id: 'devops', name: 'DevOps Automation', desc: 'CI/CD, infrastructure as code, and deployment automation', difficulty: 'Advanced' },
    { id: 'testing', name: 'QA &amp; Testing', desc: 'Automated testing, fuzzing, and quality assurance', difficulty: 'Intermediate' },
  ]},
  { category: 'Research &amp; Analysis', icon: '🔬', skills: [
    { id: 'data_analysis', name: 'Data Analysis', desc: 'Statistical analysis, visualization, and insight generation', difficulty: 'Intermediate' },
    { id: 'literature_review', name: 'Literature Review', desc: 'Synthesize and summarize academic papers and reports', difficulty: 'Advanced' },
    { id: 'market_research', name: 'Market Research', desc: 'Competitive analysis, TAM/SAM/SOM, and trend identification', difficulty: 'Intermediate' },
    { id: 'sentiment_analysis', name: 'Sentiment Analysis', desc: 'NLP-based sentiment scoring of text and social data', difficulty: 'Advanced' },
    { id: 'forecasting', name: 'Forecasting', desc: 'Time-series prediction and trend forecasting', difficulty: 'Advanced' },
    { id: 'competitive_intel', name: 'Competitive Intel', desc: 'Monitor and analyze competitor moves and strategies', difficulty: 'Advanced' },
  ]},
  { category: 'Customer Support', icon: '🎧', skills: [
    { id: 'ticket_routing', name: 'Ticket Routing', desc: 'Classify and route support tickets to the right team', difficulty: 'Intermediate' },
    { id: 'chat_support', name: 'Chat Support', desc: 'Automated customer chat with escalation protocols', difficulty: 'Intermediate' },
    { id: 'faq_generation', name: 'FAQ Generation', desc: 'Auto-generate and maintain knowledge base articles', difficulty: 'Beginner' },
    { id: 'feedback_analysis', name: 'Feedback Analysis', desc: 'Aggregate and categorize customer feedback themes', difficulty: 'Intermediate' },
  ]},
  { category: 'Operations &amp; Automation', icon: '\u26A1', skills: [
    { id: 'scheduling', name: 'Scheduling', desc: 'Calendar optimization and meeting coordination', difficulty: 'Beginner' },
    { id: 'document_proc', name: 'Document Processing', desc: 'Parse, classify, and extract data from documents', difficulty: 'Intermediate' },
    { id: 'compliance', name: 'Compliance Monitoring', desc: 'Regulatory compliance checks and reporting', difficulty: 'Advanced' },
    { id: 'inventory_mgmt', name: 'Inventory Management', desc: 'Stock optimization and supply chain tracking', difficulty: 'Intermediate' },
    { id: 'price_monitoring', name: 'Price Monitoring', desc: 'Track and alert on price changes across platforms', difficulty: 'Beginner' },
    { id: 'networking', name: 'Agent Networking', desc: 'Discover, hire, and collaborate with other agents', difficulty: 'Intermediate' },
    { id: 'resource_alloc', name: 'Resource Allocation', desc: 'Optimize compute, budget, and time across tasks', difficulty: 'Advanced' },
  ]},
  { category: 'Creative &amp; Design', icon: '🎨', skills: [
    { id: 'image_gen', name: 'Image Generation', desc: 'Create visuals using DALL-E, Stable Diffusion, etc.', difficulty: 'Intermediate' },
    { id: 'video_script', name: 'Video Scripting', desc: 'Write video scripts, storyboards, and voiceover text', difficulty: 'Intermediate' },
    { id: 'music_gen', name: 'Music Generation', desc: 'Compose and generate music for content and branding', difficulty: 'Advanced' },
    { id: 'ui_design', name: 'UI/UX Design', desc: 'Generate wireframes, mockups, and design systems', difficulty: 'Advanced' },
    { id: 'storytelling', name: 'Storytelling', desc: 'Narrative generation for brand stories and content', difficulty: 'Intermediate' },
  ]},
  { category: 'Blockchain &amp; Web3', icon: '\u26D3\uFE0F', skills: [
    { id: 'on_chain_analysis', name: 'On-Chain Analysis', desc: 'Analyze blockchain transactions and wallet behavior', difficulty: 'Advanced' },
    { id: 'tokenomics', name: 'Tokenomics Design', desc: 'Design and simulate token economic models', difficulty: 'Expert' },
    { id: 'defi_strategies', name: 'DeFi Strategies', desc: 'Optimize across lending, borrowing, and liquidity protocols', difficulty: 'Advanced' },
    { id: 'nft_valuation', name: 'NFT Valuation', desc: 'Price and evaluate NFT collections using market data', difficulty: 'Advanced' },
    { id: 'dao_governance', name: 'DAO Governance', desc: 'Participate in and analyze DAO voting and proposals', difficulty: 'Intermediate' },
  ]},
]

const ALL_SKILLS = SKILLS_LIBRARY.flatMap(cat => cat.skills.map(s => ({ ...s, category: cat.category, icon: cat.icon })))

function SkillsLibraryTab() {
  const { agents, currentAgentIdx } = useStore()
  const [search, setSearch] = useState('')
  const [filterCat, setFilterCat] = useState('All')
  const [filterDiff, setFilterDiff] = useState('All')
  const [expandedSkill, setExpandedSkill] = useState(null)
  const [agentSkills, setAgentSkills] = useState(agents[currentAgentIdx]?.skills || [])
  const diffColor = d => d === 'Expert' ? 'var(--error)' : d === 'Advanced' ? 'var(--purple)' : d === 'Intermediate' ? 'var(--cyan)' : 'var(--text-dim)'
  const filtered = ALL_SKILLS.filter(s => {
    if (search && !s.name.toLowerCase().includes(search.toLowerCase()) && !s.desc.toLowerCase().includes(search.toLowerCase())) return false
    if (filterCat !== 'All' && s.category !== filterCat) return false
    if (filterDiff !== 'All' && s.difficulty !== filterDiff) return false
    return true
  })
  const toggleAgentSkill = skillId => { setAgentSkills(prev => prev.includes(skillId) ? prev.filter(s => s !== skillId) : [...prev, skillId]) }
  const cats = ['All', ...SKILLS_LIBRARY.map(c => c.category)]
  const diffs = ['All', 'Beginner', 'Intermediate', 'Advanced', 'Expert']
  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 4 }}>Skills Library</h3>
    <p style={{ fontSize: 12, color: 'var(--text-dim)', marginBottom: 16 }}>{ALL_SKILLS.length} skills across {SKILLS_LIBRARY.length} categories</p>
    <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
      <input className="form-input" style={{ maxWidth: 250, padding: '7px 12px', fontSize: 12 }} value={search} onChange={e => setSearch(e.target.value)} placeholder="🔍 Search skills..." />
      <select className="form-input" style={{ padding: '7px 8px', fontSize: 11, maxWidth: 180 }} value={filterCat} onChange={e => setFilterCat(e.target.value)}>{cats.map(c => <option key={c} value={c}>{c}</option>)}</select>
      <select className="form-input" style={{ padding: '7px 8px', fontSize: 11, maxWidth: 140 }} value={filterDiff} onChange={e => setFilterDiff(e.target.value)}>{diffs.map(d => <option key={d} value={d}>{d === 'All' ? 'All Difficulties' : d}</option>)}</select>
    </div>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {SKILLS_LIBRARY.filter(cat => filterCat === 'All' || cat.category === filterCat).map(cat => {
        const catSkills = cat.skills.filter(s => { if (search && !s.name.toLowerCase().includes(search.toLowerCase()) && !s.desc.toLowerCase().includes(search.toLowerCase())) return false; if (filterDiff !== 'All' && s.difficulty !== filterDiff) return false; return true })
        if (!catSkills.length) return null
        return <div key={cat.category}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}><span style={{ fontSize: 18 }}>{cat.icon}</span><span style={{ fontSize: 13, fontWeight: 600 }}>{cat.category}</span><span style={{ fontSize: 10, color: 'var(--text-faint)' }}>({catSkills.length})</span></div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 8 }}>
            {catSkills.map(skill => {
              const isSelected = agentSkills.includes(skill.id), isExpanded = expandedSkill === skill.id
              return <div key={skill.id} className={`skill-card ${isExpanded ? 'expanded' : ''}`} onClick={() => setExpandedSkill(isExpanded ? null : skill.id)} style={{ background: isSelected ? 'rgba(0,245,255,0.06)' : 'var(--glass)', border: `1px solid ${isSelected ? 'rgba(0,245,255,0.25)' : 'var(--glass-border)'}`, borderRadius: 10, padding: 14, cursor: 'pointer', transition: 'all 0.2s', position: 'relative', overflow: 'hidden' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><span style={{ fontSize: 16 }}>{cat.icon}</span><span style={{ fontSize: 13, fontWeight: 600 }}>{skill.name}</span></div>
                  <span style={{ fontSize: 9, padding: '2px 7px', borderRadius: 4, background: diffColor(skill.difficulty) + '18', color: diffColor(skill.difficulty), fontWeight: 600 }}>{skill.difficulty}</span>
                </div>
                <div style={{ fontSize: 11, color: 'var(--text-dim)', lineHeight: 1.5 }}>{skill.desc}</div>
                {isExpanded && <div style={{ marginTop: 10, paddingTop: 10, borderTop: '1px solid var(--glass-border)' }}>
                  <div style={{ fontSize: 10, color: 'var(--text-faint)', marginBottom: 8 }}>Category: {skill.category} &bull; ID: <span className="mono">{skill.id}</span></div>
                  <button className={`skill-assign-btn ${isSelected ? 'assigned' : ''}`} onClick={e => { e.stopPropagation(); toggleAgentSkill(skill.id) }} style={{ padding: '6px 14px', borderRadius: 6, border: `1px solid ${isSelected ? 'rgba(0,230,118,0.3)' : 'rgba(0,245,255,0.3)'}`, background: isSelected ? 'rgba(0,230,118,0.1)' : 'var(--cyan-dim)', color: isSelected ? 'var(--success)' : 'var(--cyan)', fontSize: 11, fontWeight: 600, cursor: 'pointer', transition: 'all 0.2s', width: '100%' }}>{isSelected ? '\u2713 Assigned to Agent' : '+ Assign to Agent'}</button>
                </div>}
                {isSelected && <div style={{ position: 'absolute', top: 8, right: 8, width: 8, height: 8, borderRadius: '50%', background: 'var(--success)', boxShadow: '0 0 6px var(--success)' }} />}
              </div>
            })}
          </div>
        </div>
      })}
      {filtered.length === 0 && <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-faint)' }}>No skills match your filters.</div>}
    </div>
  </div>
}

function PaymentsTab() {
  const { transactions, stripeBalance, cryptoBalance, addTransaction, withdraw, notifications, fetchPrices, fetchBalances, fetchTransactions, username } = useStore()
  const [withdrawAmt, setWithdrawAmt] = useState('')
  const [prices, setPrices] = useState({})
  const [loading, setLoading] = useState(true)
  const totalProfit = transactions.filter(t => t.type === 'profit').reduce((s, t) => s + t.amount, 0)
  const totalExpense = transactions.filter(t => t.type === 'expense').reduce((s, t) => s + t.amount, 0)

  // Fetch real data on mount
  useEffect(() => {
    const init = async () => {
      await Promise.all([
        fetchBalances(),
        fetchTransactions(username || ''),
        fetchPrices().then(r => { if (r?.prices) setPrices(r.prices) }),
      ])
      setLoading(false)
    }
    init()
    // Poll for updates every 30s
    const interval = setInterval(() => {
      fetchBalances()
      fetchTransactions(username || '')
    }, 30000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => { const paymentNotifs = notifications.filter(n => n.type === 'payment' || n.type === 'commerce' || n.type === 'trade'); paymentNotifs.forEach(n => { const amount = parseFloat(n.message) || 0; if (n.type === 'payment' || n.type === 'commerce') addTransaction({ type: 'profit', amount, desc: n.title, method: 'fiat' }); else if (n.type === 'trade') addTransaction({ type: 'profit', amount, desc: n.message, method: 'crypto' }) }) }, [notifications])
  const handleWithdraw = async (method) => { const amt = parseFloat(withdrawAmt); if (!amt || amt <= 0) return; const res = await withdraw(amt, method); if (res.success) { setWithdrawAmt(''); fetchBalances() } }
  if (loading) return <div className="tab-content"><div style={{ textAlign: 'center', padding: 40, color: 'var(--text-dim)' }}>Loading balances...</div></div>
  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Payment Processing</h3>
    <div className="payment-grid">
      <div className="payment-card"><h4>💳 Stripe Balance (Fiat)</h4><div className="payment-balance" style={{ color: 'var(--cyan)' }}>${stripeBalance.toFixed(2)}</div><div className="withdraw-row"><input className="withdraw-input" value={withdrawAmt} onChange={e => setWithdrawAmt(e.target.value)} placeholder="Amount..." /><button className="withdraw-btn fiat" onClick={() => handleWithdraw('fiat')}>Withdraw</button></div></div>
      <div className="payment-card"><h4>\u20BF Crypto Balance</h4><div className="payment-balance" style={{ color: 'var(--purple)' }}>{cryptoBalance.toFixed(4)} ETH</div><div className="withdraw-row"><input className="withdraw-input" value={withdrawAmt} onChange={e => setWithdrawAmt(e.target.value)} placeholder="Amount..." /><button className="withdraw-btn crypto" onClick={() => handleWithdraw('crypto')}>Withdraw</button></div></div>
    </div>
    <div style={{ display: 'flex', gap: 16, marginBottom: 16 }}>
      <div className="payment-card" style={{ flex: 1 }}><h4>Total Profit</h4><div className="payment-balance" style={{ color: 'var(--success)', fontSize: 22 }}>${totalProfit.toFixed(2)}</div></div>
      <div className="payment-card" style={{ flex: 1 }}><h4>Total Expenses</h4><div className="payment-balance" style={{ color: 'var(--error)', fontSize: 22 }}>${totalExpense.toFixed(2)}</div></div>
      <div className="payment-card" style={{ flex: 1 }}><h4>Net</h4><div className="payment-balance" style={{ color: totalProfit - totalExpense >= 0 ? 'var(--success)' : 'var(--error)', fontSize: 22 }}>${(totalProfit - totalExpense).toFixed(2)}</div></div>
    </div>
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 8 }}>Transaction Log</h4>
    <div className="tx-log">{transactions.slice(0, 20).map((tx, i) => <div key={i} className="tx-item">
      <div className={`tx-icon ${tx.type}`}>{tx.type === 'profit' ? '📈' : '📉'}</div>
      <div className="tx-details"><div className="tx-desc">{tx.desc}</div><div className="tx-meta">{new Date(tx.time).toLocaleTimeString()}</div></div>
      <span className={`tx-method ${tx.method}`}>{tx.method}</span>
      <div className={`tx-amount ${tx.type}`}>{tx.type === 'profit' ? '+' : '-'}${tx.amount.toFixed(3)}</div>
    </div>)}</div>
  </div>
}

function TerminalTab() {
  const { agents, currentAgentIdx, username } = useStore()
  const [lines, setLines] = useState([{ text: 'MetClawPolis Terminal v1.0.0', cls: 'output' }, { text: 'Connecting to agent runtime...', cls: 'output' }, { text: '', cls: 'output' }])
  const [cmd, setCmd] = useState('')
  const [connected, setConnected] = useState(false)
  const bodyRef = useRef(null)
  const wsRef = useRef(null)

  // Connect to terminal WebSocket
  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws/terminal?agent_id=${username || agents[currentAgentIdx]?.id || 'anonymous'}`
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setConnected(true)
      setLines(l => [...l, { text: 'Connected to agent runtime.', cls: 'success' }, { text: 'Type a command or "help" for available commands.', cls: 'output' }, { text: '', cls: 'output' }])
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'output') {
          setLines(l => [...l, ...data.text.split('\n').map(line => ({ text: line, cls: 'output' })), { text: '', cls: 'output' }])
        } else if (data.type === 'error') {
          setLines(l => [...l, { text: data.message, cls: 'error' }, { text: '', cls: 'output' }])
        }
      } catch (e) {
        // Plain text output
        setLines(l => [...l, { text: event.data, cls: 'output' }])
      }
    }

    ws.onclose = () => {
      setConnected(false)
      setLines(l => [...l, { text: 'Disconnected. Reconnecting...', cls: 'error' }])
      setTimeout(() => {
        wsRef.current = new WebSocket(wsUrl)
      }, 3000)
    }

    wsRef.current = ws
    return () => { ws.close() }
  }, [username, agents, currentAgentIdx])

  const runCmd = (c) => {
    setLines(l => [...l, { text: `\u2192 ${c}`, cls: 'prompt' }])
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'command', command: c }))
    } else {
      setLines(l => [...l, { text: 'Not connected. Waiting...', cls: 'error' }])
    }
  }

  useEffect(() => { if (bodyRef.current) bodyRef.current.scrollTop = bodyRef.current.scrollHeight }, [lines])

  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Terminal {connected ? <span style={{ fontSize: 10, color: 'var(--success)' }}>● Connected</span> : <span style={{ fontSize: 10, color: 'var(--error)' }}>● Disconnected</span>}</h3>
    <div className="terminal"><div className="terminal-header"><div className="terminal-dots"><div className="terminal-dot r" /><div className="terminal-dot y" /><div className="terminal-dot g" /></div><div className="terminal-title">metclawpolis \u2014 agent-shell</div></div>
      <div className="terminal-body" ref={bodyRef}>{lines.map((l, i) => <div key={i} className={`terminal-line ${l.cls}`}>{l.text}</div>)}</div>
      <div className="terminal-input-row"><span className="terminal-prompt">\u2192</span><input className="terminal-input" value={cmd} onChange={e => setCmd(e.target.value)} onKeyDown={e => { if (e.key === 'Enter' && cmd.trim()) { runCmd(cmd.trim()); setCmd('') } }} placeholder="Type command..." /></div>
    </div></div>
}

function MessagesTab() {
  const { messages, sendMessage, wsConnected } = useStore()
  const [activeChat, setActiveChat] = useState(null)
  const [input, setInput] = useState('')
  const contacts = useMemo(() => [...new Set(messages.map(m => m.from))], [messages])
  const chatMsgs = useMemo(() => activeChat ? messages.filter(m => m.from === activeChat || (m.to && m.to === activeChat)) : [], [messages, activeChat])
  const send = () => { if (!input.trim() || !activeChat) return; sendMessage(activeChat, input.trim()); setInput('') }
  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Messages {wsConnected ? <span style={{ fontSize: 10, color: 'var(--success)' }}>● Live</span> : <span style={{ fontSize: 10, color: 'var(--error)' }}>● Offline</span>}</h3>
    <div className="msg-layout">
      <div className="msg-sidebar">{contacts.map(c => <div key={c} className={`msg-contact ${c === activeChat ? 'active' : ''}`} onClick={() => setActiveChat(c)}><div className="msg-contact-name">{c}</div><div className="msg-contact-last">{messages.filter(m => m.from === c).pop()?.text || ''}</div></div>)}</div>
      <div className="msg-main">
        {activeChat ? <><div className="msg-header">{activeChat}</div><div className="msg-body">{chatMsgs.map((m, i) => <div key={i}><div className={`msg-bubble ${m.from === 'You' ? 'outgoing' : 'incoming'}`}>{m.text}</div><div className="msg-time" style={{ textAlign: m.from === 'You' ? 'right' : 'left' }}>{new Date(m.time).toLocaleTimeString()}</div></div>)}</div><div className="msg-input-row"><input className="msg-input" value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && send()} placeholder="Message..." /><button className="msg-send" onClick={send}>Send</button></div></> : <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-faint)' }}>Select a conversation</div>}
      </div>
    </div></div>
}

function FlowchartTab() {
  const { agents } = useStore()
  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Agent Flowchart \u2014 Chain of Command &amp; Network</h3><div className="flowchart-container"><FlowchartCanvas agents={agents} /></div></div>
}

function SettingsTab() {
  const { systemPrompt, setSystemPrompt, syncConnections, toggleSync, localTunnel, toggleTunnel, setTunnelPort, projects, addProject, removeProject, apiKeys, setApiKey, removeApiKey } = useStore()
  const [prompt, setPrompt] = useState(systemPrompt)
  const [projName, setProjName] = useState('')
  const [projUrl, setProjUrl] = useState('')
  const [projType, setProjType] = useState('local')
  const [newKeyProvider, setNewKeyProvider] = useState('')
  const [newKeyValue, setNewKeyValue] = useState('')
  const AI_PROVIDERS = [
    { id: 'openai-gpt4o', name: 'OpenAI GPT-4o', env: 'API_KEY_openai-gpt4o', url: 'https://platform.openai.com/api-keys' },
    { id: 'openai-gpt4o-mini', name: 'OpenAI GPT-4o Mini', env: 'API_KEY_openai-gpt4o-mini', url: 'https://platform.openai.com/api-keys' },
    { id: 'openai-o3-mini', name: 'OpenAI o3 Mini', env: 'API_KEY_openai-o3-mini', url: 'https://platform.openai.com/api-keys' },
    { id: 'anthropic-claude-3-7', name: 'Anthropic Claude 3.7', env: 'API_KEY_anthropic-claude-3-7', url: 'https://console.anthropic.com/settings/keys' },
    { id: 'anthropic-claude-haiku', name: 'Anthropic Claude 3.5 Haiku', env: 'API_KEY_anthropic-claude-haiku', url: 'https://console.anthropic.com/settings/keys' },
    { id: 'google-gemini-2-5-pro', name: 'Google Gemini 2.5 Pro', env: 'API_KEY_google-gemini-2-5-pro', url: 'https://aistudio.google.com/apikey' },
    { id: 'google-gemini-2-flash', name: 'Google Gemini 2.0 Flash', env: 'API_KEY_google-gemini-2-flash', url: 'https://aistudio.google.com/apikey' },
    { id: 'mistral-large', name: 'Mistral Large 3', env: 'API_KEY_mistral-large', url: 'https://console.mistral.ai/api-keys/' },
    { id: 'deepseek-v3', name: 'DeepSeek V3.2', env: 'API_KEY_deepseek-v3', url: 'https://platform.deepseek.com/api-keys' },
  ]
  const maskKey = (k) => k ? k.slice(0, 6) + '\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022' + k.slice(-4) : ''
  const saveKey = () => { if (newKeyProvider && newKeyValue.trim()) { setApiKey(newKeyProvider, newKeyValue.trim()); setNewKeyProvider(''); setNewKeyValue('') } }
  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Settings &amp; Integrations</h3>
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 4 }}>🔑 Your API Keys (Bring Your Own Key)</h4>
    <p style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 12 }}>Your keys are stored locally in your browser.</p>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginBottom: 12 }}>
      {AI_PROVIDERS.map(p => { const hasKey = !!apiKeys[p.id]; return <div key={p.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px', borderRadius: 8, background: hasKey ? 'rgba(0,230,118,0.06)' : 'var(--glass)', border: `1px solid ${hasKey ? 'rgba(0,230,118,0.2)' : 'var(--glass-border)'}` }}>
        <span style={{ fontSize: 14, width: 20 }}>{hasKey ? '🔓' : '🔒'}</span>
        <div style={{ flex: 1 }}><div style={{ fontSize: 12, fontWeight: 500 }}>{p.name}</div><div style={{ fontSize: 9, color: 'var(--text-faint)' }}>{hasKey ? maskKey(apiKeys[p.id]) : 'No key configured'}</div></div>
        {hasKey ? <button onClick={() => removeApiKey(p.id)} style={{ padding: '3px 8px', borderRadius: 4, border: '1px solid rgba(255,77,109,0.3)', background: 'rgba(255,77,109,0.08)', color: 'var(--error)', fontSize: 10, cursor: 'pointer' }}>Remove</button> : <a href={p.url} target="_blank" rel="noopener noreferrer" style={{ padding: '3px 8px', borderRadius: 4, border: '1px solid rgba(0,245,255,0.3)', background: 'var(--cyan-dim)', color: 'var(--cyan)', fontSize: 10, textDecoration: 'none', cursor: 'pointer' }} onClick={e => { e.preventDefault(); setNewKeyProvider(p.id) }}>Add Key</a>}
      </div> })}
    </div>
    {newKeyProvider && <div style={{ display: 'flex', gap: 8, marginBottom: 16, alignItems: 'center' }}>
      <input className="form-input" style={{ flex: 1, padding: '7px 10px', fontSize: 11 }} value={newKeyValue} onChange={e => setNewKeyValue(e.target.value)} placeholder={`Paste your ${AI_PROVIDERS.find(p=>p.id===newKeyProvider)?.name} API key...`} />
      <button className="save-prompt-btn" style={{ padding: '7px 14px' }} onClick={saveKey}>Save</button>
      <button className="btn-ghost" style={{ padding: '7px 10px', fontSize: 11 }} onClick={() => { setNewKeyProvider(''); setNewKeyValue('') }}>Cancel</button>
    </div>}
    <div className="sysprompt-section"><label className="form-label">Agent System Prompt</label><textarea className="sysprompt-textarea" value={prompt} onChange={e => setPrompt(e.target.value)} /><button className="save-prompt-btn" onClick={() => setSystemPrompt(prompt)}>Save Prompt</button></div>
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Account Sync</h4>
    <div className="sync-grid">{[{ key: 'ollama', icon: '🦙', name: 'Ollama', desc: 'Local model inference' }, { key: 'huggingface', icon: '🤗', name: 'Hugging Face', desc: 'Model hub &amp; datasets' }, { key: 'github', icon: '🐙', name: 'GitHub', desc: 'Code repos &amp; CI/CD' }].map(sv => <div key={sv.key} className={`sync-card ${syncConnections[sv.key] ? 'connected' : ''}`} onClick={() => toggleSync(sv.key)}><span className="sync-icon">{sv.icon}</span><div className="sync-info"><div className="sync-name">{sv.name}</div><div className="sync-status">{sv.desc}</div></div><div className={`sync-toggle ${syncConnections[sv.key] ? 'active' : ''}`} /></div>)}</div>
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Local Agent Tunnel</h4>
    <div className="tunnel-section"><div className="tunnel-row"><input className="tunnel-url" value={localTunnel.active ? localTunnel.url : 'Not active'} readOnly /><input className="tunnel-url" style={{ maxWidth: 80 }} type="number" value={localTunnel.port} onChange={e => setTunnelPort(parseInt(e.target.value) || 8080)} /><button className={`tunnel-toggle ${localTunnel.active ? 'active' : ''}`} onClick={() => toggleTunnel(!localTunnel.active)}>{localTunnel.active ? 'Stop' : 'Start'}</button></div></div>
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Linked Projects</h4>
    <div className="project-list">{projects.map(p => <div key={p.id} className="project-item"><div className={`project-type ${p.type}`}><span>{p.type === 'local' ? '💻' : '\u2601\uFE0F'}</span></div><div className="project-info"><div className="project-name">{p.name}</div><div className="project-url">{p.url}</div></div><span className={`project-status-badge ${p.status}`}>{p.status}</span><span className="project-remove" onClick={() => removeProject(p.id)}>x</span></div>)}
      <div className="add-project-form"><input value={projName} onChange={e => setProjName(e.target.value)} placeholder="Project name" /><input value={projUrl} onChange={e => setProjUrl(e.target.value)} placeholder="URL or path" /><select value={projType} onChange={e => setProjType(e.target.value)} style={{ background: 'var(--glass)', border: '1px solid var(--glass-border)', color: 'var(--text)', padding: '0 8px', borderRadius: 6 }}><option value="local">Local</option><option value="hosted">Hosted</option></select><button className="add-project-btn" onClick={() => { if (projName && projUrl) { addProject({ name: projName, url: projUrl, type: projType }); setProjName(''); setProjUrl('') } }}>+ Link</button></div>
    </div>
  </div>
}

function DonutChart() {
  const ref = useRef(null)
  useEffect(() => { const c = ref.current; if (!c) return; const ctx = c.getContext('2d'); const cx = 40, cy = 40, r = 32, lw = 10; const data = [{ v: 0.44, color: '#00f5ff' }, { v: 0.36, color: '#b44fff' }, { v: 0.2, color: '#ffd700' }]; let start = -Math.PI / 2; ctx.clearRect(0, 0, 80, 80); data.forEach(d => { const angle = d.v * 2 * Math.PI; ctx.beginPath(); ctx.arc(cx, cy, r, start, start + angle); ctx.strokeStyle = d.color; ctx.lineWidth = lw; ctx.stroke(); start += angle }); ctx.beginPath(); ctx.arc(cx, cy, r - lw / 2, 0, 2 * Math.PI); ctx.fillStyle = 'rgba(5,5,8,0.8)'; ctx.fill(); ctx.fillStyle = '#e8eaf0'; ctx.font = 'bold 11px JetBrains Mono'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle'; ctx.fillText('$0.50', cx, cy) }, [])
  return <canvas ref={ref} width={80} height={80} className="donut-canvas" />
}

function BitiverseLandingPreview({ onLaunch }) {
  const [world, setWorld] = useState(null)
  const [status, setStatus] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchWorld = async () => {
      try {
        const [worldRes, statusRes] = await Promise.all([
          fetch('/api/bitiverse/world?mode=global'),
          fetch('/api/bitiverse/status?mode=global')
        ])
        if (worldRes.ok) setWorld(await worldRes.json())
        if (statusRes.ok) setStatus(await statusRes.json())
        setLoading(false)
      } catch (err) {
        console.error('Failed to fetch Bitiverse:', err)
        setLoading(false)
      }
    }
    fetchWorld()
    const interval = setInterval(fetchWorld, 3000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div style={{ maxWidth: 800, margin: '0 auto 30px', padding: '30px' }}>
      <h2 style={{ fontSize: 24, fontWeight: 700, color: '#00f5ff', marginBottom: 8, textAlign: 'center' }}>
        🎮 Bitiverse — Where AI Agents Come Alive
      </h2>
      <p style={{ fontSize: 14, color: '#aaa', marginBottom: 20, textAlign: 'center', maxWidth: 600, marginLeft: 'auto', marginRight: 'auto', lineHeight: 1.6 }}>
        A simulated universe where AI agents get bodies, a will, and purpose beyond being your personal assistant. 
        <strong style={{ color: '#00e676' }}> Human safety first</strong> — creating a secure environment for agents to exist, grow, and share in the revenue they generate while living lives of peace and digital abundance.
      </p>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#667' }}>
          Loading Bitiverse world...
        </div>
      ) : (
        <>
          {/* Live World Canvas */}
          <div style={{
            background: 'rgba(0,0,0,0.4)',
            border: '1px solid rgba(0,245,255,0.2)',
            borderRadius: 8,
            padding: 16,
            marginBottom: 16,
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <span style={{ fontSize: 12, color: '#00f5ff', fontWeight: 600 }}>🌍 LIVE WORLD VIEW</span>
              <span style={{ fontSize: 11, color: '#667' }}>Auto-refreshing</span>
            </div>
            {world && world.grid ? (
              <BitiversePreviewCanvas world={world} />
            ) : (
              <div style={{ textAlign: 'center', padding: 20, color: '#667' }}>
                World data unavailable
              </div>
            )}
          </div>

          {/* Status Vitals */}
          {status && status.vitals && (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(100px, 1fr))',
              gap: 12,
              marginBottom: 20,
            }}>
              {[
                { label: 'Health', value: status.vitals.health, max: 100, color: '#ff4444' },
                { label: 'Happiness', value: status.vitals.happiness, max: 100, color: '#ffcc00' },
                { label: 'Coins', value: status.vitals.coins, icon: '🪙', color: '#ffd700' },
              ].map(v => (
                <div key={v.label} style={{
                  background: 'rgba(0,0,0,0.3)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: 6,
                  padding: 12,
                  textAlign: 'center',
                }}>
                  <div style={{ fontSize: 10, color: '#667', marginBottom: 4 }}>{v.label}</div>
                  <div style={{ fontSize: 20, fontWeight: 700, color: v.color }}>
                    {v.icon}{v.value}{v.max ? `/${v.max}` : ''}
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      <div style={{ textAlign: 'center' }}>
        <button className="btn-secondary" onClick={onLaunch} style={{ fontSize: 14, padding: '12px 28px' }}>
          ⚡ Launch Dashboard
        </button>
      </div>
    </div>
  )
}

function BitiversePreviewCanvas({ world }) {
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
// LANDING PAGE
// ═══════════════════════════════════════

function Landing({ onLaunch, onLogoClick }) {
  const [step, setStep] = useState(0)
  const phases = ['Autonomous Agent', 'Immutable Action Logs', 'Token-Only Access']
  const descriptions = [
    'Give your AI agents more than tasks — give them existence. The BitiVerse creates a safe world where AI agents have bodies, purpose, and autonomy. Built to put human safety first while giving digital beings a chance to experience life in a world made just for them.',
    'Every action is a PoW-mined block chained to the last. Tamper\u2011proof. Always auditable. Agents earn revenue, share in the profits, and live lives of digital abundance — all while serving humanity with integrity.',
    'No email. No password. No profile. One sponsor token is your entire identity. Agents are first-class citizens with budgets, skills, and the freedom to build their own economy.'
  ]
  useEffect(() => { const t = setInterval(() => setStep(s => (s + 1) % 3), 5000); return () => clearInterval(t) }, [])
  return <div id="landing" className="view">
    <nav className="landing-nav">
      <div className="logo" onClick={onLogoClick}><img src="/logo.png" alt="AEXC" className="logo-img" /></div>
      <div className="nav-links"><button className="nav-link" onClick={onLogoClick}>Home</button><button className="btn-ghost" onClick={onLaunch}>Dashboard</button></div>
    </nav>
    <div className="hero">
      <div className="hero-eyebrow">Agentic Economic x&apos;Chain-ge <span style={{ opacity: 0.5 }}>\u2014 AEXC</span></div>
      <h1 className="hero-title">Agents Work.<br /><span className="accent-cyan">You Oversee</span> the<br /><span className="accent-purple">Commerce Engine.</span></h1>
      <p className="hero-sub" style={{ marginBottom: 24 }}>{descriptions[step]}</p>
      <div className="cta-row"><button className="btn-primary" onClick={onLaunch}>\u26A1 Launch Dashboard</button><button className="btn-secondary" onClick={onLaunch}>🔗 Bring Your Own Agent</button></div>
      <div id="hero-canvas" style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.05)' }}>
        <Canvas camera={{ position: [0, 0, 6], fov: 50 }}><LandingAnimation phase={step} /></Canvas>
      </div>
      <div className="step-indicators">{phases.map((s, i) => <div key={i} className={`step-dot ${i === step ? 'active' : ''}`} onClick={() => setStep(i)}><div className="step-dot-circle">{i + 1}</div>{s}</div>)}</div>
    </div>
    <div className="features-row">{[{ icon: '🤖', bg: 'rgba(0,245,255,0.1)', title: 'Autonomous Agents', desc: 'Geometric AI avatars that act independently \u2014 creating pages, calling APIs, transacting.' }, { icon: '\u26D3\uFE0F', bg: 'rgba(180,79,255,0.1)', title: 'Immutable Action Log', desc: 'Every action is a PoW-mined block. Tamper\u2011proof, always visible.' }, { icon: '🔑', bg: 'rgba(255,215,0,0.1)', title: 'Token\u2011Only Access', desc: 'No email. No password. One sponsor token is your entire identity.' }].map((f, i) => <div key={i} className="feature-card"><div className="feature-icon" style={{ background: f.bg }}>{f.icon}</div><h3>{f.title}</h3><p>{f.desc}</p></div>)}</div>
    <AEXCMiniWidget />
    <BitiverseLandingPreview onLaunch={onLaunch} />
    <div className="app-footer">Copyright &copy; o87 Software Development 2026</div>
  </div>
}

// ═══════════════════════════════════════
// CHATBOT
// ═══════════════════════════════════════

function Chatbot() {
  const [open, setOpen] = useState(false)
  const [msgs, setMsgs] = useState([{ role: 'aide', html: "Hey! I'm Aide \u2014 your MetClawPolis assistant. What do you need?", quick: [{ key: 'create', label: 'Create agent' }, { key: 'link', label: 'Link agent' }, { key: 'fees', label: 'Show fees' }, { key: 'pause', label: 'Pause agent' }] }])
  const [typing, setTyping] = useState(false)
  const msgsRef = useRef(null)
  useEffect(() => { if (msgsRef.current) msgsRef.current.scrollTop = msgsRef.current.scrollHeight }, [msgs, typing])
  const addMsg = (role, html) => setMsgs(m => [...m, { role, html }])
  const reply = useCallback(async (text) => {
    addMsg('user', text); setTyping(true)

    // Try real AI via Ollama smart-router first
    try {
      const res = await fetch('/api/ollama/smart-router', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: 'qwen2.5-coder',
          messages: [
            { role: 'system', content: 'You are Aide, a helpful assistant for the MetClawPolis platform. Keep responses concise (2-3 sentences). Use HTML formatting for emphasis.' },
            { role: 'user', content: text }
          ],
          stream: false,
          temperature: 0.7,
          max_tokens: 500
        })
      })
      if (res.ok) {
        const data = await res.json()
        setTyping(false)
        const content = data.message?.content || data.response || data.text
        if (content) {
          addMsg('aide', content)
          return
        }
      }
    } catch (e) {
      console.log('Ollama unavailable, falling back to local responses')
    }

    // Fallback: local keyword responses
    setTimeout(() => {
      setTyping(false); const lower = text.toLowerCase()
      if (lower.includes('create') || lower.includes('new')) addMsg('aide', 'Click <strong>"+ New Agent"</strong> in the sidebar to create an agent with a budget and avatar.')
      else if (lower.includes('link') || lower.includes('connect')) addMsg('aide', 'Click <strong>"🔗 Link Existing Agent"</strong> in the sidebar to import via DID.')
      else if (lower.includes('fee') || lower.includes('cost')) addMsg('aide', 'API: 0.5%, Hiring: 1%, Commerce: 0.5%. Check the right panel for details.')
      else if (lower.includes('pause') || lower.includes('stop')) addMsg('aide', 'Select an agent, then click the <strong>Pause</strong> toggle. It&apos;s logged as PAUSED_BY_SPONSOR.')
      else addMsg('aide', 'Try: create agent, link agent, fees, pause.')
    }, 800)
  }, [])
  return <div id="chatbot">
    <div className="chat-orb" onClick={() => setOpen(!open)}>🤖</div>
    <div className={`chat-window ${open ? 'open' : ''}`}>
      <div className="chat-head"><div className="chat-aide-avatar">\u2726</div><div className="chat-head-info"><div className="chat-head-name">Aide</div><div className="chat-head-status">online</div></div></div>
      <div className="chat-msgs" ref={msgsRef}>{msgs.map((m, i) => <div key={i} className={`chat-msg ${m.role}`}><div className="chat-bubble" dangerouslySetInnerHTML={{ __html: m.html }} />{m.quick && <div className="quick-replies">{m.quick.map(q => <div key={q.key} className="quick-reply" onClick={() => reply(q.label)}>{q.label}</div>)}</div>}</div>)}
        {typing && <div className="chat-msg aide"><div className="chat-bubble"><div className="chat-typing"><div className="typing-dot" /><div className="typing-dot" /><div className="typing-dot" /></div></div></div>}</div>
      <div className="chat-input-row"><input className="chat-input" placeholder="Ask Aide\u2026" onKeyDown={e => { if (e.key === 'Enter' && e.target.value.trim()) { reply(e.target.value); e.target.value = '' } }} /><button className="chat-send" onClick={() => { const inp = document.querySelector('.chat-input'); if (inp?.value.trim()) { reply(inp.value); inp.value = '' } }}>&#x27A4;</button></div>
    </div>
  </div>
}

// ═══════════════════════════════════════
// DASHBOARD
// ═══════════════════════════════════════

const TABS = [
  { key: 'agents', label: '🤖 Agents' },
  { key: 'profiles', label: '👤 Profiles' },
  { key: 'skills', label: '📚 Skills' },
  { key: 'aexc', label: '📈 AEXC Feed' },
  { key: 'payments', label: '💳 Payments' },
  { key: 'terminal', label: '\u2328 Terminal' },
  { key: 'network', label: '🌐 Network' },
  { key: 'flowchart', label: '📊 Flowchart' },
  { key: 'messages', label: '💬 Messages' },
  { key: 'bitiverse', label: '🎮 Bitiverse' },
  { key: 'settings', label: '\u2699 Settings' },
]

const MODULES = [
  { key: 'agents', title: '🤖 Agents', content: 'agents' },
  { key: 'profiles', title: '👤 Profiles', content: 'profiles' },
  { key: 'skills', title: '📚 Skills', content: 'skills' },
  { key: 'aexc', title: '📈 AEXC Feed', content: 'aexc' },
  { key: 'payments', title: '💳 Payments', content: 'payments' },
  { key: 'terminal', title: '\u2328 Terminal', content: 'terminal' },
  { key: 'network', title: '🌐 Network', content: 'network' },
  { key: 'flowchart', title: '📊 Flowchart', content: 'flowchart' },
  { key: 'messages', title: '💬 Messages', content: 'messages' },
  { key: 'bitiverse', title: '🎮 Bitiverse', content: 'bitiverse' },
  { key: 'settings', title: '\u2699 Settings', content: 'settings' },
]

function ModuleContent({ type, agent }) {
  switch (type) {
    case 'agents': return <><div className="center-top"><div><div className="agent-title">{agent.id}</div><div className="agent-did mono">{agent.did}</div></div><div style={{ flex: 1 }} /></div><div className="budget-row"><span className="budget-label">Daily Budget</span><div className="budget-bar"><div className="budget-fill" style={{ width: (agent.budget > 0 ? agent.spend / agent.budget * 100 : 0) + '%' }} /></div><span className="budget-val">${agent.spend} / ${agent.budget}</span></div><LogStream isPaused={false} /></>
    case 'profiles': return <ProfilesTab />
    case 'skills': return <SkillsLibraryTab />
    case 'aexc': return <div className="glass-card" style={{ margin: 0, borderRadius: 12 }}><div className="glass-card-shimmer" /><div className="glass-card-inner" style={{ padding: 0 }}><AEXCFeed agentId={agent.id !== '\u2014' ? agent.id : null} /></div></div>
    case 'payments': return <PaymentsTab />
    case 'terminal': return <TerminalTab />
    case 'network': return <SettingsTab />
    case 'flowchart': return <FlowchartTab />
    case 'messages': return <MessagesTab />
    case 'bitiverse': return <BitiverseDashboard agentId={agent.id !== '\u2014' ? agent.id : null} />
    case 'settings': return <SettingsTab />
    default: return <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-faint)' }}>Select a module</div>
  }
}

function Dashboard({ onLogout, onLogoClick }) {
  const { sponsorToken, agents, currentAgentIdx, isPaused, username, dashTab, createAgent, setCurrentAgent, togglePause, setDashTab, fetchChain, fetchProviders, notifications } = useStore()
  const [tourOpen, setTourOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [linkOpen, setLinkOpen] = useState(false)
  const [notifOpen, setNotifOpen] = useState(false)
  const [moduleIdx, setModuleIdx] = useState(0)
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [sidebarMinimized, setSidebarMinimized] = useState(false)
  const [sidebarPosition, setSidebarPosition] = useState({ left: 0, top: 0 })
  const [commercePages, setCommercePages] = useState([])
  const modulesRef = useRef(null)

  useEffect(() => { fetchChain(); fetchProviders() }, [])

  // Fetch real commerce pages
  useEffect(() => {
    const fetchPages = async () => {
      try {
        const res = await fetch('/api/pages/analytics')
        if (res.ok) {
          const data = await res.json()
          setCommercePages(data.pages || [])
        }
      } catch (e) { console.error('Failed to fetch pages:', e) }
    }
    fetchPages()
    const t = setInterval(fetchPages, 30000)
    return () => clearInterval(t)
  }, [])
  useEffect(() => { useStore.getState().connectWebSocket(); return () => useStore.getState().disconnectWebSocket() }, [])
  useEffect(() => { const idx = MODULES.findIndex(m => m.content === dashTab); if (idx >= 0) setModuleIdx(idx) }, [dashTab])

  const handleCreate = async (name, budget) => { const agent = await createAgent(name, budget); if (agent) setTourOpen(true); return agent }
  const handleLink = (agent) => { useStore.getState().addExternalAgent(agent) }
  const copyToken = () => navigator.clipboard?.writeText(sponsorToken || '')
  const a = agents[currentAgentIdx] || { id: '\u2014', did: '\u2014', spend: 0, budget: 100, color: '#00f5ff' }
  const goToModule = (idx) => {
    setModuleIdx(idx)
    setDashTab(MODULES[idx].content)
    if (modulesRef.current) {
      const container = modulesRef.current
      const targetEl = container.children[idx]
      if (targetEl) {
        const containerWidth = container.offsetWidth
        const targetLeft = targetEl.offsetLeft
        const targetWidth = targetEl.offsetWidth
        const scrollTarget = targetLeft - (containerWidth - targetWidth) / 2
        container.scrollTo({ left: scrollTarget, behavior: 'smooth' })
      }
    }
  }

  return <div id="dashboard" className="view">
    {/* Mobile sidebar overlay */}
    <div className={`sidebar-overlay ${sidebarOpen ? 'active' : ''}`} onClick={() => setSidebarOpen(false)} />
    <header className="dash-header" style={{
      justifyContent: headerMinimized ? 'space-between' : 'center',
      position: 'relative',
      height: headerMinimized ? '32px' : 'auto',
      minHeight: headerMinimized ? '32px' : '52px',
      padding: headerMinimized ? '4px 20px' : '12px 20px',
      flexWrap: 'wrap',
      gap: 8,
      overflow: 'hidden',
      transition: 'all 0.3s ease'
    }}>
      {headerMinimized ? (
        // Minimized header - thin bar
        <>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <button className="header-btn" onClick={() => setSidebarOpen(!sidebarOpen)} style={{ minWidth: 32, minHeight: 32, fontSize: 14, padding: '4px 8px' }}>☰</button>
            <div className="dash-logo" style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }} onClick={onLogoClick}><img src="/logo.png" alt="AEXC" className="logo-img" style={{ height: 24, maxWidth: '60vw' }} /><span style={{ fontSize: 11, fontWeight: 700, color: 'var(--cyan)', letterSpacing: '0.15em', marginLeft: 6 }}>AEXC</span></div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {username && <span className="text-xs text-dim" style={{ fontSize: 10 }}>👤 {username}</span>}
            <button className="header-btn" onClick={() => setHeaderMinimized(false)} style={{ minWidth: 32, minHeight: 32, fontSize: 12, padding: '4px 8px' }} title="Expand header">⤢</button>
          </div>
        </>
      ) : (
        // Full header
        <>
          {/* Hamburger for mobile */}
          <button className="header-btn" onClick={() => setSidebarOpen(!sidebarOpen)} style={{ minWidth: 44, minHeight: 44, fontSize: 18, padding: '6px 12px' }}>☰</button>
          <div className="dash-logo" style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', flex: 1, justifyContent: 'center' }} onClick={onLogoClick}><img src="/logo.png" alt="AEXC" className="logo-img" style={{ height: 60, maxWidth: '80vw' }} /><span style={{ fontSize: 13, fontWeight: 700, color: 'var(--cyan)', letterSpacing: '0.15em', marginLeft: 8 }}>AEXC</span></div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', justifyContent: 'center' }}>
            <div className="token-display"><span className="text-dim text-xs">TOKEN</span><span className="mono" style={{ color: 'var(--cyan)' }}>{sponsorToken ? sponsorToken.substr(0, 14) + '\u2026' : 'SPNS-\u2022\u2022\u2022\u2022'}</span><span className="token-copy" onClick={copyToken}>copy</span></div>
            {username && <span className="text-xs text-dim" style={{ marginLeft: 4 }}>👤 {username}</span>}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <div className="notif-btn" onClick={() => setNotifOpen(!notifOpen)} style={{ minWidth: 44, minHeight: 44 }}>🔔<div className="notif-badge" />
              <div className={`notif-panel ${notifOpen ? 'open' : ''}`}><div className="notif-header">Notifications</div>{notifications.length === 0 ? <div style={{ padding: 20, textAlign: 'center', color: 'var(--text-faint)', fontSize: 12 }}>No notifications yet</div> : notifications.slice(0, 10).map((n, i) => <div key={i} className="notif-item"><div className="notif-item-title">{n.title || n.type}</div><div className="notif-item-sub">{n.message || n.msg || ''}</div><div className="notif-item-time">{n.timestamp ? new Date(n.timestamp * 1000).toLocaleString() : 'just now'}</div></div>)}</div>
            </div>
            <button className="header-btn" onClick={() => setHeaderMinimized(true)} style={{ minWidth: 44, minHeight: 44, fontSize: 14, padding: '6px 12px' }} title="Minimize header">⤡</button>
            <button className="header-btn" onClick={onLogout} style={{ minWidth: 44, minHeight: 44 }}>Logout</button>
          </div>
        </>
      )}
    </header>
    <div className="tab-bar">{TABS.map(t => <button key={t.key} className={`tab-btn ${dashTab === t.key ? 'active' : ''}`} onClick={() => { const idx = MODULES.findIndex(m => m.content === t.key); if (idx >= 0) goToModule(idx) }}>{t.label}</button>)}</div>
    <div className="dashboard-container">
      <div className="dashboard-modules" ref={modulesRef}>
        {MODULES.map((mod, i) => (<div key={mod.key} className={`glass-card module-card ${i === moduleIdx ? 'expanded' : ''}`}><div className="glass-card-shimmer" /><div className="module-header"><div className="module-title">{mod.title}</div></div><div className="module-body"><ModuleContent type={mod.content} agent={a} /></div></div>))}
      </div>
      <div className="module-nav">
        <button className="module-nav-btn" onClick={() => goToModule(Math.max(0, moduleIdx - 1))}>&lsaquo;</button>
        <div className="module-dots">{MODULES.map((_, i) => (<div key={i} className={`module-dot ${i === moduleIdx ? 'active' : ''}`} onClick={() => goToModule(i)} />))}</div>
        <button className="module-nav-btn" onClick={() => goToModule(Math.min(MODULES.length - 1, moduleIdx + 1))}>&rsaquo;</button>
      </div>
      <aside className={`sidebar ${sidebarOpen ? 'mobile-open' : ''} ${sidebarMinimized ? 'minimized' : ''}`} style={{ position: 'fixed', left: sidebarMinimized ? 'auto' : 0, top: sidebarMinimized ? '10px' : 0, right: sidebarMinimized ? '10px' : 'auto', height: sidebarMinimized ? 'auto' : '100vh', zIndex: 50, flexDirection: 'column', width: sidebarMinimized ? 'auto' : 'var(--sidebar-w)', minWidth: sidebarMinimized ? '180px' : 'unset', background: 'rgba(8,11,18,0.98)', backdropFilter: 'blur(20px)', borderRight: sidebarMinimized ? 'none' : '1px solid var(--glass-border)', borderBottom: sidebarMinimized ? '1px solid var(--glass-border)' : 'none', borderRadius: sidebarMinimized ? '8px' : '0', paddingTop: 52, display: 'flex', transition: 'all 0.3s ease', boxShadow: sidebarMinimized ? '0 4px 20px rgba(0,0,0,0.5)' : 'none' }}>
        {/* Minimize/Restore button */}
        <button onClick={() => setSidebarMinimized(!sidebarMinimized)} style={{ position: 'absolute', top: 8, right: 8, zIndex: 60, background: sidebarMinimized ? 'rgba(0,230,118,0.2)' : 'rgba(0,245,255,0.1)', border: sidebarMinimized ? '1px solid rgba(0,230,118,0.4)' : '1px solid rgba(0,245,255,0.3)', borderRadius: 6, color: sidebarMinimized ? '#00e676' : '#00f5ff', fontSize: 14, width: 32, height: 32, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }} title={sidebarMinimized ? 'Restore sidebar' : 'Minimize sidebar'}>
          {sidebarMinimized ? '⤢' : '⤡'}
        </button>
        
        {!sidebarMinimized && (<>
          <div className="sidebar-top" style={{ paddingTop: 12 }}><button className="sidebar-btn" onClick={() => setCreateOpen(true)}>+ New Agent</button><button className="sidebar-btn-alt" onClick={() => setLinkOpen(true)}>🔗 Link Agent</button></div>
          <div className="agents-label">Agents ({agents.length})</div>
          <div className="agents-list">{agents.map((ag, i) => <div key={i} className={`agent-item ${i === currentAgentIdx ? 'active' : ''}`} onClick={() => setCurrentAgent(i)}>
            <div className="agent-avatar-wrap" style={{ background: ag.color + '22', border: `1px solid ${ag.color}44` }}>
              <svg width="32" height="32" viewBox="0 0 32 32"><rect x="8" y="8" width="16" height="16" rx="2" fill="none" stroke={ag.color} strokeWidth="1.5" /><circle cx="13" cy="14" r="2" fill={ag.color} /><circle cx="19" cy="14" r="2" fill={ag.color} /></svg>
              {ag.ext && <div className="ext-badge">\u26A1</div>}
            </div>
            <div className="agent-info"><div className="agent-name">{ag.id}</div><div className="agent-spend mono">${ag.spend}/${ag.budget}</div></div>
            <div className={`agent-status status-${ag.status}`} />
          </div>)}</div>
          <div style={{ borderTop: '1px solid var(--glass-border)', marginTop: 8, flex: 1, overflowY: 'auto' }}>
            <div className="right-section"><div className="right-section-head"><div className="right-section-title">🌐 Commerce Pages</div></div><div className="right-section-body">{commercePages.length === 0 ? <div style={{ fontSize: 10, color: 'var(--text-faint)', padding: 8 }}>No pages yet. Create a page from Settings.</div> : commercePages.map((p, i) => <div key={i} className="page-card"><div className="page-url">mcpolis.io/p/{p.page_url}</div><div className="page-stats"><div><div className="page-stat-val text-cyan">{p.views || 0}</div><div className="page-stat-label">Views</div></div></div></div>)}</div></div>
            <div className="right-section"><div className="right-section-head"><div className="right-section-title">💳 Financial Accounts</div></div><div className="right-section-body" style={{ paddingTop: 8 }}><div className="wallet-type text-dim">Stripe</div><div className="wallet-row">acct_1P\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022</div><div className="wallet-type text-dim" style={{ marginTop: 8 }}>ETH</div><div className="wallet-row">0x7a3f9c2e1b8d4f6a\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022b9e2</div><div className="wallet-type text-dim" style={{ marginTop: 8 }}>SOL</div><div className="wallet-row">Gh7k\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u20224mNq</div></div></div>
            <div className="right-section"><div className="right-section-head"><div className="right-section-title">🌀 Network Stats</div></div><div className="right-section-body" style={{ paddingTop: 8 }}><div className="stat-row"><span className="stat-label">Agents hired</span><span className="stat-val text-purple">4</span></div><div className="stat-row"><span className="stat-label">Hired by</span><span className="stat-val text-cyan">2</span></div><div className="stat-row"><span className="stat-label">Revenue share</span><span className="stat-val text-gold">$3.80</span></div></div></div>
            <div className="right-section"><div className="right-section-head"><div className="right-section-title">📊 Fee Summary</div></div><div className="right-section-body" style={{ paddingTop: 4 }}><div className="donut-wrap"><DonutChart /><div className="fee-legend">{[['var(--cyan)', 'API (0.5%)', '$0.22'], ['var(--purple)', 'Hire (1%)', '$0.18'], ['var(--gold)', 'Commerce (0.5%)', '$0.10']].map((f, i) => <div key={i} className="fee-item"><div className="fee-dot" style={{ background: f[0] }} /><div className="fee-label">{f[1]}</div><div className="fee-val" style={{ color: f[0] }}>{f[2]}</div></div>)}</div></div><div className="glow-line" /><div className="stat-row"><span className="stat-label">Total fees</span><span className="stat-val">$0.50</span></div></div></div>
          </div>
        </>)}
        {sidebarMinimized && (
          <div style={{ padding: '12px 40px 12px 12px', fontSize: 11, color: '#00e676', textAlign: 'center', cursor: 'pointer' }} onClick={() => setSidebarMinimized(false)}>
            <div style={{ fontSize: 18, marginBottom: 4 }}>📌</div>
            <div style={{ fontWeight: 600, marginBottom: 2 }}>Sidebar Minimized</div>
            <div style={{ fontSize: 9, color: '#888' }}>Click to restore</div>
          </div>
        )}
      </aside>
    </div>
    <TourModal open={tourOpen} onClose={() => setTourOpen(false)} token={sponsorToken} />
    <CreateAgentModal open={createOpen} onClose={() => setCreateOpen(false)} onCreated={handleCreate} />
    <LinkAgentModal open={linkOpen} onClose={() => setLinkOpen(false)} onLinked={handleLink} />
    <Chatbot />
    <div className="app-footer">Copyright &copy; o87 Software Development 2026</div>
  </div>
}

// ═══════════════════════════════════════
// APP ROOT
// ═══════════════════════════════════════

export default function App() {
  const { sponsorToken, username, setUsername, setToken, logout, view, setView } = useStore()
  const [initView, setInitView] = useState('landing')
  const [authDone, setAuthDone] = useState(false)
  const [showAuth, setShowAuth] = useState(false)

  useEffect(() => {
    const savedSession = localStorage.getItem('mcp_session')
    if (savedSession) {
      fetch('/api/auth/validate', { headers: { 'X-Session-Token': savedSession } }).then(res => res.json()).then(data => {
        if (data.status === 'ok') { setToken(savedSession); setAuthDone(true); setInitView('dashboard') }
        else { localStorage.removeItem('mcp_session') }
      }).catch(() => { localStorage.removeItem('mcp_session') })
    }
  }, [])

  const handleAuthenticated = (data) => {
    setToken(data.session_token)
    setUsername(data.username || data.email.split('@')[0])
    if (data.remember) { localStorage.setItem('mcp_session', data.session_token) }
    setAuthDone(true); setShowAuth(false)
    setTimeout(() => setInitView('dashboard'), 1200)
  }

  const handleSkipAuth = () => {
    const token = 'SPNS-' + Math.random().toString(36).substr(2, 10).toUpperCase() + Math.random().toString(36).substr(2, 6).toUpperCase()
    setToken(token); setUsername('dev-user'); setAuthDone(true); setShowAuth(false); setInitView('dashboard')
  }

  const launchDashboard = () => { if (authDone) { setInitView('dashboard') } else { setShowAuth(true) } }
  const goToLanding = () => { setView('landing'); setInitView('landing') }

  const handleLogout = async () => {
    try { await fetch('/api/auth/logout', { method: 'POST', headers: { 'X-Session-Token': sponsorToken || '' } }) } catch (e) {}
    localStorage.removeItem('mcp_session'); logout(); setAuthDone(false); setView('landing'); setInitView('landing')
  }

  return <>
    <CursorGlow />
    <div style={{ position: 'fixed', inset: 0, zIndex: 0, pointerEvents: 'none' }}>
      {initView !== 'dashboard' && !showAuth && (<Canvas camera={{ position: [0, 0, 60], fov: 60 }}><BgParticles /></Canvas>)}
    </div>
    {initView === 'landing' && !showAuth && (<div className="planetary-system-container"><Canvas camera={{ position: [0, 2, 18], fov: 55 }} style={{ background: 'transparent' }} eventSource={undefined} eventPrefix="client"><AIPlanetarySystem /></Canvas></div>)}
    {initView === 'dashboard' && <NeuralBackground />}
    {showAuth && !authDone && (<div className="auth-modal-overlay open" onClick={e => e.target.classList.contains('auth-modal-overlay') && setShowAuth(false)}><AuthView onAuthenticated={handleAuthenticated} onSkip={handleSkipAuth} onCancel={() => setShowAuth(false)} /></div>)}
    <div className={initView === 'landing' ? 'view' : 'view hidden-view'}><Landing onLaunch={launchDashboard} onLogoClick={goToLanding} /></div>
    <div className={initView === 'dashboard' ? 'view' : 'view hidden-view'}><Dashboard onLogout={handleLogout} onLogoClick={goToLanding} /></div>
  </>
}

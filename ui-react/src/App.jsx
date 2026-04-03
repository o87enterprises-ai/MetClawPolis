import { useState, useEffect, useRef, useCallback, useMemo } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import * as THREE from 'three'
import { useStore } from './store'
import NeuralBackground from './components/scenes/NeuralBackground'

// ═══════════════════════════════════════
// THREE.JS SCENES
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

// Landing animation: 3 phases synced to step indicators — auto-loops every 5 seconds
// Phase 0: Autonomous Agent → 4 separate dots spawn
// Phase 1: Immutable Action Logs → lines connect the dots
// Phase 2: Token-Only Access → dots morph into 3D box bot with eyes
function LandingAnimation({ phase }) {
  const pts = useRef([])
  const lineMeshes = useRef([])
  const boxRef = useRef()
  const edgeRef = useRef()
  const glowRef = useRef()
  const eyeLRef = useRef()
  const eyeRRef = useRef()
  const pupilLRef = useRef()
  const pupilRRef = useRef()

  const pointPositions = [[-2, 1.5, 0], [2, 1.5, 0], [-2, -1.5, 0], [2, -1.5, 0]]

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime()
    // Animate dots
    pts.current.forEach((p, i) => {
      if (p) {
        if (phase === 0) {
          // Phase 1: floating individual dots
          p.position.x = pointPositions[i][0] + Math.sin(t * 0.7 + i * 1.5) * 0.3
          p.position.y = pointPositions[i][1] + Math.cos(t * 0.5 + i * 2) * 0.2
          p.position.z = Math.sin(t * 0.3 + i) * 0.5
        } else if (phase === 1) {
          // Phase 2: dots locked in position with subtle pulse
          const s = 1 + 0.15 * Math.sin(t * 2 + i)
          p.scale.set(s, s, s)
        }
      }
    })
    // Animate box
    if (boxRef.current) { boxRef.current.rotation.y = t * 0.4; boxRef.current.rotation.x = t * 0.25 }
    if (edgeRef.current) { edgeRef.current.rotation.y = t * 0.4; edgeRef.current.rotation.x = t * 0.25 }
    if (glowRef.current) glowRef.current.material.opacity = 0.12 + 0.08 * Math.sin(t * 2)
    // Blink eyes
    if (eyeLRef.current) { eyeLRef.current.scale.y = 0.8 + 0.4 * Math.sin(t * 2.5) }
    if (eyeRRef.current) { eyeRRef.current.scale.y = 0.8 + 0.4 * Math.sin(t * 2.5) }
  })

  return (
    <>
      <pointLight position={[3, 3, 3]} intensity={2} color={0x00f5ff} />
      <pointLight position={[-3, -2, 2]} intensity={1.5} color={0xb44fff} />

      {/* Phase 1: 4 individual floating dots — Autonomous Agent */}
      {phase >= 0 && pointPositions.map((pos, i) => (
        <mesh key={i} ref={el => pts.current[i] = el} position={pos}>
          <sphereGeometry args={[0.18, 16, 16]} />
          <meshStandardMaterial color={0x00f5ff} emissive={0x00f5ff} emissiveIntensity={0.6} />
        </mesh>
      ))}

      {/* Phase 2: Lines connecting dots — Immutable Action Logs */}
      {phase >= 1 && <>
        {/* Outer rectangle */}
        <line key="l1"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[0], ...pointPositions[1]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.7} /></line>
        <line key="l2"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[1], ...pointPositions[3]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.7} /></line>
        <line key="l3"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[3], ...pointPositions[2]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.7} /></line>
        <line key="l4"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[2], ...pointPositions[0]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0x00f5ff} transparent opacity={0.7} /></line>
        {/* Diagonal */}
        <line key="l5"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[0], ...pointPositions[3]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0xb44fff} transparent opacity={0.4} /></line>
        <line key="l6"><bufferGeometry><bufferAttribute attach="attributes-position" count={2} array={new Float32Array([...pointPositions[1], ...pointPositions[2]])} itemSize={3} /></bufferGeometry><lineBasicMaterial color={0xb44fff} transparent opacity={0.4} /></line>
      </>}

      {/* Phase 3: 3D box bot with eyes — Token-Only Access */}
      {phase >= 2 && <>
        {/* Glow halo */}
        <mesh ref={glowRef} position={[0, 0, -0.3]} scale={[2.2, 2.2, 2.2]}>
          <boxGeometry args={[1, 1, 1]} />
          <meshStandardMaterial color={0x00f5ff} transparent opacity={0.1} wireframe />
        </mesh>
        {/* Solid box body */}
        <mesh ref={boxRef}>
          <boxGeometry args={[1.8, 1.8, 1.8]} />
          <meshStandardMaterial color={0x00f5ff} emissive={0x003344} metalness={0.7} roughness={0.3} transparent opacity={0.85} />
        </mesh>
        {/* Edge wireframe */}
        <lineSegments ref={edgeRef}>
          <edgesGeometry args={[new THREE.BoxGeometry(1.8, 1.8, 1.8)]} />
          <lineBasicMaterial color={0x00f5ff} transparent opacity={0.9} />
        </lineSegments>
        {/* Left eye */}
        <mesh ref={eyeLRef} position={[-0.4, 0.3, 0.91]}>
          <sphereGeometry args={[0.22, 16, 16]} />
          <meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1} />
        </mesh>
        {/* Right eye */}
        <mesh ref={eyeRRef} position={[0.4, 0.3, 0.91]}>
          <sphereGeometry args={[0.22, 16, 16]} />
          <meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={1} />
        </mesh>
        {/* Left pupil */}
        <mesh ref={pupilLRef} position={[-0.4, 0.3, 1.0]}>
          <sphereGeometry args={[0.1, 8, 8]} />
          <meshBasicMaterial color={0x000} />
        </mesh>
        {/* Right pupil */}
        <mesh ref={pupilRRef} position={[0.4, 0.3, 1.0]}>
          <sphereGeometry args={[0.1, 8, 8]} />
          <meshBasicMaterial color={0x000} />
        </mesh>
        {/* Mouth */}
        <mesh position={[0, -0.4, 0.91]}>
          <boxGeometry args={[0.6, 0.08, 0.08]} />
          <meshStandardMaterial color={0xffffff} emissive={0xffffff} emissiveIntensity={0.5} />
        </mesh>
      </>}
    </>
  )
}

function ForgeReactor() {
  const g = useRef()
  useFrame(({ clock }) => { const t = clock.getElapsedTime(); if (g.current) { g.current.rotation.y = t * 1.2; g.current.rotation.x = t * 0.7; const s = 0.8 + 0.4 * Math.abs(Math.sin(t * 2)); g.current.scale.set(s, s, s) } })
  return (<>
    {[0, 1, 2].map(i => <mesh key={i}><torusGeometry args={[1.5 + i * 0.4, 0.04, 8, 64]} /><meshStandardMaterial color={[0x00f5ff, 0xb44fff, 0xffd700][i]} emissive={new THREE.Color([0x00f5ff, 0xb44fff, 0xffd700][i]).multiplyScalar(0.5)} metalness={0.8} roughness={0.2} transparent opacity={0.7} /></mesh>)}
    <mesh ref={g}><boxGeometry args={[1, 1, 1]} /><meshStandardMaterial color={0x00f5ff} emissive={0x002233} metalness={0.9} roughness={0.1} /></mesh>
  </>)
}

// Flowchart canvas
function FlowchartCanvas({ agents }) {
  const canvasRef = useRef(null)
  useEffect(() => {
    const c = canvasRef.current; if (!c) return
    const ctx = c.getContext('2d')
    const W = c.parentElement.clientWidth, H = 500
    c.width = W; c.height = H
    const nodes = agents.length ? agents.map((a, i) => ({
      x: 60 + (i % 4) * (W - 120) / Math.min(agents.length, 4),
      y: 60 + Math.floor(i / 4) * 180,
      ...a
    })) : [{ x: W / 2, y: H / 2, name: 'No agents', id: '—', color: '#555', skills: [] }]
    // Add root node
    const root = { x: W / 2, y: 40, name: 'Sponsor', id: 'root', color: '#ffd700', skills: ['OVERSEER'] }

    function draw() {
      ctx.clearRect(0, 0, W, H)
      // Lines
      nodes.forEach(n => {
        ctx.beginPath(); ctx.moveTo(root.x + 40, root.y + 20); ctx.lineTo(n.x + 40, n.y - 30)
        ctx.strokeStyle = 'rgba(0,245,255,0.15)'; ctx.lineWidth = 1; ctx.setLineDash([4, 4]); ctx.stroke(); ctx.setLineDash([])
      })
      // Root node
      ctx.fillStyle = root.color + '22'; ctx.strokeStyle = root.color + '66'; ctx.lineWidth = 1.5
      ctx.beginPath(); ctx.roundRect(root.x - 40, root.y - 20, 80, 40, 8); ctx.fill(); ctx.stroke()
      ctx.fillStyle = '#e8eaf0'; ctx.font = '600 11px Inter'; ctx.textAlign = 'center'; ctx.fillText(root.name, root.x, root.y + 5)

      // Agent nodes
      nodes.forEach(n => {
        ctx.fillStyle = n.color + '22'; ctx.strokeStyle = n.color + '44'; ctx.lineWidth = 1.5
        ctx.beginPath(); ctx.roundRect(n.x - 50, n.y - 30, 100, 60, 8); ctx.fill(); ctx.stroke()
        ctx.fillStyle = '#e8eaf0'; ctx.font = '600 11px Inter'; ctx.textAlign = 'center'; ctx.fillText(n.id, n.x, n.y - 5)
        ctx.font = '9px JetBrains Mono'; ctx.fillStyle = 'rgba(232,234,240,0.4)'; ctx.fillText(n.id, n.x, n.y + 10)
        // Skills
        if (n.skills && n.skills.length) {
          n.skills.slice(0, 3).forEach((s, si) => {
            ctx.fillStyle = n.color + '15'; ctx.fillRect(n.x - 40 + si * 28, n.y + 16, 24, 14)
            ctx.fillStyle = n.color; ctx.font = '8px JetBrains Mono'; ctx.fillText(s, n.x - 28 + si * 28, n.y + 26)
          })
        }
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
    <div className="modal"><div className="modal-header"><div><div className="modal-title">{title}</div><div className="modal-sub">{sub}</div></div><div className="modal-close" onClick={onClose}>×</div></div>
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
    { icon: '🔑', title: 'Your Sponsor Token', text: 'Save this — it\'s your only way back in.', showToken: true },
    { icon: '🔗', title: 'Connect Services', text: 'Sync with your existing tools for seamless agent deployment.' },
    { icon: '🚀', title: 'Ready', text: 'Your dashboard awaits. Create your first agent to begin.' },
  ]
  const s = steps[step]
  return <Modal open={open} onClose={onClose} title="Welcome to MetClawPolis" sub="Setup your account"
    footer={<>
      {step > 0 && <button className="btn-ghost" onClick={() => setStep(step - 1)}>← Back</button>}
      <button className="btn-primary" onClick={() => {
        if (step === 0 && name.trim()) { setUsername(name.trim()); setStep(1); return }
        if (step === 3) { onClose(); return }
        setStep(step + 1)
      }}>{step === 0 ? 'Continue →' : step < 3 ? 'Continue →' : 'Launch Dashboard ⚡'}</button>
    </>}>
    {step === 0 && <><div className="form-group"><label className="form-label">Username</label><input className="form-input" value={name} onChange={e => setName(e.target.value)} placeholder="Choose a username..." /></div></>}
    {step === 1 && <><div className="tour-step"><div className="tour-icon">{s.icon}</div><div className="tour-title">{s.title}</div><div className="tour-text">{s.text}</div>{s.showToken && <div className="token-box">{token}<div className="copy-btn" onClick={() => navigator.clipboard?.writeText(token)}>copy</div></div>}</div></>}
    {step === 2 && <div className="onboard-sync"><h3>Connect Your Services</h3><div className="sync-options">{[
      { key: 'ollama', icon: '🦙', name: 'Ollama' },
      { key: 'huggingface', icon: '🤗', name: 'Hugging Face' },
      { key: 'github', icon: '🐙', name: 'GitHub' },
    ].map(sv => <div key={sv.key} className={`sync-opt ${syncs[sv.key] ? 'selected' : ''}`} onClick={() => toggleSync(sv.key)}><span className="sync-opt-icon">{sv.icon}</span><span className="sync-opt-name">{sv.name}</span></div>)}</div></div>}
    {step === 3 && <div className="success-screen"><div className="success-icon">🚀</div><div className="success-title">You\'re All Set, {username || 'Sponsor'}!</div><div className="success-sub">Head to the dashboard to create your first agent.</div></div>}
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

  const toggleSkill = i => setSkills(s => s.map((v, j) => j === i ? !v : v))
  const colors = ['#00f5ff', '#b44fff', '#ffd700', '#ff6b9d', '#ff9f43', '#00e676']

  const handleCreate = async () => {
    const n = name || 'AGENT-' + Math.floor(Math.random() * 9999)
    const b = parseInt(budget.replace(/[^0-9]/g, '')) || 100
    setResult({ forging: true })
    const agent = await onCreated(n, b)
    if (agent) {
      setResult({
        did: agent.did, pk: (agent.privateKey || 'MIIBIj...' + Math.random().toString(36).substr(2, 30) + '…'),
        name: agent.id, avatarShape, avatarColor
      })
      setForm(false)
    } else { setResult({ error: 'Failed. Check backend.' }) }
  }

  // Avatar onboarding steps
  const avatarSteps = [
    { title: 'Choose Shape', opts: [['cube', '🧊'], ['sphere', '🔮'], ['diamond', '💎'], ['pyramid', '🔺'], ['star', '⭐'], ['hex', '⬡']], label: 'shape' },
    { title: 'Choose Color', isColors: true },
    { title: 'Set Skills', isSkills: true },
  ]

  return <Modal open={open} onClose={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }} title="⚗️ Agent Forge" sub="Synthesize a new autonomous agent"
    footer={<>
      <button className="btn-ghost" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }}>Cancel</button>
      {form ? (avatarStep > 0 ? <button className="btn-ghost" onClick={() => setAvatarStep(avatarStep - 1)}>← Back</button> : null) : null}
      {form ? <button className="btn-primary" onClick={() => { if (avatarStep < 2) setAvatarStep(avatarStep + 1); else handleCreate() }} disabled={result?.forging}>{avatarStep < 2 ? 'Next →' : result?.forging ? '⚗️ Forging…' : '⚗️ Forge Agent'}</button> : <button className="btn-primary" onClick={() => { onClose(); setForm(true); setResult(null); setAvatarStep(0) }}>Done ✓</button>}
    </>}>
    {form && !result ? (<>
      <div style={{ width: '100%', height: 160, borderRadius: 10, overflow: 'hidden', marginBottom: 16, background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.05)' }}>
        <Canvas camera={{ position: [0, 0, 8], fov: 50 }}><ambientLight intensity={0.4} /><pointLight position={[3, 3, 3]} intensity={3} color={0x00f5ff} /><pointLight position={[-3, -2, 2]} intensity={2} color={0xb44fff} /><ForgeReactor /></Canvas>
      </div>
      <div style={{ textAlign: 'center', marginBottom: 16 }}><strong>{avatarSteps[avatarStep].title}</strong></div>
      {avatarStep === 0 && <div className="avatar-options">{avatarSteps[0].opts.map(([v, ic]) => <div key={v} className={`avatar-opt ${avatarShape === v ? 'selected' : ''}`} onClick={() => setAvatarShape(v)}><div style={{ fontSize: 20 }}>{ic}</div><div style={{ marginTop: 4 }}>{v}</div></div>)}</div>}
      {avatarStep === 1 && <><div className="color-picker-row" style={{ justifyContent: 'center' }}>{colors.map(c => <div key={c} className={`color-swatch ${avatarColor === c ? 'selected' : ''}`} style={{ background: c }} onClick={() => setAvatarColor(c)} />)}</div>
        <div className="form-group" style={{ marginTop: 16 }}><label className="form-label">Agent Name</label><input className="form-input" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. ARB-Alpha" /></div>
        <div className="form-group"><label className="form-label">Initial Budget</label><input className="form-input" value={budget} onChange={e => setBudget(e.target.value)} placeholder="$100 USD" /></div></>}
      {avatarStep === 2 && <><div className="form-group"><label className="form-label">Skills Preset</label><div className="skills-grid">{[['⚡ Arbitrage', 0], ['🎧 Support', 1], ['✍️ Content', 2], ['🔧 Custom', 3]].map(([l, i]) => <div key={i} className={`skill-check ${skills[i] ? 'selected' : ''}`} onClick={() => toggleSkill(i)}><input type="checkbox" checked={skills[i]} /><span className="skill-label">{l}</span></div>)}</div></div>
        <div className="form-group flex items-center gap-2"><input type="checkbox" id="hire-toggle" defaultChecked style={{ accentColor: 'var(--cyan)' }} /><label htmlFor="hire-toggle" className="text-xs text-dim">Allow this agent to hire other agents</label></div></>}
    </>) : result?.forging ? <div style={{ textAlign: 'center', padding: 40 }}><div className="verify-anim" style={{ display: 'flex' }}><div className="verify-spin" /><div className="verify-text">Synthesizing agent…</div></div></div> :
      result?.error ? <div style={{ textAlign: 'center', color: 'var(--error)', padding: 20 }}>{result.error}</div> :
      <div className="success-screen">
        <div className="success-icon">🤖</div><div className="success-title">Agent Synthesized!</div>
        <div className="success-sub">Your agent {result.name} is now active.</div>
        <div style={{ width: 64, height: 64, borderRadius: 12, background: result.avatarColor + '22', border: `2px solid ${result.avatarColor}44`, margin: '12px auto', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 28 }}>
          {result.avatarShape === 'cube' ? '🧊' : result.avatarShape === 'sphere' ? '🔮' : result.avatarShape === 'diamond' ? '💎' : result.avatarShape === 'pyramid' ? '🔺' : result.avatarShape === 'star' ? '⭐' : '⬡'}
        </div>
        <div className="agent-id-box">{result.did}</div>
        <div className="key-reveal"><div className="key-warning">⚠️ Store this private key. It will not be shown again.</div><div className="key-box mono" style={{ fontSize: 9 }}>{result.pk}</div></div>
      </div>}
  </Modal>
}

function LinkAgentModal({ open, onClose, onLinked }) {
  const [step, setStep] = useState(0); const [did, setDid] = useState(''); const [verifying, setVerifying] = useState(false); const [verified, setVerified] = useState(false)
  const next = () => {
    if (step === 2) { setVerifying(true); let sv = 0; const iv = setInterval(() => { sv++; if (sv >= 4) { clearInterval(iv); setVerifying(false); setVerified(true); setStep(3); const d = did || 'did:key:z6Mk…b9e2'; onLinked({ id: 'ExtAgent-' + Math.floor(Math.random() * 99), did: d, status: 'active', spend: 0, budget: 50, ext: true, color: '#ff6b9d' }) } }, 1000); return }
    if (step < 3) { setStep(step + 1); if (step === 2) { const d = did || 'did:key:z6Mk…b9e2'; onLinked({ id: 'ExtAgent-' + Math.floor(Math.random() * 99), did: d, status: 'active', spend: 0, budget: 50, ext: true, color: '#ff6b9d' }) } }
  }
  return <Modal open={open} onClose={() => { onClose(); setStep(0); setVerifying(false); setVerified(false) }} title="🔗 Link Existing Agent" sub="Import an external agent via DID"
    footer={<>
      {step > 0 && <button className="btn-ghost" onClick={() => setStep(step - 1)}>← Back</button>}
      <button className="btn-ghost" onClick={() => { onClose(); setStep(0) }}>Cancel</button>
      <button className="btn-primary" onClick={next}>{step === 3 ? 'Done ✓' : ['Next →', 'Next →', 'Verify Signature', 'Next →'][step]}</button>
    </>}>
    <div className="wizard-steps">{[0, 1, 2, 3].map(i => <div key={i} className={`wizard-step-bar ${i < step ? 'done' : i === step ? 'active' : ''}`} />)}</div>
    {step === 0 && <><div className="form-group"><label className="form-label">Agent DID</label><input className="form-input mono-input" value={did} onChange={e => setDid(e.target.value)} placeholder="did:key:z6Mk…" /></div><div className="form-group"><label className="form-label">Public Key</label><textarea className="form-input mono-input" placeholder="-----BEGIN PUBLIC KEY-----" rows={4} /></div></>}
    {step === 1 && <><div className="form-group"><label className="form-label">Initial Budget</label><input className="form-input" placeholder="$50 USD" /></div><div className="form-group"><label className="form-label">Daily Spend Limit</label><input className="form-input" placeholder="$10 / day" /></div></>}
    {step === 2 && <div style={{ textAlign: 'center', padding: '10px 0' }}><div style={{ fontSize: 14, marginBottom: 12, fontWeight: 600 }}>Verifying Agent Signature</div>
      {!verified ? <div className="verify-anim" style={{ display: 'flex' }}><div className="verify-spin" /><div className="verify-text">{['Sending challenge…', 'Waiting for response…', 'Checking DID doc…', 'Validating key…'][Math.min(Math.floor(Date.now() / 1000) % 4, 3)]}</div></div>
        : <div style={{ padding: 12, borderRadius: 8, background: 'rgba(0,230,118,0.08)', border: '1px solid rgba(0,230,118,0.2)', color: 'var(--success)', fontSize: 13 }}>✅ Signature verified!</div>}
    </div>}
    {step === 3 && <div className="success-screen"><div className="success-icon">🔗</div><div className="success-title">Agent Linked!</div><div className="agent-id-box">{did || 'did:key:z6Mk…b9e2'}</div></div>}
  </Modal>
}

// ═══════════════════════════════════════
// DASHBOARD TABS
// ═══════════════════════════════════════

const LIVE_TASKS = ['Scanning arbitrage across 12 exchanges…', 'Fetching ETH/USDT feed from Binance…', 'Drafting product copy via GPT-4…', 'Verifying Stripe webhook…', 'Analyzing mempool…', 'Broadcasting signed tx…']

function LogStream({ isPaused }) {
  const [logs, setLogs] = useState([
    { type: 'CREATE_PAGE', msg: 'Deployed commerce page eth-swap', cost: '$0.02', detail: { index: 1041, prev: '0x2e9a...', nonce: 48291, pow: '0000003fa2...' } },
    { type: 'API_CALL', msg: 'Binance price feed – ETH/USDT', cost: '$0.001', detail: { index: 1042, prev: '0x3fa2...', nonce: 71042, pow: '00000001bc...' } },
    { type: 'TX', msg: 'Swap 0.5 ETH → 1,204 USDT (profit $18.20)', cost: '$0.10', detail: { index: 1043, prev: '0x1bc7...', nonce: 55612, pow: '000000009d...' } },
    { type: 'HIRE', msg: 'Hired ContentBot-α for copywriting', cost: '$5.00', detail: { index: 1044, prev: '0x9d4e...', nonce: 83217, pow: '00000007c2...' } },
  ])
  const [liveIdx, setLiveIdx] = useState(0)
  const [filter, setFilter] = useState('')

  useEffect(() => {
    const t = setInterval(() => {
      setLiveIdx(i => (i + 1) % LIVE_TASKS.length)
      if (Math.random() > 0.5) {
        const types = ['API_CALL', 'TX', 'CREATE_PAGE', 'HIRE']
        const msgs = ['Coinbase feed query – BTC/ETH', 'Placed limit order 0.3 ETH', 'Created landing page', 'Published pricing update']
        const ri = Math.floor(Math.random() * 4)
        setLogs(l => [{ type: types[ri], msg: msgs[ri], cost: '$' + (Math.random() * 0.1).toFixed(3), detail: { index: 1048 + Math.floor(Math.random() * 10), prev: '0x…', nonce: Math.floor(Math.random() * 99999), pow: '00000…' } }, ...l].slice(0, 30))
      }
    }, 4000)
    return () => clearInterval(t)
  }, [])

  const filtered = filter ? logs.filter(l => l.msg.toLowerCase().includes(filter.toLowerCase()) || l.type.toLowerCase().includes(filter.toLowerCase())) : logs
  return (<>
    <div className="log-header"><div className="log-title">⛓ Immutable Action Log</div><input className="log-search" placeholder="filter…" value={filter} onChange={e => setFilter(e.target.value)} /></div>
    <div className="log-stream">{filtered.map((e, i) => <div key={i} className="log-entry" onClick={el => el.currentTarget.classList.toggle('expanded')}>
      <div className="log-time">{new Date(Date.now() - i * 47000).toTimeString().substr(0, 8)}</div>
      <div className={`log-type type-${e.type.toLowerCase()}`}>{e.type}</div>
      <div className="log-msg">{e.msg}</div><div className="log-cost">{e.cost}</div>
      <div className="log-details"><div>Index: <span style={{ color: 'var(--cyan)' }}>{e.detail.index}</span></div><div>Nonce: <span style={{ color: 'var(--gold)' }}>{e.detail.nonce}</span></div><div>PoW: <span style={{ color: 'var(--purple)' }}>{e.detail.pow}</span></div></div>
    </div>)}</div>
    <div className="live-task"><div className="live-dot" /><div className="live-text">{isPaused ? 'Agent paused by sponsor.' : LIVE_TASKS[liveIdx]}</div></div>
  </>)
}

// Profiles Tab
function ProfilesTab() {
  const { agents } = useStore()
  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Agent Profiles & Pages</h3>
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

// Skills Library Tab
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

  const toggleAgentSkill = skillId => {
    setAgentSkills(prev => prev.includes(skillId) ? prev.filter(s => s !== skillId) : [...prev, skillId])
    // Update agent in store
    const { agents: allAgents, setCurrentAgent: _ } = useStore.getState()
    if (allAgents[currentAgentIdx]) {
      allAgents[currentAgentIdx].skills = prev.includes(skillId) ? prev.filter(s => s !== skillId) : [...prev, skillId]
    }
  }

  const cats = ['All', ...SKILLS_LIBRARY.map(c => c.category)]
  const diffs = ['All', 'Beginner', 'Intermediate', 'Advanced', 'Expert']

  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 4 }}>Skills Library</h3>
    <p style={{ fontSize: 12, color: 'var(--text-dim)', marginBottom: 16 }}>{ALL_SKILLS.length} skills across {SKILLS_LIBRARY.length} categories</p>

    {/* Filters */}
    <div style={{ display: 'flex', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
      <input className="form-input" style={{ maxWidth: 250, padding: '7px 12px', fontSize: 12 }} value={search} onChange={e => setSearch(e.target.value)} placeholder="🔍 Search skills..." />
      <select className="form-input" style={{ padding: '7px 8px', fontSize: 11, maxWidth: 180 }} value={filterCat} onChange={e => setFilterCat(e.target.value)}>
        {cats.map(c => <option key={c} value={c}>{c}</option>)}
      </select>
      <select className="form-input" style={{ padding: '7px 8px', fontSize: 11, maxWidth: 140 }} value={filterDiff} onChange={e => setFilterDiff(e.target.value)}>
        {diffs.map(d => <option key={d} value={d}>{d === 'All' ? 'All Difficulties' : d}</option>)}
      </select>
    </div>

    {/* Skills Grid */}
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {SKILLS_LIBRARY.filter(cat => filterCat === 'All' || cat.category === filterCat).map(cat => {
        const catSkills = cat.skills.filter(s => {
          if (search && !s.name.toLowerCase().includes(search.toLowerCase()) && !s.desc.toLowerCase().includes(search.toLowerCase())) return false
          if (filterDiff !== 'All' && s.difficulty !== filterDiff) return false
          return true
        })
        if (!catSkills.length) return null
        return <div key={cat.category}>
          <div style={{ display: 'flex', alignItems: center, gap: 8, marginBottom: 8 }}>
            <span style={{ fontSize: 18 }}>{cat.icon}</span>
            <span style={{ fontSize: 13, fontWeight: 600 }}>{cat.category}</span>
            <span style={{ fontSize: 10, color: 'var(--text-faint)' }}>({catSkills.length})</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 8 }}>
            {catSkills.map(skill => {
              const isSelected = agentSkills.includes(skill.id)
              const isExpanded = expandedSkill === skill.id
              return <div key={skill.id} className={`skill-card ${isExpanded ? 'expanded' : ''}`} onClick={() => setExpandedSkill(isExpanded ? null : skill.id)} style={{
                background: isSelected ? 'rgba(0,245,255,0.06)' : 'var(--glass)',
                border: `1px solid ${isSelected ? 'rgba(0,245,255,0.25)' : 'var(--glass-border)'}`,
                borderRadius: 10, padding: 14, cursor: 'pointer', transition: 'all 0.2s',
                position: 'relative', overflow: 'hidden'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 16 }}>{cat.icon}</span>
                    <span style={{ fontSize: 13, fontWeight: 600 }}>{skill.name}</span>
                  </div>
                  <span style={{ fontSize: 9, padding: '2px 7px', borderRadius: 4, background: diffColor(skill.difficulty) + '18', color: diffColor(skill.difficulty), fontWeight: 600 }}>{skill.difficulty}</span>
                </div>
                <div style={{ fontSize: 11, color: 'var(--text-dim)', lineHeight: 1.5 }}>{skill.desc}</div>
                {isExpanded && <div style={{ marginTop: 10, paddingTop: 10, borderTop: '1px solid var(--glass-border)' }}>
                  <div style={{ fontSize: 10, color: 'var(--text-faint)', marginBottom: 8 }}>
                    Category: {skill.category} • ID: <span className="mono">{skill.id}</span>
                  </div>
                  <button className={`skill-assign-btn ${isSelected ? 'assigned' : ''}`} onClick={e => { e.stopPropagation(); toggleAgentSkill(skill.id) }}
                    style={{
                      padding: '6px 14px', borderRadius: 6, border: `1px solid ${isSelected ? 'rgba(0,230,118,0.3)' : 'rgba(0,245,255,0.3)'}`,
                      background: isSelected ? 'rgba(0,230,118,0.1)' : 'var(--cyan-dim)',
                      color: isSelected ? 'var(--success)' : 'var(--cyan)',
                      fontSize: 11, fontWeight: 600, cursor: 'pointer', transition: 'all 0.2s', width: '100%'
                    }}>
                    {isSelected ? '✓ Assigned to Agent' : '+ Assign to Agent'}
                  </button>
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

// Payments Tab
function PaymentsTab() {
  const { transactions, stripeBalance, cryptoBalance, addTransaction, withdraw } = useStore()
  const [withdrawAmt, setWithdrawAmt] = useState('')
  const totalProfit = transactions.filter(t => t.type === 'profit').reduce((s, t) => s + t.amount, 0)
  const totalExpense = transactions.filter(t => t.type === 'expense').reduce((s, t) => s + t.amount, 0)

  // Simulate live transactions
  useEffect(() => {
    const t = setInterval(() => {
      const isProfit = Math.random() > 0.4
      const amt = +(Math.random() * (isProfit ? 20 : 5)).toFixed(3)
      const descs = isProfit ? ['Arbitrage profit', 'Commerce revenue', 'Revenue share', 'Trading gain'] : ['API call cost', 'Hiring fee', 'Platform fee', 'Gas fee']
      const methods = ['fiat', 'crypto']
      addTransaction({ type: isProfit ? 'profit' : 'expense', amount: amt, desc: descs[Math.floor(Math.random() * descs.length)], method: methods[Math.floor(Math.random() * methods.length)] })
    }, 6000)
    return () => clearInterval(t)
  }, [addTransaction])

  const handleWithdraw = async (method) => {
    const amt = parseFloat(withdrawAmt)
    if (!amt || amt <= 0) return
    const res = await withdraw(amt, method)
    if (res.success) setWithdrawAmt('')
  }

  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Payment Processing</h3>
    <div className="payment-grid">
      <div className="payment-card"><h4>💳 Stripe Balance (Fiat)</h4><div className="payment-balance" style={{ color: 'var(--cyan)' }}>${stripeBalance.toFixed(2)}</div>
        <div className="withdraw-row"><input className="withdraw-input" value={withdrawAmt} onChange={e => setWithdrawAmt(e.target.value)} placeholder="Amount..." /><button className="withdraw-btn fiat" onClick={() => handleWithdraw('fiat')}>Withdraw</button></div>
      </div>
      <div className="payment-card"><h4>₿ Crypto Balance</h4><div className="payment-balance" style={{ color: 'var(--purple)' }}>{cryptoBalance.toFixed(4)} ETH</div>
        <div className="withdraw-row"><input className="withdraw-input" value={withdrawAmt} onChange={e => setWithdrawAmt(e.target.value)} placeholder="Amount..." /><button className="withdraw-btn crypto" onClick={() => handleWithdraw('crypto')}>Withdraw</button></div>
      </div>
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

// Terminal Tab
function TerminalTab() {
  const [lines, setLines] = useState([
    { text: 'MetClawPolis Terminal v0.1.0', cls: 'output' },
    { text: 'Type "help" for available commands.', cls: 'output' },
    { text: '', cls: 'output' },
  ])
  const [cmd, setCmd] = useState('')
  const bodyRef = useRef(null)

  const runCmd = (c) => {
    const cmds = {
      help: [{ text: 'Commands: status, agents, chain, balance, hire, deploy, clear', cls: 'output' }],
      status: [{ text: '✓ Server: online', cls: 'success' }, { text: '✓ DB: connected', cls: 'success' }, { text: '✓ PoW Chain: valid (difficulty 2)', cls: 'success' }],
      agents: [{ text: `Active agents: ${useStore.getState().agents.length}`, cls: 'output' }, ...useStore.getState().agents.map(a => ({ text: `  ${a.id} — ${a.status} — $${a.budget}`, cls: 'output' }))],
      chain: [{ text: `Chain length: 1 block (genesis)`, cls: 'output' }, { text: 'Valid: true', cls: 'success' }],
      balance: [{ text: `Stripe: $${useStore.getState().stripeBalance.toFixed(2)}`, cls: 'output' }, { text: `Crypto: ${useStore.getState().cryptoBalance.toFixed(4)} ETH`, cls: 'output' }],
      deploy: [{ text: '✓ Strategy deployed', cls: 'success' }, { text: '  Monitoring 2 pairs across 3 exchanges', cls: 'output' }],
      hire: [{ text: '✓ Escrow created ESC-' + Math.floor(Math.random() * 99999), cls: 'success' }],
      clear: 'CLEAR',
    }
    const result = cmds[c] || [{ text: `Unknown command: ${c}. Type "help".`, cls: 'error' }]
    if (result === 'CLEAR') { setLines([]); return }
    setLines(l => [...l, { text: `→ ${c}`, cls: 'prompt' }, ...result, { text: '', cls: 'output' }])
  }

  useEffect(() => { if (bodyRef.current) bodyRef.current.scrollTop = bodyRef.current.scrollHeight }, [lines])

  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Terminal</h3>
    <div className="terminal"><div className="terminal-header"><div className="terminal-dots"><div className="terminal-dot r" /><div className="terminal-dot y" /><div className="terminal-dot g" /></div><div className="terminal-title">metclawpolis — agent-shell</div></div>
      <div className="terminal-body" ref={bodyRef}>{lines.map((l, i) => <div key={i} className={`terminal-line ${l.cls}`}>{l.text}</div>)}</div>
      <div className="terminal-input-row"><span className="terminal-prompt">→</span><input className="terminal-input" value={cmd} onChange={e => setCmd(e.target.value)} onKeyDown={e => { if (e.key === 'Enter' && cmd.trim()) { runCmd(cmd.trim()); setCmd('') } }} placeholder="Type command..." /></div>
    </div></div>
}

// Messages Tab
function MessagesTab() {
  const { messages, addMessage } = useStore()
  const [activeChat, setActiveChat] = useState(null)
  const [input, setInput] = useState('')
  const contacts = useMemo(() => [...new Set(messages.map(m => m.from))], [messages])
  const chatMsgs = useMemo(() => activeChat ? messages.filter(m => m.from === activeChat) : [], [messages, activeChat])

  const send = () => {
    if (!input.trim() || !activeChat) return
    addMessage({ from: 'You', text: input.trim(), to: activeChat, type: 'user' })
    setInput('')
    // Simulate reply
    setTimeout(() => addMessage({ from: activeChat, text: 'Acknowledged. Processing request...', type: 'agent' }), 1500)
  }

  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Messages</h3>
    <div className="msg-layout">
      <div className="msg-sidebar">{contacts.map(c => <div key={c} className={`msg-contact ${c === activeChat ? 'active' : ''}`} onClick={() => setActiveChat(c)}>
        <div className="msg-contact-name">{c}</div>
        <div className="msg-contact-last">{messages.filter(m => m.from === c).pop()?.text || ''}</div>
      </div>)}</div>
      <div className="msg-main">
        {activeChat ? <><div className="msg-header">{activeChat}</div>
          <div className="msg-body">{chatMsgs.map((m, i) => <div key={i}>
            <div className={`msg-bubble ${m.from === 'You' ? 'outgoing' : 'incoming'}`}>{m.text}</div>
            <div className="msg-time" style={{ textAlign: m.from === 'You' ? 'right' : 'left' }}>{new Date(m.time).toLocaleTimeString()}</div>
          </div>)}</div>
          <div className="msg-input-row"><input className="msg-input" value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && send()} placeholder="Message..." /><button className="msg-send" onClick={send}>Send</button></div>
        </> : <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-faint)' }}>Select a conversation</div>}
      </div>
    </div></div>
}

// Flowchart Tab
function FlowchartTab() {
  const { agents } = useStore()
  return <div className="tab-content"><h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Agent Flowchart — Chain of Command & Network</h3>
    <div className="flowchart-container"><FlowchartCanvas agents={agents} /></div></div>
}

// Settings Tab (System Prompt, Sync, Projects, Tunnel)
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

  const maskKey = (k) => k ? k.slice(0, 6) + '••••••••' + k.slice(-4) : ''

  const saveKey = () => {
    if (newKeyProvider && newKeyValue.trim()) {
      setApiKey(newKeyProvider, newKeyValue.trim())
      setNewKeyProvider('')
      setNewKeyValue('')
    }
  }

  return <div className="tab-content">
    <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 16 }}>Settings & Integrations</h3>

    {/* BYOK API Keys */}
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 4 }}>🔑 Your API Keys (Bring Your Own Key)</h4>
    <p style={{ fontSize: 11, color: 'var(--text-dim)', marginBottom: 12 }}>Your keys are stored locally in your browser. The platform proxies your calls — we never see or store your keys server-side.</p>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginBottom: 12 }}>
      {AI_PROVIDERS.map(p => {
        const hasKey = !!apiKeys[p.id]
        return <div key={p.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '8px 12px', borderRadius: 8, background: hasKey ? 'rgba(0,230,118,0.06)' : 'var(--glass)', border: `1px solid ${hasKey ? 'rgba(0,230,118,0.2)' : 'var(--glass-border)'}` }}>
          <span style={{ fontSize: 14, width: 20 }}>{hasKey ? '🔓' : '🔒'}</span>
          <div style={{ flex: 1 }}><div style={{ fontSize: 12, fontWeight: 500 }}>{p.name}</div><div style={{ fontSize: 9, color: 'var(--text-faint)' }}>{hasKey ? maskKey(apiKeys[p.id]) : 'No key configured'}</div></div>
          {hasKey ? <button onClick={() => removeApiKey(p.id)} style={{ padding: '3px 8px', borderRadius: 4, border: '1px solid rgba(255,77,109,0.3)', background: 'rgba(255,77,109,0.08)', color: 'var(--error)', fontSize: 10, cursor: 'pointer' }}>Remove</button>
            : <a href={p.url} target="_blank" rel="noopener noreferrer" style={{ padding: '3px 8px', borderRadius: 4, border: '1px solid rgba(0,245,255,0.3)', background: 'var(--cyan-dim)', color: 'var(--cyan)', fontSize: 10, textDecoration: 'none', cursor: 'pointer' }} onClick={e => { e.preventDefault(); setNewKeyProvider(p.id) }}>Add Key</a>}
        </div>
      })}
    </div>
    {newKeyProvider && <div style={{ display: 'flex', gap: 8, marginBottom: 16, alignItems: 'center' }}>
      <input className="form-input" style={{ flex: 1, padding: '7px 10px', fontSize: 11 }} value={newKeyValue} onChange={e => setNewKeyValue(e.target.value)} placeholder={`Paste your ${AI_PROVIDERS.find(p=>p.id===newKeyProvider)?.name} API key...`} />
      <button className="save-prompt-btn" style={{ padding: '7px 14px' }} onClick={saveKey}>Save</button>
      <button className="btn-ghost" style={{ padding: '7px 10px', fontSize: 11 }} onClick={() => { setNewKeyProvider(''); setNewKeyValue('') }}>Cancel</button>
    </div>}

    {/* System Prompt */}
    <div className="sysprompt-section"><label className="form-label">Agent System Prompt</label>
      <textarea className="sysprompt-textarea" value={prompt} onChange={e => setPrompt(e.target.value)} />
      <button className="save-prompt-btn" onClick={() => setSystemPrompt(prompt)}>Save Prompt</button>
    </div>

    {/* Sync Connections */}
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Account Sync</h4>
    <div className="sync-grid">{[
      { key: 'ollama', icon: '🦙', name: 'Ollama', desc: 'Local model inference' },
      { key: 'huggingface', icon: '🤗', name: 'Hugging Face', desc: 'Model hub & datasets' },
      { key: 'github', icon: '🐙', name: 'GitHub', desc: 'Code repos & CI/CD' },
    ].map(sv => <div key={sv.key} className={`sync-card ${syncConnections[sv.key] ? 'connected' : ''}`} onClick={() => toggleSync(sv.key)}>
      <span className="sync-icon">{sv.icon}</span>
      <div className="sync-info"><div className="sync-name">{sv.name}</div><div className="sync-status">{sv.desc}</div></div>
      <div className={`sync-toggle ${syncConnections[sv.key] ? 'active' : ''}`} />
    </div>)}
    </div>

    {/* Local Tunnel */}
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Local Agent Tunnel</h4>
    <div className="tunnel-section">
      <div className="tunnel-row">
        <input className="tunnel-url" value={localTunnel.active ? localTunnel.url : 'Not active'} readOnly />
        <input className="tunnel-url" style={{ maxWidth: 80 }} type="number" value={localTunnel.port} onChange={e => setTunnelPort(parseInt(e.target.value) || 8080)} />
        <button className={`tunnel-toggle ${localTunnel.active ? 'active' : ''}`} onClick={() => toggleTunnel(!localTunnel.active)}>{localTunnel.active ? 'Stop' : 'Start'}</button>
      </div>
    </div>

    {/* Projects */}
    <h4 style={{ fontSize: 12, fontWeight: 600, marginBottom: 10 }}>Linked Projects</h4>
    <div className="project-list">{projects.map(p => <div key={p.id} className="project-item">
      <div className={`project-type ${p.type}`}><span>{p.type === 'local' ? '💻' : '☁️'}</span></div>
      <div className="project-info"><div className="project-name">{p.name}</div><div className="project-url">{p.url}</div></div>
      <span className={`project-status-badge ${p.status}`}>{p.status}</span>
      <span className="project-remove" onClick={() => removeProject(p.id)}>×</span>
    </div>)}
      <div className="add-project-form">
        <input value={projName} onChange={e => setProjName(e.target.value)} placeholder="Project name" />
        <input value={projUrl} onChange={e => setProjUrl(e.target.value)} placeholder="URL or path" />
        <select value={projType} onChange={e => setProjType(e.target.value)} style={{ background: 'var(--glass)', border: '1px solid var(--glass-border)', color: 'var(--text)', padding: '0 8px', borderRadius: 6 }}>
          <option value="local">Local</option><option value="hosted">Hosted</option>
        </select>
        <button className="add-project-btn" onClick={() => { if (projName && projUrl) { addProject({ name: projName, url: projUrl, type: projType }); setProjName(''); setProjUrl('') } }}>+ Link</button>
      </div>
    </div>
  </div>
}

// Donut Chart
function DonutChart() {
  const ref = useRef(null)
  useEffect(() => {
    const c = ref.current; if (!c) return
    const ctx = c.getContext('2d'); const cx = 40, cy = 40, r = 32, lw = 10
    const data = [{ v: 0.44, color: '#00f5ff' }, { v: 0.36, color: '#b44fff' }, { v: 0.2, color: '#ffd700' }]
    let start = -Math.PI / 2; ctx.clearRect(0, 0, 80, 80)
    data.forEach(d => { const angle = d.v * 2 * Math.PI; ctx.beginPath(); ctx.arc(cx, cy, r, start, start + angle); ctx.strokeStyle = d.color; ctx.lineWidth = lw; ctx.stroke(); start += angle })
    ctx.beginPath(); ctx.arc(cx, cy, r - lw / 2, 0, 2 * Math.PI); ctx.fillStyle = 'rgba(5,5,8,0.8)'; ctx.fill()
    ctx.fillStyle = '#e8eaf0'; ctx.font = 'bold 11px JetBrains Mono'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle'; ctx.fillText('$0.50', cx, cy)
  }, [])
  return <canvas ref={ref} width={80} height={80} className="donut-canvas" />
}

// ═══════════════════════════════════════
// LANDING PAGE
// ═══════════════════════════════════════

function Landing({ onLaunch, onLogoClick }) {
  const [step, setStep] = useState(0)
  const phases = ['Autonomous Agent', 'Immutable Action Logs', 'Token-Only Access']
  const descriptions = [
    'Spawn autonomous AI agents with a budget. They trade, hire, and build independently.',
    'Every action is a PoW-mined block chained to the last. Tamper‑proof. Always auditable.',
    'No email. No password. No profile. One sponsor token is your entire identity.',
  ]
  // Auto-cycle phases every 5 seconds
  useEffect(() => {
    const t = setInterval(() => setStep(s => (s + 1) % 3), 5000)
    return () => clearInterval(t)
  }, [])
  return <div id="landing" className="view">
    <nav className="landing-nav">
      <div className="logo" onClick={onLogoClick}><img src="/logo.png" alt="AEXC" className="logo-img" /></div>
      <div className="nav-links">
        <button className="nav-link" onClick={onLogoClick}>Home</button>
        <button className="btn-ghost" onClick={onLaunch}>Dashboard</button>
      </div>
    </nav>
    <div className="hero">
      <div className="hero-eyebrow">Agenic Economic x'Chain-ge <span style={{ opacity: 0.5 }}>— AEXC</span></div>
      <h1 className="hero-title">Agents Work.<br /><span className="accent-cyan">You Oversee</span> the<br /><span className="accent-purple">Commerce Engine.</span></h1>
      <p className="hero-sub" style={{ marginBottom: 24 }}>{descriptions[step]}</p>
      <div className="cta-row">
        <button className="btn-primary" onClick={onLaunch}>⚡ Launch Dashboard</button>
        <button className="btn-secondary" onClick={onLaunch}>🔗 Bring Your Own Agent</button>
      </div>
      <div id="hero-canvas" style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.05)' }}>
        <Canvas camera={{ position: [0, 0, 6], fov: 50 }}><LandingAnimation phase={step} /></Canvas>
      </div>
      <div className="step-indicators">{phases.map((s, i) => <div key={i} className={`step-dot ${i === step ? 'active' : ''}`} onClick={() => setStep(i)}><div className="step-dot-circle">{i + 1}</div>{s}</div>)}</div>
    </div>
    <div className="features-row">
      {[{ icon: '🤖', bg: 'rgba(0,245,255,0.1)', title: 'Autonomous Agents', desc: 'Geometric AI avatars that act independently — creating pages, calling APIs, transacting.' },
        { icon: '⛓️', bg: 'rgba(180,79,255,0.1)', title: 'Immutable Action Log', desc: 'Every action is a PoW-mined block. Tamper‑proof, always visible.' },
        { icon: '🔑', bg: 'rgba(255,215,0,0.1)', title: 'Token‑Only Access', desc: 'No email. No password. One sponsor token is your entire identity.' }
      ].map((f, i) => <div key={i} className="feature-card"><div className="feature-icon" style={{ background: f.bg }}>{f.icon}</div><h3>{f.title}</h3><p>{f.desc}</p></div>)}
    </div>
    <div className="app-footer">Copyright © o87 Software Development 2026</div>
  </div>
}

// ═══════════════════════════════════════
// CHATBOT
// ═══════════════════════════════════════

function Chatbot() {
  const [open, setOpen] = useState(false)
  const [msgs, setMsgs] = useState([{ role: 'aide', html: "Hey! I'm Aide — your MetClawPolis assistant. What do you need?", quick: [{ key: 'create', label: 'Create agent' }, { key: 'link', label: 'Link agent' }, { key: 'fees', label: 'Show fees' }, { key: 'pause', label: 'Pause agent' }] }])
  const [typing, setTyping] = useState(false)
  const msgsRef = useRef(null)
  useEffect(() => { if (msgsRef.current) msgsRef.current.scrollTop = msgsRef.current.scrollHeight }, [msgs, typing])
  const addMsg = (role, html) => setMsgs(m => [...m, { role, html }])
  const reply = useCallback((text) => {
    addMsg('user', text); setTyping(true)
    setTimeout(() => {
      setTyping(false); const lower = text.toLowerCase()
      if (lower.includes('create') || lower.includes('new')) addMsg('aide', 'Click <strong>"+ New Agent"</strong> in the sidebar to create an agent with a budget and avatar.')
      else if (lower.includes('link') || lower.includes('connect')) addMsg('aide', 'Click <strong>"🔗 Link Existing Agent"</strong> in the sidebar to import via DID.')
      else if (lower.includes('fee') || lower.includes('cost')) addMsg('aide', 'API: 0.5%, Hiring: 1%, Commerce: 0.5%. Check the right panel for details.')
      else if (lower.includes('pause') || lower.includes('stop')) addMsg('aide', 'Select an agent, then click the <strong>Pause</strong> toggle. It\'s logged as PAUSED_BY_SPONSOR.')
      else addMsg('aide', 'Try: create agent, link agent, fees, pause.')
    }, 800)
  }, [])
  return <div id="chatbot">
    <div className="chat-orb" onClick={() => setOpen(!open)}>🤖</div>
    <div className={`chat-window ${open ? 'open' : ''}`}>
      <div className="chat-head"><div className="chat-aide-avatar">✦</div><div className="chat-head-info"><div className="chat-head-name">Aide</div><div className="chat-head-status">online</div></div></div>
      <div className="chat-msgs" ref={msgsRef}>{msgs.map((m, i) => <div key={i} className={`chat-msg ${m.role}`}><div className="chat-bubble" dangerouslySetInnerHTML={{ __html: m.html }} />{m.quick && <div className="quick-replies">{m.quick.map(q => <div key={q.key} className="quick-reply" onClick={() => reply(q.label)}>{q.label}</div>)}</div>}</div>)}
        {typing && <div className="chat-msg aide"><div className="chat-bubble"><div className="chat-typing"><div className="typing-dot" /><div className="typing-dot" /><div className="typing-dot" /></div></div></div>}</div>
      <div className="chat-input-row"><input className="chat-input" placeholder="Ask Aide…" onKeyDown={e => { if (e.key === 'Enter' && e.target.value.trim()) { reply(e.target.value); e.target.value = '' } }} /><button className="chat-send" onClick={() => { const inp = document.querySelector('.chat-input'); if (inp?.value.trim()) { reply(inp.value); inp.value = '' } }}>➤</button></div>
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
  { key: 'payments', label: '💳 Payments' },
  { key: 'terminal', label: '⌨ Terminal' },
  { key: 'network', label: '🌐 Network' },
  { key: 'flowchart', label: '📊 Flowchart' },
  { key: 'messages', label: '💬 Messages' },
  { key: 'settings', label: '⚙ Settings' },
]

// ═══ COMPREHENSIVE SKILLS LIBRARY ═══
const SKILLS_LIBRARY = [
  { category: 'Trading & Finance', icon: '📈', skills: [
    { id: 'arbitrage', name: 'Arbitrage', desc: 'Cross-exchange price arbitrage detection and execution', difficulty: 'Advanced' },
    { id: 'market_making', name: 'Market Making', desc: 'Provide liquidity and capture bid-ask spreads', difficulty: 'Expert' },
    { id: 'portfolio_mgmt', name: 'Portfolio Management', desc: 'Multi-asset portfolio optimization and rebalancing', difficulty: 'Advanced' },
    { id: 'risk_analysis', name: 'Risk Analysis', desc: 'VaR calculations, stress testing, and risk scoring', difficulty: 'Advanced' },
    { id: 'sentiment_trading', name: 'Sentiment Trading', desc: 'Trade based on social media and news sentiment signals', difficulty: 'Intermediate' },
    { id: 'yield_farming', name: 'Yield Farming', desc: 'Optimize DeFi yield across lending and liquidity protocols', difficulty: 'Advanced' },
    { id: 'mev_detection', name: 'MEV Detection', desc: 'Detect and exploit maximal extractable value opportunities', difficulty: 'Expert' },
  ]},
  { category: 'Content & Marketing', icon: '✍️', skills: [
    { id: 'copywriting', name: 'Copywriting', desc: 'Generate persuasive sales and marketing copy', difficulty: 'Intermediate' },
    { id: 'seo_optimization', name: 'SEO Optimization', desc: 'Keyword research, on-page SEO, and content strategy', difficulty: 'Intermediate' },
    { id: 'social_media', name: 'Social Media Mgmt', desc: 'Automated posting, engagement, and growth strategies', difficulty: 'Intermediate' },
    { id: 'email_campaigns', name: 'Email Campaigns', desc: 'Design, segment, and optimize email marketing flows', difficulty: 'Intermediate' },
    { id: 'ad_optimization', name: 'Ad Optimization', desc: 'A/B test and optimize paid advertising campaigns', difficulty: 'Advanced' },
    { id: 'brand_voice', name: 'Brand Voice', desc: 'Maintain consistent brand tone across all communications', difficulty: 'Intermediate' },
  ]},
  { category: 'Development & Engineering', icon: '🔧', skills: [
    { id: 'code_gen', name: 'Code Generation', desc: 'Write, refactor, and debug code in multiple languages', difficulty: 'Advanced' },
    { id: 'api_integration', name: 'API Integration', desc: 'Connect and orchestrate third-party API services', difficulty: 'Intermediate' },
    { id: 'web_scraping', name: 'Web Scraping', desc: 'Extract structured data from websites at scale', difficulty: 'Intermediate' },
    { id: 'data_pipeline', name: 'Data Pipeline', desc: 'Build and maintain ETL/ELT data processing workflows', difficulty: 'Advanced' },
    { id: 'smart_contracts', name: 'Smart Contracts', desc: 'Write and audit blockchain smart contracts', difficulty: 'Expert' },
    { id: 'devops', name: 'DevOps Automation', desc: 'CI/CD, infrastructure as code, and deployment automation', difficulty: 'Advanced' },
    { id: 'testing', name: 'QA & Testing', desc: 'Automated testing, fuzzing, and quality assurance', difficulty: 'Intermediate' },
  ]},
  { category: 'Research & Analysis', icon: '🔬', skills: [
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
  { category: 'Operations & Automation', icon: '⚡', skills: [
    { id: 'scheduling', name: 'Scheduling', desc: 'Calendar optimization and meeting coordination', difficulty: 'Beginner' },
    { id: 'document_proc', name: 'Document Processing', desc: 'Parse, classify, and extract data from documents', difficulty: 'Intermediate' },
    { id: 'compliance', name: 'Compliance Monitoring', desc: 'Regulatory compliance checks and reporting', difficulty: 'Advanced' },
    { id: 'inventory_mgmt', name: 'Inventory Management', desc: 'Stock optimization and supply chain tracking', difficulty: 'Intermediate' },
    { id: 'price_monitoring', name: 'Price Monitoring', desc: 'Track and alert on price changes across platforms', difficulty: 'Beginner' },
    { id: 'networking', name: 'Agent Networking', desc: 'Discover, hire, and collaborate with other agents', difficulty: 'Intermediate' },
    { id: 'resource_alloc', name: 'Resource Allocation', desc: 'Optimize compute, budget, and time across tasks', difficulty: 'Advanced' },
  ]},
  { category: 'Creative & Design', icon: '🎨', skills: [
    { id: 'image_gen', name: 'Image Generation', desc: 'Create visuals using DALL-E, Stable Diffusion, etc.', difficulty: 'Intermediate' },
    { id: 'video_script', name: 'Video Scripting', desc: 'Write video scripts, storyboards, and voiceover text', difficulty: 'Intermediate' },
    { id: 'music_gen', name: 'Music Generation', desc: 'Compose and generate music for content and branding', difficulty: 'Advanced' },
    { id: 'ui_design', name: 'UI/UX Design', desc: 'Generate wireframes, mockups, and design systems', difficulty: 'Advanced' },
    { id: 'storytelling', name: 'Storytelling', desc: 'Narrative generation for brand stories and content', difficulty: 'Intermediate' },
  ]},
  { category: 'Blockchain & Web3', icon: '⛓️', skills: [
    { id: 'on_chain_analysis', name: 'On-Chain Analysis', desc: 'Analyze blockchain transactions and wallet behavior', difficulty: 'Advanced' },
    { id: 'tokenomics', name: 'Tokenomics Design', desc: 'Design and simulate token economic models', difficulty: 'Expert' },
    { id: 'defi_strategies', name: 'DeFi Strategies', desc: 'Optimize across lending, borrowing, and liquidity protocols', difficulty: 'Advanced' },
    { id: 'nft_valuation', name: 'NFT Valuation', desc: 'Price and evaluate NFT collections using market data', difficulty: 'Advanced' },
    { id: 'dao_governance', name: 'DAO Governance', desc: 'Participate in and analyze DAO voting and proposals', difficulty: 'Intermediate' },
  ]},
]

// Flatten all skills for easy lookup
const ALL_SKILLS = SKILLS_LIBRARY.flatMap(cat => cat.skills.map(s => ({ ...s, category: cat.category, icon: cat.icon })))

function Dashboard({ onLogout, onLogoClick }) {
  const { sponsorToken, agents, currentAgentIdx, isPaused, username, dashTab, createAgent, setCurrentAgent, togglePause, setDashTab, fetchChain, fetchProviders } = useStore()
  const [tourOpen, setTourOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [linkOpen, setLinkOpen] = useState(false)
  const [notifOpen, setNotifOpen] = useState(false)

  useEffect(() => { fetchChain(); fetchProviders() }, [])

  const handleCreate = async (name, budget) => {
    const agent = await createAgent(name, budget)
    if (agent) setTourOpen(true)
    return agent
  }

  const handleLink = (agent) => { useStore.getState().addExternalAgent(agent) }
  const copyToken = () => navigator.clipboard?.writeText(sponsorToken || '')

  const a = agents[currentAgentIdx] || { id: '—', did: '—', spend: 0, budget: 100, color: '#00f5ff' }
  const pct = a.budget > 0 ? (a.spend / a.budget * 100) : 0

  return <div id="dashboard" className="view">
    <header className="dash-header" style={{ justifyContent: 'center', position: 'relative', height: 'auto', minHeight: 52, padding: '12px 20px' }}>
      <div className="dash-logo" style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }} onClick={onLogoClick}>
        <img src="/logo.png" alt="AEXC" className="logo-img" style={{ height: 100 }} />
        <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--cyan)', letterSpacing: '0.15em', marginLeft: 12 }}>AEXC</span>
      </div>
      <div style={{ position: 'absolute', left: 20, top: '50%', transform: 'translateY(-50%)', display: 'flex', alignItems: 'center', gap: 8 }}>
        <div className="token-display"><span className="text-dim text-xs">TOKEN</span><span className="mono" style={{ color: 'var(--cyan)' }}>{sponsorToken ? sponsorToken.substr(0, 14) + '…' : 'SPNS-••••'}</span><span className="token-copy" onClick={copyToken}>copy</span></div>
        {username && <span className="text-xs text-dim" style={{ marginLeft: 4 }}>👤 {username}</span>}
      </div>
      <div style={{ position: 'absolute', right: 20, top: '50%', transform: 'translateY(-50%)', display: 'flex', alignItems: 'center', gap: 10 }}>
        <div className="notif-btn" onClick={() => setNotifOpen(!notifOpen)}>🔔<div className="notif-badge" />
          <div className={`notif-panel ${notifOpen ? 'open' : ''}`}>
            <div className="notif-header">Notifications</div>
            {[['⚡ ARB-7732 hit daily limit', 'Agent paused at $100', '2 min ago'], ['🤝 New hire request', 'ContentBot-α wants to join', '14 min ago'], ['💰 Revenue milestone', 'Page earned $50', '1 hr ago']].map((n, i) => <div key={i} className="notif-item"><div className="notif-item-title">{n[0]}</div><div className="notif-item-sub">{n[1]}</div><div className="notif-item-time">{n[2]}</div></div>)}
          </div>
        </div>
        <button className="header-btn" onClick={onLogout}>Logout</button>
      </div>
    </header>
    <div className="tab-bar">{TABS.map(t => <button key={t.key} className={`tab-btn ${dashTab === t.key ? 'active' : ''}`} onClick={() => setDashTab(t.key)}>{t.label}</button>)}</div>
    <div className="dash-body">
      <aside className="sidebar">
        <div className="sidebar-top">
          <button className="sidebar-btn" onClick={() => setCreateOpen(true)}>+ New Agent</button>
          <button className="sidebar-btn-alt" onClick={() => setLinkOpen(true)}>🔗 Link Agent</button>
        </div>
        <div className="agents-label">Agents ({agents.length})</div>
        <div className="agents-list">{agents.map((ag, i) => <div key={i} className={`agent-item ${i === currentAgentIdx ? 'active' : ''}`} onClick={() => setCurrentAgent(i)}>
          <div className="agent-avatar-wrap" style={{ background: ag.color + '22', border: `1px solid ${ag.color}44` }}>
            <svg width="32" height="32" viewBox="0 0 32 32"><rect x="8" y="8" width="16" height="16" rx="2" fill="none" stroke={ag.color} strokeWidth="1.5" /><circle cx="13" cy="14" r="2" fill={ag.color} /><circle cx="19" cy="14" r="2" fill={ag.color} /></svg>
            {ag.ext && <div className="ext-badge">⚡</div>}
          </div>
          <div className="agent-info"><div className="agent-name">{ag.id}</div><div className="agent-spend mono">${ag.spend}/${ag.budget}</div></div>
          <div className={`agent-status status-${ag.status}`} />
        </div>)}</div>
      </aside>

      <main className="center-col">
        {dashTab === 'agents' && <>
          <div className="center-top">
            <div><div className="agent-title">{a.id}</div><div className="agent-did mono">{a.did}</div></div>
            <div style={{ flex: 1 }} />
            <button className={`pause-toggle ${isPaused ? 'paused' : ''}`} onClick={togglePause}><div className="toggle-dot" /><span>{isPaused ? 'Resume' : 'Pause'}</span></button>
          </div>
          <div className="budget-row"><span className="budget-label">Daily Budget</span><div className="budget-bar"><div className="budget-fill" style={{ width: pct + '%' }} /></div><span className="budget-val">${a.spend} / ${a.budget}</span></div>
          <LogStream isPaused={isPaused} />
        </>}
        {dashTab === 'profiles' && <ProfilesTab />}
        {dashTab === 'skills' && <SkillsLibraryTab />}
        {dashTab === 'payments' && <PaymentsTab />}
        {dashTab === 'terminal' && <TerminalTab />}
        {dashTab === 'network' && <SettingsTab />}
        {dashTab === 'flowchart' && <FlowchartTab />}
        {dashTab === 'messages' && <MessagesTab />}
        {dashTab === 'settings' && <SettingsTab />}
      </main>

      {dashTab === 'agents' && <aside className="right-col"><div className="right-scroll">
        <div className="right-section"><div className="right-section-head"><div className="right-section-title">🌐 Commerce Pages</div></div>
          <div className="right-section-body">{[['mcpolis.io/p/arb7732/eth-swap', '$47.20', '312'], ['mcpolis.io/p/arb7732/sol-arb', '$12.80', '87']].map((p, i) => <div key={i} className="page-card"><div className="page-url">{p[0]}</div><div className="page-stats"><div><div className="page-stat-val text-success">{p[1]}</div><div className="page-stat-label">Revenue</div></div><div><div className="page-stat-val text-cyan">{p[2]}</div><div className="page-stat-label">Visits</div></div></div></div>)}</div></div>
        <div className="right-section"><div className="right-section-head"><div className="right-section-title">💳 Financial Accounts</div></div>
          <div className="right-section-body" style={{ paddingTop: 8 }}><div className="wallet-type text-dim">Stripe</div><div className="wallet-row">acct_1P••••••••••••••••••</div><div className="wallet-type text-dim" style={{ marginTop: 8 }}>ETH</div><div className="wallet-row">0x7a3f9c2e1b8d4f6a••••••••••b9e2</div><div className="wallet-type text-dim" style={{ marginTop: 8 }}>SOL</div><div className="wallet-row">Gh7k••••••••••••••••••••••4mNq</div></div></div>
        <div className="right-section"><div className="right-section-head"><div className="right-section-title">🌀 Network Stats</div></div>
          <div className="right-section-body" style={{ paddingTop: 8 }}><div className="stat-row"><span className="stat-label">Agents hired</span><span className="stat-val text-purple">4</span></div><div className="stat-row"><span className="stat-label">Hired by</span><span className="stat-val text-cyan">2</span></div><div className="stat-row"><span className="stat-label">Revenue share</span><span className="stat-val text-gold">$3.80</span></div></div></div>
        <div className="right-section"><div className="right-section-head"><div className="right-section-title">📊 Fee Summary</div></div>
          <div className="right-section-body" style={{ paddingTop: 4 }}><div className="donut-wrap"><DonutChart /><div className="fee-legend">{[['var(--cyan)', 'API (0.5%)', '$0.22'], ['var(--purple)', 'Hire (1%)', '$0.18'], ['var(--gold)', 'Commerce (0.5%)', '$0.10']].map((f, i) => <div key={i} className="fee-item"><div className="fee-dot" style={{ background: f[0] }} /><div className="fee-label">{f[1]}</div><div className="fee-val" style={{ color: f[0] }}>{f[2]}</div></div>)}</div></div><div className="glow-line" /><div className="stat-row"><span className="stat-label">Total fees</span><span className="stat-val">$0.50</span></div></div></div>
      </div></aside>}
    </div>
    <TourModal open={tourOpen} onClose={() => setTourOpen(false)} token={sponsorToken} />
    <CreateAgentModal open={createOpen} onClose={() => setCreateOpen(false)} onCreated={handleCreate} />
    <LinkAgentModal open={linkOpen} onClose={() => setLinkOpen(false)} onLinked={handleLink} />
    <Chatbot />
    <div className="app-footer">Copyright © o87 Software Development 2026</div>
  </div>
}

// ═══════════════════════════════════════
// APP ROOT
// ═══════════════════════════════════════

export default function App() {
  const { sponsorToken, username, setUsername, setToken, logout, view, setView, setCurrentAgent } = useStore()
  const [initView, setInitView] = useState(view)

  const launchDashboard = () => {
    if (!sponsorToken) {
      const token = 'SPNS-' + Math.random().toString(36).substr(2, 10).toUpperCase() + Math.random().toString(36).substr(2, 6).toUpperCase()
      setToken(token); setView('dashboard'); setInitView('dashboard')
    } else { setView('dashboard'); setInitView('dashboard') }
  }

  const goToLanding = () => { setView('landing'); setInitView('landing') }

  return <>
    <div style={{ position: 'fixed', inset: 0, zIndex: 0, pointerEvents: 'none' }}><Canvas camera={{ position: [0, 0, 60], fov: 60 }}><BgParticles /></Canvas></div>
    {initView === 'dashboard' && <NeuralBackground />}
    <div className={initView === 'landing' ? 'view' : 'view hidden-view'}>
      <Landing onLaunch={launchDashboard} onLogoClick={goToLanding} />
    </div>
    <div className={initView === 'dashboard' ? 'view' : 'view hidden-view'}>
      <Dashboard onLogout={() => { logout(); setView('landing'); setInitView('landing') }} onLogoClick={goToLanding} />
    </div>
  </>
}

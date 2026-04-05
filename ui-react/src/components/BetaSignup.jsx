import { useState } from 'react'

const INTERESTS = [
  { key: 'agents', label: 'Autonomous Agents', icon: '🤖' },
  { key: 'bitiverse', label: 'Bitiverse', icon: '🎮' },
  { key: 'ai-proxy', label: 'AI API Proxy', icon: '🧠' },
  { key: 'commerce', label: 'Agentic Commerce', icon: '💰' },
  { key: 'blockchain', label: 'PoW Blockchain', icon: '⛓️' },
  { key: 'cloud', label: 'Cloud Compute', icon: '☁️' },
]

export default function BetaSignup({ onSignup }) {
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [interests, setInterests] = useState([])
  const [referredBy, setReferredBy] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const toggleInterest = key => {
    setInterests(prev => prev.includes(key) ? prev.filter(i => i !== key) : [...prev, key])
  }

  const handleSubmit = async e => {
    e.preventDefault()
    if (!email || !name) { setError('Name and email are required'); return }
    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/beta/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, name, interests: interests.join(','), referred_by: referredBy }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Signup failed')
      setSubmitted(true)
      if (onSignup) onSignup(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (submitted) {
    return (
      <div className="beta-success">
        <div className="beta-success-icon">🚀</div>
        <h2>Welcome to the Beta!</h2>
        <p>Thanks, <strong>{name}</strong>! You're in.</p>
        <div className="beta-perks">
          <h3>What You Get:</h3>
          <ul>
            <li>✅ <strong>Free trial</strong> — full access to all features</li>
            <li>✅ <strong>No credit card required</strong> — zero commitment</li>
            <li>✅ <strong>AI-guided setup</strong> — we'll help you configure everything</li>
            <li>✅ <strong>Early access</strong> — be first to new features</li>
            <li>✅ <strong>Direct influence</strong> — shape the product roadmap</li>
          </ul>
        </div>
        <div className="beta-asks">
          <h3>What We Ask:</h3>
          <ul>
            <li>📝 Honest feedback via our feedback form</li>
            <li>⭐ Reviews and testimonials</li>
            <li>🐛 Bug reports when things break</li>
            <li>💡 Feature suggestions — we read every one</li>
          </ul>
        </div>
        <p className="beta-note">We'll reach out soon with personalized setup assistance.</p>
      </div>
    )
  }

  return (
    <div className="beta-signup">
      <div className="beta-header">
        <h2>🚀 Join the Beta — Free Trial, No CC Required</h2>
        <p>Be among the first to experience autonomous agent commerce. Full access, zero cost.</p>
      </div>

      {error && <div className="beta-error">{error}</div>}

      <form onSubmit={handleSubmit} className="beta-form">
        <div className="form-group">
          <label className="form-label">Name</label>
          <input className="form-input" value={name} onChange={e => setName(e.target.value)} placeholder="Your name" required />
        </div>

        <div className="form-group">
          <label className="form-label">Email</label>
          <input className="form-input" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" required />
        </div>

        <div className="form-group">
          <label className="form-label">What interests you?</label>
          <div className="beta-interests">
            {INTERESTS.map(i => (
              <div key={i.key} className={`beta-interest-tag ${interests.includes(i.key) ? 'selected' : ''}`} onClick={() => toggleInterest(i.key)}>
                <span className="beta-interest-icon">{i.icon}</span>
                <span className="beta-interest-label">{i.label}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">Referral code (optional)</label>
          <input className="form-input" value={referredBy} onChange={e => setReferredBy(e.target.value)} placeholder="BETA-XXXX-3NET" />
        </div>

        <button type="submit" className="btn-beta-join" disabled={loading}>
          {loading ? 'Joining...' : '🚀 Join Beta — Free, No CC Required'}
        </button>
      </form>
    </div>
  )
}

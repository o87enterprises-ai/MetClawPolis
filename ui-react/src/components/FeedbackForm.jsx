import { useState } from 'react'

export default function FeedbackForm({ betaSignupId, onSubmit }) {
  const [rating, setRating] = useState(0)
  const [comments, setComments] = useState('')
  const [featureReqs, setFeatureReqs] = useState('')
  const [bugs, setBugs] = useState('')
  const [hasSharedContent, setHasSharedContent] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [promoCode, setPromoCode] = useState('')
  const [discountTier, setDiscountTier] = useState('')

  const handleSubmit = async e => {
    e.preventDefault()
    if (rating === 0) { setError('Please select a rating'); return }
    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/beta/feedback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          beta_signup_id: betaSignupId,
          rating,
          comments,
          feature_requests: featureReqs,
          bugs_reported: bugs,
          has_shared_content: hasSharedContent,
        }),
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Feedback submission failed')
      setPromoCode(data.promo_code)
      setDiscountTier(data.discount)
      setSubmitted(true)
      if (onSubmit) onSubmit(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (submitted) {
    const isUltra = discountTier === '1percent'
    return (
      <div className="feedback-success">
        <div className="feedback-success-icon">🎉</div>
        <h2>Thank You for Your Feedback!</h2>
        <p>Your honest input directly shapes MetClawPolis.</p>

        <div className="promo-reveal">
          <div className="promo-reveal-label">YOUR EXCLUSIVE PROMO CODE</div>
          <div className="promo-reveal-code">{promoCode}</div>
          <div className="promo-reveal-desc">
            {isUltra
              ? 'Ultra 1% network fee (vs standard 7%) — for sharing content about MetClawPolis!'
              : 'Reduced 3% network fee (vs standard 7%) — thanks for your honest feedback!'}
          </div>
        </div>

        {hasSharedContent ? (
          <div className="feedback-bonus">
            <h3>🎬 Already shared content? You got the 1% rate!</h3>
            <p>If you haven't yet, creating a short video or live stream about MetClawPolis qualifies you for our ultra-low 1% network fee.</p>
          </div>
        ) : (
          <div className="feedback-upgrade">
            <h3>📹 Want an even better rate?</h3>
            <p>Create a short video or live stream about your MetClawPolis experience and share it on any platform. Then come back — your fee drops to just <strong>1%</strong>!</p>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="feedback-form">
      <div className="feedback-header">
        <h2>📝 Share Your MetClawPolis Experience</h2>
        <p>Honest feedback is the only requirement. No fluff needed.</p>
      </div>

      {error && <div className="feedback-error">{error}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="form-label">Overall Rating</label>
          <div className="star-rating">
            {[1, 2, 3, 4, 5].map(n => (
              <span key={n} className={`star ${n <= rating ? 'filled' : ''}`} onClick={() => setRating(n)}>
                {n <= rating ? '★' : '☆'}
              </span>
            ))}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">What worked well?</label>
          <textarea className="form-input feedback-textarea" value={comments} onChange={e => setComments(e.target.value)} placeholder="What did you love? What surprised you?" rows={4} />
        </div>

        <div className="form-group">
          <label className="form-label">What needs improvement?</label>
          <textarea className="form-input feedback-textarea" value={featureReqs} onChange={e => setFeatureReqs(e.target.value)} placeholder="What features or fixes would you prioritize?" rows={4} />
        </div>

        <div className="form-group">
          <label className="form-label">Bugs encountered</label>
          <textarea className="form-input feedback-textarea" value={bugs} onChange={e => setBugs(e.target.value)} placeholder="Describe any issues you ran into..." rows={3} />
        </div>

        <div className="form-group">
          <label className="form-label">📹 Did you create or share a short video or live stream about MetClawPolis?</label>
          <div className="content-share-toggle">
            <div className={`toggle-option ${!hasSharedContent ? 'selected' : ''}`} onClick={() => setHasSharedContent(false)}>
              <span>No</span>
              <small>Get 3% fee discount</small>
            </div>
            <div className={`toggle-option ${hasSharedContent ? 'selected' : ''}`} onClick={() => setHasSharedContent(true)}>
              <span>Yes!</span>
              <small>Get 1% fee discount</small>
            </div>
          </div>
        </div>

        <button type="submit" className="btn-feedback-submit" disabled={loading}>
          {loading ? 'Submitting...' : 'Submit Feedback & Get Promo Code →'}
        </button>
      </form>
    </div>
  )
}

import { useState, useEffect } from 'react'

export default function ProviderMarketplace({ agentId, onClose, onPurchase }) {
  const [tab, setTab] = useState('ai') // 'ai' | 'cloud'
  const [providers, setProviders] = useState([])
  const [cloudProviders, setCloudProviders] = useState([])
  const [selectedAI, setSelectedAI] = useState(null)
  const [selectedCloud, setSelectedCloud] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/providers').then(r => r.json()),
      fetch('/api/cloud/providers').then(r => r.json()),
    ]).then(([aiData, cloudData]) => {
      // Transform AI providers map to array
      const aiArr = aiData.providers ? Object.entries(aiData.providers).map(([key, val]) => ({
        key, ...val
      })) : []
      setProviders(aiArr)
      setCloudProviders(cloudData.providers || [])
      setLoading(false)
    }).catch(err => {
      setError(err.message)
      setLoading(false)
    })
  }, [])

  if (loading) return <div style={centerStyle}>Loading marketplace...</div>
  if (error) return <div style={{ ...centerStyle, color: '#ff4444' }}>{error}</div>

  return (
    <div className="provider-marketplace" style={modalStyle} onClick={e => { if (e.target === e.currentTarget) onClose() }}>
      <div style={contentStyle}>
        {/* Header */}
        <div style={headerStyle}>
          <h2 style={{ fontSize: 18, fontWeight: 800, margin: 0, color: '#00f5ff' }}>🛒 Provider Marketplace</h2>
          <p style={{ fontSize: 10, color: '#667', margin: '4px 0 0' }}>AI Models & Cloud Compute — 7% network fee</p>
          <button onClick={onClose} style={closeBtnStyle}>×</button>
        </div>

        {/* Tabs */}
        <div style={tabBarStyle}>
          <button onClick={() => setTab('ai')} style={{
            ...tabBtnStyle,
            ...(tab === 'ai' ? { background: '#00f5ff', color: '#0a0e1a' } : {}),
          }}>🧠 AI Models ({providers.length})</button>
          <button onClick={() => setTab('cloud')} style={{
            ...tabBtnStyle,
            ...(tab === 'cloud' ? { background: '#00f5ff', color: '#0a0e1a' } : {}),
          }}>☁️ Cloud Compute ({cloudProviders.length})</button>
        </div>

        {/* AI Models Tab */}
        {tab === 'ai' && (
          <div style={gridStyle}>
            {providers.map(p => (
              <div key={p.key} onClick={() => setSelectedAI(p)} style={{
                ...cardStyle,
                ...(selectedAI?.key === p.key ? { border: '2px solid #00f5ff', background: 'rgba(0,245,255,0.05)' } : {}),
              }}>
                <div style={{ fontSize: 14, fontWeight: 700, color: '#e8eaf0', marginBottom: 4 }}>{p.name}</div>
                <div style={{ fontSize: 9, color: '#667', marginBottom: 8 }}>{p.model}</div>
                <div style={priceRowStyle}>
                  <span style={priceLabelStyle}>Input</span>
                  <span style={priceValueStyle}>${p.input_cost_per_m_tokens?.toFixed(2)}</span>
                  <span style={{ ...priceLabelStyle, marginLeft: 8 }}>Output</span>
                  <span style={priceValueStyle}>${p.output_cost_per_m_tokens?.toFixed(2)}</span>
                </div>
                <div style={{ fontSize: 9, color: '#ff8800', marginTop: 8 }}>+ 7% network fee</div>
              </div>
            ))}
          </div>
        )}

        {/* Cloud Compute Tab */}
        {tab === 'cloud' && (
          <div style={gridStyle}>
            {cloudProviders.map(cp => (
              <div key={cp.name} style={cloudProviderCardStyle}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
                  <span style={{ fontSize: 24 }}>{cp.icon}</span>
                  <div>
                    <div style={{ fontSize: 14, fontWeight: 700, color: '#e8eaf0' }}>{cp.name}</div>
                    <div style={{ fontSize: 9, color: '#667' }}>{cp.description}</div>
                  </div>
                </div>
                {cp.plans.map(plan => (
                  <div key={plan.id} onClick={() => setSelectedCloud({ ...plan, provider: cp.name })} style={{
                    ...planCardStyle,
                    ...(selectedCloud?.id === plan.id ? { border: '2px solid #00f5ff', background: 'rgba(0,245,255,0.05)' } : {}),
                  }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontSize: 12, fontWeight: 600, color: '#e8eaf0' }}>{plan.name}</span>
                      <span style={{ fontSize: 12, fontWeight: 700, color: '#00f5ff' }}>
                        {plan.price_hourly > 0 ? `$${plan.price_hourly}/hr` : `$${plan.price_monthly}/mo`}
                      </span>
                    </div>
                    <div style={{ fontSize: 9, color: '#667', marginTop: 4 }}>
                      {plan.cpu} · {plan.memory} · {plan.storage} · {plan.bandwidth}
                    </div>
                    <div style={{ fontSize: 9, color: '#ff8800', marginTop: 4 }}>+ 7% network fee · {plan.type === 'on_demand' ? 'Pay-as-you-go' : 'Flat rate'}</div>
                  </div>
                ))}
              </div>
            ))}
          </div>
        )}

        {/* Transparent Pricing Banner */}
        <div style={pricingBannerStyle}>
          <div style={{ fontSize: 12, fontWeight: 700, color: '#ffd700', marginBottom: 4 }}>
            💰 7% Network Fee — What's Included
          </div>
          <div style={pricingGridStyle}>
            <div style={pricingItemStyle}>🚫 <strong>No Ads</strong><br /><span style={pricingSubStyle}>Clean, interruption-free experience</span></div>
            <div style={pricingItemStyle}>🔧 <strong>Easy API Management</strong><br /><span style={pricingSubStyle}>One dashboard for all providers</span></div>
            <div style={pricingItemStyle}>⚡ <strong>Instant Access</strong><br /><span style={pricingSubStyle}>Start using services immediately</span></div>
            <div style={pricingItemStyle}>🎛️ <strong>Full Environment Control</strong><br /><span style={pricingSubStyle}>Configure everything your way</span></div>
          </div>
          <div style={{ fontSize: 9, color: '#ff8800', marginTop: 8 }}>
            ⚠️ NOTE: Changing models while an agent is in the Bitiverse may cause delays.
          </div>
        </div>

        {/* Footer with selection */}
        {(selectedAI || selectedCloud) && (
          <div style={footerStyle}>
            <div style={{ fontSize: 11, color: '#667' }}>
              Selected: <strong style={{ color: '#e8eaf0' }}>{selectedAI?.name || selectedCloud?.provider + ' ' + selectedCloud?.name}</strong>
            </div>
            <button onClick={() => onPurchase?.(selectedAI || selectedCloud)} style={purchaseBtnStyle}>
              Purchase & Connect →
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// Styles
const modalStyle = {
  position: 'fixed', inset: 0, zIndex: 9999,
  background: 'rgba(5,8,16,0.9)', backdropFilter: 'blur(8px)',
  display: 'flex', alignItems: 'center', justifyContent: 'center',
  padding: 20,
}
const contentStyle = {
  background: 'linear-gradient(180deg, #0C1020, #080C18)',
  border: '1px solid #ffffff22', borderRadius: 16,
  maxWidth: 900, width: '100%', maxHeight: '85vh',
  display: 'flex', flexDirection: 'column', overflow: 'hidden',
}
const headerStyle = { padding: '20px 24px 12px', borderBottom: '1px solid #ffffff11', position: 'relative' }
const closeBtnStyle = {
  position: 'absolute', top: 16, right: 16, background: 'none', border: 'none',
  color: '#666', fontSize: 24, cursor: 'pointer', lineHeight: 1,
}
const tabBarStyle = { display: 'flex', gap: 4, padding: '12px 24px 0' }
const tabBtnStyle = {
  padding: '8px 16px', borderRadius: '8px 8px 0 0', border: 'none',
  background: 'rgba(255,255,255,0.05)', color: '#889', fontWeight: 600,
  fontSize: 12, cursor: 'pointer',
}
const gridStyle = { padding: 16, display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 8, overflow: 'auto', flex: 1 }
const cardStyle = {
  padding: 12, borderRadius: 8, background: 'rgba(255,255,255,0.03)',
  border: '1px solid #ffffff11', cursor: 'pointer',
  transition: 'all 0.2s',
}
const priceRowStyle = { display: 'flex', alignItems: 'center', fontSize: 10 }
const priceLabelStyle = { color: '#667' }
const priceValueStyle = { color: '#00f5ff', fontWeight: 600, fontFamily: 'monospace' }
const cloudProviderCardStyle = {
  padding: 16, borderRadius: 8, background: 'rgba(255,255,255,0.03)',
  border: '1px solid #ffffff11', marginBottom: 8,
}
const planCardStyle = {
  padding: '8px 12px', borderRadius: 6, background: 'rgba(255,255,255,0.02)',
  border: '1px solid #ffffff08', cursor: 'pointer', marginTop: 6,
}
const pricingBannerStyle = {
  margin: '0 24px', padding: 16, borderRadius: 8,
  background: 'linear-gradient(135deg, rgba(255,215,0,0.05), rgba(180,79,255,0.05))',
  border: '1px solid #ffd70033',
}
const pricingGridStyle = { display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }
const pricingItemStyle = { fontSize: 10, color: '#ccc', textAlign: 'center' }
const pricingSubStyle = { fontSize: 8, color: '#889' }
const footerStyle = {
  padding: '12px 24px', borderTop: '1px solid #ffffff11',
  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
}
const purchaseBtnStyle = {
  padding: '10px 24px', borderRadius: 8, border: 'none',
  background: 'linear-gradient(90deg, #00f5ff, #b44fff)', color: '#0a0e1a',
  fontWeight: 700, fontSize: 13, cursor: 'pointer',
}
const centerStyle = { display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', color: '#667' }

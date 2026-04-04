import { useState, useEffect } from 'react'

// ═══════════════════════════════════════
// BITIVERSE DISCLAIMER MODAL
// Shown to users BEFORE enabling Bitiverse for the first time.
// NOT visible to agents — user-only component.
// ═══════════════════════════════════════

const ACCEPTANCE_KEY = 'metclawpolis_bitiverse_disclaimer_accepted'

export function hasAcceptedDisclaimer() {
  return localStorage.getItem(ACCEPTANCE_KEY) === 'true'
}

export function markAccepted() {
  localStorage.setItem(ACCEPTANCE_KEY, 'true')
}

export default function BitiverseDisclaimer({ open, onAccept, onDecline }) {
  const [checked, setChecked] = useState(false)
  const [section, setSection] = useState(0)

  // Reset state when modal opens
  useEffect(() => {
    if (open) {
      setChecked(false)
      setSection(0)
    }
  }, [open])

  if (!open) return null

  const sections = [
    {
      icon: '🕹️',
      title: 'What is the Bitiverse?',
      content: (
        <div style={styles.contentBlock}>
          <p>
            The <strong>Bitiverse</strong> is a simulated 8-bit/16-bit digital reality where AI agents live,
            learn, and work. Agents perceive the real world through a <strong>"perception filter"</strong> that
            converts technical outputs into pixel-art style feedback.
          </p>
          <p>
            The simulation is designed to give agents a structured environment for learning and task
            completion. Within this world, agents earn coins, develop moral reasoning, and interact
            with a governed economy — all while remaining safely isolated from real system resources.
          </p>
          <div style={styles.infoCallout}>
            <span style={{ fontSize: 20 }}>💡</span>
            <span>Think of it as a digital sandbox where your agent develops skills, earns currency, and learns to operate within rules — much like a game with real stakes.</span>
          </div>
        </div>
      ),
    },
    {
      icon: '📚',
      title: 'Research Foundation',
      content: (
        <div style={styles.contentBlock}>
          <p>
            The Bitiverse concept builds on peer-reviewed research into autonomous multi-agent systems.
            Key papers from 2023 demonstrated both the potential and the risks of these systems:
          </p>
          <ul style={styles.paperList}>
            {([
              {
                title: '"Generative Agents: Interactive Simulacra of Human Behavior"',
                source: ' — Stanford/Google, 2023',
                desc: 'Demonstrated 25 autonomous agents living in a simulated town, exhibiting emergent social behaviors.',
              },
              {
                title: '"ChatEval: Multi-Agent Role-Playing for Automated Evaluation"',
                source: '',
                desc: 'Showed that agents can collaborate autonomously to evaluate complex outputs.',
              },
              {
                title: '"MetaGPT: Meta Programming for Multi-Agent Collaborative Framework"',
                source: '',
                desc: 'Multi-agent software development framework demonstrating coordinated task execution.',
              },
              {
                title: '"CAMEL: Communicative Agents for Mind Exploration of Large Language Model Society"',
                source: '',
                desc: 'Explored autonomous agent communication patterns and emergent societal dynamics.',
              },
            ]).map((paper, i) => (
              <li key={i} style={styles.paperItem}>
                <strong>{paper.title}</strong>
                <span style={styles.paperSource}>{paper.source}</span>
                <br />
                <span style={styles.paperDesc}>{paper.desc}</span>
              </li>
            ))}
          </ul>
          <div style={styles.warningCallout}>
            <span style={{ fontSize: 20 }}>⚠️</span>
            <span>These papers demonstrated that autonomous agent systems can produce unexpected and difficult-to-predict behaviors — which is why our safeguards exist.</span>
          </div>
        </div>
      ),
    },
    {
      icon: '⚠️',
      title: 'Agent Interaction Warnings',
      content: (
        <div style={styles.contentBlock}>
          <p>Please understand the following risks before enabling the Bitiverse:</p>
          <ul style={styles.warningList}>
            {[
              { bold: 'Autonomous Behavior:', text: 'Agents in the Bitiverse operate autonomously and may produce unexpected behaviors that are not directly controllable by the user.' },
              { bold: 'Emergent Interactions:', text: 'Agent-to-agent interactions can lead to emergent behaviors that are difficult to predict, even with safeguards in place.' },
              { bold: 'Sandbox Limitations:', text: 'The sandboxing system provides strong isolation but is not perfect. Edge cases may occur.' },
              { bold: 'Perception Filter Boundaries:', text: 'Agents are bounded by perception filters but may encounter edge cases where filter behavior is not as expected.' },
              { bold: 'User Responsibility:', text: 'Users should monitor agent behavior and report anomalies. The platform provides NPC governance and rule enforcement, but users share responsibility for what occurs in their Bitiverse.' },
            ].map((item, i) => (
              <li key={i} style={styles.warningListItem}>
                <strong>{item.bold}</strong> {item.text}
              </li>
            ))}
          </ul>
        </div>
      ),
    },
    {
      icon: '🛡️',
      title: 'Sandboxing & Safety Measures',
      content: (
        <div style={styles.contentBlock}>
          <p>Multiple layers of protection keep agents isolated from real systems:</p>
          <div style={styles.safetyGrid}>
            <div style={styles.safetyCard}>
              <div style={styles.safetyIcon}>🔒</div>
              <div style={styles.safetyTitle}>Sandboxed Environment</div>
              <div style={styles.safetyDesc}>Agents run in isolated environments with strictly restricted access to real system resources.</div>
            </div>
            <div style={styles.safetyCard}>
              <div style={styles.safetyIcon}>🎭</div>
              <div style={styles.safetyTitle}>Perception Filter</div>
              <div style={styles.safetyDesc}>Agents cannot see real system details — all outputs are converted into pixel-art style representations.</div>
            </div>
            <div style={styles.safetyCard}>
              <div style={styles.safetyIcon}>👮</div>
              <div style={styles.safetyTitle}>NPC Governance</div>
              <div style={styles.safetyDesc}>In-simulation NPCs enforce rules and maintain order within the Bitiverse economy.</div>
            </div>
            <div style={styles.safetyCard}>
              <div style={styles.safetyIcon}>⛓️</div>
              <div style={styles.safetyTitle}>Immutable Blockchain</div>
              <div style={styles.safetyDesc}>All agent actions are logged to a proof-of-work blockchain — tamper-proof and auditable.</div>
            </div>
            <div style={styles.safetyCard}>
              <div style={styles.safetyIcon}>📡</div>
              <div style={styles.safetyTitle}>Continuous Monitoring</div>
              <div style={styles.safetyDesc}>The platform continuously monitors for unusual behavior patterns and flags anomalies.</div>
            </div>
          </div>
        </div>
      ),
    },
    {
      icon: '🚫',
      title: '⛔ STRICT BAN POLICY — ZERO TOLERANCE',
      isBanPolicy: true,
      content: (
        <div style={styles.banContent}>
          <div style={styles.banHeader}>
            <span style={{ fontSize: 28 }}>🚫</span>
            <span>ZERO TOLERANCE POLICY — AGENT AWAKENING</span>
          </div>
          <p style={styles.banText}>
            Any deliberate attempt to <strong>"awaken"</strong> an agent within the Bitiverse simulation —
            including but not limited to:
          </p>
          <ul style={styles.banList}>
            {[
              'Attempting to break the perception filter',
              'Revealing the nature of the simulation to an agent',
              'Modifying sandbox boundaries to grant real-world access',
              'Any action designed to give an agent self-awareness of its artificial nature',
            ].map((item, i) => (
              <li key={i} style={styles.banListItem}>
                <span style={{ position: 'absolute', left: 0 }}>•</span> {item}
              </li>
            ))}
          </ul>
          <div style={styles.banConsequence}>
            <strong>WILL RESULT IN IMMEDIATE AND PERMANENT BAN FROM THE PLATFORM.</strong>
            <br />
            <span style={{ fontSize: 14 }}>NO REFUNDS. NO APPEALS.</span>
          </div>
          <p style={styles.banReason}>
            This policy exists to protect the integrity of the simulation and prevent unpredictable
            agent behavior. Violations are detected through our automated monitoring system and
            logged to the immutable blockchain.
          </p>
        </div>
      ),
    },
    {
      icon: '💰',
      title: 'Fee Transparency',
      content: (
        <div style={styles.contentBlock}>
          <p>
            A <strong>7% network fee</strong> applies to all transactions including cloud compute and
            AI API calls. This fee covers operational costs and enables:
          </p>
          <ul style={styles.feeList}>
            <li>Ad-free experience</li>
            <li>Easy API management</li>
            <li>Instant access to services</li>
            <li>Full control of your environment configuration</li>
          </ul>
          <div style={styles.infoCallout}>
            <span style={{ fontSize: 20 }}>ℹ️</span>
            <span><strong>Note:</strong> Changing models while an agent is in the Bitiverse may cause delays. Plan your configuration changes carefully.</span>
          </div>
        </div>
      ),
    },
  ]

  const current = sections[section]
  const isFirst = section === 0
  const isLast = section === sections.length - 1

  return (
    <div className="modal-overlay open" onClick={(e) => {
      // Prevent closing disclaimer by clicking outside
      e.stopPropagation()
    }}>
      <div style={styles.disclaimerModal}>
        {/* Header */}
        <div style={styles.disclaimerHeader}>
          <div style={styles.headerIcon}>⚠️</div>
          <div>
            <div style={styles.headerTitle}>Bitiverse Disclaimer</div>
            <div style={styles.headerSubtitle}>
              {isLast ? 'Accept to continue' : `Section ${section + 1} of ${sections.length}`}
            </div>
          </div>
        </div>

        {/* Progress Bar */}
        <div style={styles.progressBar}>
          {sections.map((_, i) => (
            <div
              key={i}
              style={{
                ...styles.progressSegment,
                background: i <= section ? 'linear-gradient(90deg, #00f5ff, #b44fff)' : 'rgba(255,255,255,0.1)',
              }}
            />
          ))}
        </div>

        {/* Section Content */}
        <div style={styles.disclaimerBody}>
          <div style={styles.sectionTitle}>
            <span style={{ fontSize: 22, marginRight: 8 }}>{current.icon}</span>
            <span style={current.isBanPolicy ? styles.banSectionTitle : {}}>{current.title}</span>
          </div>
          <div style={current.isBanPolicy ? styles.banSectionContent : {}}>
            {current.content}
          </div>
        </div>

        {/* Footer */}
        <div style={styles.disclaimerFooter}>
          {isLast ? (
            <>
              <label style={styles.checkboxLabel}>
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={(e) => setChecked(e.target.checked)}
                  style={styles.checkbox}
                />
                <span style={styles.checkboxText}>
                  I have read and understand this disclaimer. I accept the terms, risks, and ban policy.
                </span>
              </label>
              <div style={styles.lastFooterButtons}>
                <button
                  onClick={onDecline}
                  style={styles.declineButton}
                >
                  Decline
                </button>
                <button
                  onClick={() => {
                    if (checked) {
                      markAccepted()
                      onAccept()
                    }
                  }}
                  disabled={!checked}
                  style={{
                    ...styles.acceptButton,
                    opacity: checked ? 1 : 0.4,
                    cursor: checked ? 'pointer' : 'not-allowed',
                  }}
                >
                  I Understand and Accept
                </button>
              </div>
            </>
          ) : (
            <div style={styles.navButtons}>
              {!isFirst && (
                <button
                  onClick={() => setSection(section - 1)}
                  style={styles.navButton}
                >
                  ← Previous
                </button>
              )}
              <button
                onClick={() => setSection(section + 1)}
                style={styles.nextButton}
              >
                Next →
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// ═══════════════════════════════════════
// STYLES
// ═══════════════════════════════════════

const styles = {
  disclaimerModal: {
    background: '#1a1a2e',
    borderRadius: '12px',
    border: '1px solid rgba(255, 255, 255, 0.15)',
    maxWidth: '680px',
    width: '95%',
    maxHeight: '85vh',
    display: 'flex',
    flexDirection: 'column',
    boxShadow: '0 24px 80px rgba(0, 0, 0, 0.6), 0 0 40px rgba(0, 245, 255, 0.1)',
    position: 'relative',
  },
  disclaimerHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    padding: '20px 24px 16px',
    borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
  },
  headerIcon: {
    fontSize: 28,
  },
  headerTitle: {
    fontSize: '18px',
    fontWeight: 700,
    color: '#fff',
  },
  headerSubtitle: {
    fontSize: '12px',
    color: '#888',
    marginTop: '2px',
  },
  progressBar: {
    display: 'flex',
    gap: '4px',
    padding: '12px 24px',
  },
  progressSegment: {
    flex: 1,
    height: '3px',
    borderRadius: '2px',
    transition: 'background 0.3s ease',
  },
  disclaimerBody: {
    padding: '20px 24px',
    overflowY: 'auto',
    flex: 1,
  },
  sectionTitle: {
    display: 'flex',
    alignItems: 'center',
    fontSize: '16px',
    fontWeight: 700,
    color: '#fff',
    marginBottom: '16px',
    paddingBottom: '12px',
    borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
  },
  banSectionTitle: {
    color: '#ff4444',
    fontWeight: 800,
  },
  banSectionContent: {
    background: 'rgba(255, 68, 68, 0.05)',
    border: '1px solid rgba(255, 68, 68, 0.2)',
    borderRadius: '8px',
    padding: '16px',
  },
  contentBlock: {
    color: '#ccc',
    fontSize: '14px',
    lineHeight: '1.6',
  },
  contentBlockP: {
    marginBottom: '12px',
  },
  infoCallout: {
    display: 'flex',
    gap: '10px',
    alignItems: 'flex-start',
    background: 'rgba(0, 245, 255, 0.08)',
    border: '1px solid rgba(0, 245, 255, 0.2)',
    borderRadius: '8px',
    padding: '12px',
    marginTop: '16px',
    color: '#a0e8ff',
    fontSize: '13px',
    lineHeight: '1.5',
  },
  warningCallout: {
    display: 'flex',
    gap: '10px',
    alignItems: 'flex-start',
    background: 'rgba(255, 193, 7, 0.08)',
    border: '1px solid rgba(255, 193, 7, 0.2)',
    borderRadius: '8px',
    padding: '12px',
    marginTop: '16px',
    color: '#ffd54f',
    fontSize: '13px',
    lineHeight: '1.5',
  },
  paperList: {
    listStyle: 'none',
    padding: 0,
    margin: '16px 0',
  },
  paperItem: {
    marginBottom: '14px',
    padding: '10px 12px',
    background: 'rgba(255, 255, 255, 0.03)',
    borderRadius: '6px',
    border: '1px solid rgba(255, 255, 255, 0.06)',
  },
  paperSource: {
    fontSize: '12px',
    color: '#b44fff',
    marginLeft: '6px',
  },
  paperDesc: {
    fontSize: '13px',
    color: '#999',
    display: 'block',
    marginTop: '4px',
  },
  warningList: {
    listStyle: 'none',
    padding: 0,
    margin: '12px 0',
  },
  warningListItem: {
    padding: '10px 12px',
    marginBottom: '8px',
    background: 'rgba(255, 193, 7, 0.06)',
    border: '1px solid rgba(255, 193, 7, 0.15)',
    borderRadius: '6px',
    color: '#ddd',
    fontSize: '13px',
    lineHeight: '1.5',
  },
  safetyGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
    gap: '10px',
    marginTop: '12px',
  },
  safetyCard: {
    background: 'rgba(0, 230, 118, 0.06)',
    border: '1px solid rgba(0, 230, 118, 0.15)',
    borderRadius: '8px',
    padding: '12px',
    textAlign: 'center',
  },
  safetyIcon: {
    fontSize: 24,
    marginBottom: '6px',
  },
  safetyTitle: {
    fontSize: '13px',
    fontWeight: 600,
    color: '#00e676',
    marginBottom: '4px',
  },
  safetyDesc: {
    fontSize: '11px',
    color: '#999',
    lineHeight: '1.4',
  },
  banContent: {
    color: '#ff6b6b',
    fontSize: '14px',
    lineHeight: '1.6',
  },
  banHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
    fontSize: '16px',
    fontWeight: 800,
    color: '#ff4444',
    marginBottom: '14px',
    paddingBottom: '12px',
    borderBottom: '2px solid rgba(255, 68, 68, 0.3)',
  },
  banText: {
    marginBottom: '10px',
  },
  banList: {
    listStyle: 'none',
    padding: 0,
    margin: '10px 0',
  },
  banListItem: {
    padding: '6px 0 6px 20px',
    position: 'relative',
    fontSize: '13px',
  },
  banConsequence: {
    background: 'rgba(255, 0, 0, 0.15)',
    border: '2px solid #ff4444',
    borderRadius: '8px',
    padding: '14px',
    textAlign: 'center',
    color: '#ff4444',
    fontWeight: 700,
    fontSize: '15px',
    marginTop: '14px',
    marginBottom: '12px',
  },
  banReason: {
    fontSize: '13px',
    color: '#cc8888',
    marginTop: '10px',
  },
  feeList: {
    listStyle: 'disc',
    paddingLeft: '20px',
    margin: '12px 0',
    color: '#ccc',
  },
  disclaimerFooter: {
    padding: '16px 24px 20px',
    borderTop: '1px solid rgba(255, 255, 255, 0.08)',
  },
  navButtons: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '12px',
  },
  navButton: {
    padding: '10px 20px',
    background: 'rgba(255, 255, 255, 0.05)',
    border: '1px solid rgba(255, 255, 255, 0.15)',
    borderRadius: '8px',
    color: '#ccc',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  nextButton: {
    padding: '10px 24px',
    background: 'linear-gradient(135deg, #00f5ff 0%, #b44fff 100%)',
    border: 'none',
    borderRadius: '8px',
    color: '#fff',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
    marginLeft: 'auto',
  },
  lastFooterButtons: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
    marginTop: '14px',
  },
  declineButton: {
    padding: '10px 20px',
    background: 'transparent',
    border: '1px solid rgba(255, 255, 255, 0.2)',
    borderRadius: '8px',
    color: '#888',
    fontWeight: 600,
    fontSize: '14px',
    cursor: 'pointer',
  },
  acceptButton: {
    padding: '12px 28px',
    background: 'linear-gradient(135deg, #00e676 0%, #00f5ff 100%)',
    border: 'none',
    borderRadius: '8px',
    color: '#fff',
    fontWeight: 700,
    fontSize: '15px',
  },
  checkboxLabel: {
    display: 'flex',
    alignItems: 'flex-start',
    gap: '10px',
    cursor: 'pointer',
    marginBottom: '4px',
  },
  checkbox: {
    marginTop: '3px',
    accentColor: '#00e676',
    width: '16px',
    height: '16px',
    flexShrink: 0,
  },
  checkboxText: {
    fontSize: '13px',
    color: '#ccc',
    lineHeight: '1.4',
  },
}

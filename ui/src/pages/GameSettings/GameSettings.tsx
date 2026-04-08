import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import {
  getInviteStatus,
  generateInvite,
  revokeInvite,
} from '../../api/invite'
import type { InviteStatusResponse, InviteTokenResponse } from '../../api/invite'
import './GameSettings.css'

const EXPIRY_OPTIONS = [
  { label: '24 Hours', value: '24h' },
  { label: '7 Days', value: '7d' },
  { label: 'No Expiry', value: 'never' },
]

export default function GameSettings() {
  useDocumentTitle('Game Settings')
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()

  const [status, setStatus] = useState<InviteStatusResponse | null>(null)
  const [activeToken, setActiveToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [expiresIn, setExpiresIn] = useState('24h')
  const [generating, setGenerating] = useState(false)
  const [revoking, setRevoking] = useState(false)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!gameId) return
    let cancelled = false
    setLoading(true)
    setError(null)

    getInviteStatus(gameId)
      .then(s => {
        if (!cancelled) {
          setStatus(s)
          if (s.has_active_invite && s.token) {
            setActiveToken(s.token)
          }
          setLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load invite status.')
          setLoading(false)
        }
      })

    return () => { cancelled = true }
  }, [gameId])

  const handleGenerate = async () => {
    if (!gameId) return
    setGenerating(true)
    setError(null)
    try {
      const result: InviteTokenResponse = await generateInvite(gameId, expiresIn)
      setActiveToken(result.token)
      setStatus({
        has_active_invite: true,
        token: result.token,
        expires_at: result.expires_at ?? undefined,
        created_at: result.created_at,
      })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to generate invite link.')
    } finally {
      setGenerating(false)
    }
  }

  const handleRevoke = async () => {
    if (!gameId) return
    setRevoking(true)
    setError(null)
    try {
      await revokeInvite(gameId)
      setActiveToken(null)
      setStatus({ has_active_invite: false })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to revoke invite link.')
    } finally {
      setRevoking(false)
    }
  }

  const handleCopy = async () => {
    if (!activeToken) return
    const url = `${window.location.origin}/join/${activeToken}`
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setError('Could not copy to clipboard.')
    }
  }

  const formatExpiry = (expiresAt?: string) => {
    if (!expiresAt) return 'Never'
    return new Date(expiresAt).toLocaleString()
  }

  const formatCreated = (createdAt?: string) => {
    if (!createdAt) return '—'
    return new Date(createdAt).toLocaleString()
  }

  return (
    <div className="gsettings-page">
      <div className="gsettings-inner">
        <button className="gsettings-back-btn" onClick={() => navigate(`/games/${gameId}`)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Campaign
        </button>

        <header className="gsettings-header">
          <div className="gsettings-title-rule" aria-hidden="true">
            <span /><span className="gsettings-ornament">⚙</span><span />
          </div>
          <h1 className="gsettings-title">Campaign Settings</h1>
          <div className="gsettings-title-rule" aria-hidden="true">
            <span /><span className="gsettings-ornament">✦</span><span />
          </div>
        </header>

        {error && (
          <div className="gsettings-error" role="alert">{error}</div>
        )}

        {loading ? (
          <div className="gsettings-loading">
            <span className="gsettings-loading-rune" aria-hidden="true">᚛</span>
            <p>Loading settings…</p>
          </div>
        ) : (
          <section className="gsettings-section">
            <div className="gsettings-section-header">
              <div className="gsettings-section-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                  <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
                </svg>
              </div>
              <div>
                <h2 className="gsettings-section-title">Magic Link Invite</h2>
                <p className="gsettings-section-desc">
                  Generate a link that lets adventurers join this campaign directly.
                </p>
              </div>
            </div>

            {status?.has_active_invite && activeToken ? (
              <div className="gsettings-invite-active">
                <div className="gsettings-invite-meta">
                  <div className="gsettings-meta-row">
                    <span className="gsettings-meta-label">Created</span>
                    <span className="gsettings-meta-value">{formatCreated(status.created_at)}</span>
                  </div>
                  <div className="gsettings-meta-row">
                    <span className="gsettings-meta-label">Expires</span>
                    <span className="gsettings-meta-value">{formatExpiry(status.expires_at)}</span>
                  </div>
                </div>

                <div className="gsettings-invite-url">
                  <code className="gsettings-link-preview">
                    {window.location.origin}/join/{activeToken}
                  </code>
                </div>

                <div className="gsettings-invite-actions">
                  <button
                    className="gsettings-btn gsettings-btn--primary"
                    onClick={handleCopy}
                    disabled={revoking}
                  >
                    {copied ? (
                      <>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                        Copied!
                      </>
                    ) : (
                      <>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                        </svg>
                        Copy Link
                      </>
                    )}
                  </button>
                  <button
                    className="gsettings-btn gsettings-btn--danger"
                    onClick={handleRevoke}
                    disabled={revoking}
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="3 6 5 6 21 6" />
                      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
                      <path d="M10 11v6M14 11v6" />
                    </svg>
                    {revoking ? 'Revoking…' : 'Revoke'}
                  </button>
                </div>
              </div>
            ) : (
              <div className="gsettings-invite-generate">
                <div className="gsettings-field">
                  <label className="gsettings-label" htmlFor="expires-in">Link Expiry</label>
                  <select
                    id="expires-in"
                    className="gsettings-select"
                    value={expiresIn}
                    onChange={e => setExpiresIn(e.target.value)}
                    disabled={generating}
                  >
                    {EXPIRY_OPTIONS.map(opt => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  className="gsettings-btn gsettings-btn--primary"
                  onClick={handleGenerate}
                  disabled={generating}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="18" cy="5" r="3" />
                    <circle cx="6" cy="12" r="3" />
                    <circle cx="18" cy="19" r="3" />
                    <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" />
                    <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" />
                  </svg>
                  {generating ? 'Forging the link…' : 'Generate Link'}
                </button>
              </div>
            )}
          </section>
        )}
      </div>
    </div>
  )
}

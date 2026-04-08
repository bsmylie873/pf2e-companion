import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import { validateInvite, redeemInvite } from '../../api/invite'
import type { InviteValidationResponse } from '../../api/invite'
import './JoinGame.css'

export default function JoinGame() {
  useDocumentTitle('Join Game')
  const { token } = useParams<{ token: string }>()
  const navigate = useNavigate()
  const { isAuthenticated, isLoading } = useAuth()

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [gameInfo, setGameInfo] = useState<InviteValidationResponse | null>(null)
  const [joining, setJoining] = useState(false)
  const [alreadyMember, setAlreadyMember] = useState(false)

  // Redirect to login with returnTo if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate(`/?returnTo=/join/${token}`, { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate, token])

  // Validate the invite token once authenticated
  useEffect(() => {
    if (isLoading || !isAuthenticated || !token) return

    let cancelled = false
    setLoading(true)
    setError(null)

    validateInvite(token)
      .then(info => {
        if (!cancelled) {
          setGameInfo(info)
          setLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'This invitation is invalid or has expired.')
          setLoading(false)
        }
      })

    return () => { cancelled = true }
  }, [token, isAuthenticated, isLoading])

  const handleJoin = async () => {
    if (!token || !gameInfo) return
    setJoining(true)
    setError(null)
    try {
      const result = await redeemInvite(token)
      if (result.already_member) {
        setAlreadyMember(true)
      }
      navigate(`/games/${result.game_id}`, { replace: true })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to join the game.')
      setJoining(false)
    }
  }

  if (isLoading) return null

  return (
    <div className="join-page">
      <div className="join-bg-sigils" aria-hidden="true" />

      <div className="join-card">
        {/* Shield motif */}
        <div className="join-shield" aria-hidden="true">
          <svg viewBox="0 0 100 110" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M50 4 L90 18 L90 52 C90 75 70 96 50 106 C30 96 10 75 10 52 L10 18 Z"
              stroke="currentColor"
              strokeWidth="1.5"
              fill="none"
              opacity="0.5"
            />
            <path
              d="M50 14 L80 25 L80 52 C80 70 65 88 50 97 C35 88 20 70 20 52 L20 25 Z"
              stroke="currentColor"
              strokeWidth="0.75"
              fill="none"
              opacity="0.3"
            />
            <circle cx="50" cy="52" r="12" stroke="currentColor" strokeWidth="1" opacity="0.45" />
            <circle cx="50" cy="52" r="4" fill="currentColor" opacity="0.6" />
            <path
              d="M50 40 L50 64 M38 52 L62 52"
              stroke="currentColor"
              strokeWidth="0.75"
              opacity="0.35"
              strokeLinecap="round"
            />
          </svg>
        </div>

        {loading ? (
          <div className="join-loading">
            <div className="join-loading-rune" aria-hidden="true">᚛</div>
            <p>Consulting the arcane registry…</p>
          </div>
        ) : error ? (
          <div className="join-error-state">
            <div className="join-error-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
            </div>
            <h2 className="join-title">Invitation Unavailable</h2>
            <div className="join-title-rule" aria-hidden="true">
              <span />✦<span />
            </div>
            <p className="join-error-message">{error}</p>
            <button className="join-btn join-btn--secondary" onClick={() => navigate('/games')}>
              Return to Campaigns
            </button>
          </div>
        ) : gameInfo ? (
          <div className="join-content">
            <header className="join-header">
              <p className="join-pretitle">You have been summoned to join</p>
              <h2 className="join-title">{gameInfo.game_title}</h2>
              <div className="join-title-rule" aria-hidden="true">
                <span />✦<span />
              </div>
              <p className="join-subtitle">
                Accept this invitation to enter the realm and begin your adventure.
              </p>
            </header>

            {alreadyMember && (
              <div className="join-notice">
                You are already a member of this campaign. Redirecting…
              </div>
            )}

            {error && (
              <div className="join-error">{error}</div>
            )}

            <div className="join-actions">
              <button
                className="join-btn join-btn--primary"
                onClick={handleJoin}
                disabled={joining}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
                  <polyline points="10 17 15 12 10 7" />
                  <line x1="15" y1="12" x2="3" y2="12" />
                </svg>
                <span>{joining ? 'Entering the realm…' : 'Join Game'}</span>
              </button>
              <button
                className="join-btn join-btn--secondary"
                onClick={() => navigate('/games')}
                disabled={joining}
              >
                Decline
              </button>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  )
}

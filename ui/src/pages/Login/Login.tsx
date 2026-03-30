import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { getPreferences } from '../../api/preferences'
import './Login.css'

type Mode = 'login' | 'register'

export default function Login() {
  const navigate = useNavigate()
  const { login, register, isAuthenticated, isLoading } = useAuth()

  const [mode, setMode] = useState<Mode>('login')
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      getPreferences()
        .then(prefs => {
          if (prefs.default_game_id) {
            navigate(`/games/${prefs.default_game_id}`, { replace: true })
          } else {
            navigate('/games', { replace: true })
          }
        })
        .catch(() => navigate('/games', { replace: true }))
    }
  }, [isAuthenticated, isLoading, navigate])

  if (isLoading) return null

  const switchMode = (next: Mode) => {
    setMode(next)
    setError(null)
    setEmail('')
    setPassword('')
    setConfirmPassword('')
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (mode === 'register' && password !== confirmPassword) {
      setError('Passphrases do not match.')
      return
    }

    setIsSubmitting(true)
    try {
      if (mode === 'login') {
        await login(username, password)
      } else {
        await register(username, email, password)
      }
      try {
        const prefs = await getPreferences()
        if (prefs.default_game_id) {
          navigate(`/games/${prefs.default_game_id}`, { replace: true })
        } else {
          navigate('/games', { replace: true })
        }
      } catch {
        navigate('/games', { replace: true })
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Something went wrong.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-bg-runes" aria-hidden="true" />

      <div className="login-card">
        <div className="login-crest" aria-hidden="true">
          <svg viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="60" cy="60" r="50" stroke="currentColor" strokeWidth="0.75" strokeDasharray="6 4" opacity="0.4" />
            <circle cx="60" cy="60" r="38" stroke="currentColor" strokeWidth="1" opacity="0.5" />
            <path d="M60 10 L65 54 L60 60 L55 54 Z" fill="currentColor" opacity="0.6" />
            <path d="M110 60 L66 65 L60 60 L66 55 Z" fill="currentColor" opacity="0.4" />
            <path d="M60 110 L55 66 L60 60 L65 66 Z" fill="currentColor" opacity="0.6" />
            <path d="M10 60 L54 55 L60 60 L54 65 Z" fill="currentColor" opacity="0.4" />
            <circle cx="60" cy="60" r="8" stroke="currentColor" strokeWidth="1.5" opacity="0.7" />
            <circle cx="60" cy="60" r="3" fill="currentColor" opacity="0.9" />
            <path d="M60 2 L60 12 M60 108 L60 118 M2 60 L12 60 M108 60 L118 60"
              stroke="currentColor" strokeWidth="0.75" opacity="0.35" strokeLinecap="round" />
            <path d="M60 22 L62 36 L60 38 L58 36 Z M60 82 L58 84 L60 98 L62 84 Z
                     M22 60 L36 58 L38 60 L36 62 Z M82 60 L84 62 L98 60 L84 58 Z"
              fill="currentColor" opacity="0.35" />
          </svg>
        </div>

        <header className="login-header">
          <h1 className="login-title">PF2E Companion</h1>
          <p className="login-subtitle">{mode === 'login' ? 'Enter the Realm' : 'Forge Your Legend'}</p>
          <div className="login-title-rule" aria-hidden="true">
            <span />✦<span />
          </div>
        </header>

        <form className="login-form" onSubmit={handleSubmit}>
          <div className="login-field">
            <label className="login-label" htmlFor="username">Adventurer</label>
            <input
              id="username"
              type="text"
              className="login-input"
              placeholder="Your name, adventurer..."
              autoComplete="username"
              value={username}
              onChange={e => setUsername(e.target.value)}
              required
            />
          </div>

          {mode === 'register' && (
            <div className="login-field">
              <label className="login-label" htmlFor="email">Sending Stone</label>
              <input
                id="email"
                type="email"
                className="login-input"
                placeholder="Your email address..."
                autoComplete="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
              />
            </div>
          )}

          <div className="login-field">
            <label className="login-label" htmlFor="password">Passphrase</label>
            <input
              id="password"
              type="password"
              className="login-input"
              placeholder="Speak the arcane word..."
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
            />
          </div>

          {mode === 'register' && (
            <div className="login-field">
              <label className="login-label" htmlFor="confirmPassword">Confirm Passphrase</label>
              <input
                id="confirmPassword"
                type="password"
                className="login-input"
                placeholder="Speak it once more..."
                autoComplete="new-password"
                value={confirmPassword}
                onChange={e => setConfirmPassword(e.target.value)}
                required
              />
            </div>
          )}

          {error && <div className="login-error" role="alert">{error}</div>}

          <button type="submit" className="login-btn" disabled={isSubmitting}>
            <span className="login-btn-text">
              {isSubmitting ? 'Consulting the oracle…' : mode === 'login' ? 'Begin Your Journey' : 'Forge Your Path'}
            </span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </button>
        </form>

        <div className="login-mode-toggle">
          {mode === 'login' ? (
            <button className="login-mode-btn" type="button" onClick={() => switchMode('register')}>
              New adventurer? Forge your legend →
            </button>
          ) : (
            <button className="login-mode-btn" type="button" onClick={() => switchMode('login')}>
              ← Return to the gates
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

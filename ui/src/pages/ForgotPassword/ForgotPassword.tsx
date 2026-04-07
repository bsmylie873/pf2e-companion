import { useState } from 'react'
import { Link } from 'react-router-dom'
import { forgotPassword } from '../../api/auth'
import './ForgotPassword.css'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      await forgotPassword(email)
    } catch {
      // Intentionally swallow errors — always show success to prevent enumeration
    } finally {
      setIsSubmitting(false)
      setSubmitted(true)
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
          <p className="login-subtitle">Recover Your Passphrase</p>
          <div className="login-title-rule" aria-hidden="true">
            <span />✦<span />
          </div>
        </header>

        {submitted ? (
          <div className="login-form">
            <div
              className="login-error"
              role="status"
              style={{
                background: 'rgba(100, 140, 60, 0.08)',
                borderColor: 'rgba(100, 140, 60, 0.35)',
                color: 'var(--color-text-muted)',
              }}
            >
              The raven has been dispatched. If a matching sending stone is found in our records, a ritual of restoration shall arrive presently.
            </div>
          </div>
        ) : (
          <form className="login-form" onSubmit={handleSubmit}>
            <div className="login-field">
              <label className="login-label" htmlFor="email">Sending Stone</label>
              <input
                id="email"
                type="email"
                className="login-input"
                placeholder="Your registered email..."
                autoComplete="email"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
              />
            </div>

            <button type="submit" className="login-btn" disabled={isSubmitting}>
              <span className="login-btn-text">
                {isSubmitting ? 'Dispatching the raven…' : 'Dispatch the Raven'}
              </span>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </button>
          </form>
        )}

        <div className="login-mode-toggle">
          <Link to="/" className="login-mode-btn">← Return to the gates</Link>
        </div>
      </div>
    </div>
  )
}

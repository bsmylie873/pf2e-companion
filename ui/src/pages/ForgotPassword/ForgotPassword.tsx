import { useState } from 'react'
import { Link } from 'react-router-dom'
import { forgotPassword } from '../../api/auth'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import './ForgotPassword.css'

export default function ForgotPassword() {
  useDocumentTitle('Forgot Password')
  const [email, setEmail] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [resetToken, setResetToken] = useState<string | null>(null)
  const [submitted, setSubmitted] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    try {
      const result = await forgotPassword(email)
      setResetToken(result.token)
    } catch {
      // Intentionally swallow errors — always show success to prevent enumeration
    } finally {
      setIsSubmitting(false)
      setSubmitted(true)
    }
  }

  const resetUrl = resetToken ? `${window.location.origin}/reset-password?token=${resetToken}` : null

  const handleCopy = async () => {
    if (!resetUrl) return
    try {
      await navigator.clipboard.writeText(resetUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch { /* noop */ }
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
            {resetUrl ? (
              <>
                <div
                  className="login-error"
                  role="status"
                  style={{
                    background: 'rgba(100, 140, 60, 0.08)',
                    borderColor: 'rgba(100, 140, 60, 0.35)',
                    color: 'var(--color-text-muted)',
                  }}
                >
                  A restoration scroll has been conjured. Use the link below to reforge your passphrase. It expires in one hour.
                </div>
                <div className="gsettings-invite-url" style={{ margin: '0.75rem 0' }}>
                  <code className="gsettings-link-preview" style={{
                    display: 'block',
                    padding: '0.6rem 0.75rem',
                    background: 'var(--color-bg)',
                    border: '1px solid var(--color-border)',
                    borderRadius: '4px',
                    fontSize: '0.72rem',
                    fontFamily: 'var(--font-body)',
                    color: 'var(--color-text-muted)',
                    wordBreak: 'break-all',
                    lineHeight: '1.5',
                  }}>
                    {resetUrl}
                  </code>
                </div>
                <button type="button" className="login-btn" onClick={handleCopy}>
                  <span className="login-btn-text">
                    {copied ? 'Copied!' : 'Copy Link'}
                  </span>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                    {copied ? (
                      <polyline points="20 6 9 17 4 12" />
                    ) : (
                      <>
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                      </>
                    )}
                  </svg>
                </button>
              </>
            ) : (
              <div
                className="login-error"
                role="status"
                style={{
                  background: 'rgba(100, 140, 60, 0.08)',
                  borderColor: 'rgba(100, 140, 60, 0.35)',
                  color: 'var(--color-text-muted)',
                }}
              >
                If a matching sending stone is found in our records, a restoration scroll would have appeared here. Please verify your email and try again.
              </div>
            )}
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

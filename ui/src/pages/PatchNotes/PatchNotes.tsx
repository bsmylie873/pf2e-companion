import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Markdown from 'react-markdown'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import './PatchNotes.css'

interface PatchNotesData {
  version: string
  date: string
  notes: string
}

export default function PatchNotes() {
  useDocumentTitle('Patch Notes')
  const [data, setData] = useState<PatchNotesData | null>(null)
  const [error, setError] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/patch-notes.json')
      .then(r => {
        if (!r.ok) throw new Error('Failed to load')
        return r.json()
      })
      .then((d: PatchNotesData) => {
        setData(d)
        setLoading(false)
      })
      .catch(() => {
        setError(true)
        setLoading(false)
      })
  }, [])

  return (
    <div className="patch-notes-page">
      <div className="patch-notes-bg-runes" aria-hidden="true" />

      <div className="patch-notes-card">
        <div className="patch-notes-crest" aria-hidden="true">
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

        <header className="patch-notes-header">
          <h1 className="patch-notes-title">PF2E Companion</h1>
          <p className="patch-notes-subtitle">Chronicle of Changes</p>
          <div className="patch-notes-title-rule" aria-hidden="true">
            <span />✦<span />
          </div>
        </header>

        {loading && (
          <div className="patch-notes-loading">
            <span className="patch-notes-loading-text">Unrolling the scroll…</span>
          </div>
        )}

        {error && (
          <div className="patch-notes-error" role="alert">
            The herald's scroll could not be found. Patch notes will appear here after the next deployment.
          </div>
        )}

        {data && (
          <>
            <div className="patch-notes-meta">
              <span className="patch-notes-badge">
                v{data.version}
              </span>
              <span className="patch-notes-date">{data.date}</span>
            </div>
            <div className="patch-notes-content">
              <Markdown>{data.notes}</Markdown>
            </div>
          </>
        )}

        <div className="patch-notes-footer">
          <Link to="/" className="patch-notes-back-link">← Return to the gates</Link>
        </div>
      </div>
    </div>
  )
}

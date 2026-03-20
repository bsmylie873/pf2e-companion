import { useState, useEffect } from 'react'
import type { SessionFormData } from '../../types/session'
import '../Modal/Modal.css'
import './SessionFormModal.css'

interface SessionFormModalProps {
  mode: 'create' | 'edit'
  initial?: SessionFormData
  error: string | null
  saving: boolean
  onSave: (data: SessionFormData) => void
  onClose: () => void
}

function toLocalDatetime(utcIso: string): string {
  const d = new Date(utcIso)
  const offset = d.getTimezoneOffset()
  const local = new Date(d.getTime() - offset * 60_000)
  return local.toISOString().slice(0, 16)
}

export default function SessionFormModal({ mode, initial, error, saving, onSave, onClose }: SessionFormModalProps) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [sessionNumber, setSessionNumber] = useState<string>(
    initial?.session_number != null ? String(initial.session_number) : ''
  )
  const [scheduledAt, setScheduledAt] = useState(
    initial?.scheduled_at ? toLocalDatetime(initial.scheduled_at) : ''
  )

  useEffect(() => {
    setTitle(initial?.title ?? '')
    setSessionNumber(initial?.session_number != null ? String(initial.session_number) : '')
    setScheduledAt(initial?.scheduled_at ? toLocalDatetime(initial.scheduled_at) : '')
  }, [initial])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    onSave({
      title: title.trim(),
      session_number: sessionNumber ? Number(sessionNumber) : null,
      scheduled_at: scheduledAt ? new Date(scheduledAt).toISOString() : null,
    })
  }

  const canSave = title.trim().length > 0 && !saving

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-card" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">
            {mode === 'create' ? 'Create Session' : 'Edit Session'}
          </h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <form className="session-form" onSubmit={handleSubmit}>
          <div className="session-form-field">
            <label className="session-form-label" htmlFor="session-title">Title *</label>
            <input
              id="session-title"
              className="session-form-input"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter session title…"
              autoFocus
            />
          </div>

          <div className="session-form-field">
            <label className="session-form-label" htmlFor="session-number">Session Number</label>
            <input
              id="session-number"
              className="session-form-input"
              type="number"
              min="1"
              value={sessionNumber}
              onChange={(e) => setSessionNumber(e.target.value)}
              placeholder="Optional"
            />
          </div>

          <div className="session-form-field">
            <label className="session-form-label" htmlFor="session-scheduled">Scheduled At</label>
            <input
              id="session-scheduled"
              className="session-form-input"
              type="datetime-local"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
            />
          </div>

          {error && (
            <div className="session-form-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <p>{error}</p>
            </div>
          )}

          <button className="session-form-submit" type="submit" disabled={!canSave}>
            {saving ? 'Saving…' : mode === 'create' ? 'Create Session' : 'Save Changes'}
          </button>
        </form>
      </div>
    </div>
  )
}

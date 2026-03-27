import { useState, useEffect } from 'react'
import type { NoteFormData } from '../../types/note'
import type { Session } from '../../types/session'
import '../Modal/Modal.css'
import './NoteFormModal.css'

interface NoteFormModalProps {
  mode: 'create' | 'edit'
  initial?: NoteFormData & { visibility?: 'private' | 'shared' }
  sessions: Session[]
  error: string | null
  saving: boolean
  isAuthor?: boolean
  onSave: (data: NoteFormData) => void
  onClose: () => void
}

export default function NoteFormModal({ mode, initial, sessions, error, saving, isAuthor = true, onSave, onClose }: NoteFormModalProps) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [sessionId, setSessionId] = useState<string>(initial?.session_id ?? '')
  const [visibility, setVisibility] = useState<'private' | 'shared'>(initial?.visibility ?? 'private')

  useEffect(() => {
    setTitle(initial?.title ?? '')
    setSessionId(initial?.session_id ?? '')
    setVisibility(initial?.visibility ?? 'private')
  }, [initial])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    onSave({
      title: title.trim(),
      session_id: sessionId || null,
      visibility,
    })
  }

  const canSave = title.trim().length > 0 && !saving

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-card" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">
            {mode === 'create' ? 'New Note' : 'Edit Note'}
          </h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <form className="note-form" onSubmit={handleSubmit}>
          <div className="note-form-field">
            <label className="note-form-label" htmlFor="note-title">Title *</label>
            <input
              id="note-title"
              className="note-form-input"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter note title…"
              autoFocus
            />
          </div>

          <div className="note-form-field">
            <label className="note-form-label" htmlFor="note-session">Link to Session</label>
            <select
              id="note-session"
              className="note-form-input note-form-select"
              value={sessionId}
              onChange={(e) => setSessionId(e.target.value)}
            >
              <option value="">— Unlinked —</option>
              {sessions.map(s => (
                <option key={s.id} value={s.id}>
                  {s.session_number != null ? `#${s.session_number} — ` : ''}{s.title}
                </option>
              ))}
            </select>
          </div>

          {isAuthor && (
            <div className="note-form-field">
              <label className="note-form-label">Visibility</label>
              <div className="note-form-visibility">
                <button
                  type="button"
                  className={`note-form-vis-btn${visibility === 'private' ? ' note-form-vis-btn--active' : ''}`}
                  onClick={() => setVisibility('private')}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                  </svg>
                  Private
                </button>
                <button
                  type="button"
                  className={`note-form-vis-btn${visibility === 'shared' ? ' note-form-vis-btn--active' : ''}`}
                  onClick={() => setVisibility('shared')}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="2" y1="12" x2="22" y2="12" />
                    <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                  </svg>
                  Shared
                </button>
              </div>
            </div>
          )}

          {error && (
            <div className="note-form-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <p>{error}</p>
            </div>
          )}

          <button className="note-form-submit" type="submit" disabled={!canSave}>
            {saving ? 'Saving…' : mode === 'create' ? 'Create Note' : 'Save Changes'}
          </button>
        </form>
      </div>
    </div>
  )
}

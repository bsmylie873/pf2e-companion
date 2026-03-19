import type { Session } from '../../types/session'
import './SessionCard.css'

interface SessionCardProps {
  session: Session
  isGM: boolean
  onEdit: (session: Session) => void
  onDelete: (session: Session) => void
}

export default function SessionCard({ session, isGM, onEdit, onDelete }: SessionCardProps) {
  const formattedDate = session.scheduled_at
    ? new Intl.DateTimeFormat(undefined, {
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(new Date(session.scheduled_at))
    : null

  return (
    <article className="session-card">
      <div className="session-card-body">
        {session.session_number != null && (
          <span className="session-card-number">Session #{session.session_number}</span>
        )}
        <h3 className="session-card-title">{session.title}</h3>
        {formattedDate && (
          <time className="session-card-date">{formattedDate}</time>
        )}
      </div>
      <div className="session-card-actions">
        <button
          className="session-card-btn session-card-btn--edit"
          onClick={(e) => { e.stopPropagation(); onEdit(session) }}
          aria-label="Edit session"
          title="Edit session"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
        </button>
        {isGM && (
          <button
            className="session-card-btn session-card-btn--delete"
            onClick={(e) => { e.stopPropagation(); onDelete(session) }}
            aria-label="Delete session"
            title="Delete session"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="3 6 5 6 21 6" />
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            </svg>
          </button>
        )}
      </div>
    </article>
  )
}

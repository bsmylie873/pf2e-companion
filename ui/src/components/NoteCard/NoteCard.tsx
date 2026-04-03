import type { Note } from '../../types/note'
import { extractPreviewText } from '../../utils/contentPreview'
import './NoteCard.css'

interface NoteCardProps {
  note: Note
  sessionTitle?: string
  isGM: boolean
  isAuthor: boolean
  mode?: 'list' | 'grid'
  onEdit: (note: Note) => void
  onDelete: (note: Note) => void
  onOpen: (note: Note) => void
}

export default function NoteCard({ note, sessionTitle, isGM, isAuthor, mode = 'list', onEdit, onDelete, onOpen }: NoteCardProps) {
  if (mode === 'grid') {
    const canEdit = isAuthor || isGM || note.visibility === 'editable'
    const canDelete = isAuthor || isGM
    return (
      <article className="note-card note-card--grid">
        <div className="note-card-grid-body" onClick={() => onOpen(note)}>
          <span className="note-card-grid-icon" aria-hidden>📄</span>
          <div className="note-card-grid-title-wrapper">
            <h3 className="note-card-grid-title">{note.title}</h3>
          </div>
          <time className="note-card-grid-date">
            {new Date(note.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
          </time>
          <p className="note-card-grid-preview">{extractPreviewText(note.content)}</p>
        </div>
        <div className="note-card-grid-actions">
          {canEdit && (
            <button className="note-card-btn note-card-btn--edit"
              onClick={(e) => { e.stopPropagation(); onEdit(note) }}
              aria-label="Edit note" title="Edit note">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
          )}
          {canDelete && (
            <button className="note-card-btn note-card-btn--delete"
              onClick={(e) => { e.stopPropagation(); onDelete(note) }}
              aria-label="Delete note" title="Delete note">
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

  const canEdit = isAuthor || isGM || note.visibility === 'editable'
  const canDelete = isAuthor || isGM

  const formattedDate = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(note.created_at))

  return (
    <article className="note-card note-card--clickable" onClick={() => onOpen(note)}>
      <div className="note-card-body">
        <div className="note-card-meta">
          <span className={`note-card-visibility note-card-visibility--${note.visibility}`}>
            {note.visibility === 'private' ? (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              </svg>
            ) : note.visibility === 'visible' ? (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            ) : (
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z" />
              </svg>
            )}
            {note.visibility === 'private' ? 'Private' : note.visibility === 'visible' ? 'View Only' : 'Editable'}
          </span>
          {sessionTitle && (
            <span className="note-card-session">{sessionTitle}</span>
          )}
        </div>
        <h3 className="note-card-title">{note.title}</h3>
        <time className="note-card-date">{formattedDate}</time>
      </div>
      <div className="note-card-actions">
        {canEdit && (
          <button
            className="note-card-btn note-card-btn--edit"
            onClick={(e) => { e.stopPropagation(); onEdit(note) }}
            aria-label="Edit note"
            title="Edit note"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
          </button>
        )}
        {canDelete && (
          <button
            className="note-card-btn note-card-btn--delete"
            onClick={(e) => { e.stopPropagation(); onDelete(note) }}
            aria-label="Delete note"
            title="Delete note"
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

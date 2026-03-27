import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import type { Note } from '../../types/note'
import type { JSONContent } from '@tiptap/react'
import { getNote, updateNoteContent } from '../../api/notes'
import SessionNotesEditor from '../../components/SessionNotesEditor/SessionNotesEditor'
import './NoteEditor.css'

export default function NoteEditor() {
  const { gameId, noteId } = useParams<{ gameId: string; noteId: string }>()
  const navigate = useNavigate()

  const [note, setNote] = useState<Note | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!noteId) return
    let cancelled = false
    setLoading(true)
    setError(null)
    getNote(noteId)
      .then((data) => { if (!cancelled) { setNote(data); setLoading(false) } })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load note.')
          setLoading(false)
        }
      })
    return () => { cancelled = true }
  }, [noteId])

  const handleSave = async (content: JSONContent, version: number): Promise<Note> => {
    if (!noteId) throw new Error('No note ID')
    const updated = await updateNoteContent(noteId, { content, version })
    setNote(updated)
    return updated
  }

  return (
    <div className="note-editor-page">
      <div className="note-editor-inner">
        <button className="nep-back-btn" onClick={() => navigate(`/games/${gameId}`)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Notes
        </button>

        {loading && (
          <div className="nep-loading">
            <div className="spinner-ring" />
            <p className="spinner-label">Unfurling the scroll…</p>
          </div>
        )}

        {!loading && error && (
          <div className="nep-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <p>{error}</p>
          </div>
        )}

        {!loading && !error && note && (
          <>
            <header className="nep-header">
              <div className="nep-title-rule" aria-hidden="true">
                <span /><span className="nep-title-ornament">✦</span><span />
              </div>
              <div className="nep-meta">
                <span className={`nep-visibility nep-visibility--${note.visibility}`}>
                  {note.visibility === 'private' ? '🔒 Private' : '🌐 Shared'}
                </span>
              </div>
              <h1 className="nep-title">{note.title}</h1>
              <p className="nep-subtitle">Personal Chronicle</p>
              <div className="nep-title-rule" aria-hidden="true">
                <span /><span className="nep-title-ornament">✦</span><span />
              </div>
            </header>

            <SessionNotesEditor
              initialContent={note.content as JSONContent | null}
              version={note.version}
              onSave={handleSave}
              editable={true}
              placeholder="Begin your private chronicle…"
            />
          </>
        )}
      </div>
    </div>
  )
}

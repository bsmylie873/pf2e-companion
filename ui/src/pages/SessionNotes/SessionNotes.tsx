import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import type { Session } from '../../types/session'
import type { JSONContent } from '@tiptap/react'
import { getSession, updateSessionNotes } from '../../api/sessions'
import SessionNotesEditor from '../../components/SessionNotesEditor/SessionNotesEditor'
import { useGameSocket } from '../../hooks/useGameSocket'
import type { GameSocketEvent } from '../../hooks/useGameSocket'
import './SessionNotes.css'

export default function SessionNotes() {
  const { gameId, sessionId } = useParams<{ gameId: string; sessionId: string }>()
  const navigate = useNavigate()

  const [session, setSession] = useState<Session | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!sessionId) return
    let cancelled = false
    setLoading(true)
    setError(null)
    getSession(sessionId)
      .then((data) => { if (!cancelled) { setSession(data); setLoading(false) } })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load session.')
          setLoading(false)
        }
      })
    return () => { cancelled = true }
  }, [sessionId])

  const handleSave = async (content: JSONContent): Promise<Session> => {
    if (!sessionId) throw new Error('No session ID')
    const updated = await updateSessionNotes(sessionId!, { notes: content, version: session!.version })
    setSession(updated)
    return updated
  }

  const handleGameEvent = useCallback((event: GameSocketEvent) => {
    if (event.type === 'session_updated' && event.entity_id === sessionId) {
      getSession(sessionId!).then(setSession).catch(() => {})
    } else if (event.type === 'session_deleted' && event.entity_id === sessionId) {
      navigate(`/games/${gameId}`)
    }
  }, [sessionId, gameId, navigate])

  useGameSocket(gameId, handleGameEvent)

  return (
    <div className="session-notes-page">
      <div className="session-notes-inner">
        <button className="snp-back-btn" onClick={() => navigate(`/games/${gameId}`)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Sessions
        </button>

        {loading && (
          <div className="snp-loading">
            <div className="spinner-ring" />
            <p className="spinner-label">Unfurling the scroll…</p>
          </div>
        )}

        {!loading && error && (
          <div className="snp-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <p>{error}</p>
          </div>
        )}

        {!loading && !error && session && (
          <>
            <header className="snp-header">
              <div className="snp-title-rule" aria-hidden="true">
                <span /><span className="snp-title-ornament">✦</span><span />
              </div>
              {session.session_number != null && (
                <p className="snp-session-num">Session #{session.session_number}</p>
              )}
              <h1 className="snp-title">{session.title}</h1>
              <p className="snp-subtitle">Session Chronicle</p>
              <div className="snp-title-rule" aria-hidden="true">
                <span /><span className="snp-title-ornament">✦</span><span />
              </div>
            </header>

            <SessionNotesEditor
              initialContent={session.notes as JSONContent | null}
              onSave={handleSave}
              editable={true}
            />
          </>
        )}
      </div>
    </div>
  )
}

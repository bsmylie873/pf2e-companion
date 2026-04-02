import { useState, useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import type { JSONContent } from '@tiptap/react'
import { getSession, updateSessionNotes } from '../../api/sessions'
import { getNote, updateNoteContent } from '../../api/notes'
import { listMemberships } from '../../api/memberships'
import { useAuth } from '../../context/AuthContext'
import SessionNotesEditor from '../SessionNotesEditor/SessionNotesEditor'
import './EditorModal.css'

interface EditorModalProps {
  type: 'session' | 'note'
  itemId: string
  gameId: string
  onClose: () => void
}

export default function EditorModal({ type, itemId, gameId, onClose }: EditorModalProps) {
  const { user } = useAuth()
  const previousFocusRef = useRef<HTMLElement | null>(null)
  const backdropRef = useRef<HTMLDivElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)
  const onCloseRef = useRef(onClose)

  // Keep ref current so keydown handler always calls the latest onClose
  useEffect(() => {
    onCloseRef.current = onClose
  }, [onClose])

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [title, setTitle] = useState<string | null>(null)
  const [sessionNumber, setSessionNumber] = useState<number | null>(null)
  const [noteVisibility, setNoteVisibility] = useState<string | null>(null)
  const [editorContent, setEditorContent] = useState<JSONContent | null | undefined>(undefined)
  const [version, setVersion] = useState(1)
  const [editable, setEditable] = useState(false)

  // Save/restore focus on mount/unmount
  useEffect(() => {
    previousFocusRef.current = document.activeElement as HTMLElement
    return () => {
      previousFocusRef.current?.focus()
    }
  }, [])

  // Block wheel events from reaching the map's zoom handler.
  // Uses a native capture-phase listener so it fires before any library
  // listeners (e.g. react-zoom-pan-pinch on window/document).
  useEffect(() => {
    const backdrop = backdropRef.current
    if (!backdrop) return
    const stopWheel = (e: WheelEvent) => {
      e.stopPropagation()
    }
    backdrop.addEventListener('wheel', stopWheel, { passive: false, capture: true })
    return () => backdrop.removeEventListener('wheel', stopWheel, { capture: true })
  }, [])

  // Focus trap: Escape closes, Tab cycles within the panel
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onCloseRef.current()
        return
      }

      if (e.key === 'Tab') {
        const panel = panelRef.current
        if (!panel) return

        const focusable = Array.from(
          panel.querySelectorAll<HTMLElement>(
            'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
          )
        ).filter(el => !el.hasAttribute('disabled'))

        if (focusable.length === 0) return

        const first = focusable[0]
        const last = focusable[focusable.length - 1]

        if (e.shiftKey) {
          if (document.activeElement === first) {
            e.preventDefault()
            last.focus()
          }
        } else {
          if (document.activeElement === last) {
            e.preventDefault()
            first.focus()
          }
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  // Fetch the session or note data
  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    if (type === 'session') {
      getSession(itemId)
        .then(session => {
          if (cancelled) return
          setTitle(session.title)
          setSessionNumber(session.session_number)
          setEditorContent(session.notes)
          setVersion(session.version)
          setEditable(true)
          setLoading(false)
        })
        .catch((err: unknown) => {
          if (!cancelled) {
            setError(err instanceof Error ? err.message : 'Failed to load session.')
            setLoading(false)
          }
        })
    } else {
      Promise.all([getNote(itemId), listMemberships(gameId)])
        .then(([note, memberships]) => {
          if (cancelled) return
          const isAuthor = note.user_id === user?.id
          const isGM = memberships.some(m => m.user_id === user?.id && m.is_gm)
          const canEdit = isAuthor || isGM || note.visibility === 'editable'
          setTitle(note.title)
          setNoteVisibility(note.visibility)
          setEditorContent(note.content as JSONContent | null)
          setVersion(note.version)
          setEditable(canEdit)
          setLoading(false)
        })
        .catch((err: unknown) => {
          if (!cancelled) {
            setError(err instanceof Error ? err.message : 'Failed to load note.')
            setLoading(false)
          }
        })
    }

    return () => { cancelled = true }
  }, [type, itemId, gameId, user])

  const handleSave = async (content: JSONContent): Promise<{ version: number }> => {
    if (type === 'session') {
      const updated = await updateSessionNotes(itemId, { notes: content, version })
      setVersion(updated.version)
      return { version: updated.version }
    } else {
      const updated = await updateNoteContent(itemId, { content, version })
      setVersion(updated.version)
      return { version: updated.version }
    }
  }

  return createPortal(
    <div
      className="editor-modal-backdrop"
      ref={backdropRef}
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label={title ?? 'Editor'}
    >
      <div
        className="editor-modal-panel"
        ref={panelRef}
        onClick={e => e.stopPropagation()}
      >
        <div className="editor-modal-header">
          <div className="editor-modal-title-group">
            {type === 'session' && sessionNumber != null && (
              <span className="editor-modal-session-num">Session #{sessionNumber}</span>
            )}
            {type === 'note' && noteVisibility && (
              <span className={`editor-modal-visibility editor-modal-visibility--${noteVisibility}`}>
                {noteVisibility === 'private'
                  ? '\uD83D\uDD12 Private'
                  : noteVisibility === 'visible'
                  ? '\uD83D\uDC41 View Only'
                  : '\u270F\uFE0F Editable'}
              </span>
            )}
            <h2 className="editor-modal-title">{title ?? '\u2026'}</h2>
          </div>

          <button
            className="editor-modal-close"
            onClick={onClose}
            aria-label="Close editor"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className="editor-modal-body">
          {loading && (
            <div className="editor-modal-loading">
              <div className="spinner-ring" />
              <p className="spinner-label">Unfurling the scroll\u2026</p>
            </div>
          )}

          {!loading && error && (
            <div className="editor-modal-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" width="20" height="20">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <p>{error}</p>
            </div>
          )}

          {!loading && !error && editorContent !== undefined && (
            <SessionNotesEditor
              initialContent={editorContent}
              editable={editable}
              onSave={handleSave}
              placeholder={type === 'note' ? 'Begin your private chronicle\u2026' : undefined}
            />
          )}
        </div>
      </div>
    </div>,
    document.body
  )
}

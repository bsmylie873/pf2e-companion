import { useState, useEffect, useCallback, useRef } from 'react'
import { useParams, useLocation, useNavigate } from 'react-router-dom'
import type { Session, SessionFormData } from '../../types/session'
import type { Note, NoteFormData } from '../../types/note'
import type { GameMembership } from '../../types/membership'
import type { Game } from '../../types/game'
import { listGameSessionsPaginated, createSession, updateSession, deleteSession } from '../../api/sessions'
import { listMemberships } from '../../api/memberships'
import { listGameNotesPaginated, createNote, updateNote as updateNoteApi, deleteNote } from '../../api/notes'
import { importGameBackup } from '../../api/backup'
import type { ImportSummary } from '../../api/backup'
import { apiFetch } from '../../api/client'
import { getPreferences, updatePreferences } from '../../api/preferences'
import { listFolders } from '../../api/folders'
import type { Folder } from '../../types/folder'
import { useAuth } from '../../context/AuthContext'
import SessionCard from '../../components/SessionCard/SessionCard'
import SessionFormModal from '../../components/SessionFormModal/SessionFormModal'
import NoteCard from '../../components/NoteCard/NoteCard'
import NoteFormModal from '../../components/NoteFormModal/NoteFormModal'
import ConfirmModal from '../../components/ConfirmModal/ConfirmModal'
import Modal from '../../components/Modal/Modal'
import EditCampaignForm from '../../components/EditCampaignForm/EditCampaignForm'
import Pagination from '../../components/Pagination/Pagination'
import { usePageSize } from '../../hooks/usePageSize'
import './Editor.css'

interface LocationState {
  title?: string
}

export default function Editor() {
  const { gameId } = useParams<{ gameId: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const { user } = useAuth()
  const sessionPageSize = usePageSize('sessions')
  const notePageSize = usePageSize('notes')
  const state = location.state as LocationState | null

  const [title, setTitle] = useState(state?.title ?? '')
  const [editOpen, setEditOpen] = useState(false)
  const [isDirty, setIsDirty] = useState(false)

  useEffect(() => {
    if (!state?.title && gameId) {
      apiFetch<Game>(`/games/${gameId}`).then(game => setTitle(game.title)).catch(() => {})
    }
  }, [gameId, state?.title])

  const handleClose = useCallback(() => {
    if (isDirty && !confirm('Discard unsaved changes?')) return
    setEditOpen(false)
    setIsDirty(false)
  }, [isDirty])

  const handleSuccess = useCallback((newTitle: string) => {
    setEditOpen(false)
    setIsDirty(false)
    setTitle(newTitle)
  }, [])

  const handleMapView = useCallback(() => {
    navigate(`/games/${gameId}/map`)
  }, [gameId, navigate])

  // Tab state
  const [activeTab, setActiveTab] = useState<'sessions' | 'notes'>('sessions')
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list')

  const [sessions, setSessions] = useState<Session[]>([])
  const [notes, setNotes] = useState<Note[]>([])
  const [memberships, setMemberships] = useState<GameMembership[]>([])
  const [, setGame] = useState<Game | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionFolders, setSessionFolders] = useState<Folder[]>([])
  const [noteFolders, setNoteFolders] = useState<Folder[]>([])
  const [collapsedFolders, setCollapsedFolders] = useState<Record<string, boolean>>({})

  // Shared filter panel toggle (one panel at a time, context-aware per tab)
  const [filtersOpen, setFiltersOpen] = useState(false)

  // Session filter state
  const [filterTitle, setFilterTitle] = useState('')
  const [filterNumMin, setFilterNumMin] = useState('')
  const [filterNumMax, setFilterNumMax] = useState('')
  const [filterDateFrom, setFilterDateFrom] = useState('')
  const [filterDateTo, setFilterDateTo] = useState('')

  // Session sort state
  type SortField = 'session_number' | 'title' | 'updated_at'
  type SortDir = 'asc' | 'desc'
  const [sortField, setSortField] = useState<SortField>('session_number')
  const [sortDir, setSortDir] = useState<SortDir>('asc')

  // Note filter/sort state
  type NoteSort = 'title' | 'created_at'
  const [noteSort, setNoteSort] = useState<NoteSort>('created_at')
  const [noteSessionFilter, setNoteSessionFilter] = useState<string>('') // '' = all, 'unlinked', or session UUID

  const hasActiveNoteFilters = noteSessionFilter !== '' || noteSort !== 'created_at'

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir(prev => prev === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDir('asc')
    }
  }

  const hasActiveFilters = filterTitle !== '' || filterNumMin !== '' || filterNumMax !== '' || filterDateFrom !== '' || filterDateTo !== ''

  const clearFilters = () => {
    setFilterTitle('')
    setFilterNumMin('')
    setFilterNumMax('')
    setFilterDateFrom('')
    setFilterDateTo('')
  }

  // Session modal state
  const [formOpen, setFormOpen] = useState(false)
  const [editingSession, setEditingSession] = useState<Session | null>(null)
  const [deletingSession, setDeletingSession] = useState<Session | null>(null)
  const [mutationError, setMutationError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  // Note modal state
  const [noteFormOpen, setNoteFormOpen] = useState(false)
  const [editingNote, setEditingNote] = useState<Note | null>(null)
  const [deletingNote, setDeletingNote] = useState<Note | null>(null)
  const [noteMutationError, setNoteMutationError] = useState<string | null>(null)
  const [noteSaving, setNoteSaving] = useState(false)

  // Import modal state
  const [importOpen, setImportOpen] = useState(false)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importMode, setImportMode] = useState<'merge' | 'overwrite' | null>(null)
  const [importBusy, setImportBusy] = useState(false)
  const [importResult, setImportResult] = useState<ImportSummary | null>(null)
  const [importError, setImportError] = useState<string | null>(null)
  const importFileRef = useRef<HTMLInputElement>(null)

  // Pagination state
  const [sessionPage, setSessionPage] = useState(1)
  const [sessionTotal, setSessionTotal] = useState(0)
  const [notePage, setNotePage] = useState(1)
  const [noteTotal, setNoteTotal] = useState(0)

  // Initial load: memberships, game metadata, folders
  useEffect(() => {
    if (!gameId) return
    let cancelled = false
    setError(null)

    Promise.all([
      listMemberships(gameId),
      apiFetch<Game>(`/games/${gameId}`),
      listFolders(gameId, 'session'),
      listFolders(gameId, 'note'),
    ])
      .then(([membershipsData, gameData, sFolders, nFolders]) => {
        if (!cancelled) {
          setMemberships(membershipsData)
          setGame(gameData)
          setSessionFolders(sFolders)
          setNoteFolders(nFolders)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load data.')
        }
      })

    return () => { cancelled = true }
  }, [gameId])

  // Sessions: re-fetch when page changes
  useEffect(() => {
    if (!gameId) return
    let cancelled = false
    setLoading(true)

    listGameSessionsPaginated(gameId, { page: sessionPage, limit: sessionPageSize })
      .then((resp) => {
        if (!cancelled) {
          setSessions(resp.data)
          setSessionTotal(resp.total)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load sessions.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [gameId, sessionPage, sessionPageSize])

  // Notes: re-fetch when page, sort, or session filter changes
  useEffect(() => {
    if (!gameId) return
    let cancelled = false

    const noteParams: { sort?: string; session_id?: string; unlinked?: boolean } = {}
    if (noteSort) noteParams.sort = noteSort
    if (noteSessionFilter === 'unlinked') noteParams.unlinked = true
    else if (noteSessionFilter) noteParams.session_id = noteSessionFilter

    listGameNotesPaginated(gameId, { page: notePage, limit: notePageSize, ...noteParams })
      .then((resp) => {
        if (!cancelled) {
          setNotes(resp.data)
          setNoteTotal(resp.total)
        }
      })
      .catch(() => {})

    return () => { cancelled = true }
  }, [gameId, notePage, notePageSize, noteSort, noteSessionFilter])

  useEffect(() => {
    if (!gameId) return
    getPreferences().then(prefs => {
      if (prefs.default_view_mode && prefs.default_view_mode[gameId]) {
        setViewMode(prefs.default_view_mode[gameId])
      }
    }).catch(() => {})
  }, [gameId])

  const isGM = memberships.some(m => m.user_id === user?.id && m.is_gm)

  // ── Session computed ──────────────────────────────────────────

  const sortedSessions = [...sessions].sort((a, b) => {
    const dir = sortDir === 'asc' ? 1 : -1

    if (sortField === 'session_number') {
      if (a.session_number != null && b.session_number != null) return (a.session_number - b.session_number) * dir
      if (a.session_number != null) return -1
      if (b.session_number != null) return 1
      return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    }

    if (sortField === 'title') {
      return a.title.localeCompare(b.title) * dir
    }

    // updated_at
    return (new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime()) * dir
  })

  const filteredSessions = sortedSessions.filter(s => {
    if (filterTitle !== '') {
      if (!s.title.toLowerCase().includes(filterTitle.toLowerCase())) return false
    }
    if (filterNumMin !== '') {
      const min = Number(filterNumMin)
      if (s.session_number == null || s.session_number < min) return false
    }
    if (filterNumMax !== '') {
      const max = Number(filterNumMax)
      if (s.session_number == null || s.session_number > max) return false
    }
    if (filterDateFrom !== '') {
      const from = new Date(filterDateFrom)
      from.setHours(0, 0, 0, 0)
      if (new Date(s.updated_at) < from) return false
    }
    if (filterDateTo !== '') {
      const to = new Date(filterDateTo)
      to.setHours(23, 59, 59, 999)
      if (new Date(s.updated_at) > to) return false
    }
    return true
  })

  // ── Note computed ─────────────────────────────────────────────

  const sessionMap = Object.fromEntries(sessions.map(s => [s.id, s]))

  // Server handles note filtering/sorting; notes state already reflects current filter+sort+page
  const filteredNotes = notes

  const unfiledSessions = filteredSessions.filter(s => !s.folder_id)
  const sessionsByFolder = sessionFolders.map(folder => ({
    folder,
    items: filteredSessions.filter(s => s.folder_id === folder.id),
  }))

  const unfiledNotes = filteredNotes.filter(n => !n.folder_id)
  const notesByFolder = noteFolders.map(folder => ({
    folder,
    items: filteredNotes.filter(n => n.folder_id === folder.id),
  }))

  // ── Session handlers ──────────────────────────────────────────

  const handleCreate = async (data: SessionFormData) => {
    if (!gameId) return
    setMutationError(null)
    setSaving(true)
    try {
      const created = await createSession(gameId, data)
      setSessions(prev => [...prev, created])
      setFormOpen(false)
    } catch (err: unknown) {
      setMutationError(err instanceof Error ? err.message : 'Failed to create session.')
    } finally {
      setSaving(false)
    }
  }

  const handleEdit = async (data: SessionFormData) => {
    if (!editingSession) return
    setMutationError(null)
    setSaving(true)
    try {
      const updated = await updateSession(editingSession.id, { ...data })
      setSessions(prev => prev.map(s => s.id === updated.id ? updated : s))
      setEditingSession(null)
    } catch (err: unknown) {
      setMutationError(err instanceof Error ? err.message : 'Failed to update session.')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!deletingSession) return
    setMutationError(null)
    setSaving(true)
    try {
      await deleteSession(deletingSession.id)
      setDeletingSession(null)
      const maxPage = Math.max(1, Math.ceil((sessionTotal - 1) / sessionPageSize))
      if (sessionPage > maxPage) {
        setSessionPage(maxPage) // triggers re-fetch via useEffect
      } else {
        setSessions(prev => prev.filter(s => s.id !== deletingSession.id))
        setSessionTotal(prev => prev - 1)
      }
    } catch (err: unknown) {
      setMutationError(err instanceof Error ? err.message : 'Failed to delete session.')
    } finally {
      setSaving(false)
    }
  }

  const openCreate = () => {
    setMutationError(null)
    setFormOpen(true)
  }

  const openEdit = (session: Session) => {
    setMutationError(null)
    setEditingSession(session)
  }

  const openDelete = (session: Session) => {
    setMutationError(null)
    setDeletingSession(session)
  }

  const openNotes = (session: Session) => {
    navigate(`/games/${gameId}/sessions/${session.id}/notes`)
  }

  // ── Note handlers ─────────────────────────────────────────────

  const handleCreateNote = async (data: NoteFormData) => {
    if (!gameId) return
    setNoteMutationError(null)
    setNoteSaving(true)
    try {
      const created = await createNote(gameId, data)
      setNotes(prev => [...prev, created])
      setNoteFormOpen(false)
    } catch (err: unknown) {
      setNoteMutationError(err instanceof Error ? err.message : 'Failed to create note.')
    } finally {
      setNoteSaving(false)
    }
  }

  const handleEditNote = async (data: NoteFormData) => {
    if (!editingNote) return
    setNoteMutationError(null)
    setNoteSaving(true)
    try {
      const updated = await updateNoteApi(editingNote.id, data as unknown as Record<string, unknown>)
      setNotes(prev => prev.map(n => n.id === updated.id ? updated : n))
      setEditingNote(null)
    } catch (err: unknown) {
      setNoteMutationError(err instanceof Error ? err.message : 'Failed to update note.')
    } finally {
      setNoteSaving(false)
    }
  }

  const handleDeleteNote = async () => {
    if (!deletingNote) return
    setNoteMutationError(null)
    setNoteSaving(true)
    try {
      await deleteNote(deletingNote.id)
      setDeletingNote(null)
      const maxPage = Math.max(1, Math.ceil((noteTotal - 1) / notePageSize))
      if (notePage > maxPage) {
        setNotePage(maxPage) // triggers re-fetch via useEffect
      } else {
        setNotes(prev => prev.filter(n => n.id !== deletingNote.id))
        setNoteTotal(prev => prev - 1)
      }
    } catch (err: unknown) {
      setNoteMutationError(err instanceof Error ? err.message : 'Failed to delete note.')
    } finally {
      setNoteSaving(false)
    }
  }

  const openNoteCreate = () => {
    setNoteMutationError(null)
    setNoteFormOpen(true)
  }

  const openNoteEdit = (note: Note) => {
    setNoteMutationError(null)
    setEditingNote(note)
  }

  const openNoteDelete = (note: Note) => {
    setNoteMutationError(null)
    setDeletingNote(note)
  }

  const openNoteEditor = (note: Note) => {
    navigate(`/games/${gameId}/notes/${note.id}`)
  }

  const handleImportFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setImportError(null)
    setImportResult(null)
    if (file.size > 10 * 1024 * 1024) {
      setImportError('File exceeds 10 MB limit.')
      setImportFile(null)
      e.target.value = ''
      return
    }
    setImportFile(file)
  }

  const handleImportSubmit = async () => {
    if (!importFile || !importMode || !gameId) return
    setImportError(null)
    setImportResult(null)
    setImportBusy(true)
    try {
      const result = await importGameBackup(gameId, importFile, importMode)
      setImportResult(result)
      setImportFile(null)
      setImportMode(null)
      if (importFileRef.current) importFileRef.current.value = ''
      // Refresh sessions and notes to show imported data
      setSessionPage(1)
      setNotePage(1)
    } catch (e) {
      setImportError(e instanceof Error ? e.message : 'Import failed.')
    } finally {
      setImportBusy(false)
    }
  }

  const handleImportClose = () => {
    setImportOpen(false)
    setImportFile(null)
    setImportMode(null)
    setImportResult(null)
    setImportError(null)
    if (importFileRef.current) importFileRef.current.value = ''
  }

  const handleViewModeChange = useCallback((mode: 'list' | 'grid') => {
    setViewMode(mode)
    if (!gameId) return
    getPreferences().then(prefs => {
      updatePreferences({
        default_view_mode: { ...(prefs.default_view_mode ?? {}), [gameId]: mode },
      }).catch(() => {})
    }).catch(() => {})
  }, [gameId])

  const toggleFolder = useCallback((folderId: string) => {
    setCollapsedFolders(prev => ({ ...prev, [folderId]: !prev[folderId] }))
  }, [])

  return (
    <div className="editor-page">
      <div className="editor-inner">
        <button className="editor-back-btn" onClick={() => navigate('/games')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Campaigns
        </button>

        <header className="editor-header">
          <div className="editor-title-rule" aria-hidden="true">
            <span /><span className="editor-title-ornament">✦</span><span />
          </div>
          <h1 className="editor-title">{title}</h1>
          <button
            className="editor-edit-btn"
            onClick={() => setEditOpen(true)}
            title="Edit Campaign"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
            </svg>
            Edit Campaign
          </button>
          <div className="editor-title-rule" aria-hidden="true">
            <span /><span className="editor-title-ornament">✦</span><span />
          </div>
        </header>

        {/* ── Shared toolbar: view toggle, filter, tabs ── */}
        <div className="editor-toolbar">
          <div className="editor-toolbar-left">
            <div className="sessions-view-toggle">
              <button
                className={`sessions-view-btn${viewMode === 'list' ? ' sessions-view-btn--active' : ''}`}
                onClick={() => handleViewModeChange('list')}
                title="List view"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="8" y1="6" x2="21" y2="6" />
                  <line x1="8" y1="12" x2="21" y2="12" />
                  <line x1="8" y1="18" x2="21" y2="18" />
                  <line x1="3" y1="6" x2="3.01" y2="6" />
                  <line x1="3" y1="12" x2="3.01" y2="12" />
                  <line x1="3" y1="18" x2="3.01" y2="18" />
                </svg>
              </button>
              <button
                className={`sessions-view-btn${viewMode === 'grid' ? ' sessions-view-btn--active' : ''}`}
                onClick={() => handleViewModeChange('grid')}
                title="Grid view"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                  <rect x="3" y="3" width="7" height="7" rx="1" />
                  <rect x="14" y="3" width="7" height="7" rx="1" />
                  <rect x="3" y="14" width="7" height="7" rx="1" />
                  <rect x="14" y="14" width="7" height="7" rx="1" />
                </svg>
              </button>
            </div>
            <button
              className="editor-map-btn"
              onClick={handleMapView}
              title="Map view"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="1 6 1 22 8 18 16 22 23 18 23 2 16 6 8 2 1 6" />
                <line x1="8" y1="2" x2="8" y2="18" />
                <line x1="16" y1="6" x2="16" y2="22" />
              </svg>
              Map
            </button>
            <button
              className={`sessions-filter-toggle${filtersOpen ? ' sessions-filter-toggle--active' : ''}${(activeTab === 'sessions' ? hasActiveFilters : hasActiveNoteFilters) ? ' sessions-filter-toggle--has-filters' : ''}`}
              onClick={() => setFiltersOpen(prev => !prev)}
              aria-label="Toggle filters"
              title={activeTab === 'sessions' ? 'Filter sessions' : 'Filter notes'}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
              </svg>
              {(activeTab === 'sessions' ? hasActiveFilters : hasActiveNoteFilters) && <span className="sessions-filter-badge" />}
            </button>
          </div>
          <div className="editor-tabs">
            <button
              className={`editor-tab${activeTab === 'sessions' ? ' editor-tab--active' : ''}`}
              onClick={() => { setActiveTab('sessions'); setFiltersOpen(false) }}
            >
              Sessions
            </button>
            <button
              className={`editor-tab${activeTab === 'notes' ? ' editor-tab--active' : ''}`}
              onClick={() => { setActiveTab('notes'); setFiltersOpen(false) }}
            >
              Notes
            </button>
          </div>
          <div className="editor-toolbar-right">
            <button className="editor-import-btn" onClick={() => setImportOpen(true)} title="Import backup">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            </button>
            {activeTab === 'sessions' ? (
              <button className="sessions-new-btn" onClick={openCreate}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <line x1="12" y1="5" x2="12" y2="19" />
                  <line x1="5" y1="12" x2="19" y2="12" />
                </svg>
                New Session
              </button>
            ) : (
              <button className="sessions-new-btn" onClick={openNoteCreate}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <line x1="12" y1="5" x2="12" y2="19" />
                  <line x1="5" y1="12" x2="19" y2="12" />
                </svg>
                New Note
              </button>
            )}
          </div>
        </div>

        {/* ── Sessions tab ── */}
        {activeTab === 'sessions' && (
          <section className="sessions-section">

            {filtersOpen && (
              <div className="sessions-filters">
                <div className="sessions-filter-group sessions-filter-group--full">
                  <span className="sessions-filter-group-label">Title</span>
                  <input
                    className="sessions-filter-input"
                    type="text"
                    placeholder="Search by name…"
                    value={filterTitle}
                    onChange={e => setFilterTitle(e.target.value)}
                  />
                </div>

                <div className="sessions-filters-row">
                  <div className="sessions-filter-group">
                    <span className="sessions-filter-group-label">Session Number</span>
                    <div className="sessions-filter-range">
                      <input
                        className="sessions-filter-input"
                        type="number"
                        min="1"
                        placeholder="Min"
                        value={filterNumMin}
                        onChange={e => setFilterNumMin(e.target.value)}
                      />
                      <span className="sessions-filter-separator">&ndash;</span>
                      <input
                        className="sessions-filter-input"
                        type="number"
                        min="1"
                        placeholder="Max"
                        value={filterNumMax}
                        onChange={e => setFilterNumMax(e.target.value)}
                      />
                    </div>
                  </div>

                  <div className="sessions-filter-group">
                    <span className="sessions-filter-group-label">Last Edited</span>
                    <div className="sessions-filter-range">
                      <input
                        className="sessions-filter-input"
                        type="date"
                        value={filterDateFrom}
                        onChange={e => setFilterDateFrom(e.target.value)}
                      />
                      <span className="sessions-filter-separator">&ndash;</span>
                      <input
                        className="sessions-filter-input"
                        type="date"
                        value={filterDateTo}
                        onChange={e => setFilterDateTo(e.target.value)}
                      />
                    </div>
                  </div>
                </div>

                <div className="sessions-sort">
                  <span className="sessions-sort-label">Sort by</span>
                  <div className="sessions-sort-options">
                    {([
                      ['session_number', '#'],
                      ['title', 'Title'],
                      ['updated_at', 'Edited'],
                    ] as const).map(([field, label]) => (
                      <button
                        key={field}
                        className={`sessions-sort-btn${sortField === field ? ' sessions-sort-btn--active' : ''}`}
                        onClick={() => toggleSort(field)}
                      >
                        {label}
                        {sortField === field && (
                          <svg className={`sessions-sort-arrow${sortDir === 'desc' ? ' sessions-sort-arrow--desc' : ''}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M12 5v14M5 12l7-7 7 7" />
                          </svg>
                        )}
                      </button>
                    ))}
                  </div>
                </div>

                {hasActiveFilters && (
                  <button className="sessions-filter-clear" onClick={clearFilters}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                      <line x1="18" y1="6" x2="6" y2="18" />
                      <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                    Clear Filters
                  </button>
                )}
              </div>
            )}

            {loading && (
              <div className="sessions-loading">
                <div className="spinner-ring" />
                <p className="spinner-label">Unrolling the scrolls…</p>
              </div>
            )}

            {!loading && error && (
              <div className="sessions-error">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                  <line x1="12" y1="9" x2="12" y2="13" />
                  <line x1="12" y1="17" x2="12.01" y2="17" />
                </svg>
                <p>{error}</p>
              </div>
            )}

            {!loading && !error && sessions.length === 0 && (
              <div className="sessions-empty">
                <div className="sessions-empty-sigil" aria-hidden="true">✦</div>
                <p className="sessions-empty-message">No sessions yet.</p>
                <p className="sessions-empty-sub">Begin your chronicle — create the first session.</p>
              </div>
            )}

            {!loading && !error && sessions.length > 0 && filteredSessions.length === 0 && (
              <div className="sessions-empty">
                <div className="sessions-empty-sigil" aria-hidden="true">✦</div>
                <p className="sessions-empty-message">No matching sessions.</p>
                <p className="sessions-empty-sub">Try adjusting your filters.</p>
              </div>
            )}

            {!loading && !error && filteredSessions.length > 0 && (
              <div className="editor-folder-groups">
                {unfiledSessions.length > 0 && (
                  <div className="editor-folder-group">
                    <div className={viewMode === 'grid' ? 'sessions-grid' : 'sessions-list'}>
                      {unfiledSessions.map(session => (
                        <SessionCard key={session.id} session={session} isGM={isGM} mode={viewMode}
                          onEdit={openEdit} onDelete={openDelete} onOpen={openNotes} />
                      ))}
                    </div>
                  </div>
                )}
                {sessionsByFolder.map(({ folder, items }) => (
                  items.length > 0 && (
                    <div key={folder.id} className="editor-folder-group">
                      <button className="editor-folder-header" onClick={() => toggleFolder(folder.id)}>
                        <svg className={`editor-folder-chevron${collapsedFolders[folder.id] ? '' : ' editor-folder-chevron--open'}`}
                          viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                          <path d="M9 18l6-6-6-6" />
                        </svg>
                        <span className="editor-folder-icon" aria-hidden>📁</span>
                        <span className="editor-folder-name">{folder.name}</span>
                        <span className="editor-folder-count">{items.length}</span>
                      </button>
                      {!collapsedFolders[folder.id] && (
                        <div className={viewMode === 'grid' ? 'sessions-grid' : 'sessions-list'}>
                          {items.map(session => (
                            <SessionCard key={session.id} session={session} isGM={isGM} mode={viewMode}
                              onEdit={openEdit} onDelete={openDelete} onOpen={openNotes} />
                          ))}
                        </div>
                      )}
                    </div>
                  )
                ))}
              </div>
            )}

            {!loading && !error && sessionTotal > 0 && (
              <Pagination
                page={sessionPage}
                limit={sessionPageSize}
                total={sessionTotal}
                onPageChange={setSessionPage}
              />
            )}
          </section>
        )}

        {/* ── Notes tab ── */}
        {activeTab === 'notes' && (
          <section className="sessions-section">

            {/* Notes filter bar */}
            {filtersOpen && (
            <div className="notes-filter-bar">
              <div className="notes-filter-group">
                <span className="sessions-filter-group-label">Sort</span>
                <div className="sessions-sort-options">
                  <button
                    className={`sessions-sort-btn${noteSort === 'title' ? ' sessions-sort-btn--active' : ''}`}
                    onClick={() => { setNoteSort('title'); setNotePage(1) }}
                  >
                    Title A–Z
                  </button>
                  <button
                    className={`sessions-sort-btn${noteSort === 'created_at' ? ' sessions-sort-btn--active' : ''}`}
                    onClick={() => { setNoteSort('created_at'); setNotePage(1) }}
                  >
                    Newest
                  </button>
                </div>
              </div>
              <div className="notes-filter-group">
                <span className="sessions-filter-group-label">Session</span>
                <select
                  className="notes-filter-select"
                  value={noteSessionFilter}
                  onChange={e => { setNoteSessionFilter(e.target.value); setNotePage(1) }}
                >
                  <option value="">All Notes</option>
                  <option value="unlinked">Unlinked</option>
                  {sessions.map(s => (
                    <option key={s.id} value={s.id}>
                      {s.session_number != null ? `#${s.session_number} — ` : ''}{s.title}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            )}

            {loading && (
              <div className="sessions-loading">
                <div className="spinner-ring" />
                <p className="spinner-label">Consulting the tomes…</p>
              </div>
            )}

            {!loading && error && (
              <div className="sessions-error">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                  <line x1="12" y1="9" x2="12" y2="13" />
                  <line x1="12" y1="17" x2="12.01" y2="17" />
                </svg>
                <p>{error}</p>
              </div>
            )}

            {!loading && !error && notes.length === 0 && (
              <div className="sessions-empty">
                <div className="sessions-empty-sigil" aria-hidden="true">✦</div>
                <p className="sessions-empty-message">No notes yet.</p>
                <p className="sessions-empty-sub">Inscribe your lore — create the first note.</p>
              </div>
            )}

            {!loading && !error && notes.length > 0 && filteredNotes.length === 0 && (
              <div className="sessions-empty">
                <div className="sessions-empty-sigil" aria-hidden="true">✦</div>
                <p className="sessions-empty-message">No matching notes.</p>
                <p className="sessions-empty-sub">Try a different filter.</p>
              </div>
            )}

            {!loading && !error && filteredNotes.length > 0 && (
              <div className="editor-folder-groups">
                {unfiledNotes.length > 0 && (
                  <div className="editor-folder-group">
                    <div className={viewMode === 'grid' ? 'sessions-grid' : 'sessions-list'}>
                      {unfiledNotes.map(note => (
                        <NoteCard key={note.id} note={note}
                          sessionTitle={note.session_id ? sessionMap[note.session_id]?.title : undefined}
                          isGM={isGM} isAuthor={note.user_id === user?.id} mode={viewMode}
                          onEdit={openNoteEdit} onDelete={openNoteDelete} onOpen={openNoteEditor} />
                      ))}
                    </div>
                  </div>
                )}
                {notesByFolder.map(({ folder, items }) => (
                  items.length > 0 && (
                    <div key={folder.id} className="editor-folder-group">
                      <button className="editor-folder-header" onClick={() => toggleFolder(folder.id)}>
                        <svg className={`editor-folder-chevron${collapsedFolders[folder.id] ? '' : ' editor-folder-chevron--open'}`}
                          viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                          <path d="M9 18l6-6-6-6" />
                        </svg>
                        <span className="editor-folder-icon" aria-hidden>📁</span>
                        <span className="editor-folder-name">{folder.name}</span>
                        <span className="editor-folder-count">{items.length}</span>
                      </button>
                      {!collapsedFolders[folder.id] && (
                        <div className={viewMode === 'grid' ? 'sessions-grid' : 'sessions-list'}>
                          {items.map(note => (
                            <NoteCard key={note.id} note={note}
                              sessionTitle={note.session_id ? sessionMap[note.session_id]?.title : undefined}
                              isGM={isGM} isAuthor={note.user_id === user?.id} mode={viewMode}
                              onEdit={openNoteEdit} onDelete={openNoteDelete} onOpen={openNoteEditor} />
                          ))}
                        </div>
                      )}
                    </div>
                  )
                ))}
              </div>
            )}

            {!loading && !error && noteTotal > 0 && (
              <Pagination
                page={notePage}
                limit={notePageSize}
                total={noteTotal}
                onPageChange={setNotePage}
              />
            )}
          </section>
        )}

        {/* ── Session modals ── */}
        {formOpen && (
          <SessionFormModal
            mode="create"
            initial={{
              title: '',
              session_number: sessions.length > 0
                ? Math.max(...sessions.map(s => s.session_number ?? 0)) + 1
                : 1,
              scheduled_at: null,
              runtime_start: null,
              runtime_end: null,
            }}
            error={mutationError}
            saving={saving}
            onSave={handleCreate}
            onClose={() => { setFormOpen(false); setMutationError(null) }}
          />
        )}

        {editingSession && (
          <SessionFormModal
            mode="edit"
            initial={{
              title: editingSession.title,
              session_number: editingSession.session_number,
              scheduled_at: editingSession.scheduled_at,
              runtime_start: editingSession.runtime_start,
              runtime_end: editingSession.runtime_end,
            }}
            error={mutationError}
            saving={saving}
            onSave={handleEdit}
            onClose={() => { setEditingSession(null); setMutationError(null) }}
          />
        )}

        {deletingSession && (
          <ConfirmModal
            title="Delete Session"
            message={`Are you sure you want to delete "${deletingSession.title}"? This action cannot be undone.`}
            confirmLabel="Delete"
            error={mutationError}
            loading={saving}
            onConfirm={handleDelete}
            onCancel={() => { setDeletingSession(null); setMutationError(null) }}
          />
        )}

        {/* ── Note modals ── */}
        {noteFormOpen && (
          <NoteFormModal
            mode="create"
            sessions={sessions}
            error={noteMutationError}
            saving={noteSaving}
            onSave={handleCreateNote}
            onClose={() => { setNoteFormOpen(false); setNoteMutationError(null) }}
          />
        )}

        {editingNote && (
          <NoteFormModal
            mode="edit"
            initial={{
              title: editingNote.title,
              session_id: editingNote.session_id,
              visibility: editingNote.visibility,
            }}
            sessions={sessions}
            error={noteMutationError}
            saving={noteSaving}
            isAuthor={editingNote.user_id === user?.id}
            onSave={handleEditNote}
            onClose={() => { setEditingNote(null); setNoteMutationError(null) }}
          />
        )}

        {deletingNote && (
          <ConfirmModal
            title="Delete Note"
            message={`Are you sure you want to delete "${deletingNote.title}"? This action cannot be undone.`}
            confirmLabel="Delete"
            error={noteMutationError}
            loading={noteSaving}
            onConfirm={handleDeleteNote}
            onCancel={() => { setDeletingNote(null); setNoteMutationError(null) }}
          />
        )}
      </div>

      {editOpen && (
        <Modal title="Edit Campaign" onClose={handleClose}>
          <EditCampaignForm
            gameId={gameId!}
            onSuccess={handleSuccess}
            onDirtyChange={setIsDirty}
          />
        </Modal>
      )}

      {importOpen && (
        <Modal title="Import Backup" onClose={handleImportClose}>
          <div className="editor-import-modal">
            <div className="editor-import-field">
              <span className="editor-import-label">Backup File</span>
              <span className="editor-import-hint">Select a JSON backup file (max 10 MB)</span>
              <label className="editor-import-file-label">
                <input
                  ref={importFileRef}
                  type="file"
                  accept=".json,application/json"
                  className="editor-import-file-input"
                  onChange={handleImportFileSelect}
                  disabled={importBusy}
                />
                <span className="editor-import-file-btn">
                  <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M13 10v3a1 1 0 01-1 1H4a1 1 0 01-1-1v-3M8 2v8M5 5l3-3 3 3"/>
                  </svg>
                  {importFile ? importFile.name : 'Choose File'}
                </span>
              </label>
            </div>

            <div className="editor-import-field">
              <span className="editor-import-label">Conflict Resolution</span>
              <span className="editor-import-hint">How to handle records that already exist</span>
              <div className="editor-import-mode-toggle">
                <button
                  className={`editor-import-mode-btn${importMode === 'merge' ? ' active' : ''}`}
                  onClick={() => setImportMode('merge')}
                  disabled={importBusy}
                >
                  Merge (skip existing)
                </button>
                <button
                  className={`editor-import-mode-btn${importMode === 'overwrite' ? ' active' : ''}`}
                  onClick={() => setImportMode('overwrite')}
                  disabled={importBusy}
                >
                  Overwrite (replace)
                </button>
              </div>
            </div>

            {importError && (
              <div className="editor-import-error" role="alert">{importError}</div>
            )}

            {importResult && (
              <div className="editor-import-success" role="status">
                <span className="editor-import-success-title">Import Complete</span>
                <div className="editor-import-summary">
                  <span><strong>{importResult.sessions_created}</strong> sessions created, <strong>{importResult.sessions_skipped}</strong> skipped, <strong>{importResult.sessions_overwritten}</strong> overwritten</span>
                  <span><strong>{importResult.notes_created}</strong> notes created, <strong>{importResult.notes_skipped}</strong> skipped, <strong>{importResult.notes_overwritten}</strong> overwritten</span>
                </div>
              </div>
            )}

            <div className="editor-import-actions">
              <button
                className="editor-import-submit"
                onClick={handleImportSubmit}
                disabled={!importFile || !importMode || importBusy}
              >
                {importBusy ? 'Importing…' : 'Import'}
              </button>
              <button className="editor-import-cancel" onClick={handleImportClose}>
                {importResult ? 'Close' : 'Cancel'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}

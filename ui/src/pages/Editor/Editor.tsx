import { useState, useEffect, useCallback } from 'react'
import { useParams, useLocation, useNavigate } from 'react-router-dom'
import type { Session, SessionFormData } from '../../types/session'
import type { GameMembership } from '../../types/membership'
import { listGameSessions, createSession, updateSession, deleteSession } from '../../api/sessions'
import { listMemberships } from '../../api/memberships'
import { DEV_USER_ID } from '../../constants/dev'
import SessionCard from '../../components/SessionCard/SessionCard'
import SessionFormModal from '../../components/SessionFormModal/SessionFormModal'
import ConfirmModal from '../../components/ConfirmModal/ConfirmModal'
import Modal from '../../components/Modal/Modal'
import EditCampaignForm from '../../components/EditCampaignForm/EditCampaignForm'
import './Editor.css'

interface LocationState {
  title?: string
}

export default function Editor() {
  const { gameId } = useParams<{ gameId: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const state = location.state as LocationState | null

  const [title, setTitle] = useState(state?.title ?? `Game #${gameId}`)
  const [editOpen, setEditOpen] = useState(false)
  const [isDirty, setIsDirty] = useState(false)

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

  const [sessions, setSessions] = useState<Session[]>([])
  const [memberships, setMemberships] = useState<GameMembership[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Filter state
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [filterTitle, setFilterTitle] = useState('')
  const [filterNumMin, setFilterNumMin] = useState('')
  const [filterNumMax, setFilterNumMax] = useState('')
  const [filterDateFrom, setFilterDateFrom] = useState('')
  const [filterDateTo, setFilterDateTo] = useState('')

  // Sort state
  type SortField = 'session_number' | 'title' | 'updated_at'
  type SortDir = 'asc' | 'desc'
  const [sortField, setSortField] = useState<SortField>('session_number')
  const [sortDir, setSortDir] = useState<SortDir>('asc')

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

  // Modal state
  const [formOpen, setFormOpen] = useState(false)
  const [editingSession, setEditingSession] = useState<Session | null>(null)
  const [deletingSession, setDeletingSession] = useState<Session | null>(null)
  const [mutationError, setMutationError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!gameId) return
    let cancelled = false
    setLoading(true)
    setError(null)

    Promise.all([listGameSessions(gameId), listMemberships(gameId)])
      .then(([sessionsData, membershipsData]) => {
        if (!cancelled) {
          setSessions(sessionsData)
          setMemberships(membershipsData)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load sessions.')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [gameId])

  const isGM = memberships.some(m => m.user_id === DEV_USER_ID && m.is_gm)

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
    // Title filter
    if (filterTitle !== '') {
      if (!s.title.toLowerCase().includes(filterTitle.toLowerCase())) return false
    }
    // Session number range filter
    if (filterNumMin !== '') {
      const min = Number(filterNumMin)
      if (s.session_number == null || s.session_number < min) return false
    }
    if (filterNumMax !== '') {
      const max = Number(filterNumMax)
      if (s.session_number == null || s.session_number > max) return false
    }
    // Edit date range filter (updated_at)
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
      const updated = await updateSession(editingSession.id, data)
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
      setSessions(prev => prev.filter(s => s.id !== deletingSession.id))
      setDeletingSession(null)
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

        <section className="sessions-section">
          <div className="sessions-header">
            <h2 className="sessions-heading">Sessions</h2>
            <div className="sessions-header-actions">
              <button
                className={`sessions-filter-toggle${filtersOpen ? ' sessions-filter-toggle--active' : ''}${hasActiveFilters ? ' sessions-filter-toggle--has-filters' : ''}`}
                onClick={() => setFiltersOpen(prev => !prev)}
                aria-label="Toggle filters"
                title="Filter sessions"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                  <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
                </svg>
                {hasActiveFilters && <span className="sessions-filter-badge" />}
              </button>
              <button className="sessions-new-btn" onClick={openCreate}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <line x1="12" y1="5" x2="12" y2="19" />
                  <line x1="5" y1="12" x2="19" y2="12" />
                </svg>
                New Session
              </button>
            </div>
          </div>

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
            <div className="sessions-list">
              {filteredSessions.map(session => (
                <SessionCard
                  key={session.id}
                  session={session}
                  isGM={isGM}
                  onEdit={openEdit}
                  onDelete={openDelete}
                  onOpen={openNotes}
                />
              ))}
            </div>
          )}
        </section>

        {formOpen && (
          <SessionFormModal
            mode="create"
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
    </div>
  )
}

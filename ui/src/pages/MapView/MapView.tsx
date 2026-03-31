import { useState, useEffect, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { useParams, useNavigate } from 'react-router-dom'
import { TransformWrapper, TransformComponent } from 'react-zoom-pan-pinch'
import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import { apiFetch, BASE_URL } from '../../api/client'
import { listGameSessions, updateSession, createSession } from '../../api/sessions'
import { listGamePins, createPin, updatePin, deletePin, createGamePin } from '../../api/pins'
import { listGameNotes, updateNote, createNote } from '../../api/notes'
import { uploadMapImage, deleteMapImage } from '../../api/mapImage'
import { listMemberships } from '../../api/memberships'
import { useAuth } from '../../context/AuthContext'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import type { Game } from '../../types/game'
import type { Session } from '../../types/session'
import type { SessionPin } from '../../types/pin'
import type { GameMembership } from '../../types/membership'
import type { Note } from '../../types/note'
import type { PinGroup } from '../../types/pin'
import { listGamePinGroups, createPinGroup, addPinToGroup, removePinFromGroup, disbandPinGroup, updatePinGroup } from '../../api/pinGroups'
import { PIN_COLOURS, PIN_ICONS, COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS } from '../../constants/pins'
import type { PinColour, PinIcon } from '../../constants/pins'
import { getPreferences, updatePreferences } from '../../api/preferences'
import type { GameSidebarState } from '../../api/preferences'
import FolderSidebar from '../../components/FolderSidebar/FolderSidebar'
import './MapView.css'

/** Proximity threshold in map-percentage units (0–100). ~16px on a 1000px map. */
const GROUP_PROXIMITY_PCT = 1.5

interface MapViewState {
  scale: number
  positionX: number
  positionY: number
}

const DEFAULT_VIEW_STATE: MapViewState = {
  scale: 1,
  positionX: 0,
  positionY: 0,
}

export default function MapView() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()

  const [game, setGame] = useState<Game | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [pins, setPins] = useState<SessionPin[]>([])
  const [notes, setNotes] = useState<Note[]>([])
  const [memberships, setMemberships] = useState<GameMembership[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [uploading, setUploading] = useState(false)
  const [pendingCoords, setPendingCoords] = useState<{ x: number; y: number } | null>(null)
  const [hoveredPinId, setHoveredPinId] = useState<string | null>(null)
  const [dragging, setDragging] = useState<{ pinId: string; startX: number; startY: number } | null>(null)
  const [viewState, setViewState] = useLocalStorage<MapViewState>(
    `pf2e-map-view-${gameId}`,
    DEFAULT_VIEW_STATE,
  )
  const [panelOpen, setPanelOpen] = useState(true)
  const [displayScale, setDisplayScale] = useState(viewState.scale)

  // Colour/icon picker state
  const [pendingColour, setPendingColour] = useState<PinColour>('grey')
  const [pendingIcon, setPendingIcon] = useState<PinIcon>('position-marker')
  const [defaultPinColour, setDefaultPinColour] = useState<PinColour>('grey')
  const [defaultPinIcon, setDefaultPinIcon] = useState<PinIcon>('position-marker')

  const [pendingLabel, setPendingLabel] = useState('')
  const [pendingDescription, setPendingDescription] = useState('')
  const [pinError, setPinError] = useState<string | null>(null)

  // Edit pin popover (open/close only — changes save immediately)
  const [editingPinId, setEditingPinId] = useState<string | null>(null)

  // Search filter for session/note pickers
  const [pickerSearch, setPickerSearch] = useState('')
  const [editLinkSearch, setEditLinkSearch] = useState('')

  // Pin group state
  const [pinGroups, setPinGroups] = useState<PinGroup[]>([])
  const [activeGroupId, setActiveGroupId] = useState<string | null>(null)
  const [managingGroupId, setManagingGroupId] = useState<string | null>(null)
  const [groupingPrompt, setGroupingPrompt] = useState<{ coords: { x: number; y: number }; nearbyPins: SessionPin[]; nearbyGroups: PinGroup[] } | null>(null)
  const [pendingGroupPinIds, setPendingGroupPinIds] = useState<string[] | null>(null)
  const [pendingAddToGroupId, setPendingAddToGroupId] = useState<string | null>(null)
  const [dragGroupPrompt, setDragGroupPrompt] = useState<{ draggedPinId: string; nearbyPins: SessionPin[]; nearbyGroups: PinGroup[]; originalCoords: { x: number; y: number } } | null>(null)
  const [dropTargetIds, setDropTargetIds] = useState<Set<string>>(new Set())
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const viewportContainerRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const transformRef = useRef<ReactZoomPanPinchRef>(null)
  const wasDragRef = useRef(false)
  const sidebarStateRef = useRef<Record<string, GameSidebarState>>({})

  /** Persist transform to localStorage — called only when an interaction ends. */
  const handleTransformEnd = useCallback((ref: ReactZoomPanPinchRef) => {
    const { scale, positionX, positionY } = ref.state
    setDisplayScale(scale)
    setViewState(prev => ({ ...prev, scale, positionX, positionY }))
  }, [setViewState])

  const isGM = memberships.some(m => m.user_id === user?.id && m.is_gm)
  const pinnedSessionIds = new Set(pins.map(p => p.session_id))
  const unpinnedSessions = sessions.filter(s => !pinnedSessionIds.has(s.id))

  useEffect(() => {
    if (!gameId) return
    let cancelled = false
    setLoading(true)
    setError(null)

    Promise.all([
      apiFetch<Game>(`/games/${gameId}`),
      listGameSessions(gameId),
      listGamePins(gameId),
      listMemberships(gameId),
      listGameNotes(gameId),
      listGamePinGroups(gameId),
    ])
      .then(([gameData, sessionsData, pinsData, membershipsData, notesData, pinGroupsData]) => {
        if (!cancelled) {
          setGame(gameData)
          setSessions(sessionsData)
          setPins(pinsData)
          setMemberships(membershipsData)
          setNotes(notesData)
          setPinGroups(pinGroupsData)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load map.')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [gameId])

  useEffect(() => {
    if (!gameId) return
    getPreferences().then(prefs => {
      sidebarStateRef.current = prefs.sidebar_state ?? {}
      const gameState = sidebarStateRef.current[gameId]
      if (gameState?.panelOpen) setSidebarOpen(true)
    }).catch(() => {})
  }, [gameId])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', ' '].includes(e.key)) {
        const vp = document.querySelector('.map-viewport')
        if (vp?.contains(document.activeElement) || document.activeElement === vp) {
          e.preventDefault()
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  // Suppress the browser's default middle-click auto-scroll icon
  useEffect(() => {
    const vp = document.querySelector('.map-viewport')
    if (!vp) return
    const onMouseDown = (e: Event) => {
      if ((e as MouseEvent).button === 1) e.preventDefault()
    }
    vp.addEventListener('mousedown', onMouseDown)
    return () => vp.removeEventListener('mousedown', onMouseDown)
  }, [game?.map_image_url, loading])

  // Pre-fill pin colour/icon from account preferences
  useEffect(() => {
    getPreferences()
      .then(prefs => {
        if (prefs.default_pin_colour && PIN_COLOURS.includes(prefs.default_pin_colour as PinColour)) {
          const colour = prefs.default_pin_colour as PinColour
          setDefaultPinColour(colour)
          setPendingColour(colour)
        }
        if (prefs.default_pin_icon && PIN_ICONS.includes(prefs.default_pin_icon as PinIcon)) {
          const icon = prefs.default_pin_icon as PinIcon
          setDefaultPinIcon(icon)
          setPendingIcon(icon)
        }
      })
      .catch(() => { /* silently ignore */ })
  }, [])

  /** Convert a client-space point to percentage coords on the map. */
  const clientToMapPct = useCallback((clientX: number, clientY: number): { x: number; y: number } => {
    if (!mapContainerRef.current) return { x: 0, y: 0 }
    const rect = mapContainerRef.current.getBoundingClientRect()
    const x = ((clientX - rect.left) / rect.width) * 100
    const y = ((clientY - rect.top) / rect.height) * 100
    return {
      x: Math.min(100, Math.max(0, x)),
      y: Math.min(100, Math.max(0, y)),
    }
  }, [])

  const handleMapClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!mapContainerRef.current || dragging) return
    if ((e.target as HTMLElement).closest('.map-pin-wrapper')) return

    // If an edit popover is open, close it without placing a new pin
    if (editingPinId) {
      setEditingPinId(null)
      setEditLinkSearch('')
      return
    }

    const coords = clientToMapPct(e.clientX, e.clientY)
    setActiveGroupId(null)

    const nearbyPins: SessionPin[] = []
    const nearbyGroups: PinGroup[] = []

    pins.filter(p => p.group_id === null).forEach(p => {
      const dx = p.x - coords.x
      const dy = p.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) {
        nearbyPins.push(p)
      }
    })

    pinGroups.forEach(g => {
      const dx = g.x - coords.x
      const dy = g.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) {
        nearbyGroups.push(g)
      }
    })

    if (nearbyPins.length > 0 || nearbyGroups.length > 0) {
      setGroupingPrompt({ coords, nearbyPins, nearbyGroups })
      return
    }

    setPendingCoords(coords)

  }, [dragging, clientToMapPct, pins, pinGroups, editingPinId])

  const reloadPinGroups = useCallback(async () => {
    if (!gameId) return
    try {
      const groups = await listGamePinGroups(gameId)
      setPinGroups(groups)
    } catch (err: unknown) {
      console.error('Failed to reload pin groups', err)
    }
  }, [gameId])

  const handleSelectSession = useCallback(async (session: Session) => {
    if (!gameId || !pendingCoords) return
    const label = pendingLabel.trim() || session.title
    const desc = pendingDescription.trim() || undefined
    try {
      const pin = await createPin({
        session_id: session.id,
        x: pendingCoords.x,
        y: pendingCoords.y,
        label,
        colour: pendingColour,
        icon: pendingIcon,
        description: desc,
      })
      setPins(prev => [...prev, pin])
      setPendingCoords(null)
      setPendingLabel('')
      setPendingDescription('')
      setPendingColour(defaultPinColour)
      setPendingIcon(defaultPinIcon)

      if (pendingGroupPinIds) {
        await createPinGroup(gameId, [...pendingGroupPinIds, pin.id])
        await reloadPinGroups()
        setPendingGroupPinIds(null)
      } else if (pendingAddToGroupId) {
        await addPinToGroup(pendingAddToGroupId, pin.id)
        await reloadPinGroups()
        setPendingAddToGroupId(null)
      }
    } catch (err: unknown) {
      console.error('Failed to create pin', err)
      setPinError(err instanceof Error ? err.message : 'Failed to create pin. Please try again.')
      setPendingCoords(null)
    }
  }, [gameId, pendingCoords, pendingColour, pendingIcon, pendingLabel, pendingDescription, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

  const handleSelectNote = useCallback(async (note: Note) => {
    if (!gameId || !pendingCoords) return
    const label = pendingLabel.trim() || note.title
    const desc = pendingDescription.trim() || undefined
    try {
      const pin = await createGamePin(gameId, {
        x: pendingCoords.x,
        y: pendingCoords.y,
        label,
        colour: pendingColour,
        icon: pendingIcon,
        note_id: note.id,
        description: desc,
      })
      setPins(prev => [...prev, pin])
      setPendingCoords(null)
      setPendingLabel('')
      setPendingDescription('')
      setPendingColour(defaultPinColour)
      setPendingIcon(defaultPinIcon)

      if (pendingGroupPinIds) {
        await createPinGroup(gameId, [...pendingGroupPinIds, pin.id])
        await reloadPinGroups()
        setPendingGroupPinIds(null)
      } else if (pendingAddToGroupId) {
        await addPinToGroup(pendingAddToGroupId, pin.id)
        await reloadPinGroups()
        setPendingAddToGroupId(null)
      }
    } catch (err: unknown) {
      console.error('Failed to create pin', err)
      setPinError(err instanceof Error ? err.message : 'Failed to create pin. Please try again.')
      setPendingCoords(null)
    }
  }, [gameId, pendingCoords, pendingColour, pendingIcon, pendingLabel, pendingDescription, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

  const handleCreateMarker = useCallback(async (label: string, description: string) => {
    if (!gameId || !pendingCoords) return
    const trimmed = label.trim()
    try {
      const pin = await createGamePin(gameId, {
        x: pendingCoords.x,
        y: pendingCoords.y,
        label: trimmed,
        colour: pendingColour,
        icon: pendingIcon,
        description: description || undefined,
      })
      setPins(prev => [...prev, pin])
      setPendingCoords(null)
      setPendingLabel('')
      setPendingDescription('')
      setPendingColour(defaultPinColour)
      setPendingIcon(defaultPinIcon)

      if (pendingGroupPinIds) {
        await createPinGroup(gameId, [...pendingGroupPinIds, pin.id])
        await reloadPinGroups()
        setPendingGroupPinIds(null)
      } else if (pendingAddToGroupId) {
        await addPinToGroup(pendingAddToGroupId, pin.id)
        await reloadPinGroups()
        setPendingAddToGroupId(null)
      }
    } catch (err: unknown) {
      console.error('Failed to create marker', err)
      setPinError(err instanceof Error ? err.message : 'Failed to place marker')
    }
  }, [gameId, pendingCoords, pendingColour, pendingIcon, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

  const handlePinPointerDown = useCallback((e: React.PointerEvent, pin: SessionPin) => {
    if (pin.group_id !== null) return
    e.preventDefault()
    e.stopPropagation()
    ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
    wasDragRef.current = false
    setDragging({ pinId: pin.id, startX: e.clientX, startY: e.clientY })
  }, [])

  const handlePointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragging || !mapContainerRef.current) return

    // Check if we've exceeded the 5px drag threshold
    if (!wasDragRef.current) {
      const dx = e.clientX - dragging.startX
      const dy = e.clientY - dragging.startY
      if (Math.sqrt(dx * dx + dy * dy) <= 5) return
      wasDragRef.current = true
    }

    const coords = clientToMapPct(e.clientX, e.clientY)
    setPins(prev => prev.map(p => p.id === dragging.pinId ? { ...p, x: coords.x, y: coords.y } : p))

    // Compute which pins/groups are within drop range
    const targets = new Set<string>()
    pins.forEach(p => {
      if (p.group_id !== null || p.id === dragging.pinId) return
      const dx = p.x - coords.x
      const dy = p.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) targets.add(p.id)
    })
    pinGroups.forEach(g => {
      const dx = g.x - coords.x
      const dy = g.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) targets.add(g.id)
    })
    setDropTargetIds(targets)
  }, [dragging, clientToMapPct, pins, pinGroups])

  const handlePointerUp = useCallback(async (e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragging || !mapContainerRef.current) return
    const pinId = dragging.pinId
    const draggedPin = pins.find(p => p.id === pinId)
    setDragging(null)
    setDropTargetIds(new Set())

    // Only persist position if the pointer actually dragged
    if (!wasDragRef.current) return

    const coords = clientToMapPct(e.clientX, e.clientY)

    // Check if dropped near other standalone pins or groups
    const nearbyPins: SessionPin[] = []
    const nearbyGroups: PinGroup[] = []

    pins.filter(p => p.group_id === null && p.id !== pinId).forEach(p => {
      const dx = p.x - coords.x
      const dy = p.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) {
        nearbyPins.push(p)
      }
    })

    pinGroups.forEach(g => {
      const dx = g.x - coords.x
      const dy = g.y - coords.y
      if (Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT) {
        nearbyGroups.push(g)
      }
    })

    if (nearbyPins.length > 0 || nearbyGroups.length > 0) {
      // Revert pin to original position visually while the user decides
      if (draggedPin) {
        setPins(prev => prev.map(p => p.id === pinId ? { ...p, x: draggedPin.x, y: draggedPin.y } : p))
      }
      setDragGroupPrompt({
        draggedPinId: pinId,
        nearbyPins,
        nearbyGroups,
        originalCoords: draggedPin ? { x: draggedPin.x, y: draggedPin.y } : coords,
      })
      return
    }

    try {
      await updatePin(pinId, { x: coords.x, y: coords.y })
    } catch (err: unknown) {
      console.error('Failed to update pin', err)
    }
  }, [dragging, clientToMapPct, pins, pinGroups])

  const handleDeletePin = useCallback(async (pinId: string) => {
    try {
      await deletePin(pinId)
      setPins(prev => prev.filter(p => p.id !== pinId))
    } catch (err: unknown) {
      console.error('Failed to delete pin', err)
    }
  }, [])

  const handleEditPinField = useCallback(async (pinId: string, field: { colour?: string; icon?: string; label?: string; description?: string | null; session_id?: string | null; note_id?: string | null }) => {
    // Optimistically update local state, then persist
    setPins(prev => prev.map(p => p.id === pinId ? { ...p, ...field } : p))
    try {
      await updatePin(pinId, field)
    } catch (err: unknown) {
      console.error('Failed to update pin', err)
      if (gameId) {
        try {
          const freshPins = await listGamePins(gameId)
          setPins(freshPins)
        } catch { /* ignore */ }
      }
    }
  }, [gameId])

  const handleUploadClick = () => fileInputRef.current?.click()

  const handleFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !gameId) return
    setUploading(true)
    setUploadError(null)
    try {
      const updatedGame = await uploadMapImage(gameId, file)
      setGame(updatedGame)
    } catch (err: unknown) {
      setUploadError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }, [gameId])

  const handleDeleteMap = useCallback(async () => {
    if (!gameId || !confirm('Remove the map image? All pins will remain.')) return
    try {
      await deleteMapImage(gameId)
      setGame(prev => prev ? { ...prev, map_image_url: null } : prev)
    } catch (err: unknown) {
      console.error('Failed to delete map', err)
    }
  }, [gameId])

  const handleSidebarToggle = useCallback(() => {
    const newOpen = !sidebarOpen
    setSidebarOpen(newOpen)
    const current = sidebarStateRef.current
    const gameState: GameSidebarState = { ...(current[gameId!] ?? { panelOpen: false }), panelOpen: newOpen }
    const merged = { ...current, [gameId!]: gameState }
    sidebarStateRef.current = merged
    updatePreferences({ sidebar_state: merged }).catch(() => {})
  }, [gameId, sidebarOpen])

  const handleSessionUpdate = useCallback(async (sessionId: string, data: Record<string, unknown>) => {
    const updated = await updateSession(sessionId, data)
    setSessions(prev => prev.map(s => s.id === sessionId ? updated : s))
  }, [])

  const handleNoteUpdate = useCallback(async (noteId: string, data: Record<string, unknown>) => {
    const updated = await updateNote(noteId, data)
    setNotes(prev => prev.map(n => n.id === noteId ? updated : n))
  }, [])

  const handleCreateSession = useCallback(async (folderId: string | null) => {
    if (!gameId) return
    try {
      const title = `Session ${sessions.length + 1}`
      const created = await createSession(gameId, {
        title,
        session_number: sessions.length + 1,
        scheduled_at: null,
        runtime_start: null,
        runtime_end: null,
      })
      // Assign to folder if specified
      if (folderId) {
        const updated = await updateSession(created.id, { folder_id: folderId })
        setSessions(prev => [...prev, updated])
      } else {
        setSessions(prev => [...prev, created])
      }
    } catch (err) {
      console.error('Failed to create session from sidebar', err)
    }
  }, [gameId, sessions.length])

  const handleCreateNote = useCallback(async (folderId: string | null) => {
    if (!gameId) return
    try {
      const created = await createNote(gameId, { title: 'Untitled Note' })
      if (folderId) {
        const updated = await updateNote(created.id, { folder_id: folderId })
        setNotes(prev => [...prev, updated])
      } else {
        setNotes(prev => [...prev, created])
      }
    } catch (err) {
      console.error('Failed to create note from sidebar', err)
    }
  }, [gameId])

  const sessionForPin = (pin: SessionPin) => sessions.find(s => s.id === pin.session_id)
  const noteForPin = (pin: SessionPin) => notes.find(n => n.id === pin.note_id)

  return (
    <div className="map-view-page">
      <div className="map-view-inner">
        {(!game?.map_image_url || loading || !!error) && (
          <>
            <button className="map-back-btn" onClick={() => navigate(`/games/${gameId}`)}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
              Back to Sessions
            </button>

            <header className="map-header">
              <div className="map-title-rule" aria-hidden="true">
                <span /><span className="map-title-ornament">✦</span><span />
              </div>
              <h1 className="map-title">{game?.title ?? 'Campaign Map'}</h1>
              <div className="map-title-rule" aria-hidden="true">
                <span /><span className="map-title-ornament">✦</span><span />
              </div>
            </header>
          </>
        )}

        {loading && (
          <div className="map-spinner">
            <div className="spinner-ring" />
            <p className="spinner-label">Unrolling the map…</p>
          </div>
        )}

        {!loading && error && (
          <div className="map-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <p>{error}</p>
          </div>
        )}

        {!loading && !error && !game?.map_image_url && !isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">The Map Awaits</p>
            <p className="map-empty-sub">The Game Master has not yet unveiled the realm.</p>
          </div>
        )}

        {!loading && !error && !game?.map_image_url && isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">No Map Uploaded</p>
            <p className="map-empty-sub">Upload a map image to begin placing session markers.</p>
            {uploadError && <p className="map-upload-error">{uploadError}</p>}
            <button className="map-upload-btn" onClick={handleUploadClick} disabled={uploading}>
              {uploading ? 'Uploading…' : '+ Upload Map Image'}
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              className="map-file-input"
              onChange={handleFileChange}
            />
          </div>
        )}

        {!loading && !error && game?.map_image_url && (
          <div className="map-viewport-container" ref={viewportContainerRef}>
            {pinError && (
              <div className="map-pin-error-banner" onClick={() => setPinError(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" width="14" height="14">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                {pinError}
              </div>
            )}
            <TransformWrapper
              ref={transformRef}
              initialScale={viewState.scale}
              initialPositionX={viewState.positionX}
              initialPositionY={viewState.positionY}
              minScale={1}
              maxScale={5}
              limitToBounds={true}
              centerOnInit={false}
              panning={{
                allowLeftClickPan: false,
                allowMiddleClickPan: true,
                allowRightClickPan: false,
              }}
              wheel={{
                activationKeys: ['Control', 'Meta'],
                step: 0.25,
              }}
              doubleClick={{ disabled: true }}
              onPanningStop={handleTransformEnd}
              onZoomStop={handleTransformEnd}
              onZoom={(ref) => setDisplayScale(ref.state.scale)}
            >
              <TransformComponent
                wrapperClass="map-viewport"
                contentClass="map-container map-container--interactive"
              >
                <div
                  ref={mapContainerRef}
                  style={{ width: '100%', position: 'relative' }}
                  onClick={handleMapClick}
                  onPointerMove={handlePointerMove}
                  onPointerUp={handlePointerUp}
                >
                  <img
                    className="map-img"
                    src={`${BASE_URL}${game.map_image_url}`}
                    alt="Campaign map"
                    draggable={false}
                  />

                  {pins.filter(p => p.group_id === null).map(pin => {
                    const session = sessionForPin(pin)
                    const note = noteForPin(pin)
                    const pinColour = (pin.colour as PinColour) ?? 'grey'
                    const pinLabel = note?.title ?? session?.title ?? (pin.label || '')
                    return (
                      <div
                        key={pin.id}
                        className={`map-pin-wrapper${hoveredPinId === pin.id ? ' map-pin-wrapper--hovered' : ''}${dragging?.pinId === pin.id ? ' map-pin-wrapper--dragging' : ''}${dropTargetIds.has(pin.id) ? ' map-pin-wrapper--drop-target' : ''}`}
                        style={{ left: `${pin.x}%`, top: `${pin.y}%` }}
                        onMouseEnter={() => setHoveredPinId(pin.id)}
                        onMouseLeave={() => setHoveredPinId(null)}
                      >
                        <button
                          className="map-pin"
                          style={{ '--pin-colour': COLOUR_MAP[pinColour] ?? COLOUR_MAP.grey } as React.CSSProperties}
                          title={pinLabel}
                          onClick={e => {
                            e.stopPropagation()
                            if (!wasDragRef.current) {
                              if (pin.note_id) {
                                navigate(`/games/${gameId}/notes/${pin.note_id}`)
                              } else if (pin.session_id) {
                                navigate(`/games/${gameId}/sessions/${pin.session_id}/notes`)
                              } else {
                                { setEditingPinId(editingPinId === pin.id ? null : pin.id); setEditLinkSearch('') }
                              }
                            }
                          }}
                          onPointerDown={e => handlePinPointerDown(e, pin)}
                        >
                          {(() => {
                            const IconComp = PIN_ICON_COMPONENTS[pin.icon] ?? PIN_ICON_COMPONENTS['position-marker']
                            return <span className="map-pin__icon"><IconComp size={10} /></span>
                          })()}
                        </button>
                        {pinLabel && (
                          <span
                            className="map-pin__label"
                            onClick={e => {
                              e.stopPropagation()
                              if (pin.note_id) {
                                navigate(`/games/${gameId}/notes/${pin.note_id}`)
                              } else if (pin.session_id) {
                                navigate(`/games/${gameId}/sessions/${pin.session_id}/notes`)
                              } else {
                                setEditingPinId(editingPinId === pin.id ? null : pin.id)
                                setEditLinkSearch('')
                              }
                            }}
                          >
                            {note ? (
                              <>
                                <span className="map-pin__label-type">Note</span>
                                {note.title}
                              </>
                            ) : session ? (
                              <>
                                {session.session_number != null && (
                                  <span className="map-pin__label-num">#{session.session_number}</span>
                                )}
                                {session.title}
                              </>
                            ) : (
                              pin.label
                            )}
                          </span>
                        )}
                        <button
                          className="map-pin__edit"
                          title="Edit pin"
                          onClick={e => {
                            e.stopPropagation()
                            { setEditingPinId(editingPinId === pin.id ? null : pin.id); setEditLinkSearch('') }
                          }}
                        >
                          ✎
                        </button>
                        <button
                          className="map-pin__delete"
                          title="Remove pin"
                          onClick={e => { e.stopPropagation(); handleDeletePin(pin.id) }}
                        >
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
                        {editingPinId === pin.id && (
                          <div className="map-pin-edit-popover" onClick={e => e.stopPropagation()}>
                            <span className="map-pin-edit-popover-label">Label</span>
                            <input
                              className="map-marker-input"
                              type="text"
                              placeholder="Pin label…"
                              maxLength={100}
                              defaultValue={pin.label ?? ''}
                              onBlur={e => {
                                const val = e.target.value
                                if (val !== (pin.label ?? '')) handleEditPinField(pin.id, { label: val })
                              }}
                            />
                            <span className="map-pin-edit-popover-label">Description</span>
                            <textarea
                              className="map-marker-textarea"
                              placeholder="Optional description…"
                              rows={2}
                              maxLength={1000}
                              defaultValue={pin.description ?? ''}
                              onBlur={e => {
                                const val = e.target.value
                                if (val !== (pin.description ?? '')) handleEditPinField(pin.id, { description: val || null })
                              }}
                            />
                            <span className="map-pin-edit-popover-label">Colour</span>
                            <div className="map-pin-edit-popover-colours">
                              {PIN_COLOURS.map(c => (
                                <button
                                  key={c}
                                  className={`pin-colour-swatch${pinColour === c ? ' pin-colour-swatch--selected' : ''}`}
                                  style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                                  onClick={() => handleEditPinField(pin.id, { colour: c })}
                                  title={c}
                                />
                              ))}
                            </div>
                            <span className="map-pin-edit-popover-label">Icon</span>
                            <div className="map-pin-edit-popover-colours">
                              {PIN_ICONS.map(i => {
                                const IconComp = PIN_ICON_COMPONENTS[i]
                                return (
                                  <button
                                    key={i}
                                    className={`pin-icon-option${pin.icon === i ? ' pin-icon-option--selected' : ''}`}
                                    onClick={() => handleEditPinField(pin.id, { icon: i })}
                                    title={PIN_ICON_LABELS[i]}
                                    aria-label={PIN_ICON_LABELS[i]}
                                  >
                                    <IconComp size={14} />
                                  </button>
                                )
                              })}
                            </div>
                            <span className="map-pin-edit-popover-label">Link</span>
                            {pin.session_id ? (
                              <div className="map-pin-link-row">
                                <span className="map-pin-link-name">{sessionForPin(pin)?.title ?? pin.session_id}</span>
                                <button className="map-pin-unlink-btn" onClick={() => handleEditPinField(pin.id, { session_id: null })} title="Unlink session">Unlink</button>
                              </div>
                            ) : pin.note_id ? (
                              <div className="map-pin-link-row">
                                <span className="map-pin-link-name">{noteForPin(pin)?.title ?? pin.note_id}</span>
                                <button className="map-pin-unlink-btn" onClick={() => handleEditPinField(pin.id, { note_id: null })} title="Unlink note">Unlink</button>
                              </div>
                            ) : (
                              <div className="map-pin-link-search">
                                <input
                                  className="map-marker-input"
                                  type="text"
                                  placeholder="Search to link…"
                                  value={editLinkSearch}
                                  onChange={e => setEditLinkSearch(e.target.value)}
                                />
                                {editLinkSearch.trim() !== '' && (() => {
                                  const q = editLinkSearch.trim().toLowerCase()
                                  const matchedSessions = sessions.filter(s => s.title.toLowerCase().includes(q) || (s.session_number != null && `#${s.session_number}`.includes(q)))
                                  const matchedNotes = notes.filter(n => n.title.toLowerCase().includes(q))
                                  if (matchedSessions.length === 0 && matchedNotes.length === 0) {
                                    return <span className="map-pin-link-none">No matches</span>
                                  }
                                  return (
                                    <ul className="map-pin-link-results">
                                      {matchedSessions.slice(0, 3).map(s => (
                                        <li key={s.id}>
                                          <button className="map-picker-item" onClick={() => { handleEditPinField(pin.id, { session_id: s.id }); setEditLinkSearch('') }}>
                                            <span className="map-picker-item-type">Session</span>
                                            {s.session_number != null && <span className="map-picker-num">#{s.session_number}</span>}
                                            <span className="map-picker-name">{s.title}</span>
                                          </button>
                                        </li>
                                      ))}
                                      {matchedNotes.slice(0, 3).map(n => (
                                        <li key={n.id}>
                                          <button className="map-picker-item" onClick={() => { handleEditPinField(pin.id, { note_id: n.id }); setEditLinkSearch('') }}>
                                            <span className="map-picker-item-type">Note</span>
                                            <span className="map-picker-name">{n.title}</span>
                                          </button>
                                        </li>
                                      ))}
                                    </ul>
                                  )
                                })()}
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}

                  {pinGroups.map(group => {
                    const grpColour = (group.colour as PinColour) ?? 'grey'
                    const GroupIconComp = PIN_ICON_COMPONENTS[group.icon] ?? PIN_ICON_COMPONENTS['position-marker']
                    return (
                      <div
                        key={group.id}
                        className={`map-pin-wrapper map-pin-wrapper--group${dropTargetIds.has(group.id) ? ' map-pin-wrapper--drop-target' : ''}`}
                        style={{ left: `${group.x}%`, top: `${group.y}%` }}
                        data-group-id={group.id}
                      >
                        <button
                          className="map-pin"
                          style={{ '--pin-colour': COLOUR_MAP[grpColour] ?? COLOUR_MAP.grey } as React.CSSProperties}
                          title={`Group (${group.pin_count} pins)`}
                          onClick={e => {
                            e.stopPropagation()
                            setActiveGroupId(activeGroupId === group.id ? null : group.id)
                          }}
                        >
                          <span className="map-pin__icon"><GroupIconComp size={10} /></span>
                        </button>
                        <span className="map-pin-group-badge">{group.pin_count}</span>
                      </div>
                    )
                  })}
                </div>
              </TransformComponent>
            </TransformWrapper>

            {/* Group popover — outside TransformWrapper so it's not affected by zoom/pan */}
            {activeGroupId && (() => {
              const group = pinGroups.find(g => g.id === activeGroupId)
              if (!group || !viewportContainerRef.current) return null
              const vpRect = viewportContainerRef.current.getBoundingClientRect()
              const markerEl = viewportContainerRef.current.querySelector(`[data-group-id="${group.id}"]`)
              if (!markerEl) return null
              const markerRect = markerEl.getBoundingClientRect()
              const popoverWidth = 180
              const popoverMaxHeight = 200
              // Position relative to viewport container
              let left = markerRect.left - vpRect.left + markerRect.width / 2 - popoverWidth / 2
              let top = markerRect.top - vpRect.top - popoverMaxHeight - 8
              let flipBelow = false
              // Clamp horizontal
              if (left < 4) left = 4
              if (left + popoverWidth > vpRect.width - 4) left = vpRect.width - popoverWidth - 4
              // Flip below if not enough room above
              if (top < 4) {
                top = markerRect.top - vpRect.top + markerRect.height + 8
                flipBelow = true
              }
              return (
                <div
                  className={`map-pin-group-popover${flipBelow ? ' map-pin-group-popover--below' : ''}`}
                  style={{ left: `${left}px`, top: `${top}px`, width: `${popoverWidth}px`, maxHeight: `${popoverMaxHeight}px` }}
                  onClick={e => e.stopPropagation()}
                >
                  <div className="map-pin-group-popover-header">
                    <span>{group.pin_count} pin{group.pin_count !== 1 ? 's' : ''}</span>
                    <button onClick={() => { setManagingGroupId(group.id); setActiveGroupId(null) }}>Manage</button>
                  </div>
                  <ul className="map-pin-group-popover-list">
                    {group.pins.map(p => {
                      const s = sessions.find(sess => sess.id === p.session_id)
                      const n = notes.find(nt => nt.id === p.note_id)
                      return (
                        <li key={p.id}>
                          <button onClick={() => {
                            if (p.note_id) navigate(`/games/${gameId}/notes/${p.note_id}`)
                            else if (p.session_id) navigate(`/games/${gameId}/sessions/${p.session_id}/notes`)
                          }}>
                            {n ? <span className="map-pin-group-popover-type">Note</span> : null}
                            {n?.title ?? s?.title ?? p.label ?? '?'}
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              )
            })()}

            {/* Overlay panel — outside TransformWrapper, fixed to viewport container */}
            {!panelOpen && (
              <button
                className="map-panel-toggle"
                onClick={() => setPanelOpen(true)}
                title="Show controls"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <path d="M9 18l6-6-6-6" />
                </svg>
              </button>
            )}

            {panelOpen && (
              <div className="map-overlay-panel">
                <button
                  className="map-panel-close"
                  onClick={() => setPanelOpen(false)}
                  title="Hide controls"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                    <path d="M15 18l-6-6 6-6" />
                  </svg>
                </button>

                <button className="map-back-btn" onClick={() => navigate(`/games/${gameId}`)}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                    <path d="M19 12H5M12 19l-7-7 7-7" />
                  </svg>
                  Back to Sessions
                </button>

                <h2 className="map-panel-title">{game?.title ?? 'Campaign Map'}</h2>

                {isGM && (
                  <div className="map-toolbar">
                    {uploadError && <span className="map-upload-error">{uploadError}</span>}
                    <button className="map-upload-btn" onClick={handleUploadClick} disabled={uploading}>
                      {uploading ? 'Uploading…' : 'Replace Map'}
                    </button>
                    <button className="map-delete-btn" onClick={handleDeleteMap}>
                      Remove Map
                    </button>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      className="map-file-input"
                      onChange={handleFileChange}
                    />
                  </div>
                )}

                <p className="map-gm-hint">
                  Click anywhere on the map to place a pin. Drag pins to reposition.
                  Middle-click and drag to pan. Ctrl + scroll to zoom.
                </p>

                <div className="map-zoom-controls">
                  <button className="map-zoom-btn" onClick={() => transformRef.current?.zoomIn()} disabled={displayScale >= 5} title="Zoom in">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <line x1="12" y1="5" x2="12" y2="19" />
                      <line x1="5" y1="12" x2="19" y2="12" />
                    </svg>
                  </button>
                  <button className="map-zoom-level" onClick={() => transformRef.current?.resetTransform()} title="Reset zoom">
                    {Math.round(displayScale * 100)}%
                  </button>
                  <button className="map-zoom-btn" onClick={() => transformRef.current?.zoomOut()} disabled={displayScale <= 1} title="Zoom out">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <line x1="5" y1="12" x2="19" y2="12" />
                    </svg>
                  </button>
                </div>

                {unpinnedSessions.length === 0 && pins.length > 0 && (
                  <p className="map-all-pinned">✦ All sessions are pinned on the map.</p>
                )}
              </div>
            )}

            <button
              className="map-sidebar-toggle-btn"
              onClick={handleSidebarToggle}
              title={sidebarOpen ? 'Hide folders' : 'Show folders'}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                <path d="M3 7h18M3 12h12M3 17h9" />
              </svg>
            </button>

            {sidebarOpen && (
              <FolderSidebar
                gameId={gameId!}
                isGM={isGM}
                userId={user?.id ?? ''}
                sessions={sessions}
                notes={notes}
                onSessionClick={(id) => navigate(`/games/${gameId}/sessions/${id}/notes`)}
                onNoteClick={(id) => navigate(`/games/${gameId}/notes/${id}`)}
                onSessionUpdate={handleSessionUpdate}
                onNoteUpdate={handleNoteUpdate}
                onCreateSession={handleCreateSession}
                onCreateNote={handleCreateNote}
              />
            )}
          </div>
        )}
      </div>

      {/* Pin Picker Modal — portalled to body to avoid transform containing block issues */}
      {pendingCoords && createPortal(
        <div className="map-overlay" onClick={() => { setPendingCoords(null); setPendingLabel(''); setPendingDescription(''); setPickerSearch('') }}>
          <div className="map-session-picker" onClick={e => e.stopPropagation()}>
            <div className="map-picker-header">
              <span className="map-picker-rune" aria-hidden="true">⬡</span>
              <h3 className="map-picker-title">Mark This Location</h3>
            </div>
            <div className="map-picker-customise">
              <span className="map-picker-customise-label">Pin colour:</span>
              <div className="pin-colour-palette">
                {PIN_COLOURS.map(c => (
                  <button
                    key={c}
                    className={`pin-colour-swatch${pendingColour === c ? ' pin-colour-swatch--selected' : ''}`}
                    style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                    onClick={() => setPendingColour(c)}
                    title={c}
                  />
                ))}
              </div>
              <span className="map-picker-customise-label">Pin icon:</span>
              <div className="pin-colour-palette">
                {PIN_ICONS.map(i => {
                  const IconComp = PIN_ICON_COMPONENTS[i]
                  return (
                    <button
                      key={i}
                      className={`pin-icon-option${pendingIcon === i ? ' pin-icon-option--selected' : ''}`}
                      onClick={() => setPendingIcon(i)}
                      title={PIN_ICON_LABELS[i]}
                      aria-label={PIN_ICON_LABELS[i]}
                    >
                      <IconComp size={16} />
                    </button>
                  )
                })}
              </div>
            </div>

            <div className="map-picker-customise">
              <span className="map-picker-customise-label">Label (optional):</span>
              <input
                className="map-marker-input"
                type="text"
                placeholder="Pin label…"
                value={pendingLabel}
                onChange={e => setPendingLabel(e.target.value)}
                maxLength={100}
              />
              <span className="map-picker-customise-label">Description (optional):</span>
              <textarea
                className="map-marker-textarea"
                placeholder="Description…"
                rows={2}
                value={pendingDescription}
                onChange={e => setPendingDescription(e.target.value)}
                maxLength={1000}
              />
            </div>

            <button
              className="map-marker-submit map-marker-submit--full"
              onClick={() => handleCreateMarker(pendingLabel, pendingDescription)}
            >
              Place as Standalone Marker
            </button>

            {/* Link to session or note */}
            {(unpinnedSessions.length > 0 || notes.length > 0) && (
              <div className="map-picker-search-section">
                <span className="map-picker-customise-label">Link to a session or note:</span>
                <input
                  className="map-marker-input"
                  type="text"
                  placeholder="Search sessions &amp; notes…"
                  value={pickerSearch}
                  onChange={e => setPickerSearch(e.target.value)}
                />
                {pickerSearch.trim() !== '' && (() => {
                  const q = pickerSearch.trim().toLowerCase()
                  const matchedSessions = unpinnedSessions.filter(s =>
                    s.title.toLowerCase().includes(q) ||
                    (s.session_number != null && `#${s.session_number}`.includes(q))
                  )
                  const matchedNotes = notes.filter(n => n.title.toLowerCase().includes(q))
                  if (matchedSessions.length === 0 && matchedNotes.length === 0) {
                    return <p className="map-picker-empty">No matching sessions or notes.</p>
                  }
                  return (
                    <ul className="map-picker-list">
                      {matchedSessions.slice(0, 4).map(session => (
                        <li key={session.id}>
                          <button
                            className="map-picker-item"
                            onClick={() => handleSelectSession(session)}
                          >
                            <span className="map-picker-item-type">Session</span>
                            {session.session_number != null && (
                              <span className="map-picker-num">#{session.session_number}</span>
                            )}
                            <span className="map-picker-name">{session.title}</span>
                          </button>
                        </li>
                      ))}
                      {matchedNotes.slice(0, 4).map(note => (
                        <li key={note.id}>
                          <button
                            className="map-picker-item"
                            onClick={() => handleSelectNote(note)}
                          >
                            <span className="map-picker-item-type">Note</span>
                            <span className={`map-picker-vis map-picker-vis--${note.visibility}`}>
                              {note.visibility === 'private' ? '\uD83D\uDD12' : note.visibility === 'visible' ? '\uD83D\uDC41' : '\u270F\uFE0F'}
                            </span>
                            <span className="map-picker-name">{note.title}</span>
                          </button>
                        </li>
                      ))}
                    </ul>
                  )
                })()}
              </div>
            )}
          </div>
        </div>,
        document.body,
      )}

      {groupingPrompt && createPortal(
        <div className="map-overlay" onClick={() => setGroupingPrompt(null)}>
          <div className="map-grouping-prompt" onClick={e => e.stopPropagation()}>
            <div className="map-picker-header">
              <span className="map-picker-rune" aria-hidden="true">⬡</span>
              <h3 className="map-picker-title">Nearby Markers Detected</h3>
            </div>
            <p className="map-picker-sub">There are pins or groups nearby. How would you like to place this pin?</p>
            <ul className="map-picker-list">
              <li>
                <button className="map-picker-item" onClick={() => {
                  setPendingCoords(groupingPrompt.coords)
              
                  setGroupingPrompt(null)
                }}>Place as standalone pin</button>
              </li>
              {groupingPrompt.nearbyPins.length > 0 && (
                <li>
                  <button className="map-picker-item" onClick={() => {
                    setPendingCoords(groupingPrompt.coords)
                
                    setPendingGroupPinIds(groupingPrompt.nearbyPins.map(p => p.id))
                    setGroupingPrompt(null)
                  }}>Create new group with {groupingPrompt.nearbyPins.length} nearby pin(s)</button>
                </li>
              )}
              {groupingPrompt.nearbyGroups.map(g => (
                <li key={g.id}>
                  <button className="map-picker-item" onClick={() => {
                    setPendingCoords(groupingPrompt.coords)
                
                    setPendingAddToGroupId(g.id)
                    setGroupingPrompt(null)
                  }}>Add to group ({g.pin_count} pins)</button>
                </li>
              ))}
            </ul>
            <button className="map-picker-cancel" onClick={() => setGroupingPrompt(null)}>Cancel</button>
          </div>
        </div>,
        document.body,
      )}

      {dragGroupPrompt && createPortal(
        <div className="map-overlay" onClick={() => setDragGroupPrompt(null)}>
          <div className="map-grouping-prompt" onClick={e => e.stopPropagation()}>
            <div className="map-picker-header">
              <span className="map-picker-rune" aria-hidden="true">⬡</span>
              <h3 className="map-picker-title">Group Pins</h3>
            </div>
            <p className="map-picker-sub">You dropped this pin near other markers. Would you like to group them?</p>
            <ul className="map-picker-list">
              {dragGroupPrompt.nearbyPins.length > 0 && (
                <li>
                  <button className="map-picker-item" onClick={async () => {
                    if (!gameId) return
                    try {
                      await createPinGroup(gameId, [...dragGroupPrompt.nearbyPins.map(p => p.id), dragGroupPrompt.draggedPinId])
                      await reloadPinGroups()
                      // Reload pins so group_id is reflected
                      const updatedPins = await listGamePins(gameId)
                      setPins(updatedPins)
                    } catch (err: unknown) {
                      console.error('Failed to create group', err)
                    }
                    setDragGroupPrompt(null)
                  }}>
                    <span className="map-picker-name">
                      Create new group with {dragGroupPrompt.nearbyPins.length} nearby pin{dragGroupPrompt.nearbyPins.length !== 1 ? 's' : ''}
                    </span>
                  </button>
                </li>
              )}
              {dragGroupPrompt.nearbyGroups.map(g => (
                <li key={g.id}>
                  <button className="map-picker-item" onClick={async () => {
                    try {
                      await addPinToGroup(g.id, dragGroupPrompt.draggedPinId)
                      await reloadPinGroups()
                      if (gameId) {
                        const updatedPins = await listGamePins(gameId)
                        setPins(updatedPins)
                      }
                    } catch (err: unknown) {
                      console.error('Failed to add pin to group', err)
                    }
                    setDragGroupPrompt(null)
                  }}>
                    <span className="map-picker-name">Add to existing group ({g.pin_count} pins)</span>
                  </button>
                </li>
              ))}
              <li>
                <button className="map-picker-item" onClick={async () => {
                  // Just move the pin to where it was dropped (no grouping)
                  // We reverted visually, so we need to get the drop coords from the original drag
                  setDragGroupPrompt(null)
                }}>
                  <span className="map-picker-name">Cancel — keep pin in place</span>
                </button>
              </li>
            </ul>
          </div>
        </div>,
        document.body,
      )}

      {managingGroupId && (() => {
        const group = pinGroups.find(g => g.id === managingGroupId)
        if (!group) return null
        const mgmtColour = (group.colour as PinColour) ?? 'grey'
        const nearbyStandalonePins = pins.filter(p => {
          if (p.group_id !== null) return false
          const dx = p.x - group.x
          const dy = p.y - group.y
          return Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT * 4
        })
        return createPortal(
          <div className="map-overlay" onClick={() => setManagingGroupId(null)}>
            <div className="map-pin-group-manage" onClick={e => e.stopPropagation()}>
              <div className="map-picker-header">
                <span className="map-picker-rune" aria-hidden="true">⬡</span>
                <h3 className="map-picker-title">Manage Group</h3>
              </div>

              <span className="map-picker-customise-label">Group colour:</span>
              <div className="pin-colour-palette">
                {PIN_COLOURS.map(c => (
                  <button
                    key={c}
                    className={`pin-colour-swatch${mgmtColour === c ? ' pin-colour-swatch--selected' : ''}`}
                    style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                    onClick={async () => { await updatePinGroup(group.id, { colour: c }); await reloadPinGroups() }}
                    title={c}
                  />
                ))}
              </div>

              <span className="map-picker-customise-label">Group icon:</span>
              <div className="pin-colour-palette">
                {PIN_ICONS.map(i => {
                  const MgmtIconComp = PIN_ICON_COMPONENTS[i]
                  return (
                    <button
                      key={i}
                      className={`pin-icon-option${group.icon === i ? ' pin-icon-option--selected' : ''}`}
                      onClick={async () => { await updatePinGroup(group.id, { icon: i }); await reloadPinGroups() }}
                      title={PIN_ICON_LABELS[i]}
                      aria-label={PIN_ICON_LABELS[i]}
                    >
                      <MgmtIconComp size={14} />
                    </button>
                  )
                })}
              </div>

              <span className="map-picker-customise-label">Members ({group.pin_count}):</span>
              <ul className="map-pin-group-popover-list">
                {group.pins.map(p => {
                  const s = sessions.find(sess => sess.id === p.session_id)
                  const n = notes.find(nt => nt.id === p.note_id)
                  return (
                    <li key={p.id} className="map-pin-group-member-row">
                      <span>{n?.title ?? s?.title ?? p.label ?? '?'}</span>
                      <button
                        className="map-pin-group-remove-btn"
                        onClick={async () => {
                          await removePinFromGroup(group.id, p.id)
                          await reloadPinGroups()
                          setPins(prev => prev.map(pin => pin.id === p.id ? { ...pin, group_id: null } : pin))
                        }}
                        title="Remove from group"
                      >
                        ✕
                      </button>
                    </li>
                  )
                })}
              </ul>

              {nearbyStandalonePins.length > 0 && (
                <>
                  <span className="map-picker-customise-label">Add nearby pin:</span>
                  <ul className="map-picker-list">
                    {nearbyStandalonePins.map(p => {
                      const s = sessions.find(sess => sess.id === p.session_id)
                      const n = notes.find(nt => nt.id === p.note_id)
                      return (
                        <li key={p.id}>
                          <button
                            className="map-picker-item"
                            onClick={async () => {
                              await addPinToGroup(group.id, p.id)
                              await reloadPinGroups()
                              setPins(prev => prev.map(pin => pin.id === p.id ? { ...pin, group_id: group.id } : pin))
                            }}
                          >
                            {n?.title ?? s?.title ?? p.label ?? '?'}
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                </>
              )}

              <button
                className="map-delete-btn"
                style={{ marginTop: '0.5rem' }}
                onClick={async () => {
                  if (!confirm('Disband this group? All pins will become standalone.')) return
                  await disbandPinGroup(group.id)
                  const memberIds = new Set(group.pins.map(p => p.id))
                  setPins(prev => prev.map(p => memberIds.has(p.id) ? { ...p, group_id: null } : p))
                  await reloadPinGroups()
                  setManagingGroupId(null)
                }}
              >
                Disband Group
              </button>

              <button className="map-picker-cancel" onClick={() => setManagingGroupId(null)}>Close</button>
            </div>
          </div>,
          document.body,
        )
      })()}
    </div>
  )
}

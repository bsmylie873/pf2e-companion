import { useState, useEffect, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { useParams, useNavigate } from 'react-router-dom'
import { TransformWrapper, TransformComponent } from 'react-zoom-pan-pinch'
import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import { apiFetch, BASE_URL } from '../../api/client'
import { listGameSessions, updateSession, createSession } from '../../api/sessions'
import { listMapPins, createMapPin, createPin, updatePin, deletePin } from '../../api/pins'
import { listGameNotes, updateNote, createNote } from '../../api/notes'
import { listMaps, listArchivedMaps, uploadMapImage, archiveMap, createMap, renameMap, restoreMap, reorderMaps } from '../../api/maps'
import { listMemberships } from '../../api/memberships'
import { useAuth } from '../../context/AuthContext'
import { useMapNav } from '../../context/MapNavContext'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import type { Game } from '../../types/game'
import type { Session } from '../../types/session'
import type { SessionPin } from '../../types/pin'
import type { GameMembership } from '../../types/membership'
import type { Note } from '../../types/note'
import type { PinGroup } from '../../types/pin'
import type { GameMap } from '../../types/map'
import { listMapPinGroups, createMapPinGroup, addPinToGroup, removePinFromGroup, disbandPinGroup, updatePinGroup } from '../../api/pinGroups'
import { PIN_COLOURS, PIN_ICONS, COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS } from '../../constants/pins'
import type { PinColour, PinIcon } from '../../constants/pins'
import { getPreferences, updatePreferences } from '../../api/preferences'
import type { GameSidebarState } from '../../api/preferences'
import FolderSidebar from '../../components/FolderSidebar/FolderSidebar'
import EditorModalManager from '../../components/EditorModalManager/EditorModalManager'
import { useGameSocket } from '../../hooks/useGameSocket'
import type { GameSocketEvent } from '../../hooks/useGameSocket'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
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
  useDocumentTitle('Map')
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { register, unregister } = useMapNav()

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

  const [maps, setMaps] = useState<GameMap[]>([])
  const [activeMapId, setActiveMapId] = useLocalStorage<string | null>(
    `pf2e-map-${gameId}-last-map`,
    null,
  )
  const [archivedMaps, setArchivedMaps] = useState<GameMap[]>([])

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
  const [openItems, setOpenItems] = useState<Array<{ type: 'session' | 'note'; itemId: string; label: string }>>([])
  const [mapEditorMode, setMapEditorMode] = useState(false)

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const viewportContainerRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const transformRef = useRef<ReactZoomPanPinchRef>(null)
  const wasDragRef = useRef(false)
  const sidebarStateRef = useRef<Record<string, GameSidebarState>>({})
  const isFirstOpenRef = useRef(!window.localStorage.getItem(`pf2e-map-view-${gameId}`))
  const imageLoadedRef = useRef(false)

  /** Compute custom pan bounds so that any edge pixel can be centred in the viewport. */
  const computeBounds = useCallback((scale: number) => {
    const wrapper = transformRef.current?.instance?.wrapperComponent
    const content = mapContainerRef.current
    if (!wrapper || !content || content.offsetHeight === 0) return null

    const vpW = wrapper.offsetWidth
    const vpH = wrapper.offsetHeight
    const cW = content.offsetWidth
    const cH = content.offsetHeight

    return {
      minPosX: vpW / 2 - cW * scale,
      maxPosX: vpW / 2,
      minPosY: vpH / 2 - cH * scale,
      maxPosY: vpH / 2,
    }
  }, [])

  /** Clamp a pan position to the custom bounds. */
  const clampPosition = useCallback((positionX: number, positionY: number, scale: number) => {
    const bounds = computeBounds(scale)
    if (!bounds) return { positionX, positionY }
    return {
      positionX: Math.min(bounds.maxPosX, Math.max(bounds.minPosX, positionX)),
      positionY: Math.min(bounds.maxPosY, Math.max(bounds.minPosY, positionY)),
    }
  }, [computeBounds])

  /** Enforce custom pan bounds on every transform change (pan, zoom). */
  const handleTransformed = useCallback(
    (ref: ReactZoomPanPinchRef, state: { scale: number; positionX: number; positionY: number }) => {
      const { scale, positionX, positionY } = state
      const clamped = clampPosition(positionX, positionY, scale)
      if (
        Math.abs(clamped.positionX - positionX) > 0.5 ||
        Math.abs(clamped.positionY - positionY) > 0.5
      ) {
        ref.setTransform(clamped.positionX, clamped.positionY, scale, 0)
      }
    },
    [clampPosition],
  )

  /** After the map image loads, centre (first open) or clamp (returning user). */
  const handleImageLoad = useCallback(() => {
    imageLoadedRef.current = true
    const ref = transformRef.current
    if (!ref?.state) return

    if (isFirstOpenRef.current) {
      ref.centerView(1, 0)
      isFirstOpenRef.current = false
      const { scale, positionX, positionY } = ref.state
      setViewState({ scale, positionX, positionY })
    } else {
      const { scale, positionX, positionY } = ref.state
      const clamped = clampPosition(positionX, positionY, scale)
      if (
        Math.abs(clamped.positionX - positionX) > 0.5 ||
        Math.abs(clamped.positionY - positionY) > 0.5
      ) {
        ref.setTransform(clamped.positionX, clamped.positionY, scale, 0)
        setViewState({ scale, positionX: clamped.positionX, positionY: clamped.positionY })
      }
    }
  }, [clampPosition, setViewState])

  /** Persist transform to localStorage — called only when an interaction ends. */
  const handleTransformEnd = useCallback((ref: ReactZoomPanPinchRef) => {
    const { scale, positionX, positionY } = ref.state
    setDisplayScale(scale)
    const clamped = clampPosition(positionX, positionY, scale)
    setViewState(prev => ({
      ...prev,
      scale,
      positionX: clamped.positionX,
      positionY: clamped.positionY,
    }))
  }, [setViewState, clampPosition])

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
      listMemberships(gameId),
      listGameNotes(gameId),
      listMaps(gameId),
    ])
      .then(([gameData, sessionsData, membershipsData, notesData, mapsData]) => {
        if (!cancelled) {
          setGame(gameData)
          setSessions(sessionsData)
          setMemberships(membershipsData)
          setNotes(notesData)
          setMaps(mapsData)
          setActiveMapId((prev: string | null) => {
            if (prev && mapsData.some((m: GameMap) => m.id === prev)) return prev
            return mapsData[0]?.id ?? null
          })
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

  // Per-map data loading: reload pins + pin groups whenever the active map changes
  // Fetch archived maps for GM users
  useEffect(() => {
    if (!gameId || !isGM) return
    listArchivedMaps(gameId).then(setArchivedMaps).catch(() => {})
  }, [gameId, isGM])

  useEffect(() => {
    if (!gameId || !activeMapId) {
      setPins([])
      setPinGroups([])
      return
    }
    let cancelled = false
    Promise.all([
      listMapPins(gameId, activeMapId),
      listMapPinGroups(gameId, activeMapId),
    ]).then(([pinsData, pinGroupsData]) => {
      if (!cancelled) {
        setPins(pinsData)
        setPinGroups(pinGroupsData)
      }
    }).catch((err: unknown) => console.error('Failed to load map data', err))
    return () => { cancelled = true }
  }, [gameId, activeMapId])

  // Fetch preferences on mount and whenever the window regains focus
  // (so changes made in Settings take effect without a full refresh).
  useEffect(() => {
    if (!gameId) return
    const fetchPrefs = () => {
      getPreferences().then(prefs => {
        sidebarStateRef.current = prefs.sidebar_state ?? {}
        const gameState = sidebarStateRef.current[gameId]
        if (gameState?.panelOpen) setSidebarOpen(true)
        setMapEditorMode(prefs.map_editor_mode === 'modal')
      }).catch(() => {})
    }
    fetchPrefs()
    window.addEventListener('focus', fetchPrefs)
    return () => window.removeEventListener('focus', fetchPrefs)
  }, [gameId])

  // Real-time map updates via unified game WebSocket
  const handleGameEvent = useCallback((event: GameSocketEvent) => {
    switch (event.type) {
      case 'map_created': {
        const created = event.data as GameMap
        setMaps(prev => prev.some(m => m.id === created.id) ? prev : [...prev, created])
        break
      }
      case 'map_renamed':
      case 'map_image_updated':
        setMaps(prev => prev.map(m => m.id === (event.data as GameMap).id ? event.data as GameMap : m))
        break
      case 'map_archived': {
        const archivedId = (event.data as { id: string }).id
        setMaps(prev => {
          const remaining = prev.filter(m => m.id !== archivedId)
          setActiveMapId((current: string | null) => {
            if (current === archivedId) return remaining[0]?.id ?? null
            return current
          })
          return remaining
        })
        break
      }
      case 'map_restored': {
        const restored = event.data as GameMap
        setArchivedMaps(prev => prev.filter(m => m.id !== restored.id))
        setMaps(prev => prev.some(m => m.id === restored.id) ? prev : [...prev, restored].sort((a, b) => a.sort_order - b.sort_order))
        break
      }
      case 'map_reordered':
        if (gameId) listMaps(gameId).then(setMaps).catch(() => {})
        break
      case 'pin_created':
        setPins(prev => [...prev, event.data as SessionPin])
        break
      case 'pin_updated':
        setPins(prev => prev.map(p => p.id === (event.data as SessionPin).id ? event.data as SessionPin : p))
        break
      case 'pin_deleted':
        setPins(prev => prev.filter(p => p.id !== (event.data as { id: string }).id))
        break
      case 'pin_group_created':
        setPinGroups(prev => [...prev, event.data as PinGroup])
        break
      case 'pin_group_updated':
        setPinGroups(prev => prev.map(g => g.id === (event.data as PinGroup).id ? event.data as PinGroup : g))
        break
      case 'pin_group_disbanded':
        setPinGroups(prev => prev.filter(g => g.id !== (event.data as { id: string }).id))
        break
      case '__reconnected':
        if (gameId) {
          listMaps(gameId).then(setMaps).catch(() => {})
          listArchivedMaps(gameId).then(setArchivedMaps).catch(() => {})
          if (activeMapId) {
            listMapPins(gameId, activeMapId).then(setPins).catch(() => {})
            listMapPinGroups(gameId, activeMapId).then(setPinGroups).catch(() => {})
          }
        }
        break
    }
  }, [gameId, activeMapId])

  useGameSocket(gameId, handleGameEvent)

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
  }, [activeMapId, loading])

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

  /** Re-clamp position when the viewport is resized. */
  useEffect(() => {
    const container = viewportContainerRef.current
    if (!container) return

    const observer = new ResizeObserver(() => {
      if (!imageLoadedRef.current || !transformRef.current) return
      const ref = transformRef.current
      const { scale, positionX, positionY } = ref.state
      const clamped = clampPosition(positionX, positionY, scale)
      if (
        Math.abs(clamped.positionX - positionX) > 0.5 ||
        Math.abs(clamped.positionY - positionY) > 0.5
      ) {
        ref.setTransform(clamped.positionX, clamped.positionY, scale, 0)
      }
    })

    observer.observe(container)
    return () => observer.disconnect()
  }, [clampPosition])

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
    if (!gameId || !activeMapId) return
    try {
      const groups = await listMapPinGroups(gameId, activeMapId)
      setPinGroups(groups)
    } catch (err: unknown) {
      console.error('Failed to reload pin groups', err)
    }
  }, [gameId, activeMapId])

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
        await createMapPinGroup(gameId, activeMapId!, [...pendingGroupPinIds, pin.id])
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
  }, [gameId, activeMapId, pendingCoords, pendingColour, pendingIcon, pendingLabel, pendingDescription, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

  const handleSelectNote = useCallback(async (note: Note) => {
    if (!gameId || !pendingCoords) return
    const label = pendingLabel.trim() || note.title
    const desc = pendingDescription.trim() || undefined
    try {
      const pin = await createMapPin(gameId, activeMapId!, {
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
        await createMapPinGroup(gameId, activeMapId!, [...pendingGroupPinIds, pin.id])
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
  }, [gameId, activeMapId, pendingCoords, pendingColour, pendingIcon, pendingLabel, pendingDescription, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

  const handleCreateMarker = useCallback(async (label: string, description: string) => {
    if (!gameId || !pendingCoords) return
    const trimmed = label.trim()
    try {
      const pin = await createMapPin(gameId, activeMapId!, {
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
        await createMapPinGroup(gameId, activeMapId!, [...pendingGroupPinIds, pin.id])
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
  }, [gameId, activeMapId, pendingCoords, pendingColour, pendingIcon, pendingGroupPinIds, pendingAddToGroupId, reloadPinGroups, defaultPinColour, defaultPinIcon])

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
      if (gameId && activeMapId) {
        try {
          const freshPins = await listMapPins(gameId, activeMapId)
          setPins(freshPins)
        } catch { /* ignore */ }
      }
    }
  }, [gameId, activeMapId])

  const handleUploadClick = () => fileInputRef.current?.click()

  const handleFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !gameId || !activeMapId) return
    setUploading(true)
    setUploadError(null)
    try {
      const updatedMap = await uploadMapImage(gameId, activeMapId, file)
      setMaps(prev => prev.map(m => m.id === activeMapId ? updatedMap : m))
    } catch (err: unknown) {
      setUploadError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }, [gameId, activeMapId])

  const handleDeleteMap = useCallback(async () => {
    if (!gameId || !activeMapId || !confirm('Archive this map? All session pins and pin groups on this map will be archived. You have 24 hours to restore.')) return
    try {
      const archivedMap = maps.find(m => m.id === activeMapId)
      await archiveMap(gameId, activeMapId)
      setMaps(prev => {
        const remaining = prev.filter(m => m.id !== activeMapId)
        setActiveMapId(remaining[0]?.id ?? null)
        return remaining
      })
      if (archivedMap) setArchivedMaps(prev => [...prev, { ...archivedMap, archived_at: new Date().toISOString() }])
    } catch (err: unknown) {
      console.error('Failed to archive map', err)
    }
  }, [gameId, activeMapId, maps])

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

  const openItem = useCallback((type: 'session' | 'note', itemId: string, label: string) => {
    if (!mapEditorMode) {
      if (type === 'note') navigate(`/games/${gameId}/notes/${itemId}`)
      else navigate(`/games/${gameId}/sessions/${itemId}/notes`)
      return
    }
    setOpenItems(prev => {
      if (prev.some(i => i.itemId === itemId)) return prev
      return [...prev, { type, itemId, label }]
    })
  }, [mapEditorMode, navigate, gameId])

  const handleCreateMap = useCallback(async (name: string) => {
    if (!gameId) return
    try {
      const newMap = await createMap(gameId, { name })
      setMaps(prev => prev.some(m => m.id === newMap.id) ? prev : [...prev, newMap])
      setActiveMapId(newMap.id)
    } catch (err) { console.error('Failed to create map', err) }
  }, [gameId])

  const handleRenameMap = useCallback(async (mapId: string, name: string) => {
    if (!gameId) return
    try {
      const updated = await renameMap(gameId, mapId, { name })
      setMaps(prev => prev.map(m => m.id === mapId ? updated : m))
    } catch (err) { console.error('Failed to rename map', err) }
  }, [gameId])

  const handleArchiveMap = useCallback(async (mapId: string) => {
    if (!gameId) return
    try {
      await archiveMap(gameId, mapId)
      const archivedMap = maps.find(m => m.id === mapId)
      setMaps(prev => {
        const remaining = prev.filter(m => m.id !== mapId)
        if (activeMapId === mapId) {
          setActiveMapId(remaining[0]?.id ?? null)
        }
        return remaining
      })
      if (archivedMap) setArchivedMaps(prev => [...prev, { ...archivedMap, archived_at: new Date().toISOString() }])
    } catch (err) { console.error('Failed to archive map', err) }
  }, [gameId, activeMapId, maps])

  const handleRestoreMap = useCallback(async (mapId: string) => {
    if (!gameId) return
    try {
      const restored = await restoreMap(gameId, mapId)
      setArchivedMaps(prev => prev.filter(m => m.id !== mapId))
      setMaps(prev => prev.some(m => m.id === restored.id) ? prev : [...prev, restored].sort((a, b) => a.sort_order - b.sort_order))
    } catch (err) {
      console.error('Failed to restore map', err)
      if (err instanceof Error && err.message.includes('not found')) {
        setArchivedMaps(prev => prev.filter(m => m.id !== mapId))
      }
    }
  }, [gameId])

  const handleReorderMaps = useCallback(async (ids: string[]) => {
    if (!gameId) return
    setMaps(prev => ids.map(id => prev.find(m => m.id === id)!).filter(Boolean))
    try {
      await reorderMaps(gameId, ids)
    } catch (err) { console.error('Failed to reorder maps', err) }
  }, [gameId])

  // Register map nav state into context for TopBar breadcrumb
  useEffect(() => {
    if (!gameId || loading || error) return
    register({
      gameId,
      gameTitle: game?.title ?? 'Campaign',
      maps,
      archivedMaps,
      activeMapId,
      isGM,
      onSelectMap: (mapId: string) => setActiveMapId(mapId),
      onCreateMap: handleCreateMap,
      onRenameMap: handleRenameMap,
      onArchiveMap: handleArchiveMap,
      onUnarchiveMap: handleRestoreMap,
      onReorderMaps: handleReorderMaps,
    })
  }, [gameId, game?.title, maps, archivedMaps, activeMapId, isGM, loading, error])

  useEffect(() => {
    return () => unregister()
  }, [unregister])

  const sessionForPin = (pin: SessionPin) => sessions.find(s => s.id === pin.session_id)
  const noteForPin = (pin: SessionPin) => notes.find(n => n.id === pin.note_id)

  return (
    <div className="map-view-page">
      <div className="map-view-inner">
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

        {!loading && !error && maps.length === 0 && !isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">The Map Awaits</p>
            <p className="map-empty-sub">The Game Master has not yet unveiled the realm.</p>
          </div>
        )}

        {!loading && !error && activeMapId && !maps.find(m => m.id === activeMapId)?.image_url && isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">No Map Image</p>
            <p className="map-empty-sub">Upload an image for <em>{maps.find(m => m.id === activeMapId)?.name}</em> to begin placing session markers.</p>
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

        {!loading && !error && activeMapId && !maps.find(m => m.id === activeMapId)?.image_url && !isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">The Map Awaits</p>
            <p className="map-empty-sub">The Game Master has not yet uploaded an image for this map.</p>
          </div>
        )}

        {!loading && !error && maps.length === 0 && isGM && (
          <div className="map-empty">
            <div className="map-empty-sigil" aria-hidden="true">⊕</div>
            <p className="map-empty-title">No Maps Yet</p>
            <p className="map-empty-sub">Name your first map to get started.</p>
            <form className="map-first-create" onSubmit={(e) => {
              e.preventDefault()
              const input = e.currentTarget.querySelector('input')
              const name = input?.value.trim()
              if (name) { handleCreateMap(name); input!.value = '' }
            }}>
              <input
                className="map-first-create-input"
                type="text"
                placeholder="e.g. Otari Region, Dungeon Level 1…"
                autoFocus
                maxLength={255}
              />
              <button className="map-upload-btn" type="submit">
                + Create Map
              </button>
            </form>
          </div>
        )}

        {!loading && !error && activeMapId && maps.find(m => m.id === activeMapId)?.image_url && (
          <>
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
              minScale={0.75}
              maxScale={5}
              limitToBounds={false}
              disablePadding={true}
              centerOnInit={false}
              alignmentAnimation={{ disabled: true }}
              panning={{
                allowLeftClickPan: false,
                allowMiddleClickPan: true,
                allowRightClickPan: false,
                velocityDisabled: true,
              }}
              wheel={{
                activationKeys: ['Control', 'Meta'],
                step: 0.25,
              }}
              doubleClick={{ disabled: true }}
              onTransformed={handleTransformed}
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
                    src={`${BASE_URL}${maps.find(m => m.id === activeMapId)?.image_url ?? ''}`}
                    alt="Campaign map"
                    draggable={false}
                    onLoad={handleImageLoad}
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
                                openItem('note', pin.note_id, noteForPin(pin)?.title ?? pin.label ?? 'Note')
                              } else if (pin.session_id) {
                                openItem('session', pin.session_id, sessionForPin(pin)?.title ?? 'Session')
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
                                openItem('note', pin.note_id, noteForPin(pin)?.title ?? pin.label ?? 'Note')
                              } else if (pin.session_id) {
                                openItem('session', pin.session_id, sessionForPin(pin)?.title ?? 'Session')
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
                            if (p.note_id) openItem('note', p.note_id, notes.find(n => n.id === p.note_id)?.title ?? p.label ?? '?')
                            else if (p.session_id) openItem('session', p.session_id, sessions.find(s => s.id === p.session_id)?.title ?? '?')
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
                  <button className="map-zoom-btn" onClick={() => transformRef.current?.zoomOut()} disabled={displayScale <= 0.75} title="Zoom out">
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

          </div>
          </>
        )}

        {!loading && !error && gameId && (
          <>
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
                gameId={gameId}
                isGM={isGM}
                userId={user?.id ?? ''}
                sessions={sessions}
                notes={notes}
                onSessionClick={(id) => { const s = sessions.find(x => x.id === id); openItem('session', id, s?.title ?? 'Session') }}
                onNoteClick={(id) => { const n = notes.find(x => x.id === id); openItem('note', id, n?.title ?? 'Note') }}
                onSessionUpdate={handleSessionUpdate}
                onNoteUpdate={handleNoteUpdate}
                onCreateSession={handleCreateSession}
                onCreateNote={handleCreateNote}
              />
            )}
          </>
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
                    if (!gameId || !activeMapId) return
                    try {
                      await createMapPinGroup(gameId, activeMapId, [...dragGroupPrompt.nearbyPins.map(p => p.id), dragGroupPrompt.draggedPinId])
                      await reloadPinGroups()
                      // Reload pins so group_id is reflected
                      const updatedPins = await listMapPins(gameId, activeMapId)
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
                      if (gameId && activeMapId) {
                        const updatedPins = await listMapPins(gameId, activeMapId)
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

      {openItems.length > 0 && (
        <EditorModalManager
          items={openItems}
          gameId={gameId!}
          onClose={(itemId) => setOpenItems(prev => prev.filter(i => i.itemId !== itemId))}
          onCloseAll={() => setOpenItems([])}
        />
      )}
    </div>
  )
}

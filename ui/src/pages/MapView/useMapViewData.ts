import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import { apiFetch } from '../../api/client'
import { listGameSessions, updateSession, createSession } from '../../api/sessions'
import { listMapPins, createMapPin, createPin, updatePin, deletePin } from '../../api/pins'
import { listGameNotes, updateNote, createNote } from '../../api/notes'
import {
  listMaps,
  listArchivedMaps,
  uploadMapImage,
  archiveMap,
  createMap,
  renameMap,
  restoreMap,
  reorderMaps,
} from '../../api/maps'
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
import { listMapPinGroups, createMapPinGroup, addPinToGroup } from '../../api/pinGroups'
import { PIN_COLOURS, PIN_ICONS } from '../../constants/pins'
import type { PinColour, PinIcon } from '../../constants/pins'
import { getPreferences, updatePreferences } from '../../api/preferences'
import type { GameSidebarState } from '../../api/preferences'
import { useGameSocket } from '../../hooks/useGameSocket'
import type { GameSocketEvent } from '../../hooks/useGameSocket'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'

/** Proximity threshold in map-percentage units (0–100). ~16px on a 1000px map. */
export const GROUP_PROXIMITY_PCT = 1.5

export interface MapViewState {
  scale: number
  positionX: number
  positionY: number
}

export const DEFAULT_VIEW_STATE: MapViewState = {
  scale: 1,
  positionX: 0,
  positionY: 0,
}

export function useMapViewData(gameId: string | undefined) {
  useDocumentTitle('Map')
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

  // Sidebar-to-canvas drag-and-drop state
  const [sidebarDragOver, setSidebarDragOver] = useState(false)
  const [toastMessage, setToastMessage] = useState<string | null>(null)
  const [dropLinkedItem, setDropLinkedItem] = useState<{ type: 'session' | 'note'; id: string; label: string } | null>(null)
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

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

  // Suppress the browser's default right-click context menu on the map viewport
  useEffect(() => {
    const vp = document.querySelector('.map-viewport')
    if (!vp) return
    const onContextMenu = (e: Event) => { e.preventDefault() }
    vp.addEventListener('contextmenu', onContextMenu)
    return () => vp.removeEventListener('contextmenu', onContextMenu)
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

  /** Show a transient toast that auto-dismisses after 3 seconds. */
  const showToast = useCallback((message: string) => {
    if (toastTimerRef.current) clearTimeout(toastTimerRef.current)
    setToastMessage(message)
    toastTimerRef.current = setTimeout(() => setToastMessage(null), 3000)
  }, [])

  /** Sidebar-to-canvas drag handlers */
  const handleCanvasDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    if (!e.dataTransfer.types.includes('mapdroptype')) return
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setSidebarDragOver(true)
  }, [])

  const handleCanvasDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    // Only fire when truly leaving the container, not entering a child
    const related = e.relatedTarget as Node | null
    if (related && e.currentTarget.contains(related)) return
    setSidebarDragOver(false)
  }, [])

  const handleCanvasDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setSidebarDragOver(false)
    const dropType = e.dataTransfer.getData('mapDropType') as 'session' | 'note'
    const dropId = e.dataTransfer.getData('mapDropId')
    const dropLabel = e.dataTransfer.getData('mapDropLabel')
    if (!dropType || !dropId) return

    // Check if already pinned
    if (dropType === 'session' && pinnedSessionIds.has(dropId)) {
      showToast('This session already has a pin on the map.')
      return
    }
    if (dropType === 'note' && pins.some(p => p.note_id === dropId)) {
      showToast('This note already has a pin on the map.')
      return
    }

    // Convert drop position to map-space coordinates
    const coords = clientToMapPct(e.clientX, e.clientY)
    setPendingCoords(coords)
    setPendingLabel(dropLabel || '')
    setDropLinkedItem({ type: dropType, id: dropId, label: dropLabel || '' })
  }, [pinnedSessionIds, pins, clientToMapPct, showToast])

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

  return {
    // Data
    game,
    sessions,
    setSessions,
    pins,
    setPins,
    notes,
    setNotes,
    memberships,
    // Loading / error
    loading,
    error,
    uploadError,
    uploading,
    // Map state
    maps,
    setMaps,
    activeMapId,
    setActiveMapId,
    archivedMaps,
    // View state
    viewState,
    setViewState,
    panelOpen,
    setPanelOpen,
    displayScale,
    setDisplayScale,
    // Pin placement
    pendingCoords,
    setPendingCoords,
    pendingColour,
    setPendingColour,
    pendingIcon,
    setPendingIcon,
    defaultPinColour,
    defaultPinIcon,
    pendingLabel,
    setPendingLabel,
    pendingDescription,
    setPendingDescription,
    pinError,
    setPinError,
    // Pin interaction
    hoveredPinId,
    setHoveredPinId,
    dragging,
    editingPinId,
    setEditingPinId,
    editLinkSearch,
    setEditLinkSearch,
    dropTargetIds,
    // Search
    pickerSearch,
    setPickerSearch,
    // Pin groups
    pinGroups,
    setPinGroups,
    activeGroupId,
    setActiveGroupId,
    managingGroupId,
    setManagingGroupId,
    groupingPrompt,
    setGroupingPrompt,
    pendingGroupPinIds,
    setPendingGroupPinIds,
    pendingAddToGroupId,
    setPendingAddToGroupId,
    dragGroupPrompt,
    setDragGroupPrompt,
    // Sidebar
    sidebarOpen,
    openItems,
    setOpenItems,
    mapEditorMode,
    // Sidebar-to-canvas drag-and-drop
    sidebarDragOver,
    toastMessage,
    dropLinkedItem,
    setDropLinkedItem,
    handleCanvasDragOver,
    handleCanvasDragLeave,
    handleCanvasDrop,
    // Refs
    mapContainerRef,
    viewportContainerRef,
    fileInputRef,
    transformRef,
    wasDragRef,
    sidebarStateRef,
    // Derived
    isGM,
    pinnedSessionIds,
    unpinnedSessions,
    // Helpers
    sessionForPin,
    noteForPin,
    // Transform handlers
    handleTransformed,
    handleImageLoad,
    handleTransformEnd,
    // Map interaction
    handleMapClick,
    handlePointerMove,
    handlePointerUp,
    handlePinPointerDown,
    // Pin CRUD
    handleSelectSession,
    handleSelectNote,
    handleCreateMarker,
    handleDeletePin,
    handleEditPinField,
    reloadPinGroups,
    // Upload / map image
    handleUploadClick,
    handleFileChange,
    handleDeleteMap,
    // Sidebar / session / note handlers
    handleSidebarToggle,
    handleSessionUpdate,
    handleNoteUpdate,
    handleCreateSession,
    handleCreateNote,
    openItem,
    // Map management
    handleCreateMap,
    handleRenameMap,
    handleArchiveMap,
    handleRestoreMap,
    handleReorderMaps,
  }
}

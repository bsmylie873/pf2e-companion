import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { apiFetch, BASE_URL } from '../../api/client'
import { listGameSessions } from '../../api/sessions'
import { listGamePins, createPin, updatePin, deletePin } from '../../api/pins'
import { uploadMapImage, deleteMapImage } from '../../api/mapImage'
import { listMemberships } from '../../api/memberships'
import { useAuth } from '../../context/AuthContext'
import type { Game } from '../../types/game'
import type { Session } from '../../types/session'
import type { SessionPin } from '../../types/pin'
import type { GameMembership } from '../../types/membership'
import './MapView.css'

export default function MapView() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()

  const [game, setGame] = useState<Game | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])
  const [pins, setPins] = useState<SessionPin[]>([])
  const [memberships, setMemberships] = useState<GameMembership[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [uploading, setUploading] = useState(false)
  const [pendingCoords, setPendingCoords] = useState<{ x: number; y: number } | null>(null)
  const [hoveredPinId, setHoveredPinId] = useState<string | null>(null)
  const [dragging, setDragging] = useState<{ pinId: string; startX: number; startY: number } | null>(null)
  const [panning, setPanning] = useState<{ startX: number; startY: number; scrollLeft: number; scrollTop: number } | null>(null)
  const [zoom, setZoom] = useState(1)
  const [pinOrientation, setPinOrientation] = useState<'up' | 'down'>('down')

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const mapViewportRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const MIN_ZOOM = 1
  const MAX_ZOOM = 5
  const ZOOM_STEP = 0.25

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
    ])
      .then(([gameData, sessionsData, pinsData, membershipsData]) => {
        if (!cancelled) {
          setGame(gameData)
          setSessions(sessionsData)
          setPins(pinsData)
          setMemberships(membershipsData)
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

  /** Convert a client-space point to percentage coords on the map, accounting for zoom + scroll. */
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
    const coords = clientToMapPct(e.clientX, e.clientY)
    setPendingCoords(coords)
  }, [dragging, clientToMapPct])

  const handleSelectSession = useCallback(async (session: Session) => {
    if (!gameId || !pendingCoords) return
    try {
      const pin = await createPin({
        session_id: session.id,
        x: pendingCoords.x,
        y: pendingCoords.y,
      })
      setPins(prev => [...prev, pin])
      setPendingCoords(null)
    } catch (err: unknown) {
      console.error('Failed to create pin', err)
    }
  }, [gameId, pendingCoords])

  const handlePinPointerDown = useCallback((e: React.PointerEvent, pin: SessionPin) => {
    e.preventDefault()
    e.stopPropagation()
    ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
    setDragging({ pinId: pin.id, startX: e.clientX, startY: e.clientY })
  }, [])

  const handlePointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragging || !mapContainerRef.current) return
    const coords = clientToMapPct(e.clientX, e.clientY)
    setPins(prev => prev.map(p => p.id === dragging.pinId ? { ...p, x: coords.x, y: coords.y } : p))
  }, [dragging, clientToMapPct])

  const handlePointerUp = useCallback(async (e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragging || !mapContainerRef.current) return
    const coords = clientToMapPct(e.clientX, e.clientY)
    const pinId = dragging.pinId
    setDragging(null)
    try {
      await updatePin(pinId, { x: coords.x, y: coords.y })
    } catch (err: unknown) {
      console.error('Failed to update pin', err)
    }
  }, [dragging, clientToMapPct])

  const handleDeletePin = useCallback(async (pinId: string) => {
    try {
      await deletePin(pinId)
      setPins(prev => prev.filter(p => p.id !== pinId))
    } catch (err: unknown) {
      console.error('Failed to delete pin', err)
    }
  }, [])

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

  const handleWheel = useCallback((e: React.WheelEvent<HTMLDivElement>) => {
    if (!e.ctrlKey && !e.metaKey) return
    e.preventDefault()
    const vp = mapViewportRef.current
    if (!vp) return

    setZoom(prev => {
      const next = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round((prev + (e.deltaY < 0 ? ZOOM_STEP : -ZOOM_STEP)) * 100) / 100))
      if (next === prev) return prev

      // Point under cursor in viewport-local coords (including scroll)
      const rect = vp.getBoundingClientRect()
      const cursorX = e.clientX - rect.left + vp.scrollLeft
      const cursorY = e.clientY - rect.top + vp.scrollTop

      // The cursor points at a position in the unscaled content
      const contentX = cursorX / prev
      const contentY = cursorY / prev

      // After zoom, that same content point should stay under the cursor
      const newScrollLeft = contentX * next - (e.clientX - rect.left)
      const newScrollTop = contentY * next - (e.clientY - rect.top)

      // Apply scroll after React re-renders with the new zoom
      requestAnimationFrame(() => {
        vp.scrollLeft = Math.max(0, newScrollLeft)
        vp.scrollTop = Math.max(0, newScrollTop)
      })

      return next
    })
  }, [])

  // Also listen for native wheel to preventDefault (React passive workaround)
  useEffect(() => {
    const viewport = mapViewportRef.current
    if (!viewport) return
    const onWheel = (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) e.preventDefault()
    }
    viewport.addEventListener('wheel', onWheel, { passive: false })
    return () => viewport.removeEventListener('wheel', onWheel)
  }, [game?.map_image_url, loading])

  const zoomIn = useCallback(() => {
    setZoom(prev => Math.min(MAX_ZOOM, Math.round((prev + ZOOM_STEP) * 100) / 100))
  }, [])

  const zoomOut = useCallback(() => {
    setZoom(prev => Math.max(MIN_ZOOM, Math.round((prev - ZOOM_STEP) * 100) / 100))
  }, [])

  const zoomReset = useCallback(() => setZoom(1), [])

  // Middle-mouse-button panning
  const handleViewportPointerDown = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (e.button !== 1 || !mapViewportRef.current) return // button 1 = middle
    e.preventDefault()
    const vp = mapViewportRef.current
    vp.setPointerCapture(e.pointerId)
    setPanning({ startX: e.clientX, startY: e.clientY, scrollLeft: vp.scrollLeft, scrollTop: vp.scrollTop })
  }, [])

  const handleViewportPointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (!panning || !mapViewportRef.current) return
    const dx = e.clientX - panning.startX
    const dy = e.clientY - panning.startY
    mapViewportRef.current.scrollLeft = panning.scrollLeft - dx
    mapViewportRef.current.scrollTop = panning.scrollTop - dy
  }, [panning])

  const handleViewportPointerUp = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (!panning) return
    if (mapViewportRef.current) mapViewportRef.current.releasePointerCapture(e.pointerId)
    setPanning(null)
  }, [panning])

  // Prevent default middle-click auto-scroll icon
  useEffect(() => {
    const viewport = mapViewportRef.current
    if (!viewport) return
    const onMouseDown = (e: MouseEvent) => { if (e.button === 1) e.preventDefault() }
    viewport.addEventListener('mousedown', onMouseDown)
    return () => viewport.removeEventListener('mousedown', onMouseDown)
  }, [game?.map_image_url, loading])

  const sessionForPin = (pin: SessionPin) => sessions.find(s => s.id === pin.session_id)

  return (
    <div className="map-view-page">
      <div className="map-view-inner">
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
          <>
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
              Click anywhere on the map to place a session pin. Drag pins to reposition.
              Middle-click and drag to pan. Ctrl + scroll to zoom.
            </p>

            <div className="map-toolbar-row">
              <div className="map-pin-toggle">
                <button
                  className={`map-pin-toggle-btn${pinOrientation === 'up' ? ' map-pin-toggle-btn--active' : ''}`}
                  onClick={() => setPinOrientation('up')}
                  title="Pins point up"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M12 2 L12 16" />
                    <path d="M5 9 L12 2 L19 9" />
                  </svg>
                </button>
                <button
                  className={`map-pin-toggle-btn${pinOrientation === 'down' ? ' map-pin-toggle-btn--active' : ''}`}
                  onClick={() => setPinOrientation('down')}
                  title="Pins point down"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M12 22 L12 8" />
                    <path d="M5 15 L12 22 L19 15" />
                  </svg>
                </button>
              </div>

              <div className="map-zoom-controls">
                <button className="map-zoom-btn" onClick={zoomIn} disabled={zoom >= MAX_ZOOM} title="Zoom in">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                    <line x1="12" y1="5" x2="12" y2="19" />
                    <line x1="5" y1="12" x2="19" y2="12" />
                  </svg>
                </button>
                <button
                  className="map-zoom-level"
                  onClick={zoomReset}
                  title="Reset zoom"
                >
                  {Math.round(zoom * 100)}%
                </button>
                <button className="map-zoom-btn" onClick={zoomOut} disabled={zoom <= MIN_ZOOM} title="Zoom out">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                    <line x1="5" y1="12" x2="19" y2="12" />
                  </svg>
                </button>
              </div>
            </div>

            <div
              className={`map-viewport${panning ? ' map-viewport--panning' : ''}`}
              ref={mapViewportRef}
              onWheel={handleWheel}
              onPointerDown={handleViewportPointerDown}
              onPointerMove={handleViewportPointerMove}
              onPointerUp={handleViewportPointerUp}
            >
              <div
                className="map-container map-container--interactive"
                ref={mapContainerRef}
                style={{ transform: `scale(${zoom})`, transformOrigin: 'top left', width: '100%' }}
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

                {pins.map(pin => {
                  const session = sessionForPin(pin)
                  return (
                    <div
                      key={pin.id}
                      className={`map-pin-wrapper${pinOrientation === 'down' ? ' map-pin-wrapper--down' : ''}${hoveredPinId === pin.id ? ' map-pin-wrapper--hovered' : ''}${dragging?.pinId === pin.id ? ' map-pin-wrapper--dragging' : ''}`}
                      style={{ left: `${pin.x}%`, top: `${pin.y}%` }}
                      onMouseEnter={() => setHoveredPinId(pin.id)}
                      onMouseLeave={() => setHoveredPinId(null)}
                    >
                      <button
                        className="map-pin"
                        title={session?.title ?? 'Session'}
                        onClick={e => {
                          e.stopPropagation()
                          if (!dragging) navigate(`/games/${gameId}/sessions/${pin.session_id}/notes`)
                        }}
                        onPointerDown={e => handlePinPointerDown(e, pin)}
                      />
                      <span className="map-pin__label">
                        {session?.session_number != null && (
                          <span className="map-pin__label-num">#{session.session_number}</span>
                        )}
                        {session?.title ?? '?'}
                      </span>
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
                    </div>
                  )
                })}
              </div>
            </div>

            {unpinnedSessions.length === 0 && pins.length > 0 && (
              <p className="map-all-pinned">✦ All sessions are pinned on the map.</p>
            )}
          </>
        )}
      </div>

      {/* Session Picker Modal */}
      {pendingCoords && (
        <div className="map-overlay" onClick={() => setPendingCoords(null)}>
          <div className="map-session-picker" onClick={e => e.stopPropagation()}>
            <div className="map-picker-header">
              <span className="map-picker-rune" aria-hidden="true">⬡</span>
              <h3 className="map-picker-title">Mark This Location</h3>
            </div>
            <p className="map-picker-sub">Choose the session to pin here:</p>
            {unpinnedSessions.length === 0 ? (
              <p className="map-picker-empty">All sessions already have pins on the map.</p>
            ) : (
              <ul className="map-picker-list">
                {unpinnedSessions.map(session => (
                  <li key={session.id}>
                    <button
                      className="map-picker-item"
                      onClick={() => handleSelectSession(session)}
                    >
                      {session.session_number != null && (
                        <span className="map-picker-num">#{session.session_number}</span>
                      )}
                      <span className="map-picker-name">{session.title}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
            <button className="map-picker-cancel" onClick={() => setPendingCoords(null)}>
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

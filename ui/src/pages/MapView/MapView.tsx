import { useState, useEffect, useRef, useCallback } from 'react'
import { createPortal } from 'react-dom'
import { useParams, useNavigate } from 'react-router-dom'
import { TransformWrapper, TransformComponent } from 'react-zoom-pan-pinch'
import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import { apiFetch, BASE_URL } from '../../api/client'
import { listGameSessions } from '../../api/sessions'
import { listGamePins, createPin, updatePin, deletePin } from '../../api/pins'
import { uploadMapImage, deleteMapImage } from '../../api/mapImage'
import { listMemberships } from '../../api/memberships'
import { useAuth } from '../../context/AuthContext'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import type { Game } from '../../types/game'
import type { Session } from '../../types/session'
import type { SessionPin } from '../../types/pin'
import type { GameMembership } from '../../types/membership'
import './MapView.css'

interface MapViewState {
  scale: number
  positionX: number
  positionY: number
  defaultPinOrientation: 'up' | 'down'
}

const DEFAULT_VIEW_STATE: MapViewState = {
  scale: 1,
  positionX: 0,
  positionY: 0,
  defaultPinOrientation: 'down',
}

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
  const [viewState, setViewState] = useLocalStorage<MapViewState>(
    `pf2e-map-view-${gameId}`,
    DEFAULT_VIEW_STATE,
  )
  const [panelOpen, setPanelOpen] = useState(true)
  const [displayScale, setDisplayScale] = useState(viewState.scale)

  const mapContainerRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const transformRef = useRef<ReactZoomPanPinchRef>(null)

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
    const coords = clientToMapPct(e.clientX, e.clientY)
    setPendingCoords(coords)
  }, [dragging, clientToMapPct])

  const handleFlipPin = useCallback(async (pin: SessionPin) => {
    const newTypeId = pin.pin_type.name === 'down' ? 1 : 2 // 1=up, 2=down
    setPins(prev => prev.map(p => p.id === pin.id ? { ...p, pin_type_id: newTypeId, pin_type: { id: newTypeId, name: newTypeId === 1 ? 'up' : 'down' } } : p))
    try {
      await updatePin(pin.id, { pin_type_id: newTypeId })
    } catch (err: unknown) {
      console.error('Failed to flip pin', err)
      setPins(prev => prev.map(p => p.id === pin.id ? pin : p))
    }
  }, [])

  const handleSelectSession = useCallback(async (session: Session) => {
    if (!gameId || !pendingCoords) return
    try {
      const pin = await createPin({
        session_id: session.id,
        x: pendingCoords.x,
        y: pendingCoords.y,
        pin_type_id: viewState.defaultPinOrientation === 'up' ? 1 : 2,
      })
      setPins(prev => [...prev, pin])
      setPendingCoords(null)
    } catch (err: unknown) {
      console.error('Failed to create pin', err)
    }
  }, [gameId, pendingCoords, viewState])

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

  const sessionForPin = (pin: SessionPin) => sessions.find(s => s.id === pin.session_id)

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
          <div className="map-viewport-container">
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

                  {pins.map(pin => {
                    const session = sessionForPin(pin)
                    const isDown = pin.pin_type?.name === 'down'
                    return (
                      <div
                        key={pin.id}
                        className={`map-pin-wrapper${isDown ? ' map-pin-wrapper--down' : ''}${hoveredPinId === pin.id ? ' map-pin-wrapper--hovered' : ''}${dragging?.pinId === pin.id ? ' map-pin-wrapper--dragging' : ''}`}
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
                          className="map-pin__flip"
                          title={isDown ? 'Flip pin up' : 'Flip pin down'}
                          onClick={e => { e.stopPropagation(); handleFlipPin(pin) }}
                        >
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d={isDown ? 'M12 19V5M5 12l7-7 7 7' : 'M12 5v14M5 12l7 7 7-7'} />
                          </svg>
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
                      </div>
                    )
                  })}
                </div>
              </TransformComponent>
            </TransformWrapper>

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
                  Click anywhere on the map to place a session pin. Drag pins to reposition.
                  Middle-click and drag to pan. Ctrl + scroll to zoom.
                </p>

                <div className="map-pin-toggle">
                  <span className="map-pin-toggle-label">New pins:</span>
                  <button
                    className={`map-pin-toggle-btn${viewState.defaultPinOrientation === 'up' ? ' map-pin-toggle-btn--active' : ''}`}
                    onClick={() => setViewState(prev => ({ ...prev, defaultPinOrientation: 'up' }))}
                    title="New pins point up"
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M12 2 L12 16" />
                      <path d="M5 9 L12 2 L19 9" />
                    </svg>
                  </button>
                  <button
                    className={`map-pin-toggle-btn${viewState.defaultPinOrientation === 'down' ? ' map-pin-toggle-btn--active' : ''}`}
                    onClick={() => setViewState(prev => ({ ...prev, defaultPinOrientation: 'down' }))}
                    title="New pins point down"
                  >
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M12 22 L12 8" />
                      <path d="M5 15 L12 22 L19 15" />
                    </svg>
                  </button>
                </div>

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
          </div>
        )}
      </div>

      {/* Session Picker Modal — portalled to body to avoid transform containing block issues */}
      {pendingCoords && createPortal(
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
        </div>,
        document.body,
      )}
    </div>
  )
}

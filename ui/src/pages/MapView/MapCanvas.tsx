import React from 'react'
import { TransformWrapper, TransformComponent } from 'react-zoom-pan-pinch'
import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import { COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_COLOURS, PIN_ICONS, PIN_ICON_LABELS } from '../../constants/pins'
import type { PinColour } from '../../constants/pins'
import type { SessionPin, PinGroup } from '../../types/pin'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'
import type { MapViewState } from './useMapViewData'
import './MapCanvas.css'

interface PinFieldUpdate {
  colour?: string
  icon?: string
  label?: string
  description?: string | null
  session_id?: string | null
  note_id?: string | null
}

interface MapCanvasProps {
  activeMapId: string
  imageUrl: string
  viewState: MapViewState
  displayScale: number
  pins: SessionPin[]
  pinGroups: PinGroup[]
  sessions: Session[]
  notes: Note[]
  hoveredPinId: string | null
  flashPinId: string | null
  dragging: { pinId: string; startX: number; startY: number } | null
  editingPinId: string | null
  editLinkSearch: string
  dropTargetIds: Set<string>
  activeGroupId: string | null
  pinError: string | null
  // Refs
  mapContainerRef: React.RefObject<HTMLDivElement | null>
  viewportContainerRef: React.RefObject<HTMLDivElement | null>
  transformRef: React.RefObject<ReactZoomPanPinchRef | null>
  wasDragRef: React.RefObject<boolean>
  // Handlers
  onTransformed: (ref: ReactZoomPanPinchRef, state: { scale: number; positionX: number; positionY: number }) => void
  onTransformEnd: (ref: ReactZoomPanPinchRef) => void
  onZoom: (ref: ReactZoomPanPinchRef) => void
  onImageLoad: () => void
  onMapClick: (e: React.MouseEvent<HTMLDivElement>) => void
  onPointerMove: (e: React.PointerEvent<HTMLDivElement>) => void
  onPointerUp: (e: React.PointerEvent<HTMLDivElement>) => void
  onPinPointerDown: (e: React.PointerEvent<HTMLButtonElement>, pin: SessionPin) => void
  onHoverPin: (id: string | null) => void
  onEditPin: (id: string | null) => void
  onDeletePin: (id: string) => void
  onEditPinField: (pinId: string, field: PinFieldUpdate) => void
  onEditLinkSearchChange: (value: string) => void
  onPinClick: (pin: SessionPin) => void
  onGroupClick: (groupId: string) => void
  onManageGroup: (groupId: string) => void
  onPinErrorDismiss: () => void
  openItem: (type: 'session' | 'note', itemId: string, label: string) => void
  sessionForPin: (pin: SessionPin) => Session | undefined
  noteForPin: (pin: SessionPin) => Note | undefined
  isGM: boolean
}

export default function MapCanvas(props: MapCanvasProps) {
  const {
    viewState,
    pins,
    pinGroups,
    sessions,
    notes,
    hoveredPinId,
    dragging,
    editingPinId,
    editLinkSearch,
    dropTargetIds,
    activeGroupId,
    pinError,
    mapContainerRef,
    viewportContainerRef,
    transformRef,
    wasDragRef,
    onTransformed,
    onTransformEnd,
    onZoom,
    onImageLoad,
    onMapClick,
    onPointerMove,
    onPointerUp,
    onPinPointerDown,
    onHoverPin,
    onEditPin,
    onDeletePin,
    onEditPinField,
    onEditLinkSearchChange,
    onGroupClick,
    onManageGroup,
    onPinErrorDismiss,
    openItem,
    sessionForPin,
    noteForPin,
    imageUrl,
  } = props

  return (
    <>
      {pinError && (
        <div className="map-pin-error-banner" onClick={onPinErrorDismiss}>
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
          allowMiddleClickPan: false,
          allowRightClickPan: true,
          velocityDisabled: true,
        }}
        wheel={{
          step: 0.15,
        }}
        doubleClick={{ disabled: true }}
        onTransformed={onTransformed}
        onPanningStop={onTransformEnd}
        onZoomStop={onTransformEnd}
        onZoom={onZoom}
      >
        <TransformComponent
          wrapperClass="map-viewport"
          contentClass="map-container map-container--interactive"
        >
          <div
            ref={mapContainerRef}
            style={{ width: '100%', position: 'relative' }}
            onClick={onMapClick}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
          >
            <img
              className="map-img"
              src={imageUrl}
              alt="Campaign map"
              draggable={false}
              onLoad={onImageLoad}
            />

            {pins.filter(p => p.group_id === null).map(pin => {
              const session = sessionForPin(pin)
              const note = noteForPin(pin)
              const pinColour = (pin.colour as PinColour) ?? 'grey'
              const pinLabel = note?.title ?? session?.title ?? (pin.label || '')
              return (
                <div
                  key={pin.id}
                  className={`map-pin-wrapper${hoveredPinId === pin.id ? ' map-pin-wrapper--hovered' : ''}${props.flashPinId === pin.id ? ' map-pin-wrapper--flash' : ''}${dragging?.pinId === pin.id ? ' map-pin-wrapper--dragging' : ''}${dropTargetIds.has(pin.id) ? ' map-pin-wrapper--drop-target' : ''}`}
                  style={{ left: `${pin.x}%`, top: `${pin.y}%` }}
                  onMouseEnter={() => onHoverPin(pin.id)}
                  onMouseLeave={() => onHoverPin(null)}
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
                          onEditPin(editingPinId === pin.id ? null : pin.id)
                          onEditLinkSearchChange('')
                        }
                      }
                    }}
                    onPointerDown={e => onPinPointerDown(e, pin)}
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
                          onEditPin(editingPinId === pin.id ? null : pin.id)
                          onEditLinkSearchChange('')
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
                      onEditPin(editingPinId === pin.id ? null : pin.id)
                      onEditLinkSearchChange('')
                    }}
                  >
                    ✎
                  </button>
                  <button
                    className="map-pin__delete"
                    title="Remove pin"
                    onClick={e => { e.stopPropagation(); onDeletePin(pin.id) }}
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
                          if (val !== (pin.label ?? '')) onEditPinField(pin.id, { label: val })
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
                          if (val !== (pin.description ?? '')) onEditPinField(pin.id, { description: val || null })
                        }}
                      />
                      <span className="map-pin-edit-popover-label">Colour</span>
                      <div className="map-pin-edit-popover-colours">
                        {PIN_COLOURS.map(c => (
                          <button
                            key={c}
                            className={`pin-colour-swatch${pinColour === c ? ' pin-colour-swatch--selected' : ''}`}
                            style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                            onClick={() => onEditPinField(pin.id, { colour: c })}
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
                              onClick={() => onEditPinField(pin.id, { icon: i })}
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
                          <button className="map-pin-unlink-btn" onClick={() => onEditPinField(pin.id, { session_id: null })} title="Unlink session">Unlink</button>
                        </div>
                      ) : pin.note_id ? (
                        <div className="map-pin-link-row">
                          <span className="map-pin-link-name">{noteForPin(pin)?.title ?? pin.note_id}</span>
                          <button className="map-pin-unlink-btn" onClick={() => onEditPinField(pin.id, { note_id: null })} title="Unlink note">Unlink</button>
                        </div>
                      ) : (
                        <div className="map-pin-link-search">
                          <input
                            className="map-marker-input"
                            type="text"
                            placeholder="Search to link…"
                            value={editLinkSearch}
                            onChange={e => onEditLinkSearchChange(e.target.value)}
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
                                    <button className="map-picker-item" onClick={() => { onEditPinField(pin.id, { session_id: s.id }); onEditLinkSearchChange('') }}>
                                      <span className="map-picker-item-type">Session</span>
                                      {s.session_number != null && <span className="map-picker-num">#{s.session_number}</span>}
                                      <span className="map-picker-name">{s.title}</span>
                                    </button>
                                  </li>
                                ))}
                                {matchedNotes.slice(0, 3).map(n => (
                                  <li key={n.id}>
                                    <button className="map-picker-item" onClick={() => { onEditPinField(pin.id, { note_id: n.id }); onEditLinkSearchChange('') }}>
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
                      onGroupClick(group.id)
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
              <button onClick={() => onManageGroup(group.id)}>Manage</button>
            </div>
            <ul className="map-pin-group-popover-list">
              {group.pins.map(p => {
                const s = sessions.find(sess => sess.id === p.session_id)
                const n = notes.find(nt => nt.id === p.note_id)
                return (
                  <li key={p.id}>
                    <button onClick={() => {
                      if (p.note_id) openItem('note', p.note_id, notes.find(nt => nt.id === p.note_id)?.title ?? p.label ?? '?')
                      else if (p.session_id) openItem('session', p.session_id, sessions.find(sess => sess.id === p.session_id)?.title ?? '?')
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
    </>
  )
}

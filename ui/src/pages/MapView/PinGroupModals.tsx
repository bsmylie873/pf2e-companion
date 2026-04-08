import React from 'react'
import { createPortal } from 'react-dom'
import { PIN_COLOURS, PIN_ICONS, COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS } from '../../constants/pins'
import type { PinColour } from '../../constants/pins'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'
import type { SessionPin, PinGroup } from '../../types/pin'
import { createMapPinGroup, addPinToGroup, removePinFromGroup, disbandPinGroup, updatePinGroup } from '../../api/pinGroups'
import { listMapPins } from '../../api/pins'
import './PinGroupModals.css'

const GROUP_PROXIMITY_PCT = 1.5

interface PinGroupModalsProps {
  gameId: string
  activeMapId: string | null
  // Grouping prompt
  groupingPrompt: { coords: { x: number; y: number }; nearbyPins: SessionPin[]; nearbyGroups: PinGroup[] } | null
  onDismissGroupingPrompt: () => void
  onPlaceStandalone: (coords: { x: number; y: number }) => void
  onCreateGroupFromPrompt: (coords: { x: number; y: number }, pinIds: string[]) => void
  onAddToGroupFromPrompt: (coords: { x: number; y: number }, groupId: string) => void
  // Drag group prompt
  dragGroupPrompt: { draggedPinId: string; nearbyPins: SessionPin[]; nearbyGroups: PinGroup[]; originalCoords: { x: number; y: number } } | null
  onDismissDragGroupPrompt: () => void
  // Manage group
  managingGroupId: string | null
  pinGroups: PinGroup[]
  pins: SessionPin[]
  sessions: Session[]
  notes: Note[]
  onDismissManageGroup: () => void
  onReloadPinGroups: () => Promise<void>
  onUpdatePins: (updater: (prev: SessionPin[]) => SessionPin[]) => void
}

export default function PinGroupModals({
  gameId,
  activeMapId,
  groupingPrompt,
  onDismissGroupingPrompt,
  onPlaceStandalone,
  onCreateGroupFromPrompt,
  onAddToGroupFromPrompt,
  dragGroupPrompt,
  onDismissDragGroupPrompt,
  managingGroupId,
  pinGroups,
  pins,
  sessions,
  notes,
  onDismissManageGroup,
  onReloadPinGroups,
  onUpdatePins,
}: PinGroupModalsProps) {
  const manageGroup = managingGroupId ? pinGroups.find(g => g.id === managingGroupId) : null

  return (
    <>
      {groupingPrompt && createPortal(
        <div className='map-overlay' onClick={onDismissGroupingPrompt}>
          <div className='map-grouping-prompt' onClick={e => e.stopPropagation()}>
            <div className='map-picker-header'>
              <span className='map-picker-rune' aria-hidden='true'>⬡</span>
              <h3 className='map-picker-title'>Nearby Markers Detected</h3>
            </div>
            <p className='map-picker-sub'>There are pins or groups nearby. How would you like to place this pin?</p>
            <ul className='map-picker-list'>
              <li>
                <button className='map-picker-item' onClick={() => {
                  onPlaceStandalone(groupingPrompt.coords)
                  onDismissGroupingPrompt()
                }}>Place as standalone pin</button>
              </li>
              {groupingPrompt.nearbyPins.length > 0 && (
                <li>
                  <button className='map-picker-item' onClick={() => {
                    onCreateGroupFromPrompt(groupingPrompt.coords, groupingPrompt.nearbyPins.map(p => p.id))
                    onDismissGroupingPrompt()
                  }}>Create new group with {groupingPrompt.nearbyPins.length} nearby pin(s)</button>
                </li>
              )}
              {groupingPrompt.nearbyGroups.map(g => (
                <li key={g.id}>
                  <button className='map-picker-item' onClick={() => {
                    onAddToGroupFromPrompt(groupingPrompt.coords, g.id)
                    onDismissGroupingPrompt()
                  }}>Add to group ({g.pin_count} pins)</button>
                </li>
              ))}
            </ul>
            <button className='map-picker-cancel' onClick={onDismissGroupingPrompt}>Cancel</button>
          </div>
        </div>,
        document.body,
      )}

      {dragGroupPrompt && createPortal(
        <div className='map-overlay' onClick={onDismissDragGroupPrompt}>
          <div className='map-grouping-prompt' onClick={e => e.stopPropagation()}>
            <div className='map-picker-header'>
              <span className='map-picker-rune' aria-hidden='true'>⬡</span>
              <h3 className='map-picker-title'>Group Pins</h3>
            </div>
            <p className='map-picker-sub'>You dropped this pin near other markers. Would you like to group them?</p>
            <ul className='map-picker-list'>
              {dragGroupPrompt.nearbyPins.length > 0 && (
                <li>
                  <button className='map-picker-item' onClick={async () => {
                    if (!gameId || !activeMapId) return
                    try {
                      await createMapPinGroup(gameId, activeMapId, [...dragGroupPrompt.nearbyPins.map(p => p.id), dragGroupPrompt.draggedPinId])
                      await onReloadPinGroups()
                      const updatedPins = await listMapPins(gameId, activeMapId)
                      onUpdatePins(() => updatedPins)
                    } catch (err: unknown) {
                      console.error('Failed to create group', err)
                    }
                    onDismissDragGroupPrompt()
                  }}>
                    <span className='map-picker-name'>
                      Create new group with {dragGroupPrompt.nearbyPins.length} nearby pin{dragGroupPrompt.nearbyPins.length !== 1 ? 's' : ''}
                    </span>
                  </button>
                </li>
              )}
              {dragGroupPrompt.nearbyGroups.map(g => (
                <li key={g.id}>
                  <button className='map-picker-item' onClick={async () => {
                    try {
                      await addPinToGroup(g.id, dragGroupPrompt.draggedPinId)
                      await onReloadPinGroups()
                      if (gameId && activeMapId) {
                        const updatedPins = await listMapPins(gameId, activeMapId)
                        onUpdatePins(() => updatedPins)
                      }
                    } catch (err: unknown) {
                      console.error('Failed to add pin to group', err)
                    }
                    onDismissDragGroupPrompt()
                  }}>
                    <span className='map-picker-name'>Add to existing group ({g.pin_count} pins)</span>
                  </button>
                </li>
              ))}
              <li>
                <button className='map-picker-item' onClick={() => {
                  onDismissDragGroupPrompt()
                }}>
                  <span className='map-picker-name'>Cancel — keep pin in place</span>
                </button>
              </li>
            </ul>
          </div>
        </div>,
        document.body,
      )}

      {managingGroupId && manageGroup && (() => {
        const group = manageGroup
        const mgmtColour = (group.colour as PinColour) ?? 'grey'
        const nearbyStandalonePins = pins.filter(p => {
          if (p.group_id !== null) return false
          const dx = p.x - group.x
          const dy = p.y - group.y
          return Math.sqrt(dx * dx + dy * dy) <= GROUP_PROXIMITY_PCT * 4
        })
        return createPortal(
          <div className='map-overlay' onClick={onDismissManageGroup}>
            <div className='map-pin-group-manage' onClick={e => e.stopPropagation()}>
              <div className='map-picker-header'>
                <span className='map-picker-rune' aria-hidden='true'>⬡</span>
                <h3 className='map-picker-title'>Manage Group</h3>
              </div>

              <span className='map-picker-customise-label'>Group colour:</span>
              <div className='pin-colour-palette'>
                {PIN_COLOURS.map(c => (
                  <button
                    key={c}
                    className={`pin-colour-swatch${mgmtColour === c ? ' pin-colour-swatch--selected' : ''}`}
                    style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                    onClick={async () => { await updatePinGroup(group.id, { colour: c }); await onReloadPinGroups() }}
                    title={c}
                  />
                ))}
              </div>

              <span className='map-picker-customise-label'>Group icon:</span>
              <div className='pin-colour-palette'>
                {PIN_ICONS.map(i => {
                  const MgmtIconComp = PIN_ICON_COMPONENTS[i]
                  return (
                    <button
                      key={i}
                      className={`pin-icon-option${group.icon === i ? ' pin-icon-option--selected' : ''}`}
                      onClick={async () => { await updatePinGroup(group.id, { icon: i }); await onReloadPinGroups() }}
                      title={PIN_ICON_LABELS[i]}
                      aria-label={PIN_ICON_LABELS[i]}
                    >
                      <MgmtIconComp size={14} />
                    </button>
                  )
                })}
              </div>

              <span className='map-picker-customise-label'>Members ({group.pin_count}):</span>
              <ul className='map-pin-group-popover-list'>
                {group.pins.map(p => {
                  const s = sessions.find(sess => sess.id === p.session_id)
                  const n = notes.find(nt => nt.id === p.note_id)
                  return (
                    <li key={p.id} className='map-pin-group-member-row'>
                      <span>{n?.title ?? s?.title ?? p.label ?? '?'}</span>
                      <button
                        className='map-pin-group-remove-btn'
                        onClick={async () => {
                          await removePinFromGroup(group.id, p.id)
                          await onReloadPinGroups()
                          onUpdatePins(prev => prev.map(pin => pin.id === p.id ? { ...pin, group_id: null } : pin))
                        }}
                        title='Remove from group'
                      >
                        ✕
                      </button>
                    </li>
                  )
                })}
              </ul>

              {nearbyStandalonePins.length > 0 && (
                <>
                  <span className='map-picker-customise-label'>Add nearby pin:</span>
                  <ul className='map-picker-list'>
                    {nearbyStandalonePins.map(p => {
                      const s = sessions.find(sess => sess.id === p.session_id)
                      const n = notes.find(nt => nt.id === p.note_id)
                      return (
                        <li key={p.id}>
                          <button
                            className='map-picker-item'
                            onClick={async () => {
                              await addPinToGroup(group.id, p.id)
                              await onReloadPinGroups()
                              onUpdatePins(prev => prev.map(pin => pin.id === p.id ? { ...pin, group_id: group.id } : pin))
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
                className='map-delete-btn'
                style={{ marginTop: '0.5rem' }}
                onClick={async () => {
                  if (!confirm('Disband this group? All pins will become standalone.')) return
                  await disbandPinGroup(group.id)
                  const memberIds = new Set(group.pins.map(p => p.id))
                  onUpdatePins(prev => prev.map(p => memberIds.has(p.id) ? { ...p, group_id: null } : p))
                  await onReloadPinGroups()
                  onDismissManageGroup()
                }}
              >
                Disband Group
              </button>

              <button className='map-picker-cancel' onClick={onDismissManageGroup}>Close</button>
            </div>
          </div>,
          document.body,
        )
      })()}
    </>
  )
}

import { useState, useRef } from 'react'
import type { GameMap } from '../../types/map'
import './MapSelector.css'

interface MapSelectorProps {
  maps: GameMap[]
  activeMapId: string | null
  onSelect: (mapId: string) => void
  isGM: boolean
  onCreateMap: (name: string) => void
  onRenameMap: (mapId: string, name: string) => void
  onArchiveMap: (mapId: string) => void
  onUnarchiveMap: (mapId: string) => void
  onReorderMaps: (ids: string[]) => void
  archivedMaps: GameMap[]
}

function formatElapsed(archivedAt: string | null): string {
  if (!archivedAt) return 'Archived'
  const ms = Date.now() - new Date(archivedAt).getTime()
  const mins = Math.floor(ms / 60000)
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export default function MapSelector({
  maps,
  activeMapId,
  onSelect,
  isGM,
  onCreateMap,
  onRenameMap,
  onArchiveMap,
  onUnarchiveMap,
  onReorderMaps,
  archivedMaps,
}: MapSelectorProps) {
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [creatingMap, setCreatingMap] = useState(false)
  const [newMapName, setNewMapName] = useState('')
  const [archivedOpen, setArchivedOpen] = useState(false)
  const [dragOverId, setDragOverId] = useState<string | null>(null)
  const dragSrcId = useRef<string | null>(null)

  function handleRenameStart(map: GameMap) {
    setRenamingId(map.id)
    setRenameValue(map.name)
  }

  function handleRenameSubmit(mapId: string) {
    if (renameValue.trim()) onRenameMap(mapId, renameValue.trim())
    setRenamingId(null)
  }

  function handleCreateSubmit() {
    if (newMapName.trim()) onCreateMap(newMapName.trim())
    setNewMapName('')
    setCreatingMap(false)
  }

  function handleDragStart(e: React.DragEvent, mapId: string) {
    dragSrcId.current = mapId
    e.dataTransfer.effectAllowed = 'move'
  }

  function handleDragOver(e: React.DragEvent, mapId: string) {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverId(mapId)
  }

  function handleDragLeave() {
    setDragOverId(null)
  }

  function handleDrop(e: React.DragEvent, targetId: string) {
    e.preventDefault()
    setDragOverId(null)
    const srcId = dragSrcId.current
    if (!srcId || srcId === targetId) return
    const ids = maps.map(m => m.id)
    const srcIdx = ids.indexOf(srcId)
    const tgtIdx = ids.indexOf(targetId)
    ids.splice(srcIdx, 1)
    ids.splice(tgtIdx, 0, srcId)
    onReorderMaps(ids)
    dragSrcId.current = null
  }

  function handleDragEnd() {
    setDragOverId(null)
    dragSrcId.current = null
  }

  return (
    <div className="map-selector">
      <div className="map-selector-tabs">
        {maps.map(map => (
          <div
            key={map.id}
            className={[
              'map-tab',
              map.id === activeMapId ? 'map-tab--active' : '',
              dragOverId === map.id ? 'map-tab--drag-over' : '',
            ].filter(Boolean).join(' ')}
            draggable={isGM}
            onDragStart={isGM ? (e) => handleDragStart(e, map.id) : undefined}
            onDragOver={isGM ? (e) => handleDragOver(e, map.id) : undefined}
            onDragLeave={isGM ? handleDragLeave : undefined}
            onDrop={isGM ? (e) => handleDrop(e, map.id) : undefined}
            onDragEnd={isGM ? handleDragEnd : undefined}
          >
            {renamingId === map.id ? (
              <input
                className="map-tab-rename-input"
                value={renameValue}
                autoFocus
                onChange={e => setRenameValue(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') handleRenameSubmit(map.id)
                  if (e.key === 'Escape') setRenamingId(null)
                }}
                onBlur={() => handleRenameSubmit(map.id)}
                onClick={e => e.stopPropagation()}
              />
            ) : (
              <button
                className="map-tab-btn"
                onClick={() => onSelect(map.id)}
              >
                {map.name}
              </button>
            )}
            {isGM && renamingId !== map.id && (
              <div className="map-tab-actions">
                <button
                  className="map-tab-action-btn"
                  title="Rename map"
                  onClick={e => { e.stopPropagation(); handleRenameStart(map) }}
                  aria-label="Rename map"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" width="13" height="13">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                  </svg>
                </button>
                <button
                  className="map-tab-action-btn map-tab-action-btn--danger"
                  title="Archive map"
                  onClick={e => { e.stopPropagation(); onArchiveMap(map.id) }}
                  aria-label="Archive map"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" width="13" height="13">
                    <polyline points="21 8 21 21 3 21 3 8" />
                    <rect x="1" y="3" width="22" height="5" />
                    <line x1="10" y1="12" x2="14" y2="12" />
                  </svg>
                </button>
              </div>
            )}
          </div>
        ))}

        {isGM && (
          creatingMap ? (
            <div className="map-tab map-tab--new">
              <input
                className="map-tab-rename-input"
                value={newMapName}
                placeholder="Map name…"
                autoFocus
                onChange={e => setNewMapName(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') handleCreateSubmit()
                  if (e.key === 'Escape') { setCreatingMap(false); setNewMapName('') }
                }}
                onBlur={() => { if (!newMapName.trim()) setCreatingMap(false) }}
              />
            </div>
          ) : (
            <button
              className="map-selector-add-btn"
              onClick={() => setCreatingMap(true)}
              title="Create new map"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" width="13" height="13">
                <line x1="12" y1="5" x2="12" y2="19" />
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
              New Map
            </button>
          )
        )}
      </div>

      {isGM && archivedMaps.length > 0 && (
        <div className="map-archived-section">
          <button
            className="map-archived-toggle"
            onClick={() => setArchivedOpen(o => !o)}
          >
            <span className="map-archived-ornament">✦</span>
            {archivedOpen ? 'Hide' : 'Show'} Archived Maps ({archivedMaps.length})
            <span className="map-archived-ornament">✦</span>
          </button>
          {archivedOpen && (
            <div className="map-archived-list">
              {archivedMaps.map(map => (
                <div key={map.id} className="map-archived-item">
                  <span className="map-archived-name">{map.name}</span>
                  <span className="map-archived-time">{formatElapsed(map.archived_at)}</span>
                  <button
                    className="map-archived-restore-btn"
                    onClick={() => onUnarchiveMap(map.id)}
                  >
                    Restore
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

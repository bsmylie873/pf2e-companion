import { useState, useCallback, useRef, useEffect } from 'react'
import { createPortal } from 'react-dom'
import type { Folder } from '../../types/folder'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'
import FolderItem from './FolderItem'
import InlineNameInput from './InlineNameInput'

interface Props {
  title: string
  folders: Folder[]
  items: (Session | Note)[]
  isReadOnly: boolean
  expandedState: Record<string, boolean>
  onToggleFolder: (folderId: string) => void
  onCreateFolder: (name: string) => Promise<string | null>
  onRenameFolder: (folderId: string, name: string) => Promise<string | null>
  onDeleteFolder: (folderId: string) => void
  onReorderFolders: (orderedIds: string[]) => void
  onItemClick: (itemId: string) => void
  onAssignItem: (itemId: string, folderId: string | null) => void
  onCreateItem: (folderId: string | null) => void
  itemType: 'session' | 'note'
}

export default function FolderSection({
  title, folders, items, isReadOnly, expandedState,
  onToggleFolder, onCreateFolder, onRenameFolder, onDeleteFolder,
  onReorderFolders, onItemClick, onAssignItem, onCreateItem, itemType,
}: Props) {
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [dragOverUnfiled, setDragOverUnfiled] = useState(false)
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; folderId: string | null } | null>(null)
  const contextMenuRef = useRef<HTMLDivElement>(null)

  // Close context menu on outside click or scroll
  useEffect(() => {
    if (!contextMenu) return
    const close = () => setContextMenu(null)
    document.addEventListener('mousedown', close)
    document.addEventListener('scroll', close, true)
    return () => {
      document.removeEventListener('mousedown', close)
      document.removeEventListener('scroll', close, true)
    }
  }, [contextMenu])

  const handleContextMenu = (e: React.MouseEvent, folderId: string | null) => {
    e.preventDefault()
    e.stopPropagation()
    setContextMenu({ x: e.clientX, y: e.clientY, folderId })
  }

  const unfiledKey = `unfiled-${itemType}`
  const isUnfiledExpanded = expandedState[unfiledKey] !== false

  const itemsInFolder = useCallback((folderId: string | null) =>
    items.filter(item => ('folder_id' in item ? item.folder_id : null) === folderId),
  [items])

  const unfiledItems = itemsInFolder(null)

  const handleCreate = async (name: string) => {
    setCreateError(null)
    const err = await onCreateFolder(name)
    if (err) {
      setCreateError(err)
    } else {
      setCreating(false)
    }
  }

  const handleFolderDragStart = (e: React.DragEvent, folderId: string) => {
    e.dataTransfer.setData('dragType', 'folder-reorder')
    e.dataTransfer.setData('folderId', folderId)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleFolderDragOver = (e: React.DragEvent) => {
    if (e.dataTransfer.types.includes('dragtype') || e.dataTransfer.types.includes('dragType')) {
      e.preventDefault()
    }
  }

  const handleFolderDrop = (e: React.DragEvent, targetIdx: number) => {
    const dragType = e.dataTransfer.getData('dragType')
    if (dragType !== 'folder-reorder') return
    e.preventDefault()
    e.stopPropagation()
    const draggedId = e.dataTransfer.getData('folderId')
    if (!draggedId) return

    const currentOrder = folders.map(f => f.id)
    const fromIdx = currentOrder.indexOf(draggedId)
    if (fromIdx === -1 || fromIdx === targetIdx) return

    const newOrder = [...currentOrder]
    newOrder.splice(fromIdx, 1)
    newOrder.splice(targetIdx, 0, draggedId)
    onReorderFolders(newOrder)
  }

  const handleItemDragStart = (e: React.DragEvent, id: string) => {
    e.stopPropagation()
    e.dataTransfer.setData('itemType', itemType)
    e.dataTransfer.setData('itemId', id)
    e.dataTransfer.effectAllowed = 'move'
  }

  const handleUnfiledDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOverUnfiled(true)
  }

  const handleUnfiledDragLeave = () => setDragOverUnfiled(false)

  const handleUnfiledDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOverUnfiled(false)
    const droppedItemType = e.dataTransfer.getData('itemType')
    const itemId = e.dataTransfer.getData('itemId')
    if (itemId && droppedItemType === itemType) {
      onAssignItem(itemId, null)
    }
  }

  const getItemTitle = (item: Session | Note): string => {
    return 'title' in item ? item.title : ''
  }

  const itemLabel = itemType === 'session' ? 'Session' : 'Note'

  return (
    <div className="folder-section" onContextMenu={e => handleContextMenu(e, null)}>
      <div className="folder-section-header">
        <span className="folder-section-title">{title}</span>
        {!isReadOnly && (
          <button
            className="folder-section-add-btn"
            onClick={() => setCreating(true)}
            title="New folder"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
          </button>
        )}
      </div>

      {creating && (
        <div className="folder-new-input">
          <span className="folder-icon" aria-hidden>&#128193;</span>
          <InlineNameInput
            value=""
            onCommit={handleCreate}
            onCancel={() => { setCreating(false); setCreateError(null) }}
            error={createError}
            placeholder="Folder name"
            autoFocus
          />
        </div>
      )}

      {/* Unfiled bucket */}
      <div
        className={`folder-unfiled${dragOverUnfiled ? ' drag-over' : ''}`}
        onDragOver={handleUnfiledDragOver}
        onDragLeave={handleUnfiledDragLeave}
        onDrop={handleUnfiledDrop}
      >
        <div className="folder-unfiled-row" onClick={() => onToggleFolder(unfiledKey)} onContextMenu={e => handleContextMenu(e, null)}>
          <button
            className="folder-chevron"
            aria-label={isUnfiledExpanded ? 'Collapse' : 'Expand'}
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              style={{ transform: isUnfiledExpanded ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.15s' }}
            >
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
          <span className="folder-unfiled-label">Unfiled ({unfiledItems.length})</span>
        </div>
        {isUnfiledExpanded && unfiledItems.map(item => (
          <div
            key={item.id}
            className={`folder-content-item ${itemType}`}
            draggable
            onDragStart={e => handleItemDragStart(e, item.id)}
            onClick={() => onItemClick(item.id)}
            title={getItemTitle(item)}
          >
            <span className="folder-content-icon" aria-hidden>
              {itemType === 'session' ? '\u{1F4D6}' : '\u{1F4C4}'}
            </span>
            <span className="folder-content-label">{getItemTitle(item)}</span>
            {'visibility' in item && item.visibility === 'private' && (
              <span className="folder-content-badge" title="Private">&#128274;</span>
            )}
          </div>
        ))}
      </div>

      {/* Folder list */}
      {folders.map((folder, idx) => {
        const folderItems = itemsInFolder(folder.id)
        const isExpanded = expandedState[folder.id] !== false

        return (
          <FolderItem
            key={folder.id}
            folder={folder}
            isExpanded={isExpanded}
            onToggle={() => onToggleFolder(folder.id)}
            onRename={name => onRenameFolder(folder.id, name)}
            onDelete={() => onDeleteFolder(folder.id)}
            isReadOnly={isReadOnly}
            isEmpty={folderItems.length === 0}
            onDragStart={e => handleFolderDragStart(e, folder.id)}
            onDragOver={handleFolderDragOver}
            onDrop={e => handleFolderDrop(e, idx)}
            onItemDrop={(_itemType, itemId) => onAssignItem(itemId, folder.id)}
            onFolderContextMenu={e => handleContextMenu(e, folder.id)}
          >
            {folderItems.map(item => (
              <div
                key={item.id}
                className={`folder-content-item ${itemType}`}
                draggable
                onDragStart={e => handleItemDragStart(e, item.id)}
                onClick={() => onItemClick(item.id)}
                title={getItemTitle(item)}
              >
                <span className="folder-content-icon" aria-hidden>
                  {itemType === 'session' ? '\u{1F4D6}' : '\u{1F4C4}'}
                </span>
                <span className="folder-content-label">{getItemTitle(item)}</span>
                {'visibility' in item && item.visibility === 'private' && (
                  <span className="folder-content-badge" title="Private">&#128274;</span>
                )}
              </div>
            ))}
          </FolderItem>
        )
      })}

      {/* Right-click context menu — portalled to body to escape sidebar overflow/transform */}
      {contextMenu && createPortal(
        <div
          ref={contextMenuRef}
          className="folder-section-context-menu"
          style={{ top: contextMenu.y, left: contextMenu.x }}
          onMouseDown={e => e.stopPropagation()}
        >
          <button onClick={() => {
            onCreateItem(contextMenu.folderId)
            setContextMenu(null)
          }}>
            {itemType === 'session' ? '\u{1F4D6}' : '\u{1F4C4}'} New {itemLabel}
            {contextMenu.folderId && (
              <span className="folder-section-context-hint"> in folder</span>
            )}
          </button>
          {!isReadOnly && (
            <button onClick={() => {
              setCreating(true)
              setContextMenu(null)
            }}>
              &#128193; New Folder
            </button>
          )}
        </div>,
        document.body,
      )}
    </div>
  )
}

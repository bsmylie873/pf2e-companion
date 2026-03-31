import { useState, useRef, useEffect } from 'react'
import type { Folder } from '../../types/folder'
import InlineNameInput from './InlineNameInput'

interface Props {
  folder: Folder
  isExpanded: boolean
  onToggle: () => void
  onRename: (name: string) => Promise<string | null>
  onDelete: () => void
  isReadOnly: boolean
  isEmpty: boolean
  children?: React.ReactNode
  onDragStart: (e: React.DragEvent) => void
  onDragOver: (e: React.DragEvent) => void
  onDrop: (e: React.DragEvent) => void
  onItemDrop: (itemType: 'session' | 'note', itemId: string) => void
  onFolderContextMenu?: (e: React.MouseEvent) => void
}

export default function FolderItem({
  folder, isExpanded, onToggle, onRename, onDelete,
  isReadOnly, isEmpty, children,
  onDragStart, onDragOver, onDrop, onItemDrop, onFolderContextMenu,
}: Props) {
  const [isDragOver, setIsDragOver] = useState(false)
  const [isRenaming, setIsRenaming] = useState(false)
  const [renameError, setRenameError] = useState<string | null>(null)
  const [showMenu, setShowMenu] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!showMenu) return
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowMenu(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [showMenu])

  const handleItemDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(true)
  }

  const handleItemDragLeave = (e: React.DragEvent) => {
    if (!e.currentTarget.contains(e.relatedTarget as Node)) setIsDragOver(false)
  }

  const handleItemDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragOver(false)
    const itemType = e.dataTransfer.getData('itemType') as 'session' | 'note'
    const itemId = e.dataTransfer.getData('itemId')
    if (itemId && (itemType === 'session' || itemType === 'note')) {
      onItemDrop(itemType, itemId)
    }
  }

  const handleRename = async (name: string) => {
    setRenameError(null)
    const err = await onRename(name)
    if (err) {
      setRenameError(err)
    } else {
      setIsRenaming(false)
    }
  }

  const isPrivate = folder.visibility === 'private'

  return (
    <div
      className={`folder-item${isDragOver ? ' drag-over' : ''}`}
      onDragOver={handleItemDragOver}
      onDragLeave={handleItemDragLeave}
      onDrop={handleItemDrop}
    >
      <div
        className="folder-item-row"
        draggable={!isReadOnly}
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDrop={onDrop}
        onContextMenu={onFolderContextMenu}
      >
        <button
          className="folder-chevron"
          onClick={onToggle}
          aria-label={isExpanded ? 'Collapse' : 'Expand'}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)', transition: 'transform 0.15s' }}
          >
            <path d="M9 18l6-6-6-6" />
          </svg>
        </button>

        {isPrivate && <span className="folder-lock" title="Private">&#128274;</span>}

        <span className="folder-icon" aria-hidden>&#128193;</span>

        {isRenaming ? (
          <InlineNameInput
            value={folder.name}
            onCommit={handleRename}
            onCancel={() => { setIsRenaming(false); setRenameError(null) }}
            error={renameError}
            autoFocus
          />
        ) : (
          <span
            className="folder-name"
            onDoubleClick={() => !isReadOnly && setIsRenaming(true)}
          >
            {folder.name}
          </span>
        )}

        {isEmpty && !isExpanded && (
          <span className="folder-empty-indicator" title="Empty folder">&#8212;</span>
        )}

        {!isReadOnly && (
          <div className="folder-menu-wrap" ref={menuRef}>
            <button
              className="folder-menu-btn"
              onClick={() => setShowMenu(v => !v)}
              aria-label="Folder options"
            >
              &#8943;
            </button>
            {showMenu && (
              <div className="folder-context-menu">
                <button onClick={() => { setIsRenaming(true); setShowMenu(false) }}>Rename</button>
                <button className="danger" onClick={() => { onDelete(); setShowMenu(false) }}>Delete</button>
              </div>
            )}
          </div>
        )}
      </div>

      {isExpanded && children && (
        <div className="folder-item-children">
          {children}
        </div>
      )}
    </div>
  )
}

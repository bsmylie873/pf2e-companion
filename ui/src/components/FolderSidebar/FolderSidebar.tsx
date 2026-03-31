import { useState, useEffect, useCallback, useRef } from 'react'
import type { Folder } from '../../types/folder'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'
import type { GameSidebarState } from '../../api/preferences'
import { listFolders, createFolder, renameFolder, deleteFolder, reorderFolders } from '../../api/folders'
import { getPreferences, updatePreferences } from '../../api/preferences'
import FolderSection from './FolderSection'
import './FolderSidebar.css'

interface FolderSidebarProps {
  gameId: string
  isGM: boolean
  userId: string
  sessions: Session[]
  notes: Note[]
  onSessionClick: (sessionId: string) => void
  onNoteClick: (noteId: string) => void
  onSessionUpdate: (sessionId: string, data: Record<string, unknown>) => Promise<void>
  onNoteUpdate: (noteId: string, data: Record<string, unknown>) => Promise<void>
  onCreateSession: (folderId: string | null) => void
  onCreateNote: (folderId: string | null) => void
}

export default function FolderSidebar({
  gameId, isGM, userId: _userId,
  sessions, notes,
  onSessionClick, onNoteClick,
  onSessionUpdate, onNoteUpdate,
  onCreateSession, onCreateNote,
}: FolderSidebarProps) {
  const [sessionFolders, setSessionFolders] = useState<Folder[]>([])
  const [noteFolders, setNoteFolders] = useState<Folder[]>([])
  const [expandedState, setExpandedState] = useState<Record<string, boolean>>({})
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Load folders and sidebar state on mount
  useEffect(() => {
    listFolders(gameId, 'session').then(setSessionFolders).catch(() => {})
    listFolders(gameId, 'note').then(setNoteFolders).catch(() => {})
    getPreferences().then(prefs => {
      if (prefs.sidebar_state && prefs.sidebar_state[gameId]) {
        const gameState = prefs.sidebar_state[gameId]
        const expanded: Record<string, boolean> = {}
        for (const [key, val] of Object.entries(gameState)) {
          if (key !== 'panelOpen') expanded[key] = val as boolean
        }
        setExpandedState(expanded)
      }
    }).catch(() => {})
  }, [gameId])

  // Debounced persist of expanded state
  const persistExpandedState = useCallback((state: Record<string, boolean>) => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      getPreferences().then(prefs => {
        const sidebarState = prefs.sidebar_state ?? {}
        const gameState: GameSidebarState = {
          ...(sidebarState[gameId] ?? { panelOpen: true }),
          ...state,
        }
        updatePreferences({
          sidebar_state: { ...sidebarState, [gameId]: gameState },
        }).catch(() => {})
      }).catch(() => {})
    }, 500)
  }, [gameId])

  const handleToggleFolder = useCallback((folderId: string) => {
    setExpandedState(prev => {
      const next = { ...prev, [folderId]: prev[folderId] === false }
      persistExpandedState(next)
      return next
    })
  }, [persistExpandedState])

  // Session folder handlers
  const handleCreateSessionFolder = useCallback(async (name: string): Promise<string | null> => {
    try {
      const folder = await createFolder(gameId, { name, folder_type: 'session', visibility: 'game-wide' })
      setSessionFolders(prev => [...prev, folder])
      return null
    } catch (err: unknown) {
      return err instanceof Error ? err.message : 'Failed to create folder'
    }
  }, [gameId])

  const handleRenameSessionFolder = useCallback(async (folderId: string, name: string): Promise<string | null> => {
    try {
      const updated = await renameFolder(folderId, name)
      setSessionFolders(prev => prev.map(f => f.id === folderId ? updated : f))
      return null
    } catch (err: unknown) {
      return err instanceof Error ? err.message : 'Failed to rename folder'
    }
  }, [])

  const handleDeleteSessionFolder = useCallback(async (folderId: string) => {
    if (!confirm('Delete this folder? Items inside will become unfiled.')) return
    try {
      await deleteFolder(folderId)
      setSessionFolders(prev => prev.filter(f => f.id !== folderId))
    } catch (err) {
      console.error('Failed to delete folder', err)
    }
  }, [])

  const handleReorderSessionFolders = useCallback(async (orderedIds: string[]) => {
    // Optimistic reorder
    setSessionFolders(prev => {
      const map = new Map(prev.map(f => [f.id, f]))
      return orderedIds.map((id, i) => ({ ...map.get(id)!, position: i }))
    })
    try {
      await reorderFolders(gameId, 'session', orderedIds)
    } catch (err) {
      console.error('Failed to reorder folders', err)
      listFolders(gameId, 'session').then(setSessionFolders).catch(() => {})
    }
  }, [gameId])

  const handleAssignSession = useCallback(async (sessionId: string, folderId: string | null) => {
    try {
      await onSessionUpdate(sessionId, { folder_id: folderId })
    } catch (err) {
      console.error('Failed to assign session', err)
    }
  }, [onSessionUpdate])

  // Note folder handlers
  const handleCreateNoteFolder = useCallback(async (name: string): Promise<string | null> => {
    try {
      const folder = await createFolder(gameId, { name, folder_type: 'note', visibility: 'game-wide' })
      setNoteFolders(prev => [...prev, folder])
      return null
    } catch (err: unknown) {
      return err instanceof Error ? err.message : 'Failed to create folder'
    }
  }, [gameId])

  const handleRenameNoteFolder = useCallback(async (folderId: string, name: string): Promise<string | null> => {
    try {
      const updated = await renameFolder(folderId, name)
      setNoteFolders(prev => prev.map(f => f.id === folderId ? updated : f))
      return null
    } catch (err: unknown) {
      return err instanceof Error ? err.message : 'Failed to rename folder'
    }
  }, [])

  const handleDeleteNoteFolder = useCallback(async (folderId: string) => {
    if (!confirm('Delete this folder? Items inside will become unfiled.')) return
    try {
      await deleteFolder(folderId)
      setNoteFolders(prev => prev.filter(f => f.id !== folderId))
    } catch (err) {
      console.error('Failed to delete folder', err)
    }
  }, [])

  const handleReorderNoteFolders = useCallback(async (orderedIds: string[]) => {
    setNoteFolders(prev => {
      const map = new Map(prev.map(f => [f.id, f]))
      return orderedIds.map((id, i) => ({ ...map.get(id)!, position: i }))
    })
    try {
      await reorderFolders(gameId, 'note', orderedIds)
    } catch (err) {
      console.error('Failed to reorder folders', err)
      listFolders(gameId, 'note').then(setNoteFolders).catch(() => {})
    }
  }, [gameId])

  const handleAssignNote = useCallback(async (noteId: string, folderId: string | null) => {
    try {
      await onNoteUpdate(noteId, { folder_id: folderId })
    } catch (err) {
      console.error('Failed to assign note', err)
    }
  }, [onNoteUpdate])

  return (
    <aside className="folder-sidebar">
      <div className="folder-sidebar-header">
        <span className="folder-sidebar-title">Folders</span>
      </div>

      <div className="folder-sidebar-content">
        <FolderSection
          title="SESSION FOLDERS"
          folders={sessionFolders}
          items={sessions}
          isReadOnly={!isGM}
          expandedState={expandedState}
          onToggleFolder={handleToggleFolder}
          onCreateFolder={handleCreateSessionFolder}
          onRenameFolder={handleRenameSessionFolder}
          onDeleteFolder={handleDeleteSessionFolder}
          onReorderFolders={handleReorderSessionFolders}
          onItemClick={onSessionClick}
          onAssignItem={handleAssignSession}
          onCreateItem={onCreateSession}
          itemType="session"
        />

        <div className="folder-sidebar-divider" />

        <FolderSection
          title="NOTE FOLDERS"
          folders={noteFolders}
          items={notes}
          isReadOnly={false}
          expandedState={expandedState}
          onToggleFolder={handleToggleFolder}
          onCreateFolder={handleCreateNoteFolder}
          onRenameFolder={handleRenameNoteFolder}
          onDeleteFolder={handleDeleteNoteFolder}
          onReorderFolders={handleReorderNoteFolders}
          onItemClick={onNoteClick}
          onAssignItem={handleAssignNote}
          onCreateItem={onCreateNote}
          itemType="note"
        />
      </div>
    </aside>
  )
}

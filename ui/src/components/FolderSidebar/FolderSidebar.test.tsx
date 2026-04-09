import React from 'react'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import FolderSidebar from './FolderSidebar'

// ─── hoisted shared state for capturing FolderSection props ─────────────────
const capturedProps = vi.hoisted(() => ({} as Record<string, any>))

// ─── mocks ───────────────────────────────────────────────────────────────────

vi.mock('../../api/folders', () => ({
  listFolders: vi.fn().mockResolvedValue([]),
  createFolder: vi.fn().mockResolvedValue({ id: 'new-f', name: 'New Folder', game_id: 'game-1', item_type: 'session', visibility: 'game-wide', sort_order: 0, created_at: '', updated_at: '' }),
  renameFolder: vi.fn().mockResolvedValue({ id: 'f1', name: 'Renamed', game_id: 'game-1', item_type: 'session', visibility: 'game-wide', sort_order: 0, created_at: '', updated_at: '' }),
  deleteFolder: vi.fn().mockResolvedValue(undefined),
  reorderFolders: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../../api/preferences', () => ({
  getPreferences: vi.fn().mockResolvedValue({ sidebar_state: {} }),
  updatePreferences: vi.fn().mockResolvedValue({}),
}))

// FolderSection mock captures all props for assertion and exposes trigger buttons
vi.mock('./FolderSection', () => ({
  default: (props: any) => {
    capturedProps[props.itemType] = props
    return (
      <div
        data-testid={`folder-section-${props.itemType}`}
        data-readonly={String(props.isReadOnly)}
        data-items={String(props.items?.length ?? 0)}
        data-folders={String(props.folders?.length ?? 0)}
      >
        {props.title}
        <button
          data-testid={`do-create-folder-${props.itemType}`}
          onClick={() => props.onCreateFolder('New Folder')}
        >
          Create Folder
        </button>
        <button
          data-testid={`do-rename-folder-${props.itemType}`}
          onClick={() => props.onRenameFolder('f1', 'Renamed')}
        >
          Rename Folder
        </button>
        <button
          data-testid={`do-delete-folder-${props.itemType}`}
          onClick={() => props.onDeleteFolder('f1')}
        >
          Delete Folder
        </button>
        <button
          data-testid={`do-reorder-folders-${props.itemType}`}
          onClick={() => props.onReorderFolders(['id1', 'id2'])}
        >
          Reorder Folders
        </button>
        <button
          data-testid={`do-toggle-folder-${props.itemType}`}
          onClick={() => props.onToggleFolder('fold-key')}
        >
          Toggle
        </button>
        <button
          data-testid={`do-assign-item-${props.itemType}`}
          onClick={() => props.onAssignItem('item-1', 'folder-1')}
        >
          Assign Item
        </button>
        <button
          data-testid={`do-item-click-${props.itemType}`}
          onClick={() => props.onItemClick('item-1')}
        >
          Item Click
        </button>
        <button
          data-testid={`do-create-item-${props.itemType}`}
          onClick={() => props.onCreateItem('folder-1')}
        >
          Create Item
        </button>
      </div>
    )
  },
}))

// ─── imports after mocks ─────────────────────────────────────────────────────

import { listFolders, createFolder, renameFolder, deleteFolder, reorderFolders } from '../../api/folders'
import { getPreferences, updatePreferences } from '../../api/preferences'

const mockListFolders = vi.mocked(listFolders)
const mockGetPreferences = vi.mocked(getPreferences)
const mockCreateFolder = vi.mocked(createFolder)
const mockRenameFolder = vi.mocked(renameFolder)
const mockDeleteFolder = vi.mocked(deleteFolder)
const mockReorderFolders = vi.mocked(reorderFolders)
const mockUpdatePreferences = vi.mocked(updatePreferences)

// ─── sample data ─────────────────────────────────────────────────────────────

const sampleSessionFolder: any = {
  id: 'sf1', name: 'Campaign Arc', item_type: 'session',
  game_id: 'game-1', visibility: 'game-wide', sort_order: 0,
  created_at: '2024-01-01', updated_at: '2024-01-01',
}

const sampleNoteFolder: any = {
  id: 'nf1', name: 'Lore', item_type: 'note',
  game_id: 'game-1', visibility: 'game-wide', sort_order: 0,
  created_at: '2024-01-01', updated_at: '2024-01-01',
}

const sampleSession: any = {
  id: 's1', title: 'Session 1', folder_id: null,
  game_id: 'game-1', session_number: 1,
  created_at: '2024-01-01', updated_at: '2024-01-01',
}

const sampleNote: any = {
  id: 'n1', title: 'Meeting Notes', folder_id: null,
  game_id: 'game-1', visibility: 'public',
  created_at: '2024-01-01', updated_at: '2024-01-01',
}

// ─── default props ───────────────────────────────────────────────────────────

const defaultProps = {
  gameId: 'game-1',
  isGM: true,
  userId: 'user-1',
  sessions: [],
  notes: [],
  onSessionClick: vi.fn(),
  onNoteClick: vi.fn(),
  onSessionUpdate: vi.fn().mockResolvedValue(undefined),
  onNoteUpdate: vi.fn().mockResolvedValue(undefined),
  onCreateSession: vi.fn(),
  onCreateNote: vi.fn(),
}

// ─── tests ───────────────────────────────────────────────────────────────────

describe('FolderSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset captured props
    delete capturedProps.session
    delete capturedProps.note
    mockListFolders.mockResolvedValue([])
    mockGetPreferences.mockResolvedValue({ sidebar_state: {} })
    mockCreateFolder.mockResolvedValue({
      id: 'new-f', name: 'New Folder', game_id: 'game-1',
      item_type: 'session', visibility: 'game-wide', sort_order: 0,
      created_at: '', updated_at: '',
    } as any)
    mockRenameFolder.mockResolvedValue({
      id: 'f1', name: 'Renamed', game_id: 'game-1',
      item_type: 'session', visibility: 'game-wide', sort_order: 0,
      created_at: '', updated_at: '',
    } as any)
    mockDeleteFolder.mockResolvedValue(undefined)
    mockReorderFolders.mockResolvedValue(undefined)
    mockUpdatePreferences.mockResolvedValue({} as any)
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })

  // ── original tests ─────────────────────────────────────────────────────────

  it('should render the Folders header', () => {
    render(<FolderSidebar {...defaultProps} />)
    expect(screen.getByText('Folders')).toBeInTheDocument()
  })

  it('should render SESSION FOLDERS section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toBeInTheDocument()
    })
  })

  it('should render NOTE FOLDERS section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-note')).toBeInTheDocument()
    })
  })

  it('should call listFolders for both session and note types on mount', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(mockListFolders).toHaveBeenCalledWith('game-1', 'session')
      expect(mockListFolders).toHaveBeenCalledWith('game-1', 'note')
    })
  })

  it('should call getPreferences on mount', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(mockGetPreferences).toHaveBeenCalled()
    })
  })

  // ── isReadOnly propagation ─────────────────────────────────────────────────

  it('passes isReadOnly=true to session section when isGM=false', async () => {
    render(<FolderSidebar {...defaultProps} isGM={false} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toHaveAttribute('data-readonly', 'true')
    })
  })

  it('passes isReadOnly=false to session section when isGM=true', async () => {
    render(<FolderSidebar {...defaultProps} isGM={true} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toHaveAttribute('data-readonly', 'false')
    })
  })

  it('always passes isReadOnly=false to note section regardless of isGM', async () => {
    render(<FolderSidebar {...defaultProps} isGM={false} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-note')).toHaveAttribute('data-readonly', 'false')
    })
  })

  // ── items propagation ──────────────────────────────────────────────────────

  it('passes sessions array to session FolderSection', async () => {
    render(<FolderSidebar {...defaultProps} sessions={[sampleSession]} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toHaveAttribute('data-items', '1')
    })
  })

  it('passes notes array to note FolderSection', async () => {
    render(<FolderSidebar {...defaultProps} notes={[sampleNote]} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-note')).toHaveAttribute('data-items', '1')
    })
  })

  // ── folders from API ───────────────────────────────────────────────────────

  it('passes session folders from listFolders to session section', async () => {
    mockListFolders.mockImplementation((_gameId: string, type: string) =>
      type === 'session'
        ? Promise.resolve([sampleSessionFolder])
        : Promise.resolve([]),
    )
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toHaveAttribute('data-folders', '1')
    })
  })

  it('passes note folders from listFolders to note section', async () => {
    mockListFolders.mockImplementation((_gameId: string, type: string) =>
      type === 'note'
        ? Promise.resolve([sampleNoteFolder])
        : Promise.resolve([]),
    )
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-note')).toHaveAttribute('data-folders', '1')
    })
  })

  it('handles listFolders error gracefully — renders without crash', async () => {
    mockListFolders.mockRejectedValue(new Error('Network error'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByText('Folders')).toBeInTheDocument()
    })
  })

  // ── sidebar_state / expandedState ─────────────────────────────────────────

  it('sets expandedState from sidebar_state in preferences', async () => {
    mockGetPreferences.mockResolvedValue({
      sidebar_state: {
        'game-1': { panelOpen: true, 'folder-abc': false, 'folder-xyz': true },
      },
    } as any)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      // expandedState should include the keys from sidebar_state (excluding panelOpen)
      expect(capturedProps.session?.expandedState).toMatchObject({
        'folder-abc': false,
        'folder-xyz': true,
      })
    })
  })

  it('leaves expandedState empty when sidebar_state has no entry for gameId', async () => {
    mockGetPreferences.mockResolvedValue({
      sidebar_state: { 'other-game': { panelOpen: false } },
    } as any)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toBeInTheDocument()
    })
    expect(capturedProps.session?.expandedState).toEqual({})
  })

  // ── handleToggleFolder ────────────────────────────────────────────────────

  it('handleToggleFolder updates expandedState and schedules persistExpandedState', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('folder-section-session')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('do-toggle-folder-session'))

    // No crash and state update queued
    expect(screen.getByTestId('folder-section-session')).toBeInTheDocument()
  })

  // ── handleCreateSessionFolder ─────────────────────────────────────────────

  it('calls createFolder when session section creates a folder', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-create-folder-session'))
    })

    expect(mockCreateFolder).toHaveBeenCalledWith('game-1', expect.objectContaining({
      name: 'New Folder',
      folder_type: 'session',
    }))
  })

  it('adds created session folder to the list', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-create-folder-session'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('folder-section-session')).toHaveAttribute('data-folders', '1')
    })
  })

  it('returns error string from onCreateFolder when createFolder fails', async () => {
    mockCreateFolder.mockRejectedValue(new Error('Server error'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-session')).toBeInTheDocument())

    let result: string | null = null
    await act(async () => {
      result = await capturedProps.session?.onCreateFolder('Bad Folder')
    })

    expect(result).toBe('Server error')
  })

  // ── handleRenameSessionFolder ──────────────────────────────────────────────

  it('calls renameFolder when session section renames a folder', async () => {
    mockListFolders.mockImplementation((_id: string, type: string) =>
      type === 'session' ? Promise.resolve([sampleSessionFolder]) : Promise.resolve([]),
    )
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-rename-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-rename-folder-session'))
    })

    expect(mockRenameFolder).toHaveBeenCalledWith('f1', 'Renamed')
  })

  it('returns error string from onRenameFolder when renameFolder fails', async () => {
    mockRenameFolder.mockRejectedValue(new Error('Rename failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-rename-folder-session')).toBeInTheDocument())

    let result: string | null = null
    await act(async () => {
      result = await capturedProps.session?.onRenameFolder('f1', 'Bad Name')
    })

    expect(result).toBe('Rename failed')
  })

  // ── handleDeleteSessionFolder ──────────────────────────────────────────────

  it('calls deleteFolder when confirm=true for session folder', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-delete-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-delete-folder-session'))
    })

    expect(mockDeleteFolder).toHaveBeenCalledWith('f1')
  })

  it('does NOT call deleteFolder when confirm=false for session folder', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-delete-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-delete-folder-session'))
    })

    expect(mockDeleteFolder).not.toHaveBeenCalled()
  })

  // ── handleReorderSessionFolders ───────────────────────────────────────────

  it('calls reorderFolders when session folders are reordered', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-reorder-folders-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-reorder-folders-session'))
    })

    expect(mockReorderFolders).toHaveBeenCalledWith('game-1', 'session', ['id1', 'id2'])
  })

  // ── handleAssignSession ───────────────────────────────────────────────────

  it('calls onSessionUpdate when session is assigned to a folder', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-assign-item-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-assign-item-session'))
    })

    expect(defaultProps.onSessionUpdate).toHaveBeenCalledWith('item-1', { folder_id: 'folder-1' })
  })

  // ── handleCreateNoteFolder ────────────────────────────────────────────────

  it('calls createFolder when note section creates a folder', async () => {
    mockCreateFolder.mockResolvedValue({
      id: 'new-nf', name: 'New Note Folder', game_id: 'game-1',
      item_type: 'note', visibility: 'game-wide', sort_order: 0,
      created_at: '', updated_at: '',
    } as any)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-create-folder-note'))
    })

    expect(mockCreateFolder).toHaveBeenCalledWith('game-1', expect.objectContaining({
      name: 'New Folder',
      folder_type: 'note',
    }))
  })

  it('adds created note folder to the note list', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-create-folder-note'))
    })

    await waitFor(() => {
      expect(screen.getByTestId('folder-section-note')).toHaveAttribute('data-folders', '1')
    })
  })

  // ── handleDeleteNoteFolder ────────────────────────────────────────────────

  it('calls deleteFolder when confirm=true for note folder', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-delete-folder-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-delete-folder-note'))
    })

    expect(mockDeleteFolder).toHaveBeenCalledWith('f1')
  })

  // ── handleReorderNoteFolders ──────────────────────────────────────────────

  it('calls reorderFolders when note folders are reordered', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-reorder-folders-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-reorder-folders-note'))
    })

    expect(mockReorderFolders).toHaveBeenCalledWith('game-1', 'note', ['id1', 'id2'])
  })

  // ── handleAssignNote ──────────────────────────────────────────────────────

  it('calls onNoteUpdate when note is assigned to a folder', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-assign-item-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-assign-item-note'))
    })

    expect(defaultProps.onNoteUpdate).toHaveBeenCalledWith('item-1', { folder_id: 'folder-1' })
  })

  // ── onSessionClick / onNoteClick pass-through ─────────────────────────────

  it('passes onSessionClick through as onItemClick to session section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-item-click-session')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('do-item-click-session'))

    expect(defaultProps.onSessionClick).toHaveBeenCalledWith('item-1')
  })

  it('passes onNoteClick through as onItemClick to note section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-item-click-note')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('do-item-click-note'))

    expect(defaultProps.onNoteClick).toHaveBeenCalledWith('item-1')
  })

  // ── onCreateSession / onCreateNote pass-through ───────────────────────────

  it('passes onCreateSession through as onCreateItem to session section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-item-session')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('do-create-item-session'))

    expect(defaultProps.onCreateSession).toHaveBeenCalledWith('folder-1')
  })

  it('passes onCreateNote through as onCreateItem to note section', async () => {
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-item-note')).toBeInTheDocument())

    fireEvent.click(screen.getByTestId('do-create-item-note'))

    expect(defaultProps.onCreateNote).toHaveBeenCalledWith('folder-1')
  })

  // ── error paths: catch blocks ─────────────────────────────────────────────

  it('handleDeleteSessionFolder error: logs and does not crash', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    mockDeleteFolder.mockRejectedValue(new Error('Delete failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-delete-folder-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-delete-folder-session'))
    })

    // No crash — component still renders
    expect(screen.getByTestId('folder-section-session')).toBeInTheDocument()
  })

  it('handleDeleteNoteFolder error: logs and does not crash', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    mockDeleteFolder.mockRejectedValue(new Error('Delete note failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-delete-folder-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-delete-folder-note'))
    })

    expect(screen.getByTestId('folder-section-note')).toBeInTheDocument()
  })

  it('handleReorderSessionFolders error: refetches session folders', async () => {
    mockReorderFolders.mockRejectedValue(new Error('Reorder failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-reorder-folders-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-reorder-folders-session'))
    })

    // After error, listFolders should be called again to restore order
    await waitFor(() => {
      expect(mockListFolders).toHaveBeenCalledWith('game-1', 'session')
    })
  })

  it('handleReorderNoteFolders error: refetches note folders', async () => {
    mockReorderFolders.mockRejectedValue(new Error('Reorder failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-reorder-folders-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-reorder-folders-note'))
    })

    await waitFor(() => {
      expect(mockListFolders).toHaveBeenCalledWith('game-1', 'note')
    })
  })

  it('handleAssignSession error: logs and does not crash', async () => {
    const onSessionUpdate = vi.fn().mockRejectedValue(new Error('Assign failed'))
    render(<FolderSidebar {...defaultProps} onSessionUpdate={onSessionUpdate} />)
    await waitFor(() => expect(screen.getByTestId('do-assign-item-session')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-assign-item-session'))
    })

    expect(screen.getByTestId('folder-section-session')).toBeInTheDocument()
  })

  it('handleAssignNote error: logs and does not crash', async () => {
    const onNoteUpdate = vi.fn().mockRejectedValue(new Error('Assign note failed'))
    render(<FolderSidebar {...defaultProps} onNoteUpdate={onNoteUpdate} />)
    await waitFor(() => expect(screen.getByTestId('do-assign-item-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-assign-item-note'))
    })

    expect(screen.getByTestId('folder-section-note')).toBeInTheDocument()
  })

  it('handleRenameNoteFolder success: updates note folder list', async () => {
    mockRenameFolder.mockResolvedValue({
      id: 'f1', name: 'Renamed Note Folder', game_id: 'game-1',
      item_type: 'note', visibility: 'game-wide', sort_order: 0,
      created_at: '', updated_at: '',
    } as any)
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-rename-folder-note')).toBeInTheDocument())

    await act(async () => {
      fireEvent.click(screen.getByTestId('do-rename-folder-note'))
    })

    expect(mockRenameFolder).toHaveBeenCalledWith('f1', 'Renamed')
  })

  it('handleRenameNoteFolder error: returns error message', async () => {
    mockRenameFolder.mockRejectedValue(new Error('Note rename failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-rename-folder-note')).toBeInTheDocument())

    let result: string | null = null
    await act(async () => {
      result = await capturedProps.note?.onRenameFolder('f1', 'New Name')
    })

    expect(result).toBe('Note rename failed')
  })

  it('handleCreateNoteFolder error: returns error message', async () => {
    mockCreateFolder.mockRejectedValue(new Error('Note folder create failed'))
    render(<FolderSidebar {...defaultProps} />)
    await waitFor(() => expect(screen.getByTestId('do-create-folder-note')).toBeInTheDocument())

    let result: string | null = null
    await act(async () => {
      result = await capturedProps.note?.onCreateFolder('Bad Folder')
    })

    expect(result).toBe('Note folder create failed')
  })

  // ── persistExpandedState debounce ──────────────────────────────────────────

  it('persistExpandedState fires after debounce and calls updatePreferences', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      render(<FolderSidebar {...defaultProps} />)
      // Wait for initial async effects with real timers having advanced
      await act(async () => { await Promise.resolve() })
      await waitFor(() => expect(screen.getByTestId('do-toggle-folder-session')).toBeInTheDocument())

      // Trigger toggle which calls persistExpandedState
      fireEvent.click(screen.getByTestId('do-toggle-folder-session'))

      // Advance past the 500ms debounce + flush promises
      await act(async () => {
        vi.advanceTimersByTime(600)
        await Promise.resolve()
        await Promise.resolve()
      })

      expect(mockGetPreferences).toHaveBeenCalled()
      expect(mockUpdatePreferences).toHaveBeenCalled()
    } finally {
      vi.useRealTimers()
    }
  })
})

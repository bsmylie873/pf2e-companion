import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import FolderSection from './FolderSection'
import type { Folder } from '../../types/folder'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'

vi.mock('./FolderItem', () => ({
  default: ({ folder, onToggle, onRename, onDelete, isReadOnly, children, onFolderContextMenu, onDragStart, onDragOver, onDrop, onItemDrop }: any) => (
    <div
      data-testid={`folder-item-${folder.id}`}
      onClick={onToggle}
      onContextMenu={onFolderContextMenu}
      onDragOver={onDragOver}
      onDrop={onDrop}
    >
      {folder.name}
      {!isReadOnly && (
        <>
          <button data-testid={`rename-${folder.id}`} onClick={() => onRename('renamed name')}>Rename</button>
          <button data-testid={`delete-${folder.id}`} onClick={onDelete}>Delete</button>
        </>
      )}
      <button
        data-testid={`drag-start-${folder.id}`}
        onDragStart={onDragStart}
        draggable
      >Drag</button>
      <button
        data-testid={`item-drop-${folder.id}`}
        onClick={() => onItemDrop('session', 'item-in-folder')}
      >Drop item</button>
      {children}
    </div>
  ),
}))

vi.mock('./InlineNameInput', () => ({
  default: ({ onCommit, onCancel, error, placeholder }: any) => (
    <div data-testid="inline-name-input">
      <input
        data-testid="inline-input"
        placeholder={placeholder}
        defaultValue=""
        onChange={() => {}}
        onKeyDown={(e: any) => {
          if (e.key === 'Enter') onCommit(e.target.value || 'New Folder')
          if (e.key === 'Escape') onCancel()
        }}
      />
      <button data-testid="inline-cancel" onClick={onCancel}>Cancel</button>
      {error && <span data-testid="inline-error">{error}</span>}
    </div>
  ),
}))

const sampleFolder: Folder = {
  id: 'f1',
  name: 'Folder One',
  game_id: 'g1',
  user_id: null,
  folder_type: 'session',
  visibility: 'game-wide',
  position: 0,
  created_at: '',
  updated_at: '',
}

const sampleSession: Session = {
  id: 's1',
  title: 'Session 1',
  folder_id: null,
  game_id: 'g1',
  session_number: 1,
  scheduled_at: null,
  runtime_start: null,
  runtime_end: null,
  notes: null,
  version: 0,
  foundry_data: null,
  created_at: '',
  updated_at: '',
}

const sampleSession2: Session = {
  ...sampleSession,
  id: 's2',
  title: 'Session 2',
}

const sampleNote: Note = {
  id: 'n1',
  title: 'Note 1',
  folder_id: null,
  game_id: 'g1',
  user_id: 'u1',
  session_id: null,
  content: null,
  visibility: 'visible',
  version: 0,
  foundry_data: null,
  created_at: '',
  updated_at: '',
}

const privateNote: Note = {
  ...sampleNote,
  id: 'n2',
  title: 'Private Note',
  visibility: 'private',
}

const baseProps = {
  title: 'SESSION FOLDERS',
  folders: [] as Folder[],
  items: [] as (Session | Note)[],
  isReadOnly: false,
  expandedState: {} as Record<string, boolean>,
  onToggleFolder: vi.fn(),
  onCreateFolder: vi.fn().mockResolvedValue(null),
  onRenameFolder: vi.fn().mockResolvedValue(null),
  onDeleteFolder: vi.fn(),
  onReorderFolders: vi.fn(),
  onItemClick: vi.fn(),
  onAssignItem: vi.fn(),
  onCreateItem: vi.fn(),
  itemType: 'session' as const,
}

describe('FolderSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    baseProps.onCreateFolder.mockResolvedValue(null)
    baseProps.onRenameFolder.mockResolvedValue(null)
  })

  it('renders the section title', () => {
    render(<FolderSection {...baseProps} />)
    expect(screen.getByText('SESSION FOLDERS')).toBeInTheDocument()
  })

  it('shows add folder button when not read-only', () => {
    render(<FolderSection {...baseProps} isReadOnly={false} />)
    expect(screen.getByTitle('New folder')).toBeInTheDocument()
  })

  it('hides add folder button when read-only', () => {
    render(<FolderSection {...baseProps} isReadOnly={true} />)
    expect(screen.queryByTitle('New folder')).not.toBeInTheDocument()
  })

  it('clicking add button shows InlineNameInput', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByTitle('New folder'))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
  })

  it('InlineNameInput Enter commit calls onCreateFolder and hides input on success', async () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByTitle('New folder'))
    const input = screen.getByTestId('inline-input')
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => {
      expect(baseProps.onCreateFolder).toHaveBeenCalledWith('New Folder')
      expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
    })
  })

  it('InlineNameInput Enter with error keeps input and shows error', async () => {
    baseProps.onCreateFolder.mockResolvedValue('Name already taken')
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByTitle('New folder'))
    const input = screen.getByTestId('inline-input')
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => {
      expect(screen.getByTestId('inline-error')).toHaveTextContent('Name already taken')
    })
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
  })

  it('Cancel button hides InlineNameInput', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByTitle('New folder'))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
    fireEvent.click(screen.getByTestId('inline-cancel'))
    expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
  })

  it('shows Unfiled (0) when no items', () => {
    render(<FolderSection {...baseProps} items={[]} />)
    expect(screen.getByText('Unfiled (0)')).toBeInTheDocument()
  })

  it('shows Unfiled (2) when 2 unfiled items', () => {
    render(<FolderSection {...baseProps} items={[sampleSession, sampleSession2]} />)
    expect(screen.getByText('Unfiled (2)')).toBeInTheDocument()
  })

  it('clicking unfiled row calls onToggleFolder with unfiled-session key', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByText('Unfiled (0)'))
    expect(baseProps.onToggleFolder).toHaveBeenCalledWith('unfiled-session')
  })

  it('items shown in unfiled when expandedState is empty (default expanded)', () => {
    render(<FolderSection {...baseProps} items={[sampleSession]} expandedState={{}} />)
    expect(screen.getByText('Session 1')).toBeInTheDocument()
  })

  it('items hidden when expandedState sets unfiled-session to false', () => {
    render(<FolderSection {...baseProps} items={[sampleSession]} expandedState={{ 'unfiled-session': false }} />)
    expect(screen.queryByText('Session 1')).not.toBeInTheDocument()
  })

  it('clicking an unfiled item calls onItemClick with its id', () => {
    render(<FolderSection {...baseProps} items={[sampleSession]} />)
    fireEvent.click(screen.getByText('Session 1'))
    expect(baseProps.onItemClick).toHaveBeenCalledWith('s1')
  })

  it('private note shows lock badge with title "Private"', () => {
    render(<FolderSection {...baseProps} items={[privateNote]} itemType="note" />)
    const badge = screen.getByTitle('Private')
    expect(badge).toBeInTheDocument()
  })

  it('non-private note does not show lock badge', () => {
    render(<FolderSection {...baseProps} items={[sampleNote]} itemType="note" />)
    expect(screen.queryByTitle('Private')).not.toBeInTheDocument()
  })

  it('renders a FolderItem for each folder', () => {
    const folder2: Folder = { ...sampleFolder, id: 'f2', name: 'Folder Two' }
    render(<FolderSection {...baseProps} folders={[sampleFolder, folder2]} />)
    expect(screen.getByTestId('folder-item-f1')).toBeInTheDocument()
    expect(screen.getByTestId('folder-item-f2')).toBeInTheDocument()
  })

  it('clicking a FolderItem calls onToggleFolder with its folder id', () => {
    render(<FolderSection {...baseProps} folders={[sampleFolder]} />)
    fireEvent.click(screen.getByTestId('folder-item-f1'))
    expect(baseProps.onToggleFolder).toHaveBeenCalledWith('f1')
  })

  it('right-click on section shows context menu with New Session button', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    expect(screen.getByText(/New Session/)).toBeInTheDocument()
  })

  it('context menu New Session click calls onCreateItem and closes menu', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    fireEvent.click(screen.getByText(/New Session/))
    expect(baseProps.onCreateItem).toHaveBeenCalledWith(null)
    expect(screen.queryByText(/New Session/)).not.toBeInTheDocument()
  })

  it('context menu shows New Folder button when not read-only', () => {
    render(<FolderSection {...baseProps} isReadOnly={false} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    expect(screen.getByText(/New Folder/)).toBeInTheDocument()
  })

  it('context menu hides New Folder button when read-only', () => {
    render(<FolderSection {...baseProps} isReadOnly={true} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    expect(screen.queryByText(/New Folder/)).not.toBeInTheDocument()
  })

  it('context menu New Folder click shows InlineNameInput', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    fireEvent.click(screen.getByText(/New Folder/))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
  })

  it('context menu closes on outside mousedown', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    expect(screen.getByText(/New Session/)).toBeInTheDocument()
    fireEvent.mouseDown(document.body)
    expect(screen.queryByText(/New Session/)).not.toBeInTheDocument()
  })

  it('drag-over on unfiled area adds drag-over class', () => {
    render(<FolderSection {...baseProps} />)
    const unfiled = document.querySelector('.folder-unfiled')!
    fireEvent.dragOver(unfiled)
    expect(unfiled.classList.contains('drag-over')).toBe(true)
  })

  it('drag-leave on unfiled area removes drag-over class', () => {
    render(<FolderSection {...baseProps} />)
    const unfiled = document.querySelector('.folder-unfiled')!
    fireEvent.dragOver(unfiled)
    fireEvent.dragLeave(unfiled)
    expect(unfiled.classList.contains('drag-over')).toBe(false)
  })

  it('drop on unfiled area with matching itemType calls onAssignItem', () => {
    render(<FolderSection {...baseProps} />)
    const unfiled = document.querySelector('.folder-unfiled')!
    fireEvent.drop(unfiled, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'session'
          if (k === 'itemId') return 's1'
          return ''
        },
      },
    })
    expect(baseProps.onAssignItem).toHaveBeenCalledWith('s1', null)
  })

  it('drop on unfiled area with mismatched itemType does NOT call onAssignItem', () => {
    render(<FolderSection {...baseProps} itemType="note" />)
    const unfiled = document.querySelector('.folder-unfiled')!
    fireEvent.drop(unfiled, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'session'
          if (k === 'itemId') return 's1'
          return ''
        },
      },
    })
    expect(baseProps.onAssignItem).not.toHaveBeenCalled()
  })

  it('renders note itemType label correctly in context menu', () => {
    render(<FolderSection {...baseProps} itemType="note" />)
    fireEvent.contextMenu(screen.getByText('SESSION FOLDERS').closest('.folder-section')!)
    expect(screen.getByText(/New Note/)).toBeInTheDocument()
  })

  it('items inside a folder are rendered when folder is expanded', () => {
    const sessionInFolder: Session = { ...sampleSession, folder_id: 'f1' }
    render(
      <FolderSection
        {...baseProps}
        folders={[sampleFolder]}
        items={[sessionInFolder]}
        expandedState={{ f1: true }}
      />
    )
    expect(screen.getByText('Session 1')).toBeInTheDocument()
  })

  it('unfiled count excludes items in folders', () => {
    const sessionInFolder: Session = { ...sampleSession, id: 's3', folder_id: 'f1' }
    render(
      <FolderSection
        {...baseProps}
        folders={[sampleFolder]}
        items={[sampleSession, sessionInFolder]}
      />
    )
    // sampleSession is unfiled (folder_id: null), sessionInFolder is in f1
    expect(screen.getByText('Unfiled (1)')).toBeInTheDocument()
  })

  it('right-click on FolderItem triggers context menu with folder id', () => {
    render(<FolderSection {...baseProps} folders={[sampleFolder]} />)
    fireEvent.contextMenu(screen.getByTestId('folder-item-f1'))
    // Context menu should appear (portalled to body)
    expect(screen.getByText(/New Session/)).toBeInTheDocument()
  })

  it('onItemDrop from FolderItem mock calls onAssignItem with folder id', () => {
    render(<FolderSection {...baseProps} folders={[sampleFolder]} />)
    fireEvent.click(screen.getByTestId('item-drop-f1'))
    expect(baseProps.onAssignItem).toHaveBeenCalledWith('item-in-folder', 'f1')
  })

  it('handleFolderDrop reorders folders when valid drag data provided', () => {
    const folder2: Folder = { ...sampleFolder, id: 'f2', name: 'Folder Two', position: 1 }
    render(<FolderSection {...baseProps} folders={[sampleFolder, folder2]} />)
    const folderItemDiv = screen.getByTestId('folder-item-f1')
    fireEvent.drop(folderItemDiv, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'dragType') return 'folder-reorder'
          if (k === 'folderId') return 'f2'
          return ''
        },
      },
    })
    expect(baseProps.onReorderFolders).toHaveBeenCalledWith(['f2', 'f1'])
  })

  it('handleFolderDrop does nothing when dragType is not folder-reorder', () => {
    render(<FolderSection {...baseProps} folders={[sampleFolder]} />)
    const folderItemDiv = screen.getByTestId('folder-item-f1')
    fireEvent.drop(folderItemDiv, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'dragType') return 'item-move'
          return ''
        },
      },
    })
    expect(baseProps.onReorderFolders).not.toHaveBeenCalled()
  })

  it('handleItemDragStart fires on unfiled item drag', () => {
    const mockSetData = vi.fn()
    const { container } = render(<FolderSection {...baseProps} items={[sampleSession]} />)
    const itemDiv = container.querySelector('.folder-content-item')!
    fireEvent.dragStart(itemDiv, {
      dataTransfer: {
        setData: mockSetData,
        setDragImage: vi.fn(),
        effectAllowed: '',
      },
    })
    expect(mockSetData).toHaveBeenCalledWith('itemType', 'session')
    expect(mockSetData).toHaveBeenCalledWith('itemId', 's1')
  })

  it('handleItemDragStart sets mapDropLabel to item title', () => {
    const mockSetData = vi.fn()
    const { container } = render(<FolderSection {...baseProps} items={[sampleSession]} />)
    const itemDiv = container.querySelector('.folder-content-item')!
    fireEvent.dragStart(itemDiv, {
      dataTransfer: {
        setData: mockSetData,
        setDragImage: vi.fn(),
        effectAllowed: '',
      },
    })
    expect(mockSetData).toHaveBeenCalledWith('mapDropLabel', 'Session 1')
  })

  it('context menu on FolderItem shows "in folder" hint', () => {
    render(<FolderSection {...baseProps} folders={[sampleFolder]} />)
    fireEvent.contextMenu(screen.getByTestId('folder-item-f1'))
    // The context menu has a "in folder" span when folderId is set
    expect(screen.getByText('in folder')).toBeInTheDocument()
  })

  it('Escape key on InlineNameInput cancels folder creation', () => {
    render(<FolderSection {...baseProps} />)
    fireEvent.click(screen.getByTitle('New folder'))
    const input = screen.getByTestId('inline-input')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
  })

  it('items in folder also render with click handler', () => {
    const sessionInFolder: Session = { ...sampleSession, id: 's3', folder_id: 'f1', title: 'Folder Session' }
    render(
      <FolderSection
        {...baseProps}
        folders={[sampleFolder]}
        items={[sessionInFolder]}
        expandedState={{ f1: true }}
      />
    )
    expect(screen.getByText('Folder Session')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Folder Session'))
    expect(baseProps.onItemClick).toHaveBeenCalledWith('s3')
  })

  it('note itemType renders note icon for unfiled items', () => {
    render(<FolderSection {...baseProps} items={[sampleNote]} itemType="note" />)
    // Items render with the note emoji icon
    expect(screen.getByText('Note 1')).toBeInTheDocument()
  })

  it('unfiled section chevron shows correct rotation when expanded', () => {
    const { container } = render(<FolderSection {...baseProps} expandedState={{}} />)
    const chevronSvg = container.querySelector('.folder-unfiled .folder-chevron svg')!
    expect(chevronSvg).toHaveStyle({ transform: 'rotate(90deg)' })
  })

  it('unfiled section chevron shows correct rotation when collapsed', () => {
    const { container } = render(<FolderSection {...baseProps} expandedState={{ 'unfiled-session': false }} />)
    const chevronSvg = container.querySelector('.folder-unfiled .folder-chevron svg')!
    expect(chevronSvg).toHaveStyle({ transform: 'rotate(0deg)' })
  })
})

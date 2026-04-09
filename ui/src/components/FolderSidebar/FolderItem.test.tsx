import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import FolderItem from './FolderItem'
import type { Folder } from '../../types/folder'

vi.mock('./InlineNameInput', () => ({
  default: ({ value, onCommit, onCancel, error }: any) => (
    <div data-testid="inline-name-input">
      <input
        data-testid="inline-input"
        defaultValue={value}
        onKeyDown={(e: any) => {
          if (e.key === 'Enter') onCommit(e.target.value || value || 'renamed')
          if (e.key === 'Escape') onCancel()
        }}
      />
      <button data-testid="inline-cancel" onClick={onCancel}>Cancel</button>
      {error && <span data-testid="rename-error">{error}</span>}
    </div>
  ),
}))

const sampleFolder: Folder = {
  id: 'f1',
  name: 'My Folder',
  game_id: 'g1',
  user_id: null,
  folder_type: 'session',
  visibility: 'game-wide',
  position: 0,
  created_at: '',
  updated_at: '',
}

const baseProps = {
  folder: sampleFolder,
  isExpanded: true,
  onToggle: vi.fn(),
  onRename: vi.fn().mockResolvedValue(null),
  onDelete: vi.fn(),
  isReadOnly: false,
  isEmpty: false,
  onDragStart: vi.fn(),
  onDragOver: vi.fn(),
  onDrop: vi.fn(),
  onItemDrop: vi.fn(),
  onFolderContextMenu: vi.fn(),
}

describe('FolderItem', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    baseProps.onRename.mockResolvedValue(null)
  })

  it('renders the folder name', () => {
    render(<FolderItem {...baseProps} />)
    expect(screen.getByText('My Folder')).toBeInTheDocument()
  })

  it('clicking the chevron button calls onToggle', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Collapse'))
    expect(baseProps.onToggle).toHaveBeenCalledTimes(1)
  })

  it('chevron aria-label is "Collapse" when expanded', () => {
    render(<FolderItem {...baseProps} isExpanded={true} />)
    expect(screen.getByLabelText('Collapse')).toBeInTheDocument()
  })

  it('chevron aria-label is "Expand" when collapsed', () => {
    render(<FolderItem {...baseProps} isExpanded={false} />)
    expect(screen.getByLabelText('Expand')).toBeInTheDocument()
  })

  it('renders children when isExpanded is true', () => {
    render(
      <FolderItem {...baseProps} isExpanded={true}>
        <div data-testid="child-content">Child</div>
      </FolderItem>
    )
    expect(screen.getByTestId('child-content')).toBeInTheDocument()
  })

  it('does not render children when isExpanded is false', () => {
    render(
      <FolderItem {...baseProps} isExpanded={false}>
        <div data-testid="child-content">Child</div>
      </FolderItem>
    )
    expect(screen.queryByTestId('child-content')).not.toBeInTheDocument()
  })

  it('shows empty indicator when isEmpty=true and isExpanded=false', () => {
    render(<FolderItem {...baseProps} isEmpty={true} isExpanded={false} />)
    expect(screen.getByTitle('Empty folder')).toBeInTheDocument()
  })

  it('does not show empty indicator when isEmpty=true but isExpanded=true', () => {
    render(<FolderItem {...baseProps} isEmpty={true} isExpanded={true} />)
    expect(screen.queryByTitle('Empty folder')).not.toBeInTheDocument()
  })

  it('shows lock icon when folder visibility is private', () => {
    const privateFolder: Folder = { ...sampleFolder, visibility: 'private' }
    render(<FolderItem {...baseProps} folder={privateFolder} />)
    expect(screen.getByTitle('Private')).toBeInTheDocument()
  })

  it('does not show lock icon when folder visibility is game-wide', () => {
    render(<FolderItem {...baseProps} />)
    expect(screen.queryByTitle('Private')).not.toBeInTheDocument()
  })

  it('renders menu button with aria-label "Folder options" when not read-only', () => {
    render(<FolderItem {...baseProps} isReadOnly={false} />)
    expect(screen.getByLabelText('Folder options')).toBeInTheDocument()
  })

  it('does not render menu button when read-only', () => {
    render(<FolderItem {...baseProps} isReadOnly={true} />)
    expect(screen.queryByLabelText('Folder options')).not.toBeInTheDocument()
  })

  it('clicking menu button opens dropdown with Rename and Delete', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    expect(screen.getByText('Rename')).toBeInTheDocument()
    expect(screen.getByText('Delete')).toBeInTheDocument()
  })

  it('clicking Rename in menu shows InlineNameInput', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    fireEvent.click(screen.getByText('Rename'))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
    expect(screen.queryByText('My Folder')).not.toBeInTheDocument()
  })

  it('InlineNameInput commit success hides input and shows folder name', async () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    fireEvent.click(screen.getByText('Rename'))
    const input = screen.getByTestId('inline-input')
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => {
      expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
    })
    expect(baseProps.onRename).toHaveBeenCalled()
  })

  it('InlineNameInput commit with error shows rename error', async () => {
    baseProps.onRename.mockResolvedValue('Name is taken')
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    fireEvent.click(screen.getByText('Rename'))
    const input = screen.getByTestId('inline-input')
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => {
      expect(screen.getByTestId('rename-error')).toHaveTextContent('Name is taken')
    })
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
  })

  it('InlineNameInput cancel hides input and restores folder name', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    fireEvent.click(screen.getByText('Rename'))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
    fireEvent.click(screen.getByTestId('inline-cancel'))
    expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
    expect(screen.getByText('My Folder')).toBeInTheDocument()
  })

  it('clicking Delete in menu calls onDelete', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    fireEvent.click(screen.getByText('Delete'))
    expect(baseProps.onDelete).toHaveBeenCalledTimes(1)
  })

  it('double-clicking folder name when not read-only shows InlineNameInput', () => {
    render(<FolderItem {...baseProps} isReadOnly={false} />)
    fireEvent.doubleClick(screen.getByText('My Folder'))
    expect(screen.getByTestId('inline-name-input')).toBeInTheDocument()
  })

  it('double-clicking folder name when read-only does NOT show InlineNameInput', () => {
    render(<FolderItem {...baseProps} isReadOnly={true} />)
    fireEvent.doubleClick(screen.getByText('My Folder'))
    expect(screen.queryByTestId('inline-name-input')).not.toBeInTheDocument()
  })

  it('dragOver on folder-item adds drag-over class', () => {
    render(<FolderItem {...baseProps} />)
    const folderItem = document.querySelector('.folder-item')!
    fireEvent.dragOver(folderItem)
    expect(folderItem.classList.contains('drag-over')).toBe(true)
  })

  it('handleItemDrop with valid session type calls onItemDrop', () => {
    render(<FolderItem {...baseProps} />)
    const folderItem = document.querySelector('.folder-item')!
    fireEvent.drop(folderItem, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'session'
          if (k === 'itemId') return 'item-1'
          return ''
        },
      },
    })
    expect(baseProps.onItemDrop).toHaveBeenCalledWith('session', 'item-1')
  })

  it('handleItemDrop with valid note type calls onItemDrop', () => {
    render(<FolderItem {...baseProps} />)
    const folderItem = document.querySelector('.folder-item')!
    fireEvent.drop(folderItem, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'note'
          if (k === 'itemId') return 'note-1'
          return ''
        },
      },
    })
    expect(baseProps.onItemDrop).toHaveBeenCalledWith('note', 'note-1')
  })

  it('handleItemDrop with invalid itemType does NOT call onItemDrop', () => {
    render(<FolderItem {...baseProps} />)
    const folderItem = document.querySelector('.folder-item')!
    fireEvent.drop(folderItem, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'folder'
          if (k === 'itemId') return 'f1'
          return ''
        },
      },
    })
    expect(baseProps.onItemDrop).not.toHaveBeenCalled()
  })

  it('handleItemDrop with no itemId does NOT call onItemDrop', () => {
    render(<FolderItem {...baseProps} />)
    const folderItem = document.querySelector('.folder-item')!
    fireEvent.drop(folderItem, {
      dataTransfer: {
        getData: (k: string) => {
          if (k === 'itemType') return 'session'
          return ''
        },
      },
    })
    expect(baseProps.onItemDrop).not.toHaveBeenCalled()
  })

  it('outside mousedown closes the dropdown menu', () => {
    render(<FolderItem {...baseProps} />)
    fireEvent.click(screen.getByLabelText('Folder options'))
    expect(screen.getByText('Rename')).toBeInTheDocument()
    // Mousedown outside
    fireEvent.mouseDown(document.body)
    expect(screen.queryByText('Rename')).not.toBeInTheDocument()
  })

  it('folder-item-row is draggable when not read-only', () => {
    render(<FolderItem {...baseProps} isReadOnly={false} />)
    const row = document.querySelector('.folder-item-row')!
    expect(row).toHaveAttribute('draggable', 'true')
  })

  it('folder-item-row is not draggable when read-only', () => {
    render(<FolderItem {...baseProps} isReadOnly={true} />)
    const row = document.querySelector('.folder-item-row')!
    expect(row).toHaveAttribute('draggable', 'false')
  })

  it('onFolderContextMenu called when right-clicking the folder row', () => {
    render(<FolderItem {...baseProps} />)
    const row = document.querySelector('.folder-item-row')!
    fireEvent.contextMenu(row)
    expect(baseProps.onFolderContextMenu).toHaveBeenCalledTimes(1)
  })
})

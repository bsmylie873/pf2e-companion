import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import NoteCard from './NoteCard'

vi.mock('../../api/backup', () => ({
  exportNoteBackup: vi.fn(),
}))

vi.mock('../../utils/contentPreview', () => ({
  extractPreviewText: vi.fn().mockReturnValue('Preview text'),
}))

const mockNote = {
  id: 'note-1',
  title: 'My Private Note',
  visibility: 'private' as const,
  user_id: 'user-1',
  game_id: 'game-1',
  session_id: null,
  folder_id: null,
  content: null,
  version: 1,
  created_at: '2024-01-15T00:00:00Z',
  updated_at: '2024-01-15T00:00:00Z',
}

describe('NoteCard (list mode)', () => {
  it('should render the note title', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('My Private Note')).toBeInTheDocument()
  })

  it('should show Private visibility label', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('Private')).toBeInTheDocument()
  })

  it('should call onOpen when card is clicked', () => {
    const onOpen = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={onOpen}
      />
    )
    fireEvent.click(screen.getByRole('article'))
    expect(onOpen).toHaveBeenCalledWith(mockNote)
  })

  it('should show edit button when author', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit note' })).toBeInTheDocument()
  })

  it('should call onEdit when edit button is clicked', () => {
    const onEdit = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={onEdit}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Edit note' }))
    expect(onEdit).toHaveBeenCalledWith(mockNote)
  })

  it('should show delete button when author', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Delete note' })).toBeInTheDocument()
  })

  it('should call onDelete when delete button is clicked', () => {
    const onDelete = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={onDelete}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete note' }))
    expect(onDelete).toHaveBeenCalledWith(mockNote)
  })

  it('should not show edit/delete buttons when not author and not GM', () => {
    const viewOnlyNote = { ...mockNote, visibility: 'visible' as const }
    render(
      <NoteCard
        note={viewOnlyNote}
        isGM={false}
        isAuthor={false}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.queryByRole('button', { name: 'Edit note' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete note' })).not.toBeInTheDocument()
  })

  it('should show session title when provided', () => {
    render(
      <NoteCard
        note={mockNote}
        sessionTitle="Session #1 — The Beginning"
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('Session #1 — The Beginning')).toBeInTheDocument()
  })

  it('should show View Only label for visible notes', () => {
    const visibleNote = { ...mockNote, visibility: 'visible' as const }
    render(
      <NoteCard
        note={visibleNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('View Only')).toBeInTheDocument()
  })

  it('should show Editable label for editable notes', () => {
    const editableNote = { ...mockNote, visibility: 'editable' as const }
    render(
      <NoteCard
        note={editableNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('Editable')).toBeInTheDocument()
  })
})

describe('NoteCard (grid mode)', () => {
  it('should render in grid mode with preview text', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('My Private Note')).toBeInTheDocument()
    expect(screen.getByText('Preview text')).toBeInTheDocument()
  })

  it('should call onOpen when grid card body is clicked', () => {
    const onOpen = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={onOpen}
      />
    )
    fireEvent.click(screen.getByText('My Private Note'))
    expect(onOpen).toHaveBeenCalledWith(mockNote)
  })

  it('should show edit button in grid mode when author', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit note' })).toBeInTheDocument()
  })

  it('should call onEdit when grid edit button clicked', () => {
    const onEdit = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        mode="grid"
        onEdit={onEdit}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Edit note' }))
    expect(onEdit).toHaveBeenCalledWith(mockNote)
  })

  it('should show delete button in grid mode when GM', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={true}
        isAuthor={false}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Delete note' })).toBeInTheDocument()
  })

  it('should call onDelete when grid delete button clicked', () => {
    const onDelete = vi.fn()
    render(
      <NoteCard
        note={mockNote}
        isGM={true}
        isAuthor={false}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={onDelete}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete note' }))
    expect(onDelete).toHaveBeenCalledWith(mockNote)
  })

  it('should hide edit and delete buttons in grid mode when not author, not GM, visibility=private', () => {
    const privateNote = { ...mockNote, visibility: 'private' as const }
    render(
      <NoteCard
        note={privateNote}
        isGM={false}
        isAuthor={false}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.queryByRole('button', { name: 'Edit note' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete note' })).not.toBeInTheDocument()
  })

  it('should show edit button in grid mode for editable note when not author and not GM', () => {
    const editableNote = { ...mockNote, visibility: 'editable' as const }
    render(
      <NoteCard
        note={editableNote}
        isGM={false}
        isAuthor={false}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit note' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Delete note' })).not.toBeInTheDocument()
  })

  it('should render the formatted date in grid mode', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    // Date should be rendered inside a <time> element
    const timeEl = document.querySelector('time')
    expect(timeEl).toBeInTheDocument()
  })
})

describe('NoteCard — additional list mode coverage', () => {
  it('should show edit button when GM even if not author', () => {
    render(
      <NoteCard
        note={{ ...mockNote, visibility: 'private' as const }}
        isGM={true}
        isAuthor={false}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit note' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete note' })).toBeInTheDocument()
  })

  it('should show edit button for editable note when not author and not GM', () => {
    const editableNote = { ...mockNote, visibility: 'editable' as const }
    render(
      <NoteCard
        note={editableNote}
        isGM={false}
        isAuthor={false}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit note' })).toBeInTheDocument()
    // Non-author non-GM cannot delete
    expect(screen.queryByRole('button', { name: 'Delete note' })).not.toBeInTheDocument()
  })

  it('should have export button always visible', () => {
    render(
      <NoteCard
        note={mockNote}
        isGM={false}
        isAuthor={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Export note' })).toBeInTheDocument()
  })
})

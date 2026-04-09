import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import NoteFormModal from './NoteFormModal'

const mockSessions = [
  {
    id: 'session-1',
    title: 'The Beginning',
    session_number: 1,
    game_id: 'game-1',
    folder_id: null,
    notes: null,
    version: 1,
    scheduled_at: null,
    runtime_start: null,
    runtime_end: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

describe('NoteFormModal (create mode)', () => {
  it('should render "New Note" heading in create mode', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('New Note')).toBeInTheDocument()
  })

  it('should render title input', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByLabelText('Title *')).toBeInTheDocument()
  })

  it('should render session dropdown with options', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('#1 — The Beginning')).toBeInTheDocument()
  })

  it('should show visibility buttons when isAuthor is true', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        isAuthor={true}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: /Private/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /View Only/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Editable/ })).toBeInTheDocument()
  })

  it('should not show visibility buttons when isAuthor is false', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        isAuthor={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.queryByRole('button', { name: /Private/ })).not.toBeInTheDocument()
  })

  it('should call onSave with correct data on submit', () => {
    const onSave = vi.fn()
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.change(screen.getByLabelText('Title *'), { target: { value: 'My Test Note' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Note' }))
    expect(onSave).toHaveBeenCalledWith({
      title: 'My Test Note',
      session_id: null,
      visibility: 'private',
    })
  })

  it('should not call onSave when title is empty', () => {
    const onSave = vi.fn()
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={onSave}
        onClose={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Create Note' }))
    expect(onSave).not.toHaveBeenCalled()
  })

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn()
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={onClose}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should display error when error prop is provided', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error="Failed to save note"
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('Failed to save note')).toBeInTheDocument()
  })

  it('should disable submit button when saving', () => {
    render(
      <NoteFormModal
        mode="create"
        sessions={mockSessions}
        error={null}
        saving={true}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Saving…' })).toBeDisabled()
  })
})

describe('NoteFormModal (edit mode)', () => {
  it('should render "Edit Note" heading in edit mode', () => {
    render(
      <NoteFormModal
        mode="edit"
        initial={{ title: 'Existing Note', session_id: null, visibility: 'private' }}
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByText('Edit Note')).toBeInTheDocument()
  })

  it('should pre-fill title from initial data', () => {
    render(
      <NoteFormModal
        mode="edit"
        initial={{ title: 'Existing Note', session_id: null, visibility: 'private' }}
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument()
  })

  it('should show "Save Changes" button in edit mode', () => {
    render(
      <NoteFormModal
        mode="edit"
        initial={{ title: 'Existing Note', session_id: null, visibility: 'private' }}
        sessions={mockSessions}
        error={null}
        saving={false}
        onSave={vi.fn()}
        onClose={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Save Changes' })).toBeInTheDocument()
  })
})

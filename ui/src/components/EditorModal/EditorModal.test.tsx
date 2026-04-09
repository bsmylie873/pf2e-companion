import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import EditorModal from './EditorModal'

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'user-1', username: 'testuser', email: 'test@example.com' },
    isAuthenticated: true,
    isLoading: false,
  }),
}))

vi.mock('../../api/sessions', () => ({
  getSession: vi.fn(),
  updateSessionNotes: vi.fn(),
}))

vi.mock('../../api/notes', () => ({
  getNote: vi.fn(),
  updateNoteContent: vi.fn(),
}))

vi.mock('../../api/memberships', () => ({
  listMemberships: vi.fn(),
}))

vi.mock('../SessionNotesEditor/SessionNotesEditor', () => ({
  default: ({ editable }: { editable: boolean }) => (
    <div data-testid="session-notes-editor" data-editable={String(editable)} />
  ),
}))

import { getSession } from '../../api/sessions'
import { getNote } from '../../api/notes'
import { listMemberships } from '../../api/memberships'

const mockGetSession = getSession as ReturnType<typeof vi.fn>
const mockGetNote = getNote as ReturnType<typeof vi.fn>
const mockListMemberships = listMemberships as ReturnType<typeof vi.fn>

describe('EditorModal', () => {
  beforeEach(() => {
    mockGetSession.mockResolvedValue({
      id: 'session-1',
      title: 'The Dark Forest',
      session_number: 3,
      notes: null,
      version: 1,
    })
    mockGetNote.mockResolvedValue({
      id: 'note-1',
      title: 'My Note',
      visibility: 'private',
      user_id: 'user-1',
      content: null,
      version: 1,
    })
    mockListMemberships.mockResolvedValue([])
  })

  it('should render loading state initially for session type', () => {
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    // The portal renders in document.body
    expect(document.querySelector('.editor-modal-backdrop')).toBeInTheDocument()
  })

  it('should call getSession for session type', async () => {
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(mockGetSession).toHaveBeenCalledWith('session-1')
    })
  })

  it('should render session title after loading', async () => {
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByText('The Dark Forest')).toBeInTheDocument()
    })
  })

  it('should render session number after loading', async () => {
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByText('Session #3')).toBeInTheDocument()
    })
  })

  it('should call onClose when close button is clicked', async () => {
    const onClose = vi.fn()
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={onClose}
      />
    )
    await waitFor(() => screen.getByText('The Dark Forest'))
    fireEvent.click(screen.getByRole('button', { name: 'Close editor' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should call onClose when backdrop is clicked', async () => {
    const onClose = vi.fn()
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={onClose}
      />
    )
    await waitFor(() => screen.getByText('The Dark Forest'))
    fireEvent.click(document.querySelector('.editor-modal-backdrop')!)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should call onClose when Escape is pressed', async () => {
    const onClose = vi.fn()
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={onClose}
      />
    )
    await waitFor(() => screen.getByText('The Dark Forest'))
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('should call getNote and listMemberships for note type', async () => {
    render(
      <EditorModal
        type="note"
        itemId="note-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(mockGetNote).toHaveBeenCalledWith('note-1')
      expect(mockListMemberships).toHaveBeenCalledWith('game-1')
    })
  })

  it('should show error when session fetch fails', async () => {
    mockGetSession.mockRejectedValue(new Error('Network error'))
    render(
      <EditorModal
        type="session"
        itemId="session-1"
        gameId="game-1"
        onClose={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument()
    })
  })
})

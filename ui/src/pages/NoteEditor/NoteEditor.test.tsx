import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import NoteEditor from './NoteEditor'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ gameId: 'game-1', noteId: 'note-1' }),
  }
})

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'user-1', username: 'testuser', email: 'test@test.com' },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
    refreshUser: vi.fn(),
  }),
}))

const mockGetNote = vi.fn()

vi.mock('../../api/notes', () => ({
  getNote: (...args: unknown[]) => mockGetNote(...args),
  updateNoteContent: vi.fn().mockResolvedValue({}),
  listGameNotesPaginated: vi.fn().mockResolvedValue({ data: [], total: 0 }),
  listGameNotes: vi.fn().mockResolvedValue([]),
  createNote: vi.fn().mockResolvedValue({}),
  updateNote: vi.fn().mockResolvedValue({}),
  deleteNote: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/memberships', () => ({
  listMemberships: vi.fn().mockResolvedValue([]),
}))

vi.mock('../../hooks/useGameSocket', () => ({
  useGameSocket: vi.fn(),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

vi.mock('@tiptap/react', () => ({
  useEditor: () => null,
  EditorContent: () => <div data-testid="editor" />,
}))

vi.mock('../../components/SessionNotesEditor/SessionNotesEditor', () => ({
  default: () => <div data-testid="session-notes-editor" />,
}))

function renderNoteEditor() {
  return render(
    <MemoryRouter>
      <NoteEditor />
    </MemoryRouter>,
  )
}

describe('NoteEditor', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockGetNote.mockReset()
  })

  it('should show loading state initially', () => {
    mockGetNote.mockReturnValue(new Promise(() => {}))
    renderNoteEditor()
    expect(screen.getByText(/unfurling the scroll/i)).toBeInTheDocument()
  })

  it('should render note title after loading', async () => {
    mockGetNote.mockResolvedValue({
      id: 'note-1',
      title: 'My Adventure Notes',
      content: null,
      visibility: 'private',
      user_id: 'user-1',
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderNoteEditor()

    await waitFor(() => {
      expect(screen.getByText('My Adventure Notes')).toBeInTheDocument()
    })
  })

  it('should render visibility badge', async () => {
    mockGetNote.mockResolvedValue({
      id: 'note-1',
      title: 'Private Note',
      content: null,
      visibility: 'private',
      user_id: 'user-1',
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderNoteEditor()

    await waitFor(() => {
      // Visibility badge renders as "🔒 Private" — match the span by class
      const badges = document.querySelectorAll('.nep-visibility')
      expect(badges.length).toBeGreaterThan(0)
    })
  })

  it('should render the editor after loading', async () => {
    mockGetNote.mockResolvedValue({
      id: 'note-1',
      title: 'Test Note',
      content: null,
      visibility: 'editable',
      user_id: 'user-1',
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderNoteEditor()

    await waitFor(() => {
      expect(screen.getByTestId('session-notes-editor')).toBeInTheDocument()
    })
  })

  it('should show error when note fetch fails', async () => {
    mockGetNote.mockRejectedValue(new Error('Note not found'))
    renderNoteEditor()

    await waitFor(() => {
      expect(screen.getByText('Note not found')).toBeInTheDocument()
    })
  })

  it('should navigate back when back button clicked', async () => {
    const user = userEvent.setup()
    mockGetNote.mockResolvedValue({
      id: 'note-1',
      title: 'Test Note',
      content: null,
      visibility: 'private',
      user_id: 'user-1',
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderNoteEditor()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /back to notes/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /back to notes/i }))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1')
  })
})

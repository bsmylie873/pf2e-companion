import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import SessionNotes from './SessionNotes'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ gameId: 'game-1', sessionId: 'session-1' }),
  }
})

const mockGetSession = vi.fn()

vi.mock('../../api/sessions', () => ({
  getSession: (...args: unknown[]) => mockGetSession(...args),
  updateSessionNotes: vi.fn().mockResolvedValue({}),
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

function renderSessionNotes() {
  return render(
    <MemoryRouter>
      <SessionNotes />
    </MemoryRouter>,
  )
}

describe('SessionNotes', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockGetSession.mockReset()
  })

  it('should show loading state initially', () => {
    mockGetSession.mockReturnValue(new Promise(() => {}))
    renderSessionNotes()
    expect(screen.getByText(/unfurling the scroll/i)).toBeInTheDocument()
  })

  it('should render session title after loading', async () => {
    mockGetSession.mockResolvedValue({
      id: 'session-1',
      title: 'The Dark Forest',
      session_number: 3,
      notes: null,
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderSessionNotes()

    await waitFor(() => {
      expect(screen.getByText('The Dark Forest')).toBeInTheDocument()
    })
  })

  it('should show session number when present', async () => {
    mockGetSession.mockResolvedValue({
      id: 'session-1',
      title: 'The Dark Forest',
      session_number: 3,
      notes: null,
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderSessionNotes()

    await waitFor(() => {
      expect(screen.getByText(/session #3/i)).toBeInTheDocument()
    })
  })

  it('should render the editor after loading', async () => {
    mockGetSession.mockResolvedValue({
      id: 'session-1',
      title: 'Test Session',
      session_number: null,
      notes: null,
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderSessionNotes()

    await waitFor(() => {
      expect(screen.getByTestId('session-notes-editor')).toBeInTheDocument()
    })
  })

  it('should show error when session fetch fails', async () => {
    mockGetSession.mockRejectedValue(new Error('Session not found'))
    renderSessionNotes()

    await waitFor(() => {
      expect(screen.getByText('Session not found')).toBeInTheDocument()
    })
  })

  it('should navigate back to game when back button clicked', async () => {
    const user = userEvent.setup()
    mockGetSession.mockResolvedValue({
      id: 'session-1',
      title: 'Test Session',
      session_number: 1,
      notes: null,
      version: 1,
      game_id: 'game-1',
      created_at: '',
      updated_at: '',
    })
    renderSessionNotes()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /back to sessions/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /back to sessions/i }))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1')
  })
})

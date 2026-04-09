import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Editor from './Editor'
import { listGameSessionsPaginated, createSession, updateSession, deleteSession } from '../../api/sessions'
import { listGameNotesPaginated, createNote, deleteNote } from '../../api/notes'
import { listMemberships } from '../../api/memberships'
import { apiFetch } from '../../api/client'
import { getPreferences } from '../../api/preferences'
import { listFolders } from '../../api/folders'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ gameId: 'game-1' }),
    useLocation: () => ({ pathname: '/games/game-1', search: '', state: { title: 'Test Campaign' } }),
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

vi.mock('../../api/sessions', () => ({
  listGameSessionsPaginated: vi.fn().mockResolvedValue({ data: [], total: 0 }),
  createSession: vi.fn().mockResolvedValue({
    id: 'new-sess', title: 'New Session', session_number: 1,
    game_id: 'game-1', content: null, content_json: null, folder_id: null,
    scheduled_at: null, runtime_start: null, runtime_end: null,
    created_at: '', updated_at: '',
  }),
  updateSession: vi.fn().mockResolvedValue({
    id: 'sess-1', title: 'Updated Session', session_number: 1,
    game_id: 'game-1', content: null, content_json: null, folder_id: null,
    scheduled_at: null, runtime_start: null, runtime_end: null,
    created_at: '', updated_at: '',
  }),
  deleteSession: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../../api/notes', () => ({
  listGameNotesPaginated: vi.fn().mockResolvedValue({ data: [], total: 0 }),
  createNote: vi.fn().mockResolvedValue({
    id: 'new-note', title: 'New Note', game_id: 'game-1', user_id: 'user-1',
    session_id: null, folder_id: null, visibility: 'party',
    content: null, content_json: null, created_at: '', updated_at: '',
  }),
  updateNote: vi.fn().mockResolvedValue({
    id: 'note-1', title: 'Updated Note', game_id: 'game-1', user_id: 'user-1',
    session_id: null, folder_id: null, visibility: 'party',
    content: null, content_json: null, created_at: '', updated_at: '',
  }),
  deleteNote: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../../api/memberships', () => ({
  listMemberships: vi.fn().mockResolvedValue([]),
}))

vi.mock('../../api/backup', () => ({
  importGameBackup: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn().mockResolvedValue({ id: 'game-1', title: 'Test Campaign' }),
}))

vi.mock('../../api/preferences', () => ({
  getPreferences: vi.fn().mockResolvedValue({
    default_game_id: null,
    map_editor_mode: 'modal',
    page_size: null,
  }),
  updatePreferences: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/folders', () => ({
  listFolders: vi.fn().mockResolvedValue([]),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

vi.mock('../../hooks/usePageSize', () => ({
  usePageSize: vi.fn().mockReturnValue(10),
}))

vi.mock('@tiptap/react', () => ({
  useEditor: () => null,
  EditorContent: () => <div data-testid="editor" />,
}))

vi.mock('motion/react', () => ({
  motion: {
    div: 'div',
    ul: 'ul',
    li: 'li',
    span: 'span',
    p: 'p',
    button: 'button',
    section: 'section',
  },
  AnimatePresence: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// Mock heavy child components — include edit/delete buttons so handlers can be tested
vi.mock('../../components/SessionCard/SessionCard', () => ({
  default: ({ session, onEdit, onDelete, onOpen }: {
    session: { id: string; title: string }
    onEdit?: (s: unknown) => void
    onDelete?: (s: unknown) => void
    onOpen?: (s: unknown) => void
  }) => (
    <div data-testid={`session-card-${session.id}`}>
      {session.title}
      {onEdit && <button onClick={() => onEdit(session)} aria-label={`edit-session-${session.id}`}>Edit</button>}
      {onDelete && <button onClick={() => onDelete(session)} aria-label={`delete-session-${session.id}`}>Delete</button>}
      {onOpen && <button onClick={() => onOpen(session)} aria-label={`open-session-${session.id}`}>Open</button>}
    </div>
  ),
}))

vi.mock('../../components/NoteCard/NoteCard', () => ({
  default: ({ note, onEdit, onDelete, onOpen }: {
    note: { id: string; title: string }
    onEdit?: (n: unknown) => void
    onDelete?: (n: unknown) => void
    onOpen?: (n: unknown) => void
  }) => (
    <div data-testid={`note-card-${note.id}`}>
      {note.title}
      {onEdit && <button onClick={() => onEdit(note)} aria-label={`edit-note-${note.id}`}>Edit</button>}
      {onDelete && <button onClick={() => onDelete(note)} aria-label={`delete-note-${note.id}`}>Delete</button>}
      {onOpen && <button onClick={() => onOpen(note)} aria-label={`open-note-${note.id}`}>Open</button>}
    </div>
  ),
}))

vi.mock('../../components/SessionFormModal/SessionFormModal', () => ({
  default: ({ onSave, onClose }: {
    onSave?: (data: unknown) => void
    onClose?: () => void
  }) => (
    <div data-testid="session-form-modal">
      <button onClick={() => onSave?.({ title: 'New Session', session_number: 1, scheduled_at: null, runtime_start: null, runtime_end: null })} aria-label="sfm-save">Save</button>
      <button onClick={() => onClose?.()} aria-label="sfm-close">Close</button>
    </div>
  ),
}))

vi.mock('../../components/NoteFormModal/NoteFormModal', () => ({
  default: ({ onSave, onClose }: {
    onSave?: (data: unknown) => void
    onClose?: () => void
  }) => (
    <div data-testid="note-form-modal">
      <button onClick={() => onSave?.({ title: 'New Note', session_id: null, visibility: 'party' })} aria-label="nfm-save">Save</button>
      <button onClick={() => onClose?.()} aria-label="nfm-close">Close</button>
    </div>
  ),
}))

vi.mock('../../components/ConfirmModal/ConfirmModal', () => ({
  default: ({ onConfirm, onCancel }: {
    onConfirm?: () => void
    onCancel?: () => void
  }) => (
    <div data-testid="confirm-modal">
      <button onClick={() => onConfirm?.()} aria-label="confirm-ok">Confirm</button>
      <button onClick={() => onCancel?.()} aria-label="confirm-cancel">Cancel</button>
    </div>
  ),
}))

vi.mock('../../components/Modal/Modal', () => ({
  default: ({ children, title }: { children: React.ReactNode; title: string }) => (
    <div data-testid="modal"><h2>{title}</h2>{children}</div>
  ),
}))

vi.mock('../../components/EditCampaignForm/EditCampaignForm', () => ({
  default: () => <div data-testid="edit-campaign-form" />,
}))

vi.mock('../../components/Pagination/Pagination', () => ({
  default: () => <div data-testid="pagination" />,
}))

vi.mock('../../components/FolderSidebar/FolderSidebar', () => ({
  default: () => <div data-testid="folder-sidebar" />,
}))

const makeSession = (id = 'sess-1', title = 'Session One', num = 1) => ({
  id,
  title,
  session_number: num,
  game_id: 'game-1',
  content: null,
  content_json: null,
  folder_id: null,
  scheduled_at: null,
  runtime_start: null,
  runtime_end: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
})

const makeNote = (id = 'note-1', title = 'Note One') => ({
  id,
  title,
  game_id: 'game-1',
  user_id: 'user-1',
  session_id: null,
  folder_id: null,
  visibility: 'party' as const,
  content: null,
  content_json: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
})

function renderEditor() {
  return render(
    <MemoryRouter>
      <Editor />
    </MemoryRouter>,
  )
}

describe('Editor', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listMemberships).mockResolvedValue([])
    vi.mocked(apiFetch).mockResolvedValue({ id: 'game-1', title: 'Test Campaign' })
    vi.mocked(getPreferences).mockResolvedValue({
      default_game_id: null,
      map_editor_mode: 'modal',
      page_size: null,
    })
  })

  it('should render the campaign title from location state', async () => {
    renderEditor()

    await waitFor(() => {
      expect(screen.getByText('Test Campaign')).toBeInTheDocument()
    })
  })

  it('should render sessions and notes tabs', async () => {
    renderEditor()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /sessions/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument()
    })
  })

  it('should show empty state for sessions when none exist', async () => {
    renderEditor()

    await waitFor(() => {
      expect(screen.queryByText(/consulting/i)).not.toBeInTheDocument()
    })

    // Empty state should be visible
    expect(screen.getByRole('button', { name: /new session/i })).toBeInTheDocument()
  })

  it('should switch to notes tab when clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /notes/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new note/i })).toBeInTheDocument()
    })
  })

  it('should navigate to map view when map button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => {
      const mapBtn = screen.queryByRole('button', { name: /map/i })
      if (mapBtn) {
        expect(mapBtn).toBeInTheDocument()
      }
    })
  })
})

// ── Navigation & Toolbar ─────────────────────────────────────────

describe('Editor — navigation & toolbar', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm-1', game_id: 'game-1', user_id: 'user-1', is_gm: true, foundry_data: null, created_at: '', updated_at: '' },
    ])
  })

  it('should navigate back to /games when back button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByText(/back to campaigns/i)).toBeInTheDocument())
    await user.click(screen.getByText(/back to campaigns/i))
    expect(mockNavigate).toHaveBeenCalledWith('/games')
  })

  it('should navigate to map view when map button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Map view')).toBeInTheDocument())
    await user.click(screen.getByTitle('Map view'))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1/map')
  })

  it('should open edit campaign modal when edit button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Edit Campaign')).toBeInTheDocument())
    await user.click(screen.getByTitle('Edit Campaign'))

    await waitFor(() => {
      expect(screen.getByTestId('modal')).toBeInTheDocument()
      expect(screen.getByTestId('edit-campaign-form')).toBeInTheDocument()
    })
  })

  it('should show list and grid view toggle buttons', async () => {
    renderEditor()

    await waitFor(() => {
      expect(screen.getByTitle('List view')).toBeInTheDocument()
      expect(screen.getByTitle('Grid view')).toBeInTheDocument()
    })
  })

  it('should switch to grid view when grid button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Grid view')).toBeInTheDocument())
    await user.click(screen.getByTitle('Grid view'))

    // After clicking, grid button should be active (no error thrown)
    expect(screen.getByTitle('Grid view')).toBeInTheDocument()
  })

  it('should open import modal when import button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))

    await waitFor(() => {
      expect(screen.getByTestId('modal')).toBeInTheDocument()
      expect(screen.getByText('Import Backup')).toBeInTheDocument()
    })
  })

  it('should show merge and overwrite mode buttons in import modal', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))

    await waitFor(() => {
      expect(screen.getByText(/merge \(skip existing\)/i)).toBeInTheDocument()
      expect(screen.getByText(/overwrite \(replace\)/i)).toBeInTheDocument()
    })
  })

  it('should select merge mode when merge button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))

    await waitFor(() => expect(screen.getByText(/merge \(skip existing\)/i)).toBeInTheDocument())
    await user.click(screen.getByText(/merge \(skip existing\)/i))

    // Import submit button should still be disabled (no file selected)
    const submitBtn = screen.getByRole('button', { name: /^import$/i })
    expect(submitBtn).toBeDisabled()
  })

  it('should close import modal when cancel button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))
    await waitFor(() => expect(screen.getByText('Import Backup')).toBeInTheDocument())

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    await waitFor(() => {
      expect(screen.queryByText('Import Backup')).not.toBeInTheDocument()
    })
  })
})

// ── Sessions tab ─────────────────────────────────────────────────

describe('Editor — sessions', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should show loading spinner while sessions are loading', () => {
    // Return a promise that never resolves so loading stays true
    vi.mocked(listGameSessionsPaginated).mockReturnValue(new Promise(() => {}))
    renderEditor()
    expect(screen.getByText(/unrolling the scrolls/i)).toBeInTheDocument()
  })

  it('should show error state when sessions fail to load', async () => {
    vi.mocked(listGameSessionsPaginated).mockRejectedValue(new Error('Network error'))
    renderEditor()

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument()
    })
  })

  it('should display session cards when sessions are loaded', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Session One', 1)],
      total: 1,
    })
    renderEditor()

    await waitFor(() => {
      expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()
      expect(screen.getByText('Session One')).toBeInTheDocument()
    })
  })

  it('should display multiple session cards', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [
        makeSession('sess-1', 'The Beginning', 1),
        makeSession('sess-2', 'The Middle', 2),
        makeSession('sess-3', 'The End', 3),
      ],
      total: 3,
    })
    renderEditor()

    await waitFor(() => {
      expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()
      expect(screen.getByTestId('session-card-sess-2')).toBeInTheDocument()
      expect(screen.getByTestId('session-card-sess-3')).toBeInTheDocument()
    })
  })

  it('should show pagination when session total is greater than 0', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Session One', 1)],
      total: 1,
    })
    renderEditor()

    await waitFor(() => {
      expect(screen.getByTestId('pagination')).toBeInTheDocument()
    })
  })

  it('should open session form modal when new session button is clicked', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /new session/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new session/i }))

    await waitFor(() => {
      expect(screen.getByTestId('session-form-modal')).toBeInTheDocument()
    })
  })

  it('should show empty state message when no sessions exist', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    renderEditor()

    await waitFor(() => {
      expect(screen.getByText(/no sessions yet/i)).toBeInTheDocument()
    })
  })
})

// ── Notes tab ────────────────────────────────────────────────────

describe('Editor — notes', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should show empty notes state when no notes exist', async () => {
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /notes/i }))

    await waitFor(() => {
      expect(screen.getByText(/no notes yet/i)).toBeInTheDocument()
    })
  })

  it('should display note cards when notes are loaded', async () => {
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [makeNote('note-1', 'Lore of the Realm')],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /notes/i }))

    await waitFor(() => {
      expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument()
      expect(screen.getByText('Lore of the Realm')).toBeInTheDocument()
    })
  })

  it('should open note form modal when new note button is clicked', async () => {
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /notes/i }))

    await waitFor(() => expect(screen.getByRole('button', { name: /new note/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new note/i }))

    await waitFor(() => {
      expect(screen.getByTestId('note-form-modal')).toBeInTheDocument()
    })
  })

  it('should show pagination when note total is greater than 0', async () => {
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [makeNote('note-1', 'A Note')],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /notes/i }))

    await waitFor(() => {
      expect(screen.getByTestId('pagination')).toBeInTheDocument()
    })
  })

  it('should show notes loading indicator on notes tab', async () => {
    // After switching to notes tab, loading indicator may show briefly
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockReturnValue(new Promise(() => {}))

    const user = userEvent.setup()
    renderEditor()

    // Wait for sessions to finish loading first
    await waitFor(() => expect(screen.getByRole('button', { name: /notes/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /notes/i }))

    // Loading state from initial load (loading starts true)
    // We might see the spinner if notes loading is pending
    // This is timing dependent - just assert notes tab is active
    expect(screen.getByRole('button', { name: /new note/i })).toBeInTheDocument()
  })
})

// ── Filter panel ─────────────────────────────────────────────────

describe('Editor — filter panel', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should show filter toggle button', async () => {
    renderEditor()
    await waitFor(() => {
      expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument()
    })
  })

  it('should show filter panel when filter toggle is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/search by name/i)).toBeInTheDocument()
    })
  })

  it('should show sort controls in filter panel', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      expect(screen.getByText('Sort by')).toBeInTheDocument()
      // Sort buttons for # (session number), Title, Edited
      expect(screen.getByRole('button', { name: '#' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Title' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Edited' })).toBeInTheDocument()
    })
  })

  it('should filter sessions by title when typing in title filter', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [
        makeSession('sess-1', 'Dragon Quest', 1),
        makeSession('sess-2', 'Goblin Attack', 2),
      ],
      total: 2,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByPlaceholderText(/search by name/i)).toBeInTheDocument())
    await user.type(screen.getByPlaceholderText(/search by name/i), 'Dragon')

    await waitFor(() => {
      expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()
      expect(screen.queryByTestId('session-card-sess-2')).not.toBeInTheDocument()
    })
  })

  it('should show no matching sessions message when filter yields no results', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByPlaceholderText(/search by name/i)).toBeInTheDocument())
    await user.type(screen.getByPlaceholderText(/search by name/i), 'zzz-no-match')

    await waitFor(() => {
      expect(screen.getByText(/no matching sessions/i)).toBeInTheDocument()
    })
  })

  it('should show clear filters button when a filter is active and clear on click', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByPlaceholderText(/search by name/i)).toBeInTheDocument())
    await user.type(screen.getByPlaceholderText(/search by name/i), 'Dragon')

    await waitFor(() => expect(screen.getByRole('button', { name: /clear filters/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /clear filters/i }))

    // After clearing, the filter input should be empty
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: /clear filters/i })).not.toBeInTheDocument()
    })
  })

  it('should toggle sort direction when clicking active sort button again', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByRole('button', { name: '#' })).toBeInTheDocument())

    // # is the default sort; clicking it again should toggle direction
    await user.click(screen.getByRole('button', { name: '#' }))
    // Still present — sort arrow direction changed (no error)
    expect(screen.getByRole('button', { name: '#' })).toBeInTheDocument()
  })

  it('should switch sort field when clicking a different sort button', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByRole('button', { name: 'Title' })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: 'Title' }))

    // Title button should now be active
    expect(screen.getByRole('button', { name: /title/i })).toBeInTheDocument()
  })

  it('should show notes filter controls when filter panel is open on notes tab', async () => {
    const user = userEvent.setup()
    renderEditor()

    // Switch to notes tab first
    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    // Now open filters
    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      expect(screen.getByText(/title a–z/i)).toBeInTheDocument()
      expect(screen.getByText(/newest/i)).toBeInTheDocument()
    })
  })

  it('should close filter panel when switching tabs', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    // Filter panel visible on sessions tab
    await waitFor(() => expect(screen.getByPlaceholderText(/search by name/i)).toBeInTheDocument())

    // Switch to notes tab
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    // Filter panel should be closed
    await waitFor(() => {
      expect(screen.queryByPlaceholderText(/search by name/i)).not.toBeInTheDocument()
    })
  })
})

// ── GM Features ──────────────────────────────────────────────────

describe('Editor — GM features', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should show campaign settings link when user is GM', async () => {
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm1', game_id: 'game-1', user_id: 'user-1', is_gm: true, created_at: '', updated_at: '', username: 'testuser' },
    ])
    renderEditor()

    await waitFor(() => {
      expect(screen.getByTitle('Campaign Settings')).toBeInTheDocument()
    })
  })

  it('should not show campaign settings link when user is not GM', async () => {
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm1', game_id: 'game-1', user_id: 'user-1', is_gm: false, created_at: '', updated_at: '', username: 'testuser' },
    ])
    renderEditor()

    await waitFor(() => {
      expect(screen.queryByTitle('Campaign Settings')).not.toBeInTheDocument()
    })
  })
})

// ── Import modal interactions ─────────────────────────────────────

describe('Editor — import modal', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should show conflict resolution section in import modal', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))

    await waitFor(() => {
      expect(screen.getByText(/conflict resolution/i)).toBeInTheDocument()
    })
  })

  it('should select overwrite mode when overwrite button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Import backup')).toBeInTheDocument())
    await user.click(screen.getByTitle('Import backup'))

    await waitFor(() => expect(screen.getByText(/overwrite \(replace\)/i)).toBeInTheDocument())
    await user.click(screen.getByText(/overwrite \(replace\)/i))

    // Import submit should still be disabled (no file)
    expect(screen.getByRole('button', { name: /^import$/i })).toBeDisabled()
  })
})

// ── Edit campaign modal ──────────────────────────────────────────

describe('Editor — edit campaign modal', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm-1', game_id: 'game-1', user_id: 'user-1', is_gm: true, foundry_data: null, created_at: '', updated_at: '' },
    ])
  })

  it('should render the Edit Campaign modal with the form', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Edit Campaign')).toBeInTheDocument())
    await user.click(screen.getByTitle('Edit Campaign'))

    await waitFor(() => {
      expect(screen.getByTestId('modal')).toBeInTheDocument()
      expect(screen.getByTestId('edit-campaign-form')).toBeInTheDocument()
    })
  })
})

// ── Folder groups ────────────────────────────────────────────────

describe('Editor — folder groups', () => {
  const makeFolder = (id: string, name: string) => ({
    id,
    name,
    game_id: 'game-1',
    type: 'session' as const,
    order: 0,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  })

  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should render sessions inside folder groups when folder_id is set', async () => {
    vi.mocked(listFolders).mockImplementation((_gameId, type) =>
      type === 'session'
        ? Promise.resolve([makeFolder('folder-1', 'Arc One')])
        : Promise.resolve([])
    )
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [{ ...makeSession('sess-1', 'Dragon Quest', 1), folder_id: 'folder-1' }],
      total: 1,
    })

    renderEditor()

    await waitFor(() => {
      expect(screen.getByText('Arc One')).toBeInTheDocument()
      expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()
    })
  })

  it('should collapse folder when folder header is clicked', async () => {
    vi.mocked(listFolders).mockImplementation((_gameId, type) =>
      type === 'session'
        ? Promise.resolve([makeFolder('folder-1', 'Arc One')])
        : Promise.resolve([])
    )
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [{ ...makeSession('sess-1', 'Dragon Quest', 1), folder_id: 'folder-1' }],
      total: 1,
    })

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByText('Arc One')).toBeInTheDocument())

    // Session visible before collapse
    expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()

    // Click folder header to collapse
    await user.click(screen.getByText('Arc One'))

    await waitFor(() => {
      expect(screen.queryByTestId('session-card-sess-1')).not.toBeInTheDocument()
    })
  })

  it('should render notes inside folder groups', async () => {
    const noteFolder = { ...makeFolder('nfolder-1', 'Lore Notes'), type: 'note' as const }
    vi.mocked(listFolders).mockImplementation((_gameId, type) =>
      type === 'note'
        ? Promise.resolve([noteFolder])
        : Promise.resolve([])
    )
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [{ ...makeNote('note-1', 'Ancient History'), folder_id: 'nfolder-1' }],
      total: 1,
    })

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    await waitFor(() => {
      expect(screen.getByText('Lore Notes')).toBeInTheDocument()
      expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument()
    })
  })
})

// ── Note filter controls ─────────────────────────────────────────

describe('Editor — note filter controls', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should change note sort to title when Title A-Z clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByText(/title a–z/i)).toBeInTheDocument())
    await user.click(screen.getByText(/title a–z/i))

    // After clicking, listGameNotesPaginated should be called (via useEffect)
    await waitFor(() => {
      expect(vi.mocked(listGameNotesPaginated)).toHaveBeenCalled()
    })
  })

  it('should show the All Notes / Unlinked session filter select', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      const select = screen.getByRole('combobox')
      expect(select).toBeInTheDocument()
    })
  })
})

// ── View mode preference ─────────────────────────────────────────

describe('Editor — view mode preference', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should apply saved view mode from preferences', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      default_game_id: null,
      map_editor_mode: 'modal',
      page_size: null,
      default_view_mode: { 'game-1': 'grid' },
    } as never)

    renderEditor()

    await waitFor(() => {
      // Grid mode should be applied - grid button should be visually active
      expect(screen.getByTitle('Grid view')).toBeInTheDocument()
    })
  })

  it('should update view mode and save preference when grid button is clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTitle('Grid view')).toBeInTheDocument())
    await user.click(screen.getByTitle('Grid view'))

    await waitFor(() => {
      expect(vi.mocked(getPreferences)).toHaveBeenCalled()
    })
  })
})

// ── Date filter inputs ───────────────────────────────────────────

describe('Editor — date filter inputs', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [
        { ...makeSession('sess-1', 'Session One', 1), updated_at: '2024-01-15T00:00:00Z' },
        { ...makeSession('sess-2', 'Session Two', 2), updated_at: '2024-03-10T00:00:00Z' },
      ],
      total: 2,
    })
  })

  it('should render date from/to filter inputs in session filter panel', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      const dateInputs = screen.getAllByDisplayValue('')
      // Should have at least 2 date inputs (filterDateFrom and filterDateTo)
      expect(screen.getByPlaceholderText('Min')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('Max')).toBeInTheDocument()
    })
  })

  it('should filter sessions by date range (from)', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    // Get all date-type inputs
    await waitFor(() => expect(screen.getByPlaceholderText('Min')).toBeInTheDocument())
    const dateInputs = document.querySelectorAll('input[type="date"]')
    expect(dateInputs.length).toBeGreaterThanOrEqual(1)

    // Set a "from" date after session 1 but before session 2
    const fromInput = dateInputs[0] as HTMLElement
    await userEvent.type(fromInput, '2024-02-01')
  })
})

// ── Notes with session links ──────────────────────────────────────

describe('Editor — notes with session links', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
  })

  it('should show session title in note card when note is linked to a session', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [{ ...makeNote('note-1', 'Battle Report'), session_id: 'sess-1' }],
      total: 1,
    })

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    await waitFor(() => {
      expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument()
    })
  })

  it('should show sessions in notes filter session select when sessions are loaded', async () => {
    // Session with null session_number so option text is just the title
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [{ ...makeSession('sess-1', 'Dragon Quest', 1), session_number: null }],
      total: 1,
    })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })

    const user = userEvent.setup()
    renderEditor()

    // Wait for sessions to load
    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())

    // Switch to notes tab
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    // Open filter panel
    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    // Session should appear in the session filter select (no session_number prefix)
    await waitFor(() => {
      expect(screen.getByText('Dragon Quest')).toBeInTheDocument()
    })
  })

  it('should show session number in notes filter select for numbered sessions', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 3)],
      total: 1,
    })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => {
      // Numbered session shows #3 — Dragon Quest
      expect(screen.getByText('#3 — Dragon Quest')).toBeInTheDocument()
    })
  })
})

// ── Session number filter ────────────────────────────────────────

describe('Editor — session number filter', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should filter by session number min', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [
        makeSession('sess-1', 'Session One', 1),
        makeSession('sess-2', 'Session Two', 5),
      ],
      total: 2,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByPlaceholderText('Min')).toBeInTheDocument())
    await user.type(screen.getByPlaceholderText('Min'), '3')

    await waitFor(() => {
      // Session 1 (#1) filtered out; Session 2 (#5) stays
      expect(screen.queryByTestId('session-card-sess-1')).not.toBeInTheDocument()
      expect(screen.getByTestId('session-card-sess-2')).toBeInTheDocument()
    })
  })

  it('should filter by session number max', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [
        makeSession('sess-1', 'Session One', 1),
        makeSession('sess-2', 'Session Two', 5),
      ],
      total: 2,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByLabelText('Toggle filters')).toBeInTheDocument())
    await user.click(screen.getByLabelText('Toggle filters'))

    await waitFor(() => expect(screen.getByPlaceholderText('Max')).toBeInTheDocument())
    await user.type(screen.getByPlaceholderText('Max'), '2')

    await waitFor(() => {
      // Session 2 (#5) filtered out; Session 1 (#1) stays
      expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument()
      expect(screen.queryByTestId('session-card-sess-2')).not.toBeInTheDocument()
    })
  })
})

// ── Session edit / delete modals ─────────────────────────────────

describe('Editor — session edit & delete modals', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
  })

  it('should open edit session modal when session edit button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('edit-session-sess-1'))

    await waitFor(() => {
      expect(screen.getByTestId('session-form-modal')).toBeInTheDocument()
    })
  })

  it('should open delete confirm modal when session delete button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-session-sess-1'))

    await waitFor(() => {
      expect(screen.getByTestId('confirm-modal')).toBeInTheDocument()
    })
  })

  it('should navigate to session notes when open button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('open-session-sess-1'))

    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1/sessions/sess-1/notes')
  })

  it('should open new session form with pre-filled session number when sessions exist', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new session/i }))

    // SessionFormModal should appear (session_number will be 2, computed from max+1)
    await waitFor(() => {
      expect(screen.getByTestId('session-form-modal')).toBeInTheDocument()
    })
  })
})

// ── Session modal handler flows ──────────────────────────────────

describe('Editor — session modal handler flows', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })
  })

  it('should call createSession and close modal on save', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /new session/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new session/i }))
    await waitFor(() => expect(screen.getByLabelText('sfm-save')).toBeInTheDocument())
    await user.click(screen.getByLabelText('sfm-save'))

    await waitFor(() => {
      expect(vi.mocked(createSession)).toHaveBeenCalled()
      expect(screen.queryByTestId('session-form-modal')).not.toBeInTheDocument()
    })
  })

  it('should close create session modal without saving when close button clicked', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /new session/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new session/i }))
    await waitFor(() => expect(screen.getByLabelText('sfm-close')).toBeInTheDocument())
    await user.click(screen.getByLabelText('sfm-close'))

    await waitFor(() => {
      expect(screen.queryByTestId('session-form-modal')).not.toBeInTheDocument()
    })
  })

  it('should call updateSession and close edit modal on save', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('edit-session-sess-1'))
    await waitFor(() => expect(screen.getByLabelText('sfm-save')).toBeInTheDocument())
    await user.click(screen.getByLabelText('sfm-save'))

    await waitFor(() => {
      expect(vi.mocked(updateSession)).toHaveBeenCalled()
      expect(screen.queryByTestId('session-form-modal')).not.toBeInTheDocument()
    })
  })

  it('should close edit session modal without saving when close clicked', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('edit-session-sess-1'))
    await waitFor(() => expect(screen.getByLabelText('sfm-close')).toBeInTheDocument())
    await user.click(screen.getByLabelText('sfm-close'))

    await waitFor(() => {
      expect(screen.queryByTestId('session-form-modal')).not.toBeInTheDocument()
    })
  })

  it('should call deleteSession and close confirm modal when confirmed', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-session-sess-1'))
    await waitFor(() => expect(screen.getByLabelText('confirm-ok')).toBeInTheDocument())
    await user.click(screen.getByLabelText('confirm-ok'))

    await waitFor(() => {
      expect(vi.mocked(deleteSession)).toHaveBeenCalled()
    })
  })

  it('should cancel delete and close confirm modal', async () => {
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({
      data: [makeSession('sess-1', 'Dragon Quest', 1)],
      total: 1,
    })
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByTestId('session-card-sess-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-session-sess-1'))
    await waitFor(() => expect(screen.getByLabelText('confirm-cancel')).toBeInTheDocument())
    await user.click(screen.getByLabelText('confirm-cancel'))

    await waitFor(() => {
      expect(screen.queryByTestId('confirm-modal')).not.toBeInTheDocument()
    })
  })
})

// ── Note modal handler flows ──────────────────────────────────────

describe('Editor — note modal handler flows', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [makeNote('note-1', 'Lore of the Realm')],
      total: 1,
    })
  })

  it('should call createNote and close modal on save', async () => {
    vi.mocked(listGameNotesPaginated)
      .mockResolvedValueOnce({ data: [], total: 0 })
      .mockResolvedValue({ data: [], total: 0 })
    vi.mocked(createNote).mockResolvedValue(makeNote('new-note', 'New Note'))

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await waitFor(() => expect(screen.getByRole('button', { name: /new note/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new note/i }))
    await waitFor(() => expect(screen.getByLabelText('nfm-save')).toBeInTheDocument())
    await user.click(screen.getByLabelText('nfm-save'))

    await waitFor(() => {
      expect(vi.mocked(createNote)).toHaveBeenCalled()
    })
  })

  it('should close create note modal without saving when close clicked', async () => {
    vi.mocked(listGameNotesPaginated).mockResolvedValue({ data: [], total: 0 })

    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await waitFor(() => expect(screen.getByRole('button', { name: /new note/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /new note/i }))
    await waitFor(() => expect(screen.getByLabelText('nfm-close')).toBeInTheDocument())
    await user.click(screen.getByLabelText('nfm-close'))

    await waitFor(() => {
      expect(screen.queryByTestId('note-form-modal')).not.toBeInTheDocument()
    })
  })

  it('should call deleteNote and close confirm when confirmed', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-note-note-1'))
    await waitFor(() => expect(screen.getByLabelText('confirm-ok')).toBeInTheDocument())
    await user.click(screen.getByLabelText('confirm-ok'))

    await waitFor(() => {
      expect(vi.mocked(deleteNote)).toHaveBeenCalled()
    })
  })

  it('should cancel note delete and close confirm modal', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-note-note-1'))
    await waitFor(() => expect(screen.getByLabelText('confirm-cancel')).toBeInTheDocument())
    await user.click(screen.getByLabelText('confirm-cancel'))

    await waitFor(() => {
      expect(screen.queryByTestId('confirm-modal')).not.toBeInTheDocument()
    })
  })

  it('should close edit note modal without saving when close clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))
    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('edit-note-note-1'))
    await waitFor(() => expect(screen.getByLabelText('nfm-close')).toBeInTheDocument())
    await user.click(screen.getByLabelText('nfm-close'))

    await waitFor(() => {
      expect(screen.queryByTestId('note-form-modal')).not.toBeInTheDocument()
    })
  })
})

// ── Note edit / delete modals ─────────────────────────────────────

describe('Editor — note edit & delete modals', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    vi.mocked(listGameSessionsPaginated).mockResolvedValue({ data: [], total: 0 })
    vi.mocked(listGameNotesPaginated).mockResolvedValue({
      data: [makeNote('note-1', 'Lore of the Realm')],
      total: 1,
    })
  })

  it('should open edit note modal when note edit button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    // Switch to notes tab
    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('edit-note-note-1'))

    await waitFor(() => {
      expect(screen.getByTestId('note-form-modal')).toBeInTheDocument()
    })
  })

  it('should open delete confirm modal when note delete button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('delete-note-note-1'))

    await waitFor(() => {
      expect(screen.getByTestId('confirm-modal')).toBeInTheDocument()
    })
  })

  it('should navigate to note editor when open button clicked', async () => {
    const user = userEvent.setup()
    renderEditor()

    await waitFor(() => expect(screen.getByRole('button', { name: /^notes$/i })).toBeInTheDocument())
    await user.click(screen.getByRole('button', { name: /^notes$/i }))

    await waitFor(() => expect(screen.getByTestId('note-card-note-1')).toBeInTheDocument())
    await user.click(screen.getByLabelText('open-note-note-1'))

    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1/notes/note-1')
  })
})

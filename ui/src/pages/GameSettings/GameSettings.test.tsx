import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import GameSettings from './GameSettings'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ gameId: 'game-123' }),
  }
})

const mockGetInviteStatus = vi.fn()
const mockGenerateInvite = vi.fn()
const mockRevokeInvite = vi.fn()

vi.mock('../../api/invite', () => ({
  getInviteStatus: (...args: unknown[]) => mockGetInviteStatus(...args),
  generateInvite: (...args: unknown[]) => mockGenerateInvite(...args),
  revokeInvite: (...args: unknown[]) => mockRevokeInvite(...args),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderGameSettings() {
  return render(
    <MemoryRouter>
      <GameSettings />
    </MemoryRouter>,
  )
}

describe('GameSettings', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockGetInviteStatus.mockReset()
    mockGenerateInvite.mockReset()
    mockRevokeInvite.mockReset()
  })

  it('should show loading state initially', () => {
    mockGetInviteStatus.mockReturnValue(new Promise(() => {}))
    renderGameSettings()
    expect(screen.getByText(/loading settings/i)).toBeInTheDocument()
  })

  it('should render page heading after loading', async () => {
    mockGetInviteStatus.mockResolvedValue({ has_active_invite: false })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByText('Campaign Settings')).toBeInTheDocument()
    })
  })

  it('should render generate link form when no active invite', async () => {
    mockGetInviteStatus.mockResolvedValue({ has_active_invite: false })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /generate link/i })).toBeInTheDocument()
    })
  })

  it('should show active invite details when invite exists', async () => {
    mockGetInviteStatus.mockResolvedValue({
      has_active_invite: true,
      token: 'invite-token-xyz',
      created_at: '2024-01-01T00:00:00Z',
    })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByText(/invite-token-xyz/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /revoke/i })).toBeInTheDocument()
    })
  })

  it('should call generateInvite with gameId and expiry when generate button clicked', async () => {
    const user = userEvent.setup()
    mockGetInviteStatus.mockResolvedValue({ has_active_invite: false })
    mockGenerateInvite.mockResolvedValue({
      token: 'new-token',
      created_at: '2024-01-01T00:00:00Z',
      expires_at: null,
    })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /generate link/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /generate link/i }))

    await waitFor(() => {
      expect(mockGenerateInvite).toHaveBeenCalledWith('game-123', '24h')
    })
  })

  it('should call revokeInvite when revoke button clicked', async () => {
    const user = userEvent.setup()
    mockGetInviteStatus.mockResolvedValue({
      has_active_invite: true,
      token: 'invite-token-xyz',
      created_at: '2024-01-01T00:00:00Z',
    })
    mockRevokeInvite.mockResolvedValue(undefined)
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /revoke/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /revoke/i }))

    await waitFor(() => {
      expect(mockRevokeInvite).toHaveBeenCalledWith('game-123')
    })
  })

  it('should show error when invite status fetch fails', async () => {
    mockGetInviteStatus.mockRejectedValue(new Error('Failed to load'))
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load')
    })
  })

  it('should navigate back to game when back button clicked', async () => {
    const user = userEvent.setup()
    mockGetInviteStatus.mockResolvedValue({ has_active_invite: false })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /back to campaign/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /back to campaign/i }))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-123')
  })

  it('should allow changing expiry option before generating', async () => {
    const user = userEvent.setup()
    mockGetInviteStatus.mockResolvedValue({ has_active_invite: false })
    mockGenerateInvite.mockResolvedValue({
      token: 'new-token',
      created_at: '2024-01-01T00:00:00Z',
    })
    renderGameSettings()

    await waitFor(() => {
      expect(screen.getByRole('combobox', { name: /link expiry/i })).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByRole('combobox', { name: /link expiry/i }), '7d')
    await user.click(screen.getByRole('button', { name: /generate link/i }))

    await waitFor(() => {
      expect(mockGenerateInvite).toHaveBeenCalledWith('game-123', '7d')
    })
  })
})

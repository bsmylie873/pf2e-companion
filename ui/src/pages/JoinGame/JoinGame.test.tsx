import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import JoinGame from './JoinGame'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ token: 'invite-token-abc' }),
  }
})

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'u1', username: 'testuser', email: 'test@test.com' },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
    refreshUser: vi.fn(),
  }),
}))

const mockValidateInvite = vi.fn()
const mockRedeemInvite = vi.fn()

vi.mock('../../api/invite', () => ({
  validateInvite: (...args: unknown[]) => mockValidateInvite(...args),
  redeemInvite: (...args: unknown[]) => mockRedeemInvite(...args),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderJoinGame() {
  return render(
    <MemoryRouter>
      <JoinGame />
    </MemoryRouter>,
  )
}

describe('JoinGame', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockValidateInvite.mockReset()
    mockRedeemInvite.mockReset()
  })

  it('should show loading state while validating invite', () => {
    mockValidateInvite.mockReturnValue(new Promise(() => {}))
    renderJoinGame()
    expect(screen.getByText(/consulting the arcane registry/i)).toBeInTheDocument()
  })

  it('should show game info when invite is valid', async () => {
    mockValidateInvite.mockResolvedValue({
      game_id: 'game-1',
      game_title: 'Epic Adventure',
    })
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByText('Epic Adventure')).toBeInTheDocument()
    })
  })

  it('should show Join Game button when invite is valid', async () => {
    mockValidateInvite.mockResolvedValue({
      game_id: 'game-1',
      game_title: 'Epic Adventure',
    })
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /join game/i })).toBeInTheDocument()
    })
  })

  it('should show error when invite is invalid', async () => {
    mockValidateInvite.mockRejectedValue(new Error('This invitation is invalid or has expired.'))
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByText('This invitation is invalid or has expired.')).toBeInTheDocument()
      expect(screen.getByText('Invitation Unavailable')).toBeInTheDocument()
    })
  })

  it('should call redeemInvite and navigate on join', async () => {
    const user = userEvent.setup()
    mockValidateInvite.mockResolvedValue({ game_id: 'game-1', game_title: 'Epic Adventure' })
    mockRedeemInvite.mockResolvedValue({ game_id: 'game-1', already_member: false })
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /join game/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /join game/i }))

    await waitFor(() => {
      expect(mockRedeemInvite).toHaveBeenCalledWith('invite-token-abc')
      expect(mockNavigate).toHaveBeenCalledWith('/games/game-1', { replace: true })
    })
  })

  it('should show decline button that navigates to games', async () => {
    const user = userEvent.setup()
    mockValidateInvite.mockResolvedValue({ game_id: 'game-1', game_title: 'Epic Adventure' })
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /decline/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /decline/i }))
    expect(mockNavigate).toHaveBeenCalledWith('/games')
  })

  it('should show error when redeem fails', async () => {
    const user = userEvent.setup()
    mockValidateInvite.mockResolvedValue({ game_id: 'game-1', game_title: 'Epic Adventure' })
    mockRedeemInvite.mockRejectedValue(new Error('Failed to join the game.'))
    renderJoinGame()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /join game/i })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /join game/i }))

    await waitFor(() => {
      expect(screen.getByText('Failed to join the game.')).toBeInTheDocument()
    })
  })
})

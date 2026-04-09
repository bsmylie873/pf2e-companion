import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Profile from './Profile'

const mockNavigate = vi.fn()
const mockRefreshUser = vi.fn()
const mockLogout = vi.fn()
const mockApiFetch = vi.fn()

const mockUser = {
  id: 'user-1',
  username: 'testuser',
  email: 'test@example.com',
  avatar_url: undefined,
  description: undefined,
  location: undefined,
}

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: mockUser,
    isAuthenticated: true,
    isLoading: false,
    logout: mockLogout,
    refreshUser: mockRefreshUser,
    login: vi.fn(),
    register: vi.fn(),
  }),
}))

vi.mock('../../api/client', () => ({
  apiFetch: (...args: unknown[]) => mockApiFetch(...args),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderProfile() {
  return render(
    <MemoryRouter>
      <Profile />
    </MemoryRouter>,
  )
}

describe('Profile', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockRefreshUser.mockReset()
    mockLogout.mockReset()
    mockApiFetch.mockReset()
  })

  it('should render user profile information', () => {
    renderProfile()
    // Username appears in multiple places (heading + field), check at least one exists
    expect(screen.getAllByText('testuser').length).toBeGreaterThan(0)
    expect(screen.getByText('test@example.com')).toBeInTheDocument()
  })

  it('should render the back button', () => {
    renderProfile()
    expect(screen.getByRole('button', { name: /back to campaigns/i })).toBeInTheDocument()
  })

  it('should navigate to games when back button clicked', async () => {
    const user = userEvent.setup()
    renderProfile()
    await user.click(screen.getByRole('button', { name: /back to campaigns/i }))
    expect(mockNavigate).toHaveBeenCalledWith('/games')
  })

  it('should show edit form when Edit button clicked', async () => {
    const user = userEvent.setup()
    renderProfile()

    await user.click(screen.getByRole('button', { name: /edit/i }))

    expect(screen.getByRole('button', { name: /inscribe changes/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/adventurer/i)).toBeInTheDocument()
  })

  it('should call apiFetch on profile save', async () => {
    const user = userEvent.setup()
    mockApiFetch.mockResolvedValue({})
    mockRefreshUser.mockResolvedValue(undefined)
    renderProfile()

    await user.click(screen.getByRole('button', { name: /edit/i }))
    await user.click(screen.getByRole('button', { name: /inscribe changes/i }))

    await waitFor(() => {
      expect(mockApiFetch).toHaveBeenCalledWith('/users/user-1', expect.objectContaining({ method: 'PATCH' }))
    })
  })

  it('should show error when profile save fails', async () => {
    const user = userEvent.setup()
    mockApiFetch.mockRejectedValue(new Error('Update failed'))
    renderProfile()

    await user.click(screen.getByRole('button', { name: /edit/i }))
    await user.click(screen.getByRole('button', { name: /inscribe changes/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Update failed')
    })
  })

  it('should show Change Passphrase section', () => {
    renderProfile()
    expect(screen.getByText('Change Passphrase')).toBeInTheDocument()
  })

  it('should show password form when Change button is clicked', async () => {
    const user = userEvent.setup()
    renderProfile()

    await user.click(screen.getByRole('button', { name: /change/i }))

    expect(screen.getByLabelText('Current Passphrase')).toBeInTheDocument()
    expect(screen.getByLabelText('New Passphrase')).toBeInTheDocument()
    expect(screen.getByLabelText('Confirm Passphrase')).toBeInTheDocument()
  })

  it('should show passphrase mismatch error in password form', async () => {
    const user = userEvent.setup()
    renderProfile()

    await user.click(screen.getByRole('button', { name: /change/i }))
    await user.type(screen.getByLabelText('Current Passphrase'), 'oldpass')
    await user.type(screen.getByLabelText('New Passphrase'), 'newpass')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'different')
    await user.click(screen.getByRole('button', { name: /seal the scroll/i }))

    expect(screen.getByRole('alert')).toHaveTextContent('New passphrases do not match.')
  })

  it('should call logout after successful password change', async () => {
    const user = userEvent.setup()
    mockApiFetch.mockResolvedValue({})
    mockLogout.mockResolvedValue(undefined)
    renderProfile()

    await user.click(screen.getByRole('button', { name: /change/i }))
    await user.type(screen.getByLabelText('Current Passphrase'), 'oldpass')
    await user.type(screen.getByLabelText('New Passphrase'), 'newpass123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'newpass123')
    await user.click(screen.getByRole('button', { name: /seal the scroll/i }))

    await waitFor(() => {
      expect(mockLogout).toHaveBeenCalled()
    })
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Login from './Login'

const mockNavigate = vi.fn()
const mockLogin = vi.fn()
const mockRegister = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login: mockLogin,
    register: mockRegister,
    logout: vi.fn(),
    refreshUser: vi.fn(),
  }),
}))

vi.mock('../../api/preferences', () => ({
  getPreferences: vi.fn().mockResolvedValue({ default_game_id: null }),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderLogin(search = '') {
  return render(
    <MemoryRouter initialEntries={[`/${search}`]}>
      <Login />
    </MemoryRouter>,
  )
}

describe('Login', () => {
  beforeEach(() => {
    mockNavigate.mockReset()
    mockLogin.mockReset()
    mockRegister.mockReset()
  })

  it('should render login form with username and password fields', () => {
    renderLogin()
    expect(screen.getByLabelText('Adventurer')).toBeInTheDocument()
    expect(screen.getByLabelText('Passphrase')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /begin your journey/i })).toBeInTheDocument()
  })

  it('should render the app title', () => {
    renderLogin()
    expect(screen.getByText('PF2E Companion')).toBeInTheDocument()
  })

  it('should show register mode when toggle button is clicked', async () => {
    const user = userEvent.setup()
    renderLogin()

    await user.click(screen.getByRole('button', { name: /new adventurer/i }))

    expect(screen.getByLabelText('Sending Stone')).toBeInTheDocument()
    expect(screen.getByLabelText('Confirm Passphrase')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /forge your path/i })).toBeInTheDocument()
  })

  it('should call login with username and password on submit', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValue(undefined)
    renderLogin()

    await user.type(screen.getByLabelText('Adventurer'), 'testuser')
    await user.type(screen.getByLabelText('Passphrase'), 'password123')
    await user.click(screen.getByRole('button', { name: /begin your journey/i }))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('testuser', 'password123')
    })
  })

  it('should show error message when login fails', async () => {
    const user = userEvent.setup()
    mockLogin.mockRejectedValue(new Error('Invalid credentials'))
    renderLogin()

    await user.type(screen.getByLabelText('Adventurer'), 'testuser')
    await user.type(screen.getByLabelText('Passphrase'), 'wrongpass')
    await user.click(screen.getByRole('button', { name: /begin your journey/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Invalid credentials')
    })
  })

  it('should show password mismatch error in register mode', async () => {
    const user = userEvent.setup()
    renderLogin()

    await user.click(screen.getByRole('button', { name: /new adventurer/i }))
    await user.type(screen.getByLabelText('Adventurer'), 'newuser')
    await user.type(screen.getByLabelText('Sending Stone'), 'new@test.com')
    await user.type(screen.getByLabelText('Passphrase'), 'password123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'different')
    await user.click(screen.getByRole('button', { name: /forge your path/i }))

    expect(screen.getByRole('alert')).toHaveTextContent('Passphrases do not match')
    expect(mockRegister).not.toHaveBeenCalled()
  })

  it('should call register with correct args in register mode', async () => {
    const user = userEvent.setup()
    mockRegister.mockResolvedValue(undefined)
    renderLogin()

    await user.click(screen.getByRole('button', { name: /new adventurer/i }))
    await user.type(screen.getByLabelText('Adventurer'), 'newuser')
    await user.type(screen.getByLabelText('Sending Stone'), 'new@test.com')
    await user.type(screen.getByLabelText('Passphrase'), 'password123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'password123')
    await user.click(screen.getByRole('button', { name: /forge your path/i }))

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith('newuser', 'new@test.com', 'password123')
    })
  })

  it('should navigate to games after successful login', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValue(undefined)
    renderLogin()

    await user.type(screen.getByLabelText('Adventurer'), 'testuser')
    await user.type(screen.getByLabelText('Passphrase'), 'password123')
    await user.click(screen.getByRole('button', { name: /begin your journey/i }))

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/games', { replace: true })
    })
  })

  it('should show forgot password link in login mode', () => {
    renderLogin()
    expect(screen.getByText(/forgot your passphrase/i)).toBeInTheDocument()
  })

  it('should switch back to login mode from register mode', async () => {
    const user = userEvent.setup()
    renderLogin()

    await user.click(screen.getByRole('button', { name: /new adventurer/i }))
    await user.click(screen.getByRole('button', { name: /return to the gates/i }))

    expect(screen.queryByLabelText('Sending Stone')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /begin your journey/i })).toBeInTheDocument()
  })
})

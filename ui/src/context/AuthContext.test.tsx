import { render, screen, waitFor, act } from '@testing-library/react'
import { AuthProvider, useAuth } from './AuthContext'

vi.mock('../api/auth', () => ({
  getMe: vi.fn(),
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
}))

import { getMe, login as apiLogin, logout as apiLogout, register as apiRegister } from '../api/auth'

const mockUser = { id: 'user-1', username: 'testuser', email: 'test@example.com' }

// A simple consumer that exposes context values via data-testid attributes
function TestConsumer() {
  const { user, isAuthenticated, isLoading, login, logout, register } = useAuth()
  return (
    <div>
      <div data-testid="loading">{String(isLoading)}</div>
      <div data-testid="authenticated">{String(isAuthenticated)}</div>
      <div data-testid="user">{user ? user.username : 'null'}</div>
      <button onClick={() => void login('testuser', 'pass')}>Login</button>
      <button onClick={() => void logout()}>Logout</button>
      <button onClick={() => void register('newuser', 'new@example.com', 'pass')}>Register</button>
    </div>
  )
}

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Prevent JSDOM errors from window.location.href assignment inside logout
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: { href: '/' },
    })
  })

  it('should start in the loading state while getMe is pending', () => {
    vi.mocked(getMe).mockImplementation(() => new Promise(() => {})) // never resolves

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('loading').textContent).toBe('true')
    expect(screen.getByTestId('user').textContent).toBe('null')
  })

  it('should populate user and clear loading after getMe resolves', async () => {
    vi.mocked(getMe).mockResolvedValue(mockUser)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'))
    expect(screen.getByTestId('user').textContent).toBe('testuser')
    expect(screen.getByTestId('authenticated').textContent).toBe('true')
  })

  it('should set user to null and clear loading when getMe rejects', async () => {
    vi.mocked(getMe).mockRejectedValue(new Error('Unauthorized'))

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'))
    expect(screen.getByTestId('user').textContent).toBe('null')
    expect(screen.getByTestId('authenticated').textContent).toBe('false')
  })

  it('should update user state after a successful login', async () => {
    // getMe returns null initially (not logged in)
    vi.mocked(getMe).mockRejectedValue(new Error('Not authenticated'))
    vi.mocked(apiLogin).mockResolvedValue(mockUser)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'))

    await act(async () => {
      screen.getByRole('button', { name: 'Login' }).click()
    })

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('testuser'))
    expect(screen.getByTestId('authenticated').textContent).toBe('true')
  })

  it('should clear user state after logout', async () => {
    vi.mocked(getMe).mockResolvedValue(mockUser)
    vi.mocked(apiLogout).mockResolvedValue(undefined)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('testuser'))

    await act(async () => {
      screen.getByRole('button', { name: 'Logout' }).click()
    })

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('null'))
    expect(screen.getByTestId('authenticated').textContent).toBe('false')
  })

  it('should clear user state after logout even when apiLogout throws', async () => {
    vi.mocked(getMe).mockResolvedValue(mockUser)
    vi.mocked(apiLogout).mockRejectedValue(new Error('Server error'))

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('testuser'))

    await act(async () => {
      screen.getByRole('button', { name: 'Logout' }).click()
    })

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('null'))
  })

  it('should update user state after a successful register', async () => {
    vi.mocked(getMe).mockRejectedValue(new Error('Not authenticated'))
    vi.mocked(apiRegister).mockResolvedValue(mockUser)

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => expect(screen.getByTestId('loading').textContent).toBe('false'))

    await act(async () => {
      screen.getByRole('button', { name: 'Register' }).click()
    })

    await waitFor(() => expect(screen.getByTestId('user').textContent).toBe('testuser'))
  })

  it('should throw when useAuth is used outside of AuthProvider', () => {
    function Orphan() {
      useAuth()
      return null
    }

    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    expect(() => render(<Orphan />)).toThrow('useAuth must be used within AuthProvider')
    consoleSpy.mockRestore()
  })
})

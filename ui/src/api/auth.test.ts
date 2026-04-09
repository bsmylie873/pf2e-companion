import { describe, it, expect, vi, beforeEach } from 'vitest'
import { login, register, logout, getMe, forgotPassword, resetPassword } from './auth'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('login', () => {
  it('should call apiFetch POST /auth/login with credentials', async () => {
    const mockUser = { id: '1', username: 'testuser', email: 'test@example.com' }
    mockApiFetch.mockResolvedValueOnce(mockUser)

    const payload = { username: 'testuser', password: 'secret' }
    const result = await login(payload)

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    expect(result).toEqual(mockUser)
  })
})

describe('register', () => {
  it('should call apiFetch POST /auth/register with registration data', async () => {
    const mockUser = { id: '2', username: 'newuser', email: 'new@example.com' }
    mockApiFetch.mockResolvedValueOnce(mockUser)

    const payload = { username: 'newuser', email: 'new@example.com', password: 'password123' }
    const result = await register(payload)

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
    expect(result).toEqual(mockUser)
  })
})

describe('logout', () => {
  it('should call apiFetch POST /auth/logout', async () => {
    mockApiFetch.mockResolvedValueOnce({ message: 'logged out' })

    const result = await logout()

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/logout', { method: 'POST' })
    expect(result).toEqual({ message: 'logged out' })
  })
})

describe('getMe', () => {
  it('should call apiFetch GET /auth/me and return the current user', async () => {
    const mockUser = { id: '1', username: 'testuser', email: 'test@example.com' }
    mockApiFetch.mockResolvedValueOnce(mockUser)

    const result = await getMe()

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/me')
    expect(result).toEqual(mockUser)
  })
})

describe('forgotPassword', () => {
  it('should call apiFetch POST /auth/forgot-password with email', async () => {
    const response = { token: 'reset-token-123' }
    mockApiFetch.mockResolvedValueOnce(response)

    const result = await forgotPassword('user@example.com')

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email: 'user@example.com' }),
    })
    expect(result).toEqual(response)
  })

  it('should return null token when no reset token is issued', async () => {
    mockApiFetch.mockResolvedValueOnce({ token: null })

    const result = await forgotPassword('nonexistent@example.com')

    expect(result.token).toBeNull()
  })
})

describe('resetPassword', () => {
  it('should call apiFetch POST /auth/reset-password with token and new password', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await resetPassword('reset-token-abc', 'newpassword123')

    expect(mockApiFetch).toHaveBeenCalledWith('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token: 'reset-token-abc', new_password: 'newpassword123' }),
    })
  })
})

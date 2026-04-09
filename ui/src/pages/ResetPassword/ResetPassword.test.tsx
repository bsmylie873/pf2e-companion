import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import ResetPassword from './ResetPassword'

const mockResetPassword = vi.fn()

vi.mock('../../api/auth', () => ({
  resetPassword: (...args: unknown[]) => mockResetPassword(...args),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderWithToken(token = '') {
  return render(
    <MemoryRouter initialEntries={[`/reset-password${token ? `?token=${token}` : ''}`]}>
      <ResetPassword />
    </MemoryRouter>,
  )
}

describe('ResetPassword', () => {
  beforeEach(() => {
    mockResetPassword.mockReset()
  })

  it('should show error when no token is provided', () => {
    renderWithToken('')
    expect(screen.getByRole('alert')).toHaveTextContent(/missing its seal/i)
    expect(screen.getByText(/request a new scroll/i)).toBeInTheDocument()
  })

  it('should render password form when token is provided', () => {
    renderWithToken('valid-token-123')
    expect(screen.getByLabelText('New Passphrase')).toBeInTheDocument()
    expect(screen.getByLabelText('Confirm Passphrase')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /reforge my passphrase/i })).toBeInTheDocument()
  })

  it('should show mismatch error when passwords differ', async () => {
    const user = userEvent.setup()
    renderWithToken('valid-token-123')

    await user.type(screen.getByLabelText('New Passphrase'), 'password123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'different')
    await user.click(screen.getByRole('button', { name: /reforge my passphrase/i }))

    expect(screen.getByRole('alert')).toHaveTextContent('Passphrases do not match.')
    expect(mockResetPassword).not.toHaveBeenCalled()
  })

  it('should call resetPassword with token and password on valid submit', async () => {
    const user = userEvent.setup()
    mockResetPassword.mockResolvedValue(undefined)
    renderWithToken('valid-token-123')

    await user.type(screen.getByLabelText('New Passphrase'), 'newpass123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'newpass123')
    await user.click(screen.getByRole('button', { name: /reforge my passphrase/i }))

    await waitFor(() => {
      expect(mockResetPassword).toHaveBeenCalledWith('valid-token-123', 'newpass123')
    })
  })

  it('should show success state after successful reset', async () => {
    const user = userEvent.setup()
    mockResetPassword.mockResolvedValue(undefined)
    renderWithToken('valid-token-123')

    await user.type(screen.getByLabelText('New Passphrase'), 'newpass123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'newpass123')
    await user.click(screen.getByRole('button', { name: /reforge my passphrase/i }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(/passphrase has been reforged/i)
    })
  })

  it('should show error message on reset failure', async () => {
    const user = userEvent.setup()
    mockResetPassword.mockRejectedValue(new Error('Invalid or expired token'))
    renderWithToken('bad-token')

    await user.type(screen.getByLabelText('New Passphrase'), 'newpass123')
    await user.type(screen.getByLabelText('Confirm Passphrase'), 'newpass123')
    await user.click(screen.getByRole('button', { name: /reforge my passphrase/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Invalid or expired token')
    })
  })

  it('should show return link when token present', () => {
    renderWithToken('valid-token-123')
    expect(screen.getByText(/return to the gates/i)).toBeInTheDocument()
  })
})

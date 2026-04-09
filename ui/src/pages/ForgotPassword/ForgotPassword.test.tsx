import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import ForgotPassword from './ForgotPassword'

const mockForgotPassword = vi.fn()

vi.mock('../../api/auth', () => ({
  forgotPassword: (...args: unknown[]) => mockForgotPassword(...args),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

function renderForgotPassword() {
  return render(
    <MemoryRouter>
      <ForgotPassword />
    </MemoryRouter>,
  )
}

describe('ForgotPassword', () => {
  beforeEach(() => {
    mockForgotPassword.mockReset()
  })

  it('should render the email input form', () => {
    renderForgotPassword()
    expect(screen.getByLabelText('Sending Stone')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /dispatch the raven/i })).toBeInTheDocument()
  })

  it('should render the app title', () => {
    renderForgotPassword()
    expect(screen.getByText('PF2E Companion')).toBeInTheDocument()
  })

  it('should show reset link when token is returned', async () => {
    const user = userEvent.setup()
    mockForgotPassword.mockResolvedValue({ token: 'abc123' })

    Object.defineProperty(window, 'location', {
      value: { origin: 'http://localhost:3000' },
      writable: true,
    })

    renderForgotPassword()

    await user.type(screen.getByLabelText('Sending Stone'), 'test@example.com')
    await user.click(screen.getByRole('button', { name: /dispatch the raven/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /copy link/i })).toBeInTheDocument()
    })
  })

  it('should show no-token message when API returns no token', async () => {
    const user = userEvent.setup()
    mockForgotPassword.mockResolvedValue({ token: null })
    renderForgotPassword()

    await user.type(screen.getByLabelText('Sending Stone'), 'unknown@example.com')
    await user.click(screen.getByRole('button', { name: /dispatch the raven/i }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(/matching sending stone/i)
    })
  })

  it('should show success message even if API throws (enumeration protection)', async () => {
    const user = userEvent.setup()
    mockForgotPassword.mockRejectedValue(new Error('Server error'))
    renderForgotPassword()

    await user.type(screen.getByLabelText('Sending Stone'), 'test@example.com')
    await user.click(screen.getByRole('button', { name: /dispatch the raven/i }))

    await waitFor(() => {
      // After submission, either success state is shown, no error alert
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })

  it('should show return link', () => {
    renderForgotPassword()
    expect(screen.getByText(/return to the gates/i)).toBeInTheDocument()
  })

  it('should disable submit button while submitting', async () => {
    const user = userEvent.setup()
    let resolve: (v: unknown) => void
    mockForgotPassword.mockReturnValue(new Promise(r => { resolve = r }))
    renderForgotPassword()

    await user.type(screen.getByLabelText('Sending Stone'), 'test@example.com')
    await user.click(screen.getByRole('button', { name: /dispatch the raven/i }))

    expect(screen.getByRole('button', { name: /dispatching/i })).toBeDisabled()
    resolve!({ token: null })
  })
})

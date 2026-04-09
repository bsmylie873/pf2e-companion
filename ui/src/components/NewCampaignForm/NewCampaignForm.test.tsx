import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import NewCampaignForm from './NewCampaignForm'

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'user-1', username: 'creator', email: 'creator@example.com' },
    isAuthenticated: true,
    isLoading: false,
  }),
}))

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn(),
}))

vi.mock('../UserSearch/UserSearch', () => ({
  default: ({ onSelect }: { onSelect: (user: unknown) => void }) => (
    <button
      data-testid="user-search"
      onClick={() => onSelect({ id: 'user-2', username: 'newmember', email: 'new@example.com' })}
    >
      Add Member
    </button>
  ),
}))

import { apiFetch } from '../../api/client'

const mockApiFetch = apiFetch as ReturnType<typeof vi.fn>

describe('NewCampaignForm', () => {
  beforeEach(() => {
    mockApiFetch.mockResolvedValue({
      id: 'game-new',
      title: 'The Dark Campaign',
    })
  })

  it('should render the form with title input', () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    expect(screen.getByLabelText('Title')).toBeInTheDocument()
  })

  it('should render description and splash URL fields', () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    expect(screen.getByLabelText(/Description/)).toBeInTheDocument()
    expect(screen.getByLabelText(/Splash Image URL/)).toBeInTheDocument()
  })

  it('should show the creator in member list', () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    expect(screen.getByText('creator')).toBeInTheDocument()
    expect(screen.getByText('Creator')).toBeInTheDocument()
  })

  it('should show validation error when title is empty on submit', async () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: 'Create Campaign' }))
    expect(await screen.findByText('Title is required')).toBeInTheDocument()
  })

  it('should show validation error for invalid splash URL', async () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Title'), { target: { value: 'My Campaign' } })
    fireEvent.change(screen.getByLabelText(/Splash Image URL/), { target: { value: 'not-a-url' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Campaign' }))
    expect(await screen.findByText(/Must be a valid URL/)).toBeInTheDocument()
  })

  it('should call apiFetch on successful form submission', async () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Title'), { target: { value: 'The Dark Campaign' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Campaign' }))
    await waitFor(() => {
      expect(mockApiFetch).toHaveBeenCalledWith(
        '/games',
        expect.objectContaining({ method: 'POST' })
      )
    })
  })

  it('should call onSuccess with gameId and title after successful creation', async () => {
    const onSuccess = vi.fn()
    render(<NewCampaignForm onSuccess={onSuccess} onDirtyChange={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Title'), { target: { value: 'The Dark Campaign' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Campaign' }))
    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('game-new', 'The Dark Campaign')
    })
  })

  it('should show api error on failed submission', async () => {
    mockApiFetch.mockRejectedValue(new Error('Server error'))
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Title'), { target: { value: 'My Campaign' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create Campaign' }))
    expect(await screen.findByText('Server error')).toBeInTheDocument()
  })

  it('should add a new member when UserSearch onSelect is triggered', () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    fireEvent.click(screen.getByTestId('user-search'))
    expect(screen.getByText('newmember')).toBeInTheDocument()
  })

  it('should not allow removing the creator', () => {
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={vi.fn()} />)
    const removeCreatorBtn = screen.getByRole('button', { name: /The game creator must remain a GM/ })
    expect(removeCreatorBtn).toBeDisabled()
  })

  it('should call onDirtyChange when title is filled', () => {
    const onDirtyChange = vi.fn()
    render(<NewCampaignForm onSuccess={vi.fn()} onDirtyChange={onDirtyChange} />)
    fireEvent.change(screen.getByLabelText('Title'), { target: { value: 'Something' } })
    expect(onDirtyChange).toHaveBeenCalledWith(true)
  })
})

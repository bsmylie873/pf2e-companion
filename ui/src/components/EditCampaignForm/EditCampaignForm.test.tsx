import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import EditCampaignForm from './EditCampaignForm'

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn(),
}))

vi.mock('../UserSearch/UserSearch', () => ({
  default: ({ onSelect }: { onSelect: (user: unknown) => void }) => (
    <button
      data-testid="user-search"
      onClick={() => onSelect({ id: 'user-new', username: 'newmember', email: 'new@example.com' })}
    >
      Add Member
    </button>
  ),
}))

import { apiFetch } from '../../api/client'

const mockApiFetch = apiFetch as ReturnType<typeof vi.fn>

const mockGame = {
  id: 'game-1',
  title: 'The Shattered Throne',
  description: 'Epic campaign',
  splash_image_url: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockMemberships = [
  { id: 'mem-1', game_id: 'game-1', user_id: 'user-1', is_gm: true, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
]

const mockUsers = [
  { id: 'user-1', username: 'gmuser', email: 'gm@example.com' },
]

describe('EditCampaignForm', () => {
  beforeEach(() => {
    // apiFetch is called in sequence: game, memberships, users
    mockApiFetch
      .mockResolvedValueOnce(mockGame)
      .mockResolvedValueOnce(mockMemberships)
      .mockResolvedValueOnce(mockUsers)
  })

  it('should show loading state initially', () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    expect(screen.getByText('Loading campaign...')).toBeInTheDocument()
  })

  it('should render form after loading', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByDisplayValue('The Shattered Throne')).toBeInTheDocument()
    })
  })

  it('should pre-fill title from game data', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByDisplayValue('The Shattered Throne')).toBeInTheDocument()
    })
  })

  it('should pre-fill description from game data', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByDisplayValue('Epic campaign')).toBeInTheDocument()
    })
  })

  it('should show existing members', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => {
      expect(screen.getByText('gmuser')).toBeInTheDocument()
    })
  })

  it('should show validation error when title is cleared', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => screen.getByDisplayValue('The Shattered Throne'))
    fireEvent.change(screen.getByDisplayValue('The Shattered Throne'), { target: { value: '' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save Changes' }))
    expect(await screen.findByText('Title is required')).toBeInTheDocument()
  })

  it('should call apiFetch to update game on save', async () => {
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={vi.fn()}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => screen.getByDisplayValue('The Shattered Throne'))

    // Change the title
    fireEvent.change(screen.getByDisplayValue('The Shattered Throne'), {
      target: { value: 'New Title' },
    })

    // Reset mock for the save call
    mockApiFetch.mockResolvedValueOnce({})

    fireEvent.click(screen.getByRole('button', { name: 'Save Changes' }))

    await waitFor(() => {
      expect(mockApiFetch).toHaveBeenCalledWith(
        '/games/game-1',
        expect.objectContaining({ method: 'PATCH' })
      )
    })
  })

  it('should call onSuccess after successful save', async () => {
    const onSuccess = vi.fn()
    render(
      <EditCampaignForm
        gameId="game-1"
        onSuccess={onSuccess}
        onDirtyChange={vi.fn()}
      />
    )
    await waitFor(() => screen.getByDisplayValue('The Shattered Throne'))

    fireEvent.change(screen.getByDisplayValue('The Shattered Throne'), {
      target: { value: 'Updated Title' },
    })

    mockApiFetch.mockResolvedValueOnce({})

    fireEvent.click(screen.getByRole('button', { name: 'Save Changes' }))

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('Updated Title')
    })
  })
})

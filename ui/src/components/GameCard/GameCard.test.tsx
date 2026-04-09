import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import GameCard from './GameCard'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockGame = {
  id: 'game-1',
  title: 'The Shattered Throne',
  description: 'A tale of heroes and shadows.',
  splash_image_url: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('GameCard (grid mode)', () => {
  it('should render the game title', () => {
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" />
      </MemoryRouter>
    )
    expect(screen.getByText('The Shattered Throne')).toBeInTheDocument()
  })

  it('should render the game description', () => {
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" />
      </MemoryRouter>
    )
    expect(screen.getByText('A tale of heroes and shadows.')).toBeInTheDocument()
  })

  it('should navigate to game page on click', () => {
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button'))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1', { state: { title: 'The Shattered Throne' } })
  })

  it('should render image when splash_image_url is provided', () => {
    const gameWithImage = { ...mockGame, splash_image_url: 'https://example.com/img.jpg' }
    render(
      <MemoryRouter>
        <GameCard game={gameWithImage} mode="grid" />
      </MemoryRouter>
    )
    expect(screen.getByRole('img', { name: 'The Shattered Throne' })).toBeInTheDocument()
  })

  it('should show delete button when onDelete is provided', () => {
    const onDelete = vi.fn()
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" onDelete={onDelete} />
      </MemoryRouter>
    )
    expect(screen.getByRole('button', { name: 'Delete campaign' })).toBeInTheDocument()
  })

  it('should call onDelete when delete button is clicked', () => {
    const onDelete = vi.fn()
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" onDelete={onDelete} />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete campaign' }))
    expect(onDelete).toHaveBeenCalledWith('game-1')
  })

  it('should not navigate when delete button is clicked', () => {
    mockNavigate.mockClear()
    const onDelete = vi.fn()
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="grid" onDelete={onDelete} />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete campaign' }))
    expect(mockNavigate).not.toHaveBeenCalled()
  })
})

describe('GameCard (list mode)', () => {
  it('should render in list mode', () => {
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="list" />
      </MemoryRouter>
    )
    expect(screen.getByText('The Shattered Throne')).toBeInTheDocument()
  })

  it('should navigate to game page on click in list mode', () => {
    render(
      <MemoryRouter>
        <GameCard game={mockGame} mode="list" />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: /The Shattered Throne/ }))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1', { state: { title: 'The Shattered Throne' } })
  })
})

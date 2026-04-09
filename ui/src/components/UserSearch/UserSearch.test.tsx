import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import userEvent from '@testing-library/user-event'
import UserSearch from './UserSearch'

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from '../../api/client'

const mockApiFetch = apiFetch as ReturnType<typeof vi.fn>

const mockUsers = [
  { id: 'user-1', username: 'alice', email: 'alice@example.com' },
  { id: 'user-2', username: 'bob', email: 'bob@example.com' },
  { id: 'user-3', username: 'charlie', email: 'charlie@example.com' },
]

describe('UserSearch', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    mockApiFetch.mockResolvedValue(mockUsers)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should render the search input', () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    expect(screen.getByRole('combobox')).toBeInTheDocument()
  })

  it('should show results after typing and debounce delay', async () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => {
      expect(screen.getByText('alice')).toBeInTheDocument()
    })
  })

  it('should filter out excluded user IDs', async () => {
    render(<UserSearch excludeIds={['user-1']} onSelect={vi.fn()} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => {
      expect(mockApiFetch).toHaveBeenCalled()
    })
    expect(screen.queryByText('alice')).not.toBeInTheDocument()
  })

  it('should call onSelect when a user is clicked', async () => {
    const onSelect = vi.fn()
    render(<UserSearch excludeIds={[]} onSelect={onSelect} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByText('alice'))
    fireEvent.mouseDown(screen.getByText('alice'))
    expect(onSelect).toHaveBeenCalledWith(mockUsers[0])
  })

  it('should clear input after selecting a user', async () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    const input = screen.getByRole('combobox')
    fireEvent.change(input, { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByText('alice'))
    fireEvent.mouseDown(screen.getByText('alice'))
    expect(input).toHaveValue('')
  })

  it('should not show dropdown when query is empty', () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('should navigate results with arrow keys', async () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    const input = screen.getByRole('combobox')
    fireEvent.change(input, { target: { value: 'b' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByText('bob'))
    fireEvent.keyDown(input, { key: 'ArrowDown' })
    expect(screen.getByRole('option', { name: /bob/ })).toHaveAttribute('aria-selected', 'true')
  })

  it('should close dropdown on Escape key', async () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    const input = screen.getByRole('combobox')
    fireEvent.change(input, { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByRole('listbox'))
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('should reuse cached users on subsequent queries', async () => {
    render(<UserSearch excludeIds={[]} onSelect={vi.fn()} />)
    const input = screen.getByRole('combobox')

    // First query - loads users from API
    fireEvent.change(input, { target: { value: 'ali' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByText('alice'))
    const callsAfterFirst = mockApiFetch.mock.calls.length

    // Second query - should reuse cache, not call API again
    fireEvent.change(input, { target: { value: 'bob' } })
    vi.advanceTimersByTime(300)
    await waitFor(() => screen.getByText('bob'))
    expect(mockApiFetch.mock.calls.length).toBe(callsAfterFirst)
  })
})

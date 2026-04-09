import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMemberships } from './memberships'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('listMemberships', () => {
  it('should call apiFetch GET /memberships with game_id query param', async () => {
    const mockMemberships = [
      {
        id: 'mem-1',
        game_id: 'game-1',
        user_id: 'user-1',
        role: 'player',
        created_at: '2024-01-01',
      },
    ]
    mockApiFetch.mockResolvedValueOnce(mockMemberships)

    const result = await listMemberships('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/memberships?game_id=game-1')
    expect(result).toEqual(mockMemberships)
  })

  it('should return an empty array when there are no memberships', async () => {
    mockApiFetch.mockResolvedValueOnce([])

    const result = await listMemberships('game-empty')

    expect(mockApiFetch).toHaveBeenCalledWith('/memberships?game_id=game-empty')
    expect(result).toEqual([])
  })

  it('should use the correct game ID in the query string', async () => {
    mockApiFetch.mockResolvedValueOnce([])

    await listMemberships('game-abc-xyz')

    expect(mockApiFetch).toHaveBeenCalledWith('/memberships?game_id=game-abc-xyz')
  })
})

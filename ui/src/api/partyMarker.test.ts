import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getPartyMarker, upsertPartyMarker, deletePartyMarker } from './partyMarker'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
  BASE_URL: 'http://localhost:8080',
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

const mockMarker = {
  id: 'marker-1',
  game_id: 'game-1',
  map_id: 'map-1',
  x: 50,
  y: 50,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('getPartyMarker', () => {
  it('calls apiFetch GET /games/:gameId/party-marker', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMarker)
    const result = await getPartyMarker('game-1')
    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/party-marker')
    expect(result).toEqual(mockMarker)
  })

  it('returns null when no marker exists', async () => {
    mockApiFetch.mockResolvedValueOnce(null)
    const result = await getPartyMarker('game-1')
    expect(result).toBeNull()
  })
})

describe('upsertPartyMarker', () => {
  it('calls apiFetch PUT /games/:gameId/party-marker with data', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMarker)
    const result = await upsertPartyMarker('game-1', { map_id: 'map-1', x: 50, y: 50 })
    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/party-marker', {
      method: 'PUT',
      body: JSON.stringify({ map_id: 'map-1', x: 50, y: 50 }),
    })
    expect(result).toEqual(mockMarker)
  })
})

describe('deletePartyMarker', () => {
  it('calls apiFetch DELETE /games/:gameId/party-marker', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)
    await deletePartyMarker('game-1')
    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/party-marker', { method: 'DELETE' })
  })
})

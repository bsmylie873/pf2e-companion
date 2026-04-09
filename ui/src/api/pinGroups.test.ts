import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listMapPinGroups,
  createMapPinGroup,
  updatePinGroup,
  addPinToGroup,
  removePinFromGroup,
  disbandPinGroup,
} from './pinGroups'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

const mockPinGroup = {
  id: 'group-1',
  colour: '#ff0000',
  icon: 'circle',
  pin_ids: ['pin-1', 'pin-2'],
}

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('listMapPinGroups', () => {
  it('should call apiFetch GET /games/:gameId/maps/:mapId/pin-groups', async () => {
    mockApiFetch.mockResolvedValueOnce([mockPinGroup])

    const result = await listMapPinGroups('game-1', 'map-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1/pin-groups')
    expect(result).toEqual([mockPinGroup])
  })
})

describe('createMapPinGroup', () => {
  it('should call apiFetch POST with pin_ids array', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPinGroup)

    const pinIds = ['pin-1', 'pin-2']
    const result = await createMapPinGroup('game-1', 'map-1', pinIds)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1/pin-groups', {
      method: 'POST',
      body: JSON.stringify({ pin_ids: pinIds }),
    })
    expect(result).toEqual(mockPinGroup)
  })
})

describe('updatePinGroup', () => {
  it('should call apiFetch PATCH /pin-groups/:groupId with colour', async () => {
    const updatedGroup = { ...mockPinGroup, colour: '#00ff00' }
    mockApiFetch.mockResolvedValueOnce(updatedGroup)

    const result = await updatePinGroup('group-1', { colour: '#00ff00' })

    expect(mockApiFetch).toHaveBeenCalledWith('/pin-groups/group-1', {
      method: 'PATCH',
      body: JSON.stringify({ colour: '#00ff00' }),
    })
    expect(result).toEqual(updatedGroup)
  })

  it('should update icon only when provided', async () => {
    mockApiFetch.mockResolvedValueOnce({ ...mockPinGroup, icon: 'star' })

    await updatePinGroup('group-1', { icon: 'star' })

    expect(mockApiFetch).toHaveBeenCalledWith('/pin-groups/group-1', {
      method: 'PATCH',
      body: JSON.stringify({ icon: 'star' }),
    })
  })
})

describe('addPinToGroup', () => {
  it('should call apiFetch POST /pin-groups/:groupId/pins with pin_id', async () => {
    const updatedGroup = { ...mockPinGroup, pin_ids: ['pin-1', 'pin-2', 'pin-3'] }
    mockApiFetch.mockResolvedValueOnce(updatedGroup)

    const result = await addPinToGroup('group-1', 'pin-3')

    expect(mockApiFetch).toHaveBeenCalledWith('/pin-groups/group-1/pins', {
      method: 'POST',
      body: JSON.stringify({ pin_id: 'pin-3' }),
    })
    expect(result).toEqual(updatedGroup)
  })
})

describe('removePinFromGroup', () => {
  it('should call apiFetch DELETE /pin-groups/:groupId/pins/:pinId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await removePinFromGroup('group-1', 'pin-2')

    expect(mockApiFetch).toHaveBeenCalledWith('/pin-groups/group-1/pins/pin-2', { method: 'DELETE' })
  })
})

describe('disbandPinGroup', () => {
  it('should call apiFetch DELETE /pin-groups/:groupId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await disbandPinGroup('group-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/pin-groups/group-1', { method: 'DELETE' })
  })
})

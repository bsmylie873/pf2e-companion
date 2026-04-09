import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listMaps,
  listArchivedMaps,
  createMap,
  renameMap,
  reorderMaps,
  archiveMap,
  restoreMap,
  uploadMapImage,
} from './maps'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
  BASE_URL: 'http://localhost:8080',
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

const mockMap = {
  id: 'map-1',
  name: 'World Map',
  description: null,
  image_url: null,
  game_id: 'game-1',
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

beforeEach(() => {
  mockApiFetch.mockReset()
  vi.restoreAllMocks()
})

describe('listMaps', () => {
  it('should call apiFetch GET /games/:gameId/maps', async () => {
    mockApiFetch.mockResolvedValueOnce([mockMap])

    const result = await listMaps('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps')
    expect(result).toEqual([mockMap])
  })
})

describe('listArchivedMaps', () => {
  it('should call apiFetch GET /games/:gameId/maps/archived', async () => {
    mockApiFetch.mockResolvedValueOnce([mockMap])

    const result = await listArchivedMaps('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/archived')
    expect(result).toEqual([mockMap])
  })
})

describe('createMap', () => {
  it('should call apiFetch POST /games/:gameId/maps with name and description', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMap)

    const data = { name: 'World Map', description: 'The overworld' }
    const result = await createMap('game-1', data)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    expect(result).toEqual(mockMap)
  })

  it('should create a map with just a name', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMap)

    await createMap('game-1', { name: 'Dungeon' })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps', {
      method: 'POST',
      body: JSON.stringify({ name: 'Dungeon' }),
    })
  })
})

describe('renameMap', () => {
  it('should call apiFetch PATCH /games/:gameId/maps/:mapId with new name', async () => {
    const updatedMap = { ...mockMap, name: 'Renamed Map' }
    mockApiFetch.mockResolvedValueOnce(updatedMap)

    const result = await renameMap('game-1', 'map-1', { name: 'Renamed Map' })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1', {
      method: 'PATCH',
      body: JSON.stringify({ name: 'Renamed Map' }),
    })
    expect(result.name).toBe('Renamed Map')
  })

  it('should patch only description if only description is provided', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMap)

    await renameMap('game-1', 'map-1', { description: 'Updated description' })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1', {
      method: 'PATCH',
      body: JSON.stringify({ description: 'Updated description' }),
    })
  })
})

describe('reorderMaps', () => {
  it('should call apiFetch PATCH /games/:gameId/maps/order with map_ids', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    const mapIds = ['map-3', 'map-1', 'map-2']
    await reorderMaps('game-1', mapIds)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/order', {
      method: 'PATCH',
      body: JSON.stringify({ map_ids: mapIds }),
    })
  })
})

describe('archiveMap', () => {
  it('should call apiFetch DELETE /games/:gameId/maps/:mapId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await archiveMap('game-1', 'map-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1', { method: 'DELETE' })
  })
})

describe('restoreMap', () => {
  it('should call apiFetch POST /games/:gameId/maps/:mapId/restore', async () => {
    mockApiFetch.mockResolvedValueOnce(mockMap)

    const result = await restoreMap('game-1', 'map-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1/restore', { method: 'POST' })
    expect(result).toEqual(mockMap)
  })
})

describe('uploadMapImage', () => {
  it('should POST a FormData with the file to the image upload endpoint', async () => {
    const updatedMap = { ...mockMap, image_url: 'https://cdn.example.com/map.png' }
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: updatedMap }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const file = new File(['image-bytes'], 'map.png', { type: 'image/png' })
    const result = await uploadMapImage('game-1', 'map-1', file)

    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/games/game-1/maps/map-1/image',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
      }),
    )
    // Verify FormData was passed (not JSON)
    const callArgs = fetchSpy.mock.calls[0][1]
    expect(callArgs?.body).toBeInstanceOf(FormData)
    expect(result).toEqual(updatedMap)
  })

  it('should throw with server message when upload fails', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'File too large' }), {
        status: 413,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const file = new File(['big-image'], 'large.png', { type: 'image/png' })
    await expect(uploadMapImage('game-1', 'map-1', file)).rejects.toThrow('File too large')
  })

  it('should throw with Upload failed message when no message in error response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({}), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const file = new File(['bytes'], 'map.png', { type: 'image/png' })
    await expect(uploadMapImage('game-1', 'map-1', file)).rejects.toThrow('Upload failed')
  })
})

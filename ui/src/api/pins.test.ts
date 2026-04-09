import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listMapPins, createMapPin, createPin, updatePin, deletePin } from './pins'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

const mockPin = {
  id: 'pin-1',
  label: 'Tavern',
  x: 100,
  y: 200,
  colour: '#ff0000',
  icon: 'circle',
  map_id: 'map-1',
  note_id: null,
  session_id: null,
  description: null,
}

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('listMapPins', () => {
  it('should call apiFetch GET /games/:gameId/maps/:mapId/pins', async () => {
    mockApiFetch.mockResolvedValueOnce([mockPin])

    const result = await listMapPins('game-1', 'map-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1/pins')
    expect(result).toEqual([mockPin])
  })

  it('should return empty array when no pins exist', async () => {
    mockApiFetch.mockResolvedValueOnce([])

    const result = await listMapPins('game-1', 'map-empty')

    expect(result).toEqual([])
  })
})

describe('createMapPin', () => {
  it('should call apiFetch POST /games/:gameId/maps/:mapId/pins with pin data', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    const pinData = {
      x: 100,
      y: 200,
      label: 'Tavern',
      colour: '#ff0000',
      icon: 'circle',
    }
    const result = await createMapPin('game-1', 'map-1', pinData)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/maps/map-1/pins', {
      method: 'POST',
      body: JSON.stringify({
        label: 'Tavern',
        x: 100,
        y: 200,
        colour: '#ff0000',
        icon: 'circle',
        note_id: undefined,
        description: undefined,
        session_id: undefined,
      }),
    })
    expect(result).toEqual(mockPin)
  })

  it('should default label to empty string when not provided', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    await createMapPin('game-1', 'map-1', { x: 50, y: 75 })

    const body = JSON.parse(mockApiFetch.mock.calls[0][1]?.body as string)
    expect(body.label).toBe('')
  })

  it('should include optional note_id, description, and session_id when provided', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    await createMapPin('game-1', 'map-1', {
      x: 10,
      y: 20,
      note_id: 'note-abc',
      description: 'A cool place',
      session_id: 'session-xyz',
    })

    const body = JSON.parse(mockApiFetch.mock.calls[0][1]?.body as string)
    expect(body.note_id).toBe('note-abc')
    expect(body.description).toBe('A cool place')
    expect(body.session_id).toBe('session-xyz')
  })
})

describe('createPin', () => {
  it('should call apiFetch POST /sessions/:sessionId/pins', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    const formData = {
      session_id: 'session-1',
      label: 'Camp',
      x: 300,
      y: 400,
      colour: '#0000ff',
      icon: 'star',
      description: 'Base camp',
    }
    const result = await createPin(formData)

    expect(mockApiFetch).toHaveBeenCalledWith('/sessions/session-1/pins', {
      method: 'POST',
      body: JSON.stringify({
        label: 'Camp',
        x: 300,
        y: 400,
        colour: '#0000ff',
        icon: 'star',
        description: 'Base camp',
      }),
    })
    expect(result).toEqual(mockPin)
  })

  it('should default label to empty string when not provided', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    await createPin({ session_id: 'session-1', x: 10, y: 20 })

    const body = JSON.parse(mockApiFetch.mock.calls[0][1]?.body as string)
    expect(body.label).toBe('')
  })
})

describe('updatePin', () => {
  it('should call apiFetch PATCH /pins/:pinId with update data', async () => {
    const updatedPin = { ...mockPin, label: 'Updated Tavern', x: 150 }
    mockApiFetch.mockResolvedValueOnce(updatedPin)

    const result = await updatePin('pin-1', { label: 'Updated Tavern', x: 150 })

    expect(mockApiFetch).toHaveBeenCalledWith('/pins/pin-1', {
      method: 'PATCH',
      body: JSON.stringify({ label: 'Updated Tavern', x: 150 }),
    })
    expect(result).toEqual(updatedPin)
  })

  it('should support setting note_id to null', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPin)

    await updatePin('pin-1', { note_id: null })

    const body = JSON.parse(mockApiFetch.mock.calls[0][1]?.body as string)
    expect(body.note_id).toBeNull()
  })
})

describe('deletePin', () => {
  it('should call apiFetch DELETE /pins/:pinId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await deletePin('pin-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/pins/pin-1', { method: 'DELETE' })
  })
})

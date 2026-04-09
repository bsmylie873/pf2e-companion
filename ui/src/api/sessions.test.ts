import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listGameSessions,
  getSession,
  createSession,
  updateSession,
  deleteSession,
  updateSessionNotes,
  listGameSessionsPaginated,
} from './sessions'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
  apiFetchRaw: vi.fn(),
}))

import { apiFetch, apiFetchRaw } from './client'

const mockApiFetch = vi.mocked(apiFetch)
const mockApiFetchRaw = vi.mocked(apiFetchRaw)

const mockSession = {
  id: 'session-1',
  title: 'The Dragon Awakens',
  description: 'First encounter',
  game_id: 'game-1',
  notes: null,
  version: 1,
  folder_id: null,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

beforeEach(() => {
  mockApiFetch.mockReset()
  mockApiFetchRaw.mockReset()
})

describe('listGameSessions', () => {
  it('should call apiFetch GET /games/:gameId/sessions', async () => {
    mockApiFetch.mockResolvedValueOnce([mockSession])

    const result = await listGameSessions('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/sessions')
    expect(result).toEqual([mockSession])
  })

  it('should return empty array when no sessions exist', async () => {
    mockApiFetch.mockResolvedValueOnce([])

    const result = await listGameSessions('game-empty')

    expect(result).toEqual([])
  })
})

describe('getSession', () => {
  it('should call apiFetch GET /sessions/:sessionId', async () => {
    mockApiFetch.mockResolvedValueOnce(mockSession)

    const result = await getSession('session-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/sessions/session-1')
    expect(result).toEqual(mockSession)
  })
})

describe('createSession', () => {
  it('should call apiFetch POST /games/:gameId/sessions with form data', async () => {
    mockApiFetch.mockResolvedValueOnce(mockSession)

    const formData = { title: 'The Dragon Awakens', description: 'First encounter' }
    const result = await createSession('game-1', formData)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/sessions', {
      method: 'POST',
      body: JSON.stringify(formData),
    })
    expect(result).toEqual(mockSession)
  })
})

describe('updateSession', () => {
  it('should call apiFetch PATCH /sessions/:sessionId with update data', async () => {
    const updatedSession = { ...mockSession, title: 'Updated Title' }
    mockApiFetch.mockResolvedValueOnce(updatedSession)

    const result = await updateSession('session-1', { title: 'Updated Title' })

    expect(mockApiFetch).toHaveBeenCalledWith('/sessions/session-1', {
      method: 'PATCH',
      body: JSON.stringify({ title: 'Updated Title' }),
    })
    expect(result).toEqual(updatedSession)
  })

  it('should support updating folder_id', async () => {
    mockApiFetch.mockResolvedValueOnce({ ...mockSession, folder_id: 'folder-5' })

    await updateSession('session-1', { folder_id: 'folder-5' })

    const body = JSON.parse(mockApiFetch.mock.calls[0][1]?.body as string)
    expect(body.folder_id).toBe('folder-5')
  })
})

describe('deleteSession', () => {
  it('should call apiFetch DELETE /sessions/:sessionId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await deleteSession('session-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/sessions/session-1', { method: 'DELETE' })
  })
})

describe('updateSessionNotes', () => {
  it('should call apiFetch PATCH /sessions/:sessionId with notes content and version', async () => {
    const updatedSession = { ...mockSession, version: 2 }
    mockApiFetch.mockResolvedValueOnce(updatedSession)

    const notes = { type: 'doc', content: [{ type: 'paragraph', content: [] }] }
    const result = await updateSessionNotes('session-1', { notes, version: 1 })

    expect(mockApiFetch).toHaveBeenCalledWith('/sessions/session-1', {
      method: 'PATCH',
      body: JSON.stringify({ notes, version: 1 }),
    })
    expect(result).toEqual(updatedSession)
  })
})

describe('listGameSessionsPaginated', () => {
  it('should call apiFetchRaw with page and limit params', async () => {
    const mockResponse = { data: [mockSession], total: 1, page: 1, limit: 10 }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    const result = await listGameSessionsPaginated('game-1', { page: 1, limit: 10 })

    const url = mockApiFetchRaw.mock.calls[0][0] as string
    expect(url).toContain('/games/game-1/sessions?')
    expect(url).toContain('page=1')
    expect(url).toContain('limit=10')
    expect(result).toEqual(mockResponse)
  })

  it('should return the full paginated response structure', async () => {
    const mockResponse = { data: [], total: 50, page: 3, limit: 5 }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    const result = await listGameSessionsPaginated('game-1', { page: 3, limit: 5 })

    expect(result.total).toBe(50)
    expect(result.page).toBe(3)
    expect(result.limit).toBe(5)
  })
})

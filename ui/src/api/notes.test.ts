import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  listGameNotes,
  getNote,
  createNote,
  updateNote,
  deleteNote,
  updateNoteContent,
  listGameNotesPaginated,
} from './notes'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
  apiFetchRaw: vi.fn(),
}))

import { apiFetch, apiFetchRaw } from './client'

const mockApiFetch = vi.mocked(apiFetch)
const mockApiFetchRaw = vi.mocked(apiFetchRaw)

const mockNote = {
  id: 'note-1',
  title: 'Session Notes',
  content: null,
  game_id: 'game-1',
  session_id: null,
  folder_id: null,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

beforeEach(() => {
  mockApiFetch.mockReset()
  mockApiFetchRaw.mockReset()
})

describe('listGameNotes', () => {
  it('should call apiFetch GET /games/:gameId/notes without params', async () => {
    mockApiFetch.mockResolvedValueOnce([mockNote])

    const result = await listGameNotes('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/notes')
    expect(result).toEqual([mockNote])
  })

  it('should include sort param in query string when provided', async () => {
    mockApiFetch.mockResolvedValueOnce([mockNote])

    await listGameNotes('game-1', { sort: 'title' })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/notes?sort=title')
  })

  it('should include session_id param when provided', async () => {
    mockApiFetch.mockResolvedValueOnce([mockNote])

    await listGameNotes('game-1', { session_id: 'session-42' })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/notes?session_id=session-42')
  })

  it('should include unlinked=true when unlinked flag is set', async () => {
    mockApiFetch.mockResolvedValueOnce([mockNote])

    await listGameNotes('game-1', { unlinked: true })

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/notes?unlinked=true')
  })

  it('should combine multiple params in the query string', async () => {
    mockApiFetch.mockResolvedValueOnce([mockNote])

    await listGameNotes('game-1', { sort: 'updated_at', unlinked: true })

    const call = mockApiFetch.mock.calls[0][0] as string
    expect(call).toContain('sort=updated_at')
    expect(call).toContain('unlinked=true')
  })
})

describe('getNote', () => {
  it('should call apiFetch GET /notes/:noteId', async () => {
    mockApiFetch.mockResolvedValueOnce(mockNote)

    const result = await getNote('note-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/notes/note-1')
    expect(result).toEqual(mockNote)
  })
})

describe('createNote', () => {
  it('should call apiFetch POST /games/:gameId/notes with form data', async () => {
    mockApiFetch.mockResolvedValueOnce(mockNote)

    const formData = { title: 'New Note', session_id: 'session-1', folder_id: null }
    const result = await createNote('game-1', formData)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/notes', {
      method: 'POST',
      body: JSON.stringify(formData),
    })
    expect(result).toEqual(mockNote)
  })
})

describe('updateNote', () => {
  it('should call apiFetch PATCH /notes/:noteId with update data', async () => {
    const updatedNote = { ...mockNote, title: 'Updated Title' }
    mockApiFetch.mockResolvedValueOnce(updatedNote)

    const result = await updateNote('note-1', { title: 'Updated Title' })

    expect(mockApiFetch).toHaveBeenCalledWith('/notes/note-1', {
      method: 'PATCH',
      body: JSON.stringify({ title: 'Updated Title' }),
    })
    expect(result).toEqual(updatedNote)
  })
})

describe('deleteNote', () => {
  it('should call apiFetch DELETE /notes/:noteId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await deleteNote('note-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/notes/note-1', { method: 'DELETE' })
  })
})

describe('updateNoteContent', () => {
  it('should call apiFetch PATCH /notes/:noteId with content and version', async () => {
    mockApiFetch.mockResolvedValueOnce(mockNote)

    const content = { type: 'doc', content: [] }
    const result = await updateNoteContent('note-1', { content, version: 5 })

    expect(mockApiFetch).toHaveBeenCalledWith('/notes/note-1', {
      method: 'PATCH',
      body: JSON.stringify({ content, version: 5 }),
    })
    expect(result).toEqual(mockNote)
  })
})

describe('listGameNotesPaginated', () => {
  it('should call apiFetchRaw with page and limit params', async () => {
    const mockResponse = { data: [mockNote], total: 1, page: 1, limit: 10 }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    const result = await listGameNotesPaginated('game-1', { page: 1, limit: 10 })

    expect(mockApiFetchRaw).toHaveBeenCalledWith(
      expect.stringContaining('/games/game-1/notes?'),
    )
    const url = mockApiFetchRaw.mock.calls[0][0] as string
    expect(url).toContain('page=1')
    expect(url).toContain('limit=10')
    expect(result).toEqual(mockResponse)
  })

  it('should include optional sort param when provided', async () => {
    const mockResponse = { data: [], total: 0, page: 1, limit: 20 }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    await listGameNotesPaginated('game-1', { page: 1, limit: 20, sort: 'title' })

    const url = mockApiFetchRaw.mock.calls[0][0] as string
    expect(url).toContain('sort=title')
  })

  it('should include session_id param when provided', async () => {
    mockApiFetchRaw.mockResolvedValueOnce({ data: [], total: 0, page: 1, limit: 10 })

    await listGameNotesPaginated('game-1', { page: 1, limit: 10, session_id: 'sess-99' })

    const url = mockApiFetchRaw.mock.calls[0][0] as string
    expect(url).toContain('session_id=sess-99')
  })

  it('should include unlinked param when set to true', async () => {
    mockApiFetchRaw.mockResolvedValueOnce({ data: [], total: 0, page: 1, limit: 10 })

    await listGameNotesPaginated('game-1', { page: 1, limit: 10, unlinked: true })

    const url = mockApiFetchRaw.mock.calls[0][0] as string
    expect(url).toContain('unlinked=true')
  })
})

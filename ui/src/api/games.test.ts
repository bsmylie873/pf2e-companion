import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listGamesPaginated } from './games'

vi.mock('./client', () => ({
  apiFetchRaw: vi.fn(),
}))

import { apiFetchRaw } from './client'

const mockApiFetchRaw = vi.mocked(apiFetchRaw)

beforeEach(() => {
  mockApiFetchRaw.mockReset()
})

describe('listGamesPaginated', () => {
  it('should call apiFetchRaw with correct page and limit query params', async () => {
    const mockResponse = {
      data: [{ id: 'g1', title: 'My Campaign', description: null, splash_image_url: null, foundry_data: null, created_at: '2024-01-01', updated_at: '2024-01-01' }],
      total: 1,
      page: 1,
      limit: 10,
    }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    const result = await listGamesPaginated({ page: 1, limit: 10 })

    expect(mockApiFetchRaw).toHaveBeenCalledWith('/games?page=1&limit=10')
    expect(result).toEqual(mockResponse)
  })

  it('should pass the correct page number in the query string', async () => {
    const mockResponse = { data: [], total: 0, page: 3, limit: 5 }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    await listGamesPaginated({ page: 3, limit: 5 })

    expect(mockApiFetchRaw).toHaveBeenCalledWith('/games?page=3&limit=5')
  })

  it('should return the full paginated response including total, page, and limit', async () => {
    const mockResponse = {
      data: [],
      total: 42,
      page: 2,
      limit: 20,
    }
    mockApiFetchRaw.mockResolvedValueOnce(mockResponse)

    const result = await listGamesPaginated({ page: 2, limit: 20 })

    expect(result.total).toBe(42)
    expect(result.page).toBe(2)
    expect(result.limit).toBe(20)
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'

// We import after setting up mocks to avoid stale module-level state issues.
// Re-import apiFetch/apiFetchRaw dynamically where needed.

describe('BASE_URL and WS_BASE_URL', () => {
  it('should export BASE_URL as a string', async () => {
    const { BASE_URL } = await import('./client')
    expect(typeof BASE_URL).toBe('string')
    expect(BASE_URL.length).toBeGreaterThan(0)
  })

  it('should export WS_BASE_URL as a string starting with ws', async () => {
    const { WS_BASE_URL } = await import('./client')
    expect(typeof WS_BASE_URL).toBe('string')
    expect(WS_BASE_URL).toMatch(/^ws/)
  })
})

describe('apiFetch', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('should make a fetch call with credentials and Content-Type header', async () => {
    const { apiFetch, BASE_URL } = await import('./client')
    const mockData = { id: '1', name: 'test' }
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: mockData }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await apiFetch('/test')

    expect(fetchSpy).toHaveBeenCalledWith(
      `${BASE_URL}/test`,
      expect.objectContaining({
        credentials: 'include',
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    )
  })

  it('should return the data property from the JSON response', async () => {
    const { apiFetch } = await import('./client')
    const mockData = { id: '42', value: 'hello' }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: mockData }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await apiFetch<typeof mockData>('/test')

    expect(result).toEqual(mockData)
  })

  it('should pass method and body to fetch', async () => {
    const { apiFetch, BASE_URL } = await import('./client')
    const body = { name: 'campaign' }
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { id: '1' } }), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await apiFetch('/items', { method: 'POST', body: JSON.stringify(body) })

    expect(fetchSpy).toHaveBeenCalledWith(
      `${BASE_URL}/items`,
      expect.objectContaining({ method: 'POST', body: JSON.stringify(body) }),
    )
  })

  it('should throw an error when response is not ok', async () => {
    const { apiFetch } = await import('./client')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'Not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(apiFetch('/missing')).rejects.toThrow('Not found')
  })

  it('should throw with generic message when error response has no message', async () => {
    const { apiFetch } = await import('./client')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({}), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(apiFetch('/error')).rejects.toThrow('Request failed: 500')
  })

  it('should throw immediately on 401 for /auth/ paths without attempting refresh', async () => {
    const { apiFetch } = await import('./client')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'Unauthorized' }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(apiFetch('/auth/me')).rejects.toThrow()
    // fetch should only have been called once (no refresh attempt)
    expect(vi.mocked(globalThis.fetch)).toHaveBeenCalledTimes(1)
  })

  it('should attempt refresh on 401 for non-auth paths and retry on success', async () => {
    const { apiFetch } = await import('./client')
    const mockData = { id: '1' }

    vi.spyOn(globalThis, 'fetch')
      // First call: original request returns 401
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'Unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      // Second call: refresh succeeds
      .mockResolvedValueOnce(new Response(null, { status: 200 }))
      // Third call: retry succeeds
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ data: mockData }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )

    const result = await apiFetch<typeof mockData>('/games')

    expect(result).toEqual(mockData)
    expect(vi.mocked(globalThis.fetch)).toHaveBeenCalledTimes(3)
  })

  it('should redirect and throw when refresh fails on 401', async () => {
    const { apiFetch } = await import('./client')

    // Mock window.location
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      value: { href: '' },
      writable: true,
      configurable: true,
    })

    vi.spyOn(globalThis, 'fetch')
      // First call: 401
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'Unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      // Second call: refresh fails
      .mockResolvedValueOnce(new Response(null, { status: 401 }))

    await expect(apiFetch('/games')).rejects.toThrow('Session expired')
    expect(window.location.href).toBe('/?expired=true')

    // Restore
    Object.defineProperty(window, 'location', { value: originalLocation, writable: true, configurable: true })
  })
})

describe('apiFetchRaw', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('should return the full JSON response (not just .data)', async () => {
    const { apiFetchRaw } = await import('./client')
    const fullResponse = { data: [{ id: '1' }], total: 1, page: 1, limit: 10 }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(fullResponse), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await apiFetchRaw<typeof fullResponse>('/games')

    expect(result).toEqual(fullResponse)
    // Should NOT unwrap .data
    expect((result as typeof fullResponse).total).toBe(1)
  })

  it('should throw an error when response is not ok', async () => {
    const { apiFetchRaw } = await import('./client')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'Forbidden' }), {
        status: 403,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(apiFetchRaw('/protected')).rejects.toThrow('Forbidden')
  })

  it('should include credentials and Content-Type in the request', async () => {
    const { apiFetchRaw, BASE_URL } = await import('./client')
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: [] }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await apiFetchRaw('/items')

    expect(fetchSpy).toHaveBeenCalledWith(
      `${BASE_URL}/items`,
      expect.objectContaining({
        credentials: 'include',
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    )
  })
})

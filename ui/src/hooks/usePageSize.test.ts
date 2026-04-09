import { renderHook, waitFor } from '@testing-library/react'
import { usePageSize } from './usePageSize'

vi.mock('../api/preferences', () => ({
  getPreferences: vi.fn(),
}))

import { getPreferences } from '../api/preferences'

const basePrefs = {
  default_game_id: null,
  default_pin_colour: null,
  default_pin_icon: null,
  sidebar_state: null,
  default_view_mode: null,
  map_editor_mode: 'modal' as const,
}

describe('usePageSize', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return the fallback page size of 10 before preferences load', () => {
    vi.mocked(getPreferences).mockImplementation(() => new Promise(() => {}))
    const { result } = renderHook(() => usePageSize('campaigns'))
    expect(result.current).toBe(10)
  })

  it('should return 10 when page_size preference is null', async () => {
    vi.mocked(getPreferences).mockResolvedValue({ ...basePrefs, page_size: null })
    const { result } = renderHook(() => usePageSize('notes'))
    await waitFor(() => expect(vi.mocked(getPreferences)).toHaveBeenCalled())
    expect(result.current).toBe(10)
  })

  it('should return the resource-specific page size when defined', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...basePrefs,
      page_size: { default: 10, campaigns: 25 },
    })
    const { result } = renderHook(() => usePageSize('campaigns'))
    await waitFor(() => expect(result.current).toBe(25))
  })

  it('should fall back to the default page size when the resource override is null', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...basePrefs,
      page_size: { default: 20, campaigns: null },
    })
    const { result } = renderHook(() => usePageSize('campaigns'))
    await waitFor(() => expect(result.current).toBe(20))
  })

  it('should fall back to the default page size when the resource key is absent', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...basePrefs,
      page_size: { default: 15 },
    })
    const { result } = renderHook(() => usePageSize('sessions'))
    await waitFor(() => expect(result.current).toBe(15))
  })

  it('should stay at the fallback when both resource and default page sizes are 0 or missing', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...basePrefs,
      page_size: { default: 0 },
    })
    const { result } = renderHook(() => usePageSize('notes'))
    await waitFor(() => expect(vi.mocked(getPreferences)).toHaveBeenCalled())
    expect(result.current).toBe(10)
  })

  it('should stay at the fallback when getPreferences throws', async () => {
    vi.mocked(getPreferences).mockRejectedValue(new Error('Network error'))
    const { result } = renderHook(() => usePageSize('sessions'))
    await waitFor(() => expect(vi.mocked(getPreferences)).toHaveBeenCalled())
    expect(result.current).toBe(10)
  })

  it('should re-fetch preferences when the resource changes', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...basePrefs,
      page_size: { default: 10, campaigns: 30, sessions: 5 },
    })

    const { result, rerender } = renderHook(
      ({ resource }: { resource: 'campaigns' | 'sessions' | 'notes' }) =>
        usePageSize(resource),
      { initialProps: { resource: 'campaigns' } },
    )

    await waitFor(() => expect(result.current).toBe(30))

    rerender({ resource: 'sessions' })
    await waitFor(() => expect(result.current).toBe(5))
  })
})

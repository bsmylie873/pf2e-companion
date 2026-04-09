import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getPreferences, updatePreferences } from './preferences'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

const mockPreferences = {
  default_game_id: 'game-1',
  default_pin_colour: '#ff0000',
  default_pin_icon: 'circle',
  sidebar_state: null,
  default_view_mode: null,
  map_editor_mode: 'modal' as const,
  page_size: null,
}

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('getPreferences', () => {
  it('should call apiFetch GET /preferences', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPreferences)

    const result = await getPreferences()

    expect(mockApiFetch).toHaveBeenCalledWith('/preferences')
    expect(result).toEqual(mockPreferences)
  })

  it('should return preferences with all expected fields', async () => {
    mockApiFetch.mockResolvedValueOnce(mockPreferences)

    const result = await getPreferences()

    expect(result).toHaveProperty('default_game_id')
    expect(result).toHaveProperty('default_pin_colour')
    expect(result).toHaveProperty('default_pin_icon')
    expect(result).toHaveProperty('map_editor_mode')
  })
})

describe('updatePreferences', () => {
  it('should call apiFetch PATCH /preferences with partial updates', async () => {
    const updatedPrefs = { ...mockPreferences, default_pin_colour: '#00ff00' }
    mockApiFetch.mockResolvedValueOnce(updatedPrefs)

    const result = await updatePreferences({ default_pin_colour: '#00ff00' })

    expect(mockApiFetch).toHaveBeenCalledWith('/preferences', {
      method: 'PATCH',
      body: JSON.stringify({ default_pin_colour: '#00ff00' }),
    })
    expect(result).toEqual(updatedPrefs)
  })

  it('should support updating map_editor_mode', async () => {
    const updatedPrefs = { ...mockPreferences, map_editor_mode: 'navigate' as const }
    mockApiFetch.mockResolvedValueOnce(updatedPrefs)

    const result = await updatePreferences({ map_editor_mode: 'navigate' })

    expect(mockApiFetch).toHaveBeenCalledWith('/preferences', {
      method: 'PATCH',
      body: JSON.stringify({ map_editor_mode: 'navigate' }),
    })
    expect(result.map_editor_mode).toBe('navigate')
  })

  it('should support updating page_size preferences', async () => {
    const pageSizeUpdate = { page_size: { default: 20, sessions: 10, notes: 15, campaigns: null } }
    const updatedPrefs = { ...mockPreferences, ...pageSizeUpdate }
    mockApiFetch.mockResolvedValueOnce(updatedPrefs)

    await updatePreferences(pageSizeUpdate)

    expect(mockApiFetch).toHaveBeenCalledWith('/preferences', {
      method: 'PATCH',
      body: JSON.stringify(pageSizeUpdate),
    })
  })

  it('should support updating sidebar_state', async () => {
    const sidebarUpdate = {
      sidebar_state: {
        'game-1': { panelOpen: true, 'folder-1': true },
      },
    }
    mockApiFetch.mockResolvedValueOnce({ ...mockPreferences, ...sidebarUpdate })

    await updatePreferences(sidebarUpdate)

    expect(mockApiFetch).toHaveBeenCalledWith('/preferences', {
      method: 'PATCH',
      body: JSON.stringify(sidebarUpdate),
    })
  })
})

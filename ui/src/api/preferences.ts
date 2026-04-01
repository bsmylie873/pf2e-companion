import { apiFetch } from './client'

/** Per-game sidebar UI state, keyed by folder ID for expanded booleans. */
export interface GameSidebarState {
  panelOpen: boolean
  [folderId: string]: boolean
}

export interface UserPreferences {
  default_game_id: string | null
  default_pin_colour: string | null
  default_pin_icon: string | null
  sidebar_state: Record<string, GameSidebarState> | null
  default_view_mode: Record<string, 'list' | 'grid'> | null
}

export async function getPreferences(): Promise<UserPreferences> {
  return apiFetch<UserPreferences>('/preferences')
}

export async function updatePreferences(updates: Partial<UserPreferences>): Promise<UserPreferences> {
  return apiFetch<UserPreferences>('/preferences', { method: 'PATCH', body: JSON.stringify(updates) })
}

import { apiFetch } from './client'

export interface UserPreferences {
  default_game_id: string | null
  default_pin_colour: string | null
  default_pin_icon: string | null
}

export async function getPreferences(): Promise<UserPreferences> {
  return apiFetch<UserPreferences>('/preferences')
}

export async function updatePreferences(updates: Partial<UserPreferences>): Promise<UserPreferences> {
  return apiFetch<UserPreferences>('/preferences', { method: 'PATCH', body: JSON.stringify(updates) })
}

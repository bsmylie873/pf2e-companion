import { apiFetch } from './client'
import type { PartyMarker } from '../types/map'

export function getPartyMarker(gameId: string): Promise<PartyMarker | null> {
  return apiFetch<PartyMarker | null>(`/games/${gameId}/party-marker`)
}

export function upsertPartyMarker(
  gameId: string,
  data: { map_id: string; x: number; y: number },
): Promise<PartyMarker> {
  return apiFetch<PartyMarker>(`/games/${gameId}/party-marker`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deletePartyMarker(gameId: string): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/party-marker`, { method: 'DELETE' })
}

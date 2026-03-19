import { apiFetch } from './client'
import type { GameMembership } from '../types/membership'

export function listMemberships(gameId: string): Promise<GameMembership[]> {
  return apiFetch<GameMembership[]>(`/memberships?game_id=${gameId}`)
}

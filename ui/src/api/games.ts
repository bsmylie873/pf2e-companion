import { apiFetchRaw } from './client'
import type { Game } from '../types/game'
import type { PaginatedResponse } from '../types/pagination'

export function listGamesPaginated(params: { page: number; limit: number }): Promise<PaginatedResponse<Game>> {
  const query = new URLSearchParams({ page: String(params.page), limit: String(params.limit) })
  return apiFetchRaw<PaginatedResponse<Game>>(`/games?${query}`)
}

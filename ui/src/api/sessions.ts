import { apiFetch, apiFetchRaw } from './client'
import type { Session, SessionFormData } from '../types/session'
import type { JSONContent } from '@tiptap/react'
import type { PaginatedResponse } from '../types/pagination'

export function listGameSessions(gameId: string): Promise<Session[]> {
  return apiFetch<Session[]>(`/games/${gameId}/sessions`)
}

export function getSession(sessionId: string): Promise<Session> {
  return apiFetch<Session>(`/sessions/${sessionId}`)
}

export function createSession(gameId: string, data: SessionFormData): Promise<Session> {
  return apiFetch<Session>(`/games/${gameId}/sessions`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateSession(sessionId: string, data: Record<string, unknown>): Promise<Session> {
  return apiFetch<Session>(`/sessions/${sessionId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function deleteSession(sessionId: string): Promise<void> {
  return apiFetch<void>(`/sessions/${sessionId}`, { method: 'DELETE' })
}

export function updateSessionNotes(sessionId: string, data: { notes: JSONContent; version: number }): Promise<Session> {
  return apiFetch<Session>(`/sessions/${sessionId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function listGameSessionsPaginated(
  gameId: string,
  params: { page: number; limit: number },
): Promise<PaginatedResponse<Session>> {
  const query = new URLSearchParams({ page: String(params.page), limit: String(params.limit) })
  return apiFetchRaw<PaginatedResponse<Session>>(`/games/${gameId}/sessions?${query}`)
}

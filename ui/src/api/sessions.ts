import { apiFetch } from './client'
import type { Session, SessionFormData } from '../types/session'
import type { JSONContent } from '@tiptap/react'

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

export function updateSessionNotes(sessionId: string, data: { notes: JSONContent }): Promise<Session> {
  return apiFetch<Session>(`/sessions/${sessionId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

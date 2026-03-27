import { apiFetch } from './client'
import type { SessionPin, SessionPinFormData } from '../types/pin'

export function listGamePins(gameId: string): Promise<SessionPin[]> {
  return apiFetch<SessionPin[]>(`/games/${gameId}/pins`)
}

export function createPin(data: SessionPinFormData): Promise<SessionPin> {
  return apiFetch<SessionPin>(`/sessions/${data.session_id}/pins`, {
    method: 'POST',
    body: JSON.stringify({ label: data.label ?? '', x: data.x, y: data.y, colour: data.colour, icon: data.icon }),
  })
}

export function createGamePin(gameId: string, data: { x: number; y: number; label?: string; colour?: string; icon?: string; note_id?: string }): Promise<SessionPin> {
  return apiFetch<SessionPin>(`/games/${gameId}/pins`, {
    method: 'POST',
    body: JSON.stringify({ label: data.label ?? '', x: data.x, y: data.y, colour: data.colour, icon: data.icon, note_id: data.note_id }),
  })
}

export function updatePin(pinId: string, data: { x?: number; y?: number; label?: string; colour?: string; icon?: string; note_id?: string | null }): Promise<SessionPin> {
  return apiFetch<SessionPin>(`/pins/${pinId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function deletePin(pinId: string): Promise<void> {
  return apiFetch<void>(`/pins/${pinId}`, { method: 'DELETE' })
}

import { apiFetch } from './client'
import type { SessionPin, SessionPinFormData } from '../types/pin'

export function listGamePins(gameId: string): Promise<SessionPin[]> {
  return apiFetch<SessionPin[]>(`/games/${gameId}/pins`)
}

export function createPin(data: SessionPinFormData): Promise<SessionPin> {
  return apiFetch<SessionPin>(`/sessions/${data.session_id}/pins`, {
    method: 'POST',
    body: JSON.stringify({ label: data.label ?? '', x: data.x, y: data.y, pin_type_id: data.pin_type_id }),
  })
}

export function updatePin(pinId: string, data: { x?: number; y?: number; label?: string; pin_type_id?: number }): Promise<SessionPin> {
  return apiFetch<SessionPin>(`/pins/${pinId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function deletePin(pinId: string): Promise<void> {
  return apiFetch<void>(`/pins/${pinId}`, { method: 'DELETE' })
}

import { apiFetch } from './client'
import type { PinGroup } from '../types/pin'

export function listGamePinGroups(gameId: string): Promise<PinGroup[]> {
  return apiFetch<PinGroup[]>(`/games/${gameId}/pin-groups`)
}
export function createPinGroup(gameId: string, pinIds: string[]): Promise<PinGroup> {
  return apiFetch<PinGroup>(`/games/${gameId}/pin-groups`, { method: 'POST', body: JSON.stringify({ pin_ids: pinIds }) })
}
export function updatePinGroup(groupId: string, data: { colour?: string; icon?: string }): Promise<PinGroup> {
  return apiFetch<PinGroup>(`/pin-groups/${groupId}`, { method: 'PATCH', body: JSON.stringify(data) })
}
export function addPinToGroup(groupId: string, pinId: string): Promise<PinGroup> {
  return apiFetch<PinGroup>(`/pin-groups/${groupId}/pins`, { method: 'POST', body: JSON.stringify({ pin_id: pinId }) })
}
export function removePinFromGroup(groupId: string, pinId: string): Promise<void> {
  return apiFetch<void>(`/pin-groups/${groupId}/pins/${pinId}`, { method: 'DELETE' })
}
export function disbandPinGroup(groupId: string): Promise<void> {
  return apiFetch<void>(`/pin-groups/${groupId}`, { method: 'DELETE' })
}

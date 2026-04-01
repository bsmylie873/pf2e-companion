import { apiFetch, BASE_URL } from './client'
import type { GameMap } from '../types/map'

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

export function listMaps(gameId: string): Promise<GameMap[]> {
  return apiFetch<GameMap[]>(`/games/${gameId}/maps`)
}

export function listArchivedMaps(gameId: string): Promise<GameMap[]> {
  return apiFetch<GameMap[]>(`/games/${gameId}/maps/archived`)
}

export function createMap(gameId: string, data: { name: string; description?: string }): Promise<GameMap> {
  return apiFetch<GameMap>(`/games/${gameId}/maps`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function renameMap(gameId: string, mapId: string, data: { name?: string; description?: string }): Promise<GameMap> {
  return apiFetch<GameMap>(`/games/${gameId}/maps/${mapId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function reorderMaps(gameId: string, mapIds: string[]): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/maps/order`, {
    method: 'PATCH',
    body: JSON.stringify({ map_ids: mapIds }),
  })
}

export function archiveMap(gameId: string, mapId: string): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/maps/${mapId}`, { method: 'DELETE' })
}

export function restoreMap(gameId: string, mapId: string): Promise<GameMap> {
  return apiFetch<GameMap>(`/games/${gameId}/maps/${mapId}/restore`, { method: 'POST' })
}

export async function uploadMapImage(gameId: string, mapId: string, file: File): Promise<GameMap> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`${BASE_URL}/games/${gameId}/maps/${mapId}/image`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'X-CSRF-Token': getCsrfToken() },
    body: formData,
  })
  if (!res.ok) {
    const json = await res.json()
    throw new Error(json.message ?? 'Upload failed')
  }
  const json = await res.json()
  return json.data as GameMap
}

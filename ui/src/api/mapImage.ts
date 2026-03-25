import { apiFetch, BASE_URL } from './client'
import type { Game } from '../types/game'

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

export async function uploadMapImage(gameId: string, file: File): Promise<Game> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`${BASE_URL}/games/${gameId}/map-image`, {
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
  return json.data as Game
}

export function deleteMapImage(gameId: string): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/map-image`, { method: 'DELETE' })
}

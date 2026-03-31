import { apiFetch } from './client'
import type { Folder } from '../types/folder'

export function listFolders(gameId: string, type: 'session' | 'note'): Promise<Folder[]> {
  return apiFetch<Folder[]>(`/games/${gameId}/folders?type=${type}`)
}

export function createFolder(gameId: string, data: { name: string; folder_type: string; visibility?: string }): Promise<Folder> {
  return apiFetch<Folder>(`/games/${gameId}/folders`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function renameFolder(folderId: string, name: string): Promise<Folder> {
  return apiFetch<Folder>(`/folders/${folderId}`, {
    method: 'PATCH',
    body: JSON.stringify({ name }),
  })
}

export function deleteFolder(folderId: string): Promise<void> {
  return apiFetch<void>(`/folders/${folderId}`, { method: 'DELETE' })
}

export function reorderFolders(gameId: string, folderType: string, folderIds: string[]): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/folders/reorder`, {
    method: 'PUT',
    body: JSON.stringify({ folder_type: folderType, folder_ids: folderIds }),
  })
}

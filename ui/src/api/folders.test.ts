import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listFolders, createFolder, renameFolder, deleteFolder, reorderFolders } from './folders'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('listFolders', () => {
  it('should call apiFetch GET with gameId and type query param', async () => {
    const mockFolders = [{ id: 'f1', name: 'Chapter 1', folder_type: 'session' }]
    mockApiFetch.mockResolvedValueOnce(mockFolders)

    const result = await listFolders('game-1', 'session')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/folders?type=session')
    expect(result).toEqual(mockFolders)
  })

  it('should support note folder type', async () => {
    mockApiFetch.mockResolvedValueOnce([])

    await listFolders('game-1', 'note')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/folders?type=note')
  })
})

describe('createFolder', () => {
  it('should call apiFetch POST with folder data', async () => {
    const mockFolder = { id: 'f2', name: 'New Folder', folder_type: 'note' }
    mockApiFetch.mockResolvedValueOnce(mockFolder)

    const data = { name: 'New Folder', folder_type: 'note', visibility: 'public' }
    const result = await createFolder('game-1', data)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/folders', {
      method: 'POST',
      body: JSON.stringify(data),
    })
    expect(result).toEqual(mockFolder)
  })

  it('should create a folder without visibility field', async () => {
    const mockFolder = { id: 'f3', name: 'Private Folder', folder_type: 'session' }
    mockApiFetch.mockResolvedValueOnce(mockFolder)

    await createFolder('game-2', { name: 'Private Folder', folder_type: 'session' })

    expect(mockApiFetch).toHaveBeenCalledWith(
      '/games/game-2/folders',
      expect.objectContaining({ method: 'POST' }),
    )
  })
})

describe('renameFolder', () => {
  it('should call apiFetch PATCH /folders/:folderId with new name', async () => {
    const updatedFolder = { id: 'f1', name: 'Renamed Folder', folder_type: 'session' }
    mockApiFetch.mockResolvedValueOnce(updatedFolder)

    const result = await renameFolder('f1', 'Renamed Folder')

    expect(mockApiFetch).toHaveBeenCalledWith('/folders/f1', {
      method: 'PATCH',
      body: JSON.stringify({ name: 'Renamed Folder' }),
    })
    expect(result).toEqual(updatedFolder)
  })
})

describe('deleteFolder', () => {
  it('should call apiFetch DELETE /folders/:folderId', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await deleteFolder('f1')

    expect(mockApiFetch).toHaveBeenCalledWith('/folders/f1', { method: 'DELETE' })
  })
})

describe('reorderFolders', () => {
  it('should call apiFetch PUT with folder IDs in correct order', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    const folderIds = ['f3', 'f1', 'f2']
    await reorderFolders('game-1', 'session', folderIds)

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/folders/reorder', {
      method: 'PUT',
      body: JSON.stringify({ folder_type: 'session', folder_ids: folderIds }),
    })
  })
})

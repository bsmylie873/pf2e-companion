import { BASE_URL } from './client'

export interface ImportSummary {
  sessions_created: number
  sessions_skipped: number
  sessions_overwritten: number
  notes_created: number
  notes_skipped: number
  notes_overwritten: number
}

async function handleExportResponse(res: Response, fallbackFilename: string): Promise<void> {
  if (res.status === 401) {
    window.location.href = '/'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const json = await res.json().catch(() => ({}))
    throw new Error((json as { message?: string }).message ?? 'Export failed')
  }

  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = res.headers.get('Content-Disposition')
    ?.split('filename=')[1]?.replace(/"/g, '')
    ?? fallbackFilename
  a.click()
  URL.revokeObjectURL(url)
}

export async function exportGameBackup(gameId: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/games/${gameId}/backup/export`, {
    credentials: 'include',
  })
  await handleExportResponse(res, `game-backup-${gameId}.json`)
}

export async function exportSessionBackup(sessionId: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/backup/export`, {
    credentials: 'include',
  })
  await handleExportResponse(res, `session-backup-${sessionId}.json`)
}

export async function exportNoteBackup(noteId: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/notes/${noteId}/backup/export`, {
    credentials: 'include',
  })
  await handleExportResponse(res, `note-backup-${noteId}.json`)
}

export async function importGameBackup(
  gameId: string,
  file: File,
  mode: 'merge' | 'overwrite'
): Promise<ImportSummary> {
  const res = await fetch(
    `${BASE_URL}/games/${gameId}/backup/import?mode=${mode}`,
    {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: await file.text(),
    }
  )

  if (res.status === 401) {
    window.location.href = '/'
    throw new Error('Unauthorized')
  }

  const json = await res.json()
  if (!res.ok) {
    throw new Error((json as { message?: string }).message ?? 'Import failed')
  }
  return json.data as ImportSummary
}

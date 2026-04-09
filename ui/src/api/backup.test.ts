import { describe, it, expect, vi, beforeEach } from 'vitest'
import { exportGameBackup, exportSessionBackup, exportNoteBackup, importGameBackup } from './backup'

// backup.ts uses BASE_URL from ./client and raw fetch directly
vi.mock('./client', () => ({
  BASE_URL: 'http://localhost:8080',
}))

beforeEach(() => {
  vi.restoreAllMocks()

  // Mock DOM APIs used by handleExportResponse
  vi.stubGlobal('URL', {
    createObjectURL: vi.fn(() => 'blob:mock-url'),
    revokeObjectURL: vi.fn(),
  })

  const mockAnchor = {
    href: '',
    download: '',
    click: vi.fn(),
  }
  vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as unknown as HTMLElement)
})

describe('exportGameBackup', () => {
  it('should fetch the correct export URL with credentials', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{"data":{}}'], { type: 'application/json' }), {
        status: 200,
        headers: { 'Content-Disposition': 'attachment; filename="game-backup.json"' },
      }),
    )

    await exportGameBackup('game-123')

    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/games/game-123/backup/export',
      expect.objectContaining({ credentials: 'include' }),
    )
  })

  it('should trigger a download by clicking an anchor element', async () => {
    const mockAnchor = { href: '', download: '', click: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as unknown as HTMLElement)

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/json' }), {
        status: 200,
        headers: {},
      }),
    )

    await exportGameBackup('game-123')

    expect(mockAnchor.click).toHaveBeenCalled()
    expect(mockAnchor.href).toBe('blob:mock-url')
  })

  it('should use Content-Disposition filename when provided', async () => {
    const mockAnchor = { href: '', download: '', click: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as unknown as HTMLElement)

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/json' }), {
        status: 200,
        headers: { 'Content-Disposition': 'attachment; filename="my-backup.json"' },
      }),
    )

    await exportGameBackup('game-123')

    expect(mockAnchor.download).toBe('my-backup.json')
  })

  it('should fall back to default filename when Content-Disposition is missing', async () => {
    const mockAnchor = { href: '', download: '', click: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as unknown as HTMLElement)

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/octet-stream' }), {
        status: 200,
        headers: {},
      }),
    )

    await exportGameBackup('game-abc')

    expect(mockAnchor.download).toBe('game-backup-game-abc.json')
  })

  it('should redirect and throw on 401 response', async () => {
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      value: { href: '' },
      writable: true,
      configurable: true,
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(null, { status: 401 }),
    )

    await expect(exportGameBackup('game-123')).rejects.toThrow('Unauthorized')
    expect(window.location.href).toBe('/')

    Object.defineProperty(window, 'location', { value: originalLocation, writable: true, configurable: true })
  })

  it('should throw with error message on non-ok response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'Export failed due to server error' }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(exportGameBackup('game-123')).rejects.toThrow('Export failed due to server error')
  })
})

describe('exportSessionBackup', () => {
  it('should fetch the correct session export URL', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/json' }), {
        status: 200,
        headers: {},
      }),
    )

    await exportSessionBackup('session-456')

    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/sessions/session-456/backup/export',
      expect.objectContaining({ credentials: 'include' }),
    )
  })

  it('should use default filename for session backup', async () => {
    const mockAnchor = { href: '', download: '', click: vi.fn() }
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as unknown as HTMLElement)

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/json' }), {
        status: 200,
        headers: {},
      }),
    )

    await exportSessionBackup('session-xyz')

    expect(mockAnchor.download).toBe('session-backup-session-xyz.json')
  })
})

describe('exportNoteBackup', () => {
  it('should fetch the correct note export URL', async () => {
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(new Blob(['{}'], { type: 'application/json' }), {
        status: 200,
        headers: {},
      }),
    )

    await exportNoteBackup('note-789')

    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/notes/note-789/backup/export',
      expect.objectContaining({ credentials: 'include' }),
    )
  })
})

describe('importGameBackup', () => {
  it('should POST file contents to the import URL with mode query param', async () => {
    const fileContent = JSON.stringify({ sessions: [], notes: [] })
    const file = new File([fileContent], 'backup.json', { type: 'application/json' })

    const importSummary = {
      sessions_created: 2,
      sessions_skipped: 0,
      sessions_overwritten: 0,
      notes_created: 5,
      notes_skipped: 1,
      notes_overwritten: 0,
    }
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: importSummary }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await importGameBackup('game-123', file, 'merge')

    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/games/game-123/backup/import?mode=merge',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: fileContent,
      }),
    )
    expect(result).toEqual(importSummary)
  })

  it('should support overwrite mode', async () => {
    const file = new File(['{}'], 'backup.json', { type: 'application/json' })
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { sessions_created: 0, sessions_skipped: 0, sessions_overwritten: 1, notes_created: 0, notes_skipped: 0, notes_overwritten: 0 } }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await importGameBackup('game-123', file, 'overwrite')

    expect(fetchSpy).toHaveBeenCalledWith(
      expect.stringContaining('mode=overwrite'),
      expect.anything(),
    )
  })

  it('should redirect and throw on 401 response', async () => {
    const originalLocation = window.location
    Object.defineProperty(window, 'location', {
      value: { href: '' },
      writable: true,
      configurable: true,
    })

    const file = new File(['{}'], 'backup.json')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(null, { status: 401 }),
    )

    await expect(importGameBackup('game-123', file, 'merge')).rejects.toThrow('Unauthorized')
    expect(window.location.href).toBe('/')

    Object.defineProperty(window, 'location', { value: originalLocation, writable: true, configurable: true })
  })

  it('should throw with server error message on non-ok response', async () => {
    const file = new File(['{}'], 'backup.json')
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ message: 'Invalid backup format' }), {
        status: 422,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(importGameBackup('game-123', file, 'merge')).rejects.toThrow('Invalid backup format')
  })

  it('should return the ImportSummary data on success', async () => {
    const file = new File(['{}'], 'backup.json')
    const summary = {
      sessions_created: 3,
      sessions_skipped: 1,
      sessions_overwritten: 0,
      notes_created: 10,
      notes_skipped: 2,
      notes_overwritten: 0,
    }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ data: summary }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await importGameBackup('game-abc', file, 'merge')

    expect(result).toEqual(summary)
  })
})

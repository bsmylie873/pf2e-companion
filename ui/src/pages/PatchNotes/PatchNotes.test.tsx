import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import PatchNotes from './PatchNotes'

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

const MOCK_PATCH_NOTES = {
  version: 'abc1234',
  date: '2026-04-29',
  notes: '# What\'s New\n\n## ✨ New Features\n- Added patch notes page\n\n## 🐛 Bug Fixes\n- Fixed a thing',
}

function renderPatchNotes() {
  return render(
    <MemoryRouter initialEntries={['/patch-notes']}>
      <PatchNotes />
    </MemoryRouter>,
  )
}

describe('PatchNotes', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('should show loading state initially', () => {
    // Never-resolving fetch to keep the loading state
    vi.spyOn(global, 'fetch').mockReturnValue(new Promise(() => {}))
    renderPatchNotes()
    expect(screen.getByText(/unrolling the scroll/i)).toBeInTheDocument()
  })

  it('should render patch notes content on successful fetch', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(MOCK_PATCH_NOTES),
    } as Response)

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByText('vabc1234')).toBeInTheDocument()
    })
    expect(screen.getByText('2026-04-29')).toBeInTheDocument()
    expect(screen.getByText('Added patch notes page')).toBeInTheDocument()
    expect(screen.getByText('Fixed a thing')).toBeInTheDocument()
  })

  it('should render the page title and subtitle', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(MOCK_PATCH_NOTES),
    } as Response)

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByText('PF2E Companion')).toBeInTheDocument()
    })
    expect(screen.getByText('Chronicle of Changes')).toBeInTheDocument()
  })

  it('should show error state when fetch fails', async () => {
    vi.spyOn(global, 'fetch').mockRejectedValue(new Error('Network error'))

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/herald's scroll could not be found/i)
    })
  })

  it('should show error state when response is not ok', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({}),
    } as Response)

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/herald's scroll could not be found/i)
    })
  })

  it('should render a back link to the home page', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(MOCK_PATCH_NOTES),
    } as Response)

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByText(/return to the gates/i)).toBeInTheDocument()
    })
    const link = screen.getByText(/return to the gates/i)
    expect(link.closest('a')).toHaveAttribute('href', '/')
  })

  it('should not show loading or error when data loads successfully', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(MOCK_PATCH_NOTES),
    } as Response)

    renderPatchNotes()

    await waitFor(() => {
      expect(screen.getByText('vabc1234')).toBeInTheDocument()
    })
    expect(screen.queryByText(/unrolling the scroll/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})

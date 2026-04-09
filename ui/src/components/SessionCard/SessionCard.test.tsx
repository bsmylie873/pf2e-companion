import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import SessionCard from './SessionCard'

vi.mock('../../api/backup', () => ({
  exportSessionBackup: vi.fn(),
}))

vi.mock('../../utils/contentPreview', () => ({
  extractPreviewText: vi.fn().mockReturnValue('Session preview text'),
}))

const mockSession = {
  id: 'session-1',
  title: 'The Dragon Awakens',
  session_number: 1,
  game_id: 'game-1',
  folder_id: null,
  notes: null,
  version: 1,
  scheduled_at: null,
  runtime_start: null,
  runtime_end: null,
  created_at: '2024-01-15T00:00:00Z',
  updated_at: '2024-01-15T00:00:00Z',
}

describe('SessionCard (list mode)', () => {
  it('should render the session title', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByText('The Dragon Awakens')).toBeInTheDocument()
  })

  it('should render the session number', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByText('Session #1')).toBeInTheDocument()
  })

  it('should show edit button', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Edit session' })).toBeInTheDocument()
  })

  it('should call onEdit when edit button is clicked', () => {
    const onEdit = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={onEdit}
        onDelete={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Edit session' }))
    expect(onEdit).toHaveBeenCalledWith(mockSession)
  })

  it('should show delete button for GM', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Delete session' })).toBeInTheDocument()
  })

  it('should not show delete button for non-GM', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={false}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.queryByRole('button', { name: 'Delete session' })).not.toBeInTheDocument()
  })

  it('should call onDelete when delete button is clicked', () => {
    const onDelete = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={onDelete}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete session' }))
    expect(onDelete).toHaveBeenCalledWith(mockSession)
  })

  it('should call onOpen when card is clicked and onOpen is provided', () => {
    const onOpen = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={onOpen}
      />
    )
    fireEvent.click(screen.getByRole('article'))
    expect(onOpen).toHaveBeenCalledWith(mockSession)
  })

  it('should render scheduled date when provided', () => {
    const sessionWithDate = { ...mockSession, scheduled_at: '2024-03-15T18:00:00Z' }
    render(
      <SessionCard
        session={sessionWithDate}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    const timeEl = document.querySelector('time')
    expect(timeEl).toBeInTheDocument()
  })

  it('should render runtime label when start and end are provided', () => {
    const sessionWithRuntime = {
      ...mockSession,
      runtime_start: '2024-03-15T18:00:00Z',
      runtime_end: '2024-03-15T21:00:00Z',
    }
    render(
      <SessionCard
        session={sessionWithRuntime}
        isGM={true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />
    )
    expect(screen.getByText('3h')).toBeInTheDocument()
  })
})

describe('SessionCard (grid mode)', () => {
  it('should render in grid mode with preview text', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('The Dragon Awakens')).toBeInTheDocument()
    expect(screen.getByText('Session preview text')).toBeInTheDocument()
  })

  it('should call onOpen when grid card body is clicked', () => {
    const onOpen = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={onOpen}
      />
    )
    fireEvent.click(screen.getByText('The Dragon Awakens'))
    expect(onOpen).toHaveBeenCalledWith(mockSession)
  })

  it('should show delete button in grid mode for GM', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByRole('button', { name: 'Delete session' })).toBeInTheDocument()
  })

  it('should not show delete button in grid mode for non-GM', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={false}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.queryByRole('button', { name: 'Delete session' })).not.toBeInTheDocument()
  })

  it('should call onEdit when grid edit button clicked', () => {
    const onEdit = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={onEdit}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Edit session' }))
    expect(onEdit).toHaveBeenCalledWith(mockSession)
  })

  it('should call onDelete when grid delete button clicked', () => {
    const onDelete = vi.fn()
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={onDelete}
        onOpen={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Delete session' }))
    expect(onDelete).toHaveBeenCalledWith(mockSession)
  })

  it('should show session number in grid mode', () => {
    render(
      <SessionCard
        session={mockSession}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.getByText('Session #1')).toBeInTheDocument()
  })

  it('should render non-breaking space for null session_number in grid mode', () => {
    const sessionNoNum = { ...mockSession, session_number: null }
    render(
      <SessionCard
        session={sessionNoNum}
        isGM={true}
        mode="grid"
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onOpen={vi.fn()}
      />
    )
    expect(screen.queryByText(/Session #/)).not.toBeInTheDocument()
  })
})

describe('SessionCard — runtime label edge cases', () => {
  it('should show hours and minutes label (e.g., 1h 30m)', () => {
    const session = {
      ...mockSession,
      runtime_start: '2024-03-15T18:00:00Z',
      runtime_end: '2024-03-15T19:30:00Z',
    }
    render(<SessionCard session={session} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.getByText('1h 30m')).toBeInTheDocument()
  })

  it('should show only minutes label when runtime < 1 hour', () => {
    const session = {
      ...mockSession,
      runtime_start: '2024-03-15T18:00:00Z',
      runtime_end: '2024-03-15T18:45:00Z',
    }
    render(<SessionCard session={session} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.getByText('45m')).toBeInTheDocument()
  })

  it('should not show runtime label when end is before start (diffMs <= 0)', () => {
    const session = {
      ...mockSession,
      runtime_start: '2024-03-15T19:00:00Z',
      runtime_end: '2024-03-15T18:00:00Z', // end before start
    }
    render(<SessionCard session={session} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    // No runtime label should be rendered
    expect(document.querySelector('.session-card-runtime')).not.toBeInTheDocument()
  })

  it('should not show runtime label when only runtime_start is set', () => {
    const session = { ...mockSession, runtime_start: '2024-03-15T18:00:00Z', runtime_end: null }
    render(<SessionCard session={session} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    expect(document.querySelector('.session-card-runtime')).not.toBeInTheDocument()
  })
})

describe('SessionCard — additional list mode coverage', () => {
  it('should not show session number when session_number is null', () => {
    const sessionNoNum = { ...mockSession, session_number: null }
    render(<SessionCard session={sessionNoNum} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.queryByText(/Session #/)).not.toBeInTheDocument()
  })

  it('should show export button in list mode', () => {
    render(<SessionCard session={mockSession} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Export session' })).toBeInTheDocument()
  })

  it('should not show clickable class when no onOpen provided', () => {
    render(<SessionCard session={mockSession} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} />)
    const article = screen.getByRole('article')
    expect(article).not.toHaveClass('session-card--clickable')
  })

  it('should show clickable class when onOpen is provided', () => {
    render(<SessionCard session={mockSession} isGM={true} onEdit={vi.fn()} onDelete={vi.fn()} onOpen={vi.fn()} />)
    const article = screen.getByRole('article')
    expect(article).toHaveClass('session-card--clickable')
  })
})

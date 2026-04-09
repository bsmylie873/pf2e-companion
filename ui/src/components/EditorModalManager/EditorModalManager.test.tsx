import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import EditorModalManager from './EditorModalManager'

// Mock EditorModal to avoid complex dependencies
vi.mock('../EditorModal/EditorModal', () => ({
  default: ({ type, itemId, onClose }: { type: string; itemId: string; onClose: () => void }) => (
    <div data-testid={`editor-modal-${itemId}`} data-type={type}>
      <button onClick={onClose} aria-label={`Close ${itemId}`}>Close</button>
    </div>
  ),
}))

const mockItems = [
  { type: 'session' as const, itemId: 'session-1', label: 'Session One' },
  { type: 'note' as const, itemId: 'note-1', label: 'Note One' },
]

describe('EditorModalManager', () => {
  it('should return null when items list is empty', () => {
    const { container } = render(
      <EditorModalManager
        items={[]}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(container.firstChild).toBeNull()
  })

  it('should render single item without tab strip', () => {
    render(
      <EditorModalManager
        items={[mockItems[0]]}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(screen.queryByRole('tablist')).not.toBeInTheDocument()
    expect(screen.getByTestId('editor-modal-session-1')).toBeInTheDocument()
  })

  it('should render tab strip when multiple items', () => {
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(screen.getByRole('tablist')).toBeInTheDocument()
  })

  it('should render tabs for each item', () => {
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(screen.getByText('Session One')).toBeInTheDocument()
    expect(screen.getByText('Note One')).toBeInTheDocument()
  })

  it('should switch active tab when tab is clicked', () => {
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    fireEvent.click(screen.getByText('Note One'))
    const noteTab = screen.getByRole('tab', { name: /Note One/ })
    expect(noteTab).toHaveAttribute('aria-selected', 'true')
  })

  it('should call onClose when a tab close button is clicked', () => {
    const onClose = vi.fn()
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={onClose}
        onCloseAll={vi.fn()}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Close Session One' }))
    expect(onClose).toHaveBeenCalledWith('session-1')
  })

  it('should call onCloseAll when "Close all" is clicked', () => {
    const onCloseAll = vi.fn()
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={onCloseAll}
      />
    )
    fireEvent.click(screen.getByRole('button', { name: 'Close all' }))
    expect(onCloseAll).toHaveBeenCalledTimes(1)
  })

  it('should show session icon for session type tab', () => {
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(screen.getByText('⚔')).toBeInTheDocument()
  })

  it('should show note icon for note type tab', () => {
    render(
      <EditorModalManager
        items={mockItems}
        gameId="game-1"
        onClose={vi.fn()}
        onCloseAll={vi.fn()}
      />
    )
    expect(screen.getByText('📜')).toBeInTheDocument()
  })
})

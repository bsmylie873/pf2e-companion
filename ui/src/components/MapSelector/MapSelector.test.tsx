import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import MapSelector from './MapSelector'

const mockMaps = [
  { id: 'map-1', name: 'World Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
  { id: 'map-2', name: 'Dungeon Level 1', game_id: 'game-1', archived_at: null, image_url: null, position: 1, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
]

const defaultProps = {
  maps: mockMaps,
  activeMapId: 'map-1',
  onSelect: vi.fn(),
  isGM: true,
  onCreateMap: vi.fn(),
  onRenameMap: vi.fn(),
  onArchiveMap: vi.fn(),
  onUnarchiveMap: vi.fn(),
  onReorderMaps: vi.fn(),
  archivedMaps: [],
}

describe('MapSelector', () => {
  it('should render all map tabs', () => {
    render(<MapSelector {...defaultProps} />)
    expect(screen.getByRole('button', { name: 'World Map' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Dungeon Level 1' })).toBeInTheDocument()
  })

  it('should call onSelect when a map tab is clicked', () => {
    const onSelect = vi.fn()
    render(<MapSelector {...defaultProps} onSelect={onSelect} />)
    fireEvent.click(screen.getByRole('button', { name: 'Dungeon Level 1' }))
    expect(onSelect).toHaveBeenCalledWith('map-2')
  })

  it('should show GM action buttons when isGM is true', () => {
    render(<MapSelector {...defaultProps} />)
    expect(screen.getAllByRole('button', { name: 'Rename map' })).toHaveLength(2)
    expect(screen.getAllByRole('button', { name: 'Archive map' })).toHaveLength(2)
  })

  it('should not show GM action buttons when isGM is false', () => {
    render(<MapSelector {...defaultProps} isGM={false} />)
    expect(screen.queryByRole('button', { name: 'Rename map' })).not.toBeInTheDocument()
  })

  it('should show "New Map" button for GMs', () => {
    render(<MapSelector {...defaultProps} />)
    expect(screen.getByRole('button', { name: /New Map/ })).toBeInTheDocument()
  })

  it('should not show "New Map" button for non-GMs', () => {
    render(<MapSelector {...defaultProps} isGM={false} />)
    expect(screen.queryByRole('button', { name: /New Map/ })).not.toBeInTheDocument()
  })

  it('should show new map input when "New Map" button is clicked', () => {
    render(<MapSelector {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /New Map/ }))
    expect(screen.getByPlaceholderText('Map name…')).toBeInTheDocument()
  })

  it('should call onCreateMap when submitting new map name', () => {
    const onCreateMap = vi.fn()
    render(<MapSelector {...defaultProps} onCreateMap={onCreateMap} />)
    fireEvent.click(screen.getByRole('button', { name: /New Map/ }))
    const input = screen.getByPlaceholderText('Map name…')
    fireEvent.change(input, { target: { value: 'New Map Name' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCreateMap).toHaveBeenCalledWith('New Map Name')
  })

  it('should call onArchiveMap when archive button is clicked', () => {
    const onArchiveMap = vi.fn()
    render(<MapSelector {...defaultProps} onArchiveMap={onArchiveMap} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Archive map' })[0])
    expect(onArchiveMap).toHaveBeenCalledWith('map-1')
  })

  it('should show archived maps section when there are archived maps', () => {
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: '2024-01-01T00:00:00Z', image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    expect(screen.getByText(/Show Archived Maps \(1\)/)).toBeInTheDocument()
  })

  it('should reveal archived maps when toggle is clicked', () => {
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: '2024-01-01T00:00:00Z', image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText('Old Map')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Restore' })).toBeInTheDocument()
  })

  it('should call onUnarchiveMap when Restore is clicked', () => {
    const onUnarchiveMap = vi.fn()
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: '2024-01-01T00:00:00Z', image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} onUnarchiveMap={onUnarchiveMap} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    fireEvent.click(screen.getByRole('button', { name: 'Restore' }))
    expect(onUnarchiveMap).toHaveBeenCalledWith('map-arch')
  })

  // ── Rename flow ─────────────────────────────────────────────────
  it('should show rename input when Rename map is clicked', () => {
    render(<MapSelector {...defaultProps} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Rename map' })[0])
    expect(screen.getByDisplayValue('World Map')).toBeInTheDocument()
  })

  it('should call onRenameMap and hide input when Enter pressed', () => {
    const onRenameMap = vi.fn()
    render(<MapSelector {...defaultProps} onRenameMap={onRenameMap} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Rename map' })[0])
    const input = screen.getByDisplayValue('World Map')
    fireEvent.change(input, { target: { value: 'New World Map' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onRenameMap).toHaveBeenCalledWith('map-1', 'New World Map')
    expect(screen.queryByDisplayValue('New World Map')).not.toBeInTheDocument()
  })

  it('should call onRenameMap when rename input loses focus', () => {
    const onRenameMap = vi.fn()
    render(<MapSelector {...defaultProps} onRenameMap={onRenameMap} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Rename map' })[0])
    const input = screen.getByDisplayValue('World Map')
    fireEvent.change(input, { target: { value: 'Renamed Map' } })
    fireEvent.blur(input)
    expect(onRenameMap).toHaveBeenCalledWith('map-1', 'Renamed Map')
  })

  it('should cancel rename and NOT call onRenameMap when Escape is pressed', () => {
    const onRenameMap = vi.fn()
    render(<MapSelector {...defaultProps} onRenameMap={onRenameMap} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Rename map' })[0])
    const input = screen.getByDisplayValue('World Map')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(onRenameMap).not.toHaveBeenCalled()
    expect(screen.queryByDisplayValue('World Map')).not.toBeInTheDocument()
    // Original tab buttons should be back
    expect(screen.getByRole('button', { name: 'World Map' })).toBeInTheDocument()
  })

  it('should NOT call onRenameMap when rename submitted with empty value', () => {
    const onRenameMap = vi.fn()
    render(<MapSelector {...defaultProps} onRenameMap={onRenameMap} />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Rename map' })[0])
    const input = screen.getByDisplayValue('World Map')
    fireEvent.change(input, { target: { value: '' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onRenameMap).not.toHaveBeenCalled()
  })

  // ── New Map cancel flows ─────────────────────────────────────────
  it('should cancel new map creation when Escape is pressed', () => {
    render(<MapSelector {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /New Map/ }))
    const input = screen.getByPlaceholderText('Map name…')
    fireEvent.keyDown(input, { key: 'Escape' })
    expect(screen.queryByPlaceholderText('Map name…')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /New Map/ })).toBeInTheDocument()
  })

  it('should cancel new map creation on blur if name is empty', () => {
    render(<MapSelector {...defaultProps} />)
    fireEvent.click(screen.getByRole('button', { name: /New Map/ }))
    const input = screen.getByPlaceholderText('Map name…')
    // Blur without typing anything — should close
    fireEvent.blur(input)
    expect(screen.queryByPlaceholderText('Map name…')).not.toBeInTheDocument()
  })

  it('should NOT call onCreateMap when new map name is only whitespace', () => {
    const onCreateMap = vi.fn()
    render(<MapSelector {...defaultProps} onCreateMap={onCreateMap} />)
    fireEvent.click(screen.getByRole('button', { name: /New Map/ }))
    const input = screen.getByPlaceholderText('Map name…')
    fireEvent.change(input, { target: { value: '   ' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    expect(onCreateMap).not.toHaveBeenCalled()
  })

  // ── Archived section toggle ─────────────────────────────────────
  it('should hide archived maps list when toggle clicked twice', () => {
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: '2024-01-01T00:00:00Z', image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    // Show
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText('Old Map')).toBeInTheDocument()
    // Hide
    fireEvent.click(screen.getByText(/Hide Archived Maps/))
    expect(screen.queryByText('Old Map')).not.toBeInTheDocument()
  })

  it('should not show archived section when isGM is false', () => {
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: '2024-01-01T00:00:00Z', image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} isGM={false} archivedMaps={archivedMaps} />)
    expect(screen.queryByText(/Archived Maps/)).not.toBeInTheDocument()
  })

  // ── formatElapsed edge cases ────────────────────────────────────
  it('should show "Archived" when archived_at is null', () => {
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText('Archived')).toBeInTheDocument()
  })

  it('should show minutes ago for recent archives', () => {
    const recentDate = new Date(Date.now() - 30 * 60 * 1000).toISOString() // 30 min ago
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: recentDate, image_url: null, position: 0, created_at: '', updated_at: '' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText(/\d+m ago/)).toBeInTheDocument()
  })

  it('should show hours ago for archives older than 1 hour', () => {
    const hoursAgo = new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString() // 3 hours ago
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: hoursAgo, image_url: null, position: 0, created_at: '', updated_at: '' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText('3h ago')).toBeInTheDocument()
  })

  it('should show days ago for archives older than 1 day', () => {
    const daysAgo = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString() // 3 days ago
    const archivedMaps = [
      { id: 'map-arch', name: 'Old Map', game_id: 'game-1', archived_at: daysAgo, image_url: null, position: 0, created_at: '', updated_at: '' },
    ]
    render(<MapSelector {...defaultProps} archivedMaps={archivedMaps} />)
    fireEvent.click(screen.getByText(/Show Archived Maps/))
    expect(screen.getByText('3d ago')).toBeInTheDocument()
  })
})

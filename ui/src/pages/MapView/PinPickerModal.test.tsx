import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PinPickerModal from './PinPickerModal'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'

vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom')
  return {
    ...(actual as object),
    createPortal: (node: React.ReactNode) => node,
  }
})

vi.mock('../../constants/pins', () => ({
  PIN_COLOURS: ['red', 'blue'],
  PIN_ICONS: ['star', 'circle'],
  COLOUR_MAP: { red: '#ff0000', blue: '#0000ff' },
  PIN_ICON_COMPONENTS: {
    star: ({ size }: { size: number }) => <span data-testid="icon-star" style={{ fontSize: size }}>★</span>,
    circle: ({ size }: { size: number }) => <span data-testid="icon-circle" style={{ fontSize: size }}>●</span>,
  },
  PIN_ICON_LABELS: { star: 'Star', circle: 'Circle' },
}))

const mockSession: Session = {
  id: 'sess-1',
  game_id: 'game-1',
  title: 'The Dragon Encounter',
  session_number: 5,
  scheduled_at: null,
  runtime_start: null,
  runtime_end: null,
  folder_id: null,
  notes: null,
  version: 1,
  foundry_data: null,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

const mockNote: Note = {
  id: 'note-1',
  game_id: 'game-1',
  user_id: 'user-1',
  session_id: null,
  folder_id: null,
  title: 'The Ancient Library',
  content: null,
  visibility: 'visible',
  version: 1,
  foundry_data: null,
  created_at: '2024-01-01',
  updated_at: '2024-01-01',
}

const baseProps = {
  pendingColour: 'red' as const,
  pendingIcon: 'star' as const,
  pendingLabel: '',
  pendingDescription: '',
  pickerSearch: '',
  unpinnedSessions: [] as Session[],
  notes: [] as Note[],
  onClose: vi.fn(),
  onColourChange: vi.fn(),
  onIconChange: vi.fn(),
  onLabelChange: vi.fn(),
  onDescriptionChange: vi.fn(),
  onSearchChange: vi.fn(),
  onCreateMarker: vi.fn(),
  onSelectSession: vi.fn(),
  onSelectNote: vi.fn(),
}

describe('PinPickerModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders modal with title', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByText('Mark This Location')).toBeInTheDocument()
  })

  it('renders colour palette buttons for each colour', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByTitle('red')).toBeInTheDocument()
    expect(screen.getByTitle('blue')).toBeInTheDocument()
  })

  it('calls onColourChange when colour button is clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    fireEvent.click(screen.getByTitle('red'))
    expect(baseProps.onColourChange).toHaveBeenCalledWith('red')
  })

  it('calls onColourChange with blue when blue swatch clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    fireEvent.click(screen.getByTitle('blue'))
    expect(baseProps.onColourChange).toHaveBeenCalledWith('blue')
  })

  it('renders icon buttons with aria-labels', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByRole('button', { name: 'Star' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Circle' })).toBeInTheDocument()
  })

  it('calls onIconChange when icon button is clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    fireEvent.click(screen.getByRole('button', { name: 'Star' }))
    expect(baseProps.onIconChange).toHaveBeenCalledWith('star')
  })

  it('calls onIconChange with circle when circle icon clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    fireEvent.click(screen.getByRole('button', { name: 'Circle' }))
    expect(baseProps.onIconChange).toHaveBeenCalledWith('circle')
  })

  it('renders label input', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByPlaceholderText('Pin label…')).toBeInTheDocument()
  })

  it('renders description textarea', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByPlaceholderText('Description…')).toBeInTheDocument()
  })

  it('calls onLabelChange when label input changes', async () => {
    const user = userEvent.setup()
    render(<PinPickerModal {...baseProps} />)
    await user.type(screen.getByPlaceholderText('Pin label…'), 'A')
    expect(baseProps.onLabelChange).toHaveBeenCalledWith('A')
  })

  it('calls onDescriptionChange when description textarea changes', async () => {
    const user = userEvent.setup()
    render(<PinPickerModal {...baseProps} />)
    await user.type(screen.getByPlaceholderText('Description…'), 'D')
    expect(baseProps.onDescriptionChange).toHaveBeenCalledWith('D')
  })

  it('renders create marker button', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.getByText('Place as Standalone Marker')).toBeInTheDocument()
  })

  it('calls onCreateMarker with label and description when create marker clicked', () => {
    render(<PinPickerModal {...baseProps} pendingLabel="My Pin" pendingDescription="Desc" />)
    fireEvent.click(screen.getByText('Place as Standalone Marker'))
    expect(baseProps.onCreateMarker).toHaveBeenCalledWith('My Pin', 'Desc')
  })

  it('calls onClose when overlay background is clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    const overlay = document.querySelector('.map-overlay')!
    fireEvent.click(overlay)
    expect(baseProps.onClose).toHaveBeenCalledTimes(1)
  })

  it('does not call onClose when modal body is clicked', () => {
    render(<PinPickerModal {...baseProps} />)
    const body = document.querySelector('.map-session-picker')!
    fireEvent.click(body)
    expect(baseProps.onClose).not.toHaveBeenCalled()
  })

  it('does not show search section when no sessions or notes', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[]} notes={[]} />)
    expect(screen.queryByPlaceholderText(/Search sessions/)).not.toBeInTheDocument()
  })

  it('shows search section when there are unpinned sessions', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} />)
    expect(screen.getByPlaceholderText(/Search sessions/)).toBeInTheDocument()
  })

  it('shows search section when there are notes', () => {
    render(<PinPickerModal {...baseProps} notes={[mockNote]} />)
    expect(screen.getByPlaceholderText(/Search sessions/)).toBeInTheDocument()
  })

  it('shows matching sessions when search query matches', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} pickerSearch="Dragon" />)
    expect(screen.getByText('The Dragon Encounter')).toBeInTheDocument()
  })

  it('shows session number in search results', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} pickerSearch="Dragon" />)
    expect(screen.getByText('#5')).toBeInTheDocument()
  })

  it('calls onSelectSession when session result is clicked', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} pickerSearch="Dragon" />)
    const sessionBtn = screen.getAllByRole('button').find(b => b.textContent?.includes('The Dragon Encounter'))
    fireEvent.click(sessionBtn!)
    expect(baseProps.onSelectSession).toHaveBeenCalledWith(mockSession)
  })

  it('shows matching notes when search query matches', () => {
    render(<PinPickerModal {...baseProps} notes={[mockNote]} pickerSearch="Library" />)
    expect(screen.getByText('The Ancient Library')).toBeInTheDocument()
  })

  it('calls onSelectNote when note result is clicked', () => {
    render(<PinPickerModal {...baseProps} notes={[mockNote]} pickerSearch="Library" />)
    const noteBtn = screen.getAllByRole('button').find(b => b.textContent?.includes('The Ancient Library'))
    fireEvent.click(noteBtn!)
    expect(baseProps.onSelectNote).toHaveBeenCalledWith(mockNote)
  })

  it('shows no match message when search has no results', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} pickerSearch="xyzzy" />)
    expect(screen.getByText('No matching sessions or notes.')).toBeInTheDocument()
  })

  it('does not show results when search is empty', () => {
    render(<PinPickerModal {...baseProps} unpinnedSessions={[mockSession]} pickerSearch="" />)
    expect(screen.queryByText('The Dragon Encounter')).not.toBeInTheDocument()
  })

  it('shows dropLinkedItem section when provided', () => {
    render(
      <PinPickerModal
        {...baseProps}
        dropLinkedItem={{ type: 'session', id: 'sess-1', label: 'My Session' }}
      />
    )
    expect(screen.getByText(/My Session/)).toBeInTheDocument()
    expect(screen.getByText(/Place Pin & Link Session/)).toBeInTheDocument()
  })

  it('shows note linking when dropLinkedItem is a note', () => {
    render(
      <PinPickerModal
        {...baseProps}
        dropLinkedItem={{ type: 'note', id: 'note-1', label: 'My Note' }}
      />
    )
    expect(screen.getByText(/My Note/)).toBeInTheDocument()
    expect(screen.getByText(/Place Pin & Link Note/)).toBeInTheDocument()
  })

  it('calls onSelectSession when dropLinkedItem session link button clicked', () => {
    render(
      <PinPickerModal
        {...baseProps}
        unpinnedSessions={[mockSession]}
        dropLinkedItem={{ type: 'session', id: 'sess-1', label: 'The Dragon Encounter' }}
      />
    )
    fireEvent.click(screen.getByText(/Place Pin & Link Session/))
    expect(baseProps.onSelectSession).toHaveBeenCalledWith(mockSession)
  })

  it('calls onSelectNote when dropLinkedItem note link button clicked', () => {
    render(
      <PinPickerModal
        {...baseProps}
        notes={[mockNote]}
        dropLinkedItem={{ type: 'note', id: 'note-1', label: 'The Ancient Library' }}
      />
    )
    fireEvent.click(screen.getByText(/Place Pin & Link Note/))
    expect(baseProps.onSelectNote).toHaveBeenCalledWith(mockNote)
  })

  it('does not show dropLinkedItem section when not provided', () => {
    render(<PinPickerModal {...baseProps} />)
    expect(screen.queryByText(/Place Pin & Link/)).not.toBeInTheDocument()
  })

  it('shows correct visibility icon for visible notes', () => {
    render(<PinPickerModal {...baseProps} notes={[mockNote]} pickerSearch="Library" />)
    expect(screen.getByText('👁')).toBeInTheDocument()
  })

  it('shows correct visibility icon for private notes', () => {
    const privateNote = { ...mockNote, visibility: 'private' as const }
    render(<PinPickerModal {...baseProps} notes={[privateNote]} pickerSearch="Library" />)
    expect(screen.getByText('🔒')).toBeInTheDocument()
  })
})

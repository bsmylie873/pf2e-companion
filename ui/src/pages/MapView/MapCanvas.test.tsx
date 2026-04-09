import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import MapCanvas from './MapCanvas'
import type { SessionPin, PinGroup } from '../../types/pin'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'

vi.mock('react-zoom-pan-pinch', () => ({
  TransformWrapper: ({ children }: { children: React.ReactNode | ((...args: unknown[]) => React.ReactNode) }) => (
    <div data-testid="transform-wrapper">
      {typeof children === 'function'
        ? children({ zoomIn: vi.fn(), zoomOut: vi.fn(), resetTransform: vi.fn() })
        : children}
    </div>
  ),
  TransformComponent: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="transform-component">{children}</div>
  ),
}))

vi.mock('../../constants/pins', () => ({
  PIN_COLOURS: ['red', 'blue'],
  PIN_ICONS: ['star', 'circle'],
  COLOUR_MAP: {
    red: '#ff0000',
    blue: '#0000ff',
    grey: '#888888',
  },
  PIN_ICON_COMPONENTS: {
    'position-marker': ({ size }: { size: number }) => <span data-testid="icon-position-marker" style={{ fontSize: size }}>📍</span>,
    star: ({ size }: { size: number }) => <span data-testid="icon-star" style={{ fontSize: size }}>★</span>,
    circle: ({ size }: { size: number }) => <span data-testid="icon-circle" style={{ fontSize: size }}>●</span>,
  },
  PIN_ICON_LABELS: { star: 'Star', circle: 'Circle', 'position-marker': 'Position Marker' },
}))

const makePin = (id: string, overrides: Partial<SessionPin> = {}): SessionPin => ({
  id,
  game_id: 'game-1',
  session_id: null,
  note_id: null,
  group_id: null,
  map_id: 'map-1',
  label: `Pin ${id}`,
  x: 10,
  y: 20,
  colour: 'red',
  icon: 'star',
  description: null,
  created_at: '',
  updated_at: '',
  ...overrides,
})

const makeGroup = (id: string, overrides: Partial<PinGroup> = {}): PinGroup => ({
  id,
  game_id: 'game-1',
  map_id: 'map-1',
  x: 50,
  y: 50,
  colour: 'red',
  icon: 'star',
  pin_count: 2,
  pins: [makePin('gp-1'), makePin('gp-2')],
  created_at: '',
  updated_at: '',
  ...overrides,
})

const mockSession: Session = {
  id: 'sess-1',
  game_id: 'game-1',
  title: 'Session One',
  session_number: 1,
  scheduled_at: null,
  runtime_start: null,
  runtime_end: null,
  folder_id: null,
  notes: null,
  version: 1,
  foundry_data: null,
  created_at: '',
  updated_at: '',
}

const mockNote: Note = {
  id: 'note-1',
  game_id: 'game-1',
  user_id: 'user-1',
  session_id: null,
  folder_id: null,
  title: 'Note One',
  content: null,
  visibility: 'visible',
  version: 1,
  foundry_data: null,
  created_at: '',
  updated_at: '',
}

const baseHandlers = {
  onTransformed: vi.fn(),
  onTransformEnd: vi.fn(),
  onZoom: vi.fn(),
  onImageLoad: vi.fn(),
  onMapClick: vi.fn(),
  onPointerMove: vi.fn(),
  onPointerUp: vi.fn(),
  onPinPointerDown: vi.fn(),
  onHoverPin: vi.fn(),
  onEditPin: vi.fn(),
  onDeletePin: vi.fn(),
  onEditPinField: vi.fn(),
  onEditLinkSearchChange: vi.fn(),
  onPinClick: vi.fn(),
  onGroupClick: vi.fn(),
  onManageGroup: vi.fn(),
  onPinErrorDismiss: vi.fn(),
  openItem: vi.fn(),
  sessionForPin: vi.fn().mockReturnValue(undefined),
  noteForPin: vi.fn().mockReturnValue(undefined),
}

const baseProps = {
  activeMapId: 'map-1',
  imageUrl: '/test-map.jpg',
  viewState: { scale: 1, positionX: 0, positionY: 0 },
  displayScale: 1,
  pins: [] as SessionPin[],
  pinGroups: [] as PinGroup[],
  sessions: [] as Session[],
  notes: [] as Note[],
  hoveredPinId: null,
  flashPinId: null,
  dragging: null,
  editingPinId: null,
  editLinkSearch: '',
  dropTargetIds: new Set<string>(),
  activeGroupId: null,
  pinError: null,
  mapContainerRef: React.createRef() as React.RefObject<HTMLDivElement | null>,
  viewportContainerRef: React.createRef() as React.RefObject<HTMLDivElement | null>,
  transformRef: React.createRef() as React.RefObject<null>,
  wasDragRef: { current: false },
  isGM: false,
  ...baseHandlers,
}

describe('MapCanvas', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    baseHandlers.sessionForPin.mockReturnValue(undefined)
    baseHandlers.noteForPin.mockReturnValue(undefined)
  })

  it('renders TransformWrapper', () => {
    render(<MapCanvas {...baseProps} />)
    expect(screen.getByTestId('transform-wrapper')).toBeInTheDocument()
  })

  it('renders TransformComponent', () => {
    render(<MapCanvas {...baseProps} />)
    expect(screen.getByTestId('transform-component')).toBeInTheDocument()
  })

  it('renders map image with correct src', () => {
    render(<MapCanvas {...baseProps} />)
    const img = screen.getByAltText('Campaign map')
    expect(img).toBeInTheDocument()
    expect(img).toHaveAttribute('src', '/test-map.jpg')
  })

  it('calls onImageLoad when image loads', () => {
    render(<MapCanvas {...baseProps} />)
    fireEvent.load(screen.getByAltText('Campaign map'))
    expect(baseHandlers.onImageLoad).toHaveBeenCalled()
  })

  it('does not render pin error banner when pinError is null', () => {
    render(<MapCanvas {...baseProps} pinError={null} />)
    expect(screen.queryByText(/error/i)).not.toBeInTheDocument()
  })

  it('shows pin error banner when pinError is set', () => {
    render(<MapCanvas {...baseProps} pinError="Failed to place pin" />)
    expect(screen.getByText('Failed to place pin')).toBeInTheDocument()
  })

  it('calls onPinErrorDismiss when error banner is clicked', () => {
    render(<MapCanvas {...baseProps} pinError="Failed to place pin" />)
    fireEvent.click(screen.getByText('Failed to place pin'))
    expect(baseHandlers.onPinErrorDismiss).toHaveBeenCalled()
  })

  it('renders no pins when pins array is empty', () => {
    render(<MapCanvas {...baseProps} pins={[]} />)
    expect(screen.queryByTitle(/Pin/)).not.toBeInTheDocument()
  })

  it('renders a pin with its label', () => {
    const pin = makePin('p1', { label: 'Dragon Cave' })
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    expect(screen.getByText('Dragon Cave')).toBeInTheDocument()
  })

  it('calls onHoverPin when mouse enters pin wrapper', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    const pinWrapper = document.querySelector('.map-pin-wrapper')!
    fireEvent.mouseEnter(pinWrapper)
    expect(baseHandlers.onHoverPin).toHaveBeenCalledWith('p1')
  })

  it('calls onHoverPin with null when mouse leaves pin wrapper', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    const pinWrapper = document.querySelector('.map-pin-wrapper')!
    fireEvent.mouseLeave(pinWrapper)
    expect(baseHandlers.onHoverPin).toHaveBeenCalledWith(null)
  })

  it('calls onDeletePin when delete button is clicked', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    fireEvent.click(screen.getByTitle('Remove pin'))
    expect(baseHandlers.onDeletePin).toHaveBeenCalledWith('p1')
  })

  it('calls onEditPin when edit button is clicked', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    fireEvent.click(screen.getByTitle('Edit pin'))
    expect(baseHandlers.onEditPin).toHaveBeenCalledWith('p1')
  })

  it('calls onEditPin to close when already editing and edit button clicked', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" />)
    fireEvent.click(screen.getByTitle('Edit pin'))
    expect(baseHandlers.onEditPin).toHaveBeenCalledWith(null)
  })

  it('shows edit popover when editingPinId matches pin', () => {
    const pin = makePin('p1', { label: 'PopoverPin' })
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" />)
    expect(screen.getByPlaceholderText('Pin label…')).toBeInTheDocument()
  })

  it('does not show edit popover when editingPinId does not match', () => {
    const pin = makePin('p1', { label: 'PopoverPin' })
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="other-id" />)
    expect(screen.queryByPlaceholderText('Pin label…')).not.toBeInTheDocument()
  })

  it('shows session title as pin label when pin has session_id', () => {
    const pin = makePin('p1', { session_id: 'sess-1', label: 'ignored' })
    baseHandlers.sessionForPin.mockReturnValue(mockSession)
    render(<MapCanvas {...baseProps} pins={[pin]} sessions={[mockSession]} />)
    expect(screen.getByText('Session One')).toBeInTheDocument()
  })

  it('shows note title as pin label when pin has note_id', () => {
    const pin = makePin('p1', { note_id: 'note-1', label: 'ignored' })
    baseHandlers.noteForPin.mockReturnValue(mockNote)
    render(<MapCanvas {...baseProps} pins={[pin]} notes={[mockNote]} />)
    expect(screen.getByText('Note One')).toBeInTheDocument()
  })

  it('adds hovered class when hoveredPinId matches pin', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} hoveredPinId="p1" />)
    const wrapper = document.querySelector('.map-pin-wrapper')!
    expect(wrapper.classList.contains('map-pin-wrapper--hovered')).toBe(true)
  })

  it('adds flash class when flashPinId matches pin', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} flashPinId="p1" />)
    const wrapper = document.querySelector('.map-pin-wrapper')!
    expect(wrapper.classList.contains('map-pin-wrapper--flash')).toBe(true)
  })

  it('adds drop-target class when pin id is in dropTargetIds', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} dropTargetIds={new Set(['p1'])} />)
    const wrapper = document.querySelector('.map-pin-wrapper')!
    expect(wrapper.classList.contains('map-pin-wrapper--drop-target')).toBe(true)
  })

  it('does not render pins with group_id set (grouped pins)', () => {
    const groupedPin = makePin('p1', { group_id: 'group-1', label: 'GroupedPin' })
    render(<MapCanvas {...baseProps} pins={[groupedPin]} />)
    expect(screen.queryByText('GroupedPin')).not.toBeInTheDocument()
  })

  it('renders pin groups', () => {
    const group = makeGroup('g1')
    render(<MapCanvas {...baseProps} pinGroups={[group]} />)
    expect(screen.getByTitle('Group (2 pins)')).toBeInTheDocument()
  })

  it('shows group badge with pin count', () => {
    const group = makeGroup('g1', { pin_count: 5 })
    render(<MapCanvas {...baseProps} pinGroups={[group]} />)
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('calls onGroupClick when group pin is clicked', () => {
    const group = makeGroup('g1')
    render(<MapCanvas {...baseProps} pinGroups={[group]} />)
    fireEvent.click(screen.getByTitle('Group (2 pins)'))
    expect(baseHandlers.onGroupClick).toHaveBeenCalledWith('g1')
  })

  it('calls openItem with note type when pin with note_id is clicked', () => {
    const pin = makePin('p1', { note_id: 'note-1' })
    baseHandlers.noteForPin.mockReturnValue(mockNote)
    render(<MapCanvas {...baseProps} pins={[pin]} notes={[mockNote]} />)
    const pinButton = document.querySelector('.map-pin') as HTMLElement
    fireEvent.click(pinButton)
    expect(baseHandlers.openItem).toHaveBeenCalledWith('note', 'note-1', 'Note One')
  })

  it('calls openItem with session type when pin with session_id is clicked', () => {
    const pin = makePin('p1', { session_id: 'sess-1' })
    baseHandlers.sessionForPin.mockReturnValue(mockSession)
    render(<MapCanvas {...baseProps} pins={[pin]} sessions={[mockSession]} />)
    const pinButton = document.querySelector('.map-pin') as HTMLElement
    fireEvent.click(pinButton)
    expect(baseHandlers.openItem).toHaveBeenCalledWith('session', 'sess-1', 'Session One')
  })

  it('shows link search input in edit popover for unlinked pin', () => {
    const pin = makePin('p1', { session_id: null, note_id: null })
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" />)
    expect(screen.getByPlaceholderText('Search to link…')).toBeInTheDocument()
  })

  it('shows session link search results when editLinkSearch matches', () => {
    const pin = makePin('p1', { session_id: null, note_id: null })
    render(
      <MapCanvas
        {...baseProps}
        pins={[pin]}
        editingPinId="p1"
        sessions={[mockSession]}
        editLinkSearch="Session"
      />
    )
    expect(screen.getByText('Session One')).toBeInTheDocument()
  })

  it('shows no matches when editLinkSearch has no results', () => {
    const pin = makePin('p1', { session_id: null, note_id: null })
    render(
      <MapCanvas
        {...baseProps}
        pins={[pin]}
        editingPinId="p1"
        sessions={[mockSession]}
        editLinkSearch="xyzzy"
      />
    )
    expect(screen.getByText('No matches')).toBeInTheDocument()
  })

  it('shows unlink button when pin has session_id and editing', () => {
    const pin = makePin('p1', { session_id: 'sess-1' })
    baseHandlers.sessionForPin.mockReturnValue(mockSession)
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" sessions={[mockSession]} />)
    expect(screen.getByText('Unlink')).toBeInTheDocument()
  })

  it('shows unlink button when pin has note_id and editing', () => {
    const pin = makePin('p1', { note_id: 'note-1' })
    baseHandlers.noteForPin.mockReturnValue(mockNote)
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" notes={[mockNote]} />)
    expect(screen.getByText('Unlink')).toBeInTheDocument()
  })

  it('calls onEditPinField with null session_id when Unlink session clicked', () => {
    const pin = makePin('p1', { session_id: 'sess-1' })
    baseHandlers.sessionForPin.mockReturnValue(mockSession)
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" sessions={[mockSession]} />)
    fireEvent.click(screen.getByText('Unlink'))
    expect(baseHandlers.onEditPinField).toHaveBeenCalledWith('p1', { session_id: null })
  })

  it('renders colour swatches in edit popover', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} editingPinId="p1" />)
    // Both red and blue colour swatches should be visible in the popover
    const redSwatches = screen.getAllByTitle('red')
    expect(redSwatches.length).toBeGreaterThan(0)
  })

  it('calls onPointerDown handler when pointer down on pin', () => {
    const pin = makePin('p1')
    render(<MapCanvas {...baseProps} pins={[pin]} />)
    const pinButton = document.querySelector('.map-pin') as HTMLElement
    fireEvent.pointerDown(pinButton)
    expect(baseHandlers.onPinPointerDown).toHaveBeenCalled()
  })
})

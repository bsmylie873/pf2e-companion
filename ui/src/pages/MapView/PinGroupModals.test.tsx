import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import PinGroupModals from './PinGroupModals'
import type { SessionPin, PinGroup } from '../../types/pin'
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

vi.mock('../../api/pinGroups', () => ({
  createMapPinGroup: vi.fn().mockResolvedValue({ id: 'group-new' }),
  addPinToGroup: vi.fn().mockResolvedValue({}),
  removePinFromGroup: vi.fn().mockResolvedValue({}),
  disbandPinGroup: vi.fn().mockResolvedValue({}),
  updatePinGroup: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/pins', () => ({
  listMapPins: vi.fn().mockResolvedValue([]),
}))

const makePin = (id: string, overrides: Partial<SessionPin> = {}): SessionPin => ({
  id,
  game_id: 'game-1',
  session_id: null,
  note_id: null,
  group_id: null,
  map_id: 'map-1',
  label: `Pin ${id}`,
  x: 50,
  y: 50,
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
  title: 'Dragon Encounter',
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
  title: 'Ancient Library',
  content: null,
  visibility: 'visible',
  version: 1,
  foundry_data: null,
  created_at: '',
  updated_at: '',
}

const noopProps = {
  gameId: 'game-1',
  activeMapId: 'map-1',
  groupingPrompt: null,
  onDismissGroupingPrompt: vi.fn(),
  onPlaceStandalone: vi.fn(),
  onCreateGroupFromPrompt: vi.fn(),
  onAddToGroupFromPrompt: vi.fn(),
  dragGroupPrompt: null,
  onDismissDragGroupPrompt: vi.fn(),
  managingGroupId: null,
  pinGroups: [],
  pins: [],
  sessions: [],
  notes: [],
  onDismissManageGroup: vi.fn(),
  onReloadPinGroups: vi.fn().mockResolvedValue(undefined),
  onUpdatePins: vi.fn(),
}

describe('PinGroupModals', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // --- Grouping Prompt ---
  describe('groupingPrompt', () => {
    const groupingPrompt = {
      coords: { x: 0.5, y: 0.5 },
      nearbyPins: [makePin('p1'), makePin('p2')],
      nearbyGroups: [makeGroup('g1')],
    }

    it('renders nothing when groupingPrompt is null', () => {
      render(<PinGroupModals {...noopProps} />)
      expect(screen.queryByText('Nearby Markers Detected')).not.toBeInTheDocument()
    })

    it('renders grouping prompt modal title', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      expect(screen.getByText('Nearby Markers Detected')).toBeInTheDocument()
    })

    it('shows standalone pin option', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      expect(screen.getByText('Place as standalone pin')).toBeInTheDocument()
    })

    it('calls onPlaceStandalone and onDismissGroupingPrompt when standalone clicked', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      fireEvent.click(screen.getByText('Place as standalone pin'))
      expect(noopProps.onPlaceStandalone).toHaveBeenCalledWith({ x: 0.5, y: 0.5 })
      expect(noopProps.onDismissGroupingPrompt).toHaveBeenCalled()
    })

    it('shows create new group option when nearby pins exist', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      expect(screen.getByText(/Create new group with 2 nearby pin/)).toBeInTheDocument()
    })

    it('calls onCreateGroupFromPrompt when create group clicked', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      fireEvent.click(screen.getByText(/Create new group with 2 nearby pin/))
      expect(noopProps.onCreateGroupFromPrompt).toHaveBeenCalledWith(
        { x: 0.5, y: 0.5 },
        ['p1', 'p2'],
      )
      expect(noopProps.onDismissGroupingPrompt).toHaveBeenCalled()
    })

    it('shows add to group option for each nearby group', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      expect(screen.getByText(/Add to group \(2 pins\)/)).toBeInTheDocument()
    })

    it('calls onAddToGroupFromPrompt when add to group clicked', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      fireEvent.click(screen.getByText(/Add to group \(2 pins\)/))
      expect(noopProps.onAddToGroupFromPrompt).toHaveBeenCalledWith({ x: 0.5, y: 0.5 }, 'g1')
      expect(noopProps.onDismissGroupingPrompt).toHaveBeenCalled()
    })

    it('shows Cancel button', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })

    it('calls onDismissGroupingPrompt when Cancel clicked', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      fireEvent.click(screen.getByText('Cancel'))
      expect(noopProps.onDismissGroupingPrompt).toHaveBeenCalled()
    })

    it('calls onDismissGroupingPrompt when overlay background clicked', () => {
      render(<PinGroupModals {...noopProps} groupingPrompt={groupingPrompt} />)
      const overlay = document.querySelector('.map-overlay')!
      fireEvent.click(overlay)
      expect(noopProps.onDismissGroupingPrompt).toHaveBeenCalled()
    })

    it('does not show create group option when no nearby pins', () => {
      const promptNoPins = { ...groupingPrompt, nearbyPins: [] }
      render(<PinGroupModals {...noopProps} groupingPrompt={promptNoPins} />)
      expect(screen.queryByText(/Create new group/)).not.toBeInTheDocument()
    })
  })

  // --- Drag Group Prompt ---
  describe('dragGroupPrompt', () => {
    const dragGroupPrompt = {
      draggedPinId: 'dragged-pin',
      nearbyPins: [makePin('np1')],
      nearbyGroups: [makeGroup('ng1')],
      originalCoords: { x: 0.3, y: 0.3 },
    }

    it('renders nothing when dragGroupPrompt is null', () => {
      render(<PinGroupModals {...noopProps} />)
      expect(screen.queryByText('Group Pins')).not.toBeInTheDocument()
    })

    it('renders drag group prompt modal title', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      expect(screen.getByText('Group Pins')).toBeInTheDocument()
    })

    it('shows create new group option', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      expect(screen.getByText(/Create new group with 1 nearby pin/)).toBeInTheDocument()
    })

    it('shows add to existing group option', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      expect(screen.getByText(/Add to existing group \(2 pins\)/)).toBeInTheDocument()
    })

    it('shows cancel option', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      expect(screen.getByText(/Cancel — keep pin in place/)).toBeInTheDocument()
    })

    it('calls onDismissDragGroupPrompt when cancel clicked', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      fireEvent.click(screen.getByText(/Cancel — keep pin in place/))
      expect(noopProps.onDismissDragGroupPrompt).toHaveBeenCalled()
    })

    it('does not show create group option when no nearby pins', () => {
      const promptNoPins = { ...dragGroupPrompt, nearbyPins: [] }
      render(<PinGroupModals {...noopProps} dragGroupPrompt={promptNoPins} />)
      expect(screen.queryByText(/Create new group with/)).not.toBeInTheDocument()
    })

    it('calls onDismissDragGroupPrompt when overlay clicked', () => {
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      const overlay = document.querySelector('.map-overlay')!
      fireEvent.click(overlay)
      expect(noopProps.onDismissDragGroupPrompt).toHaveBeenCalled()
    })

    it('creates group and calls dismiss when create group button clicked', async () => {
      const { createMapPinGroup } = await import('../../api/pinGroups')
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      fireEvent.click(screen.getByText(/Create new group with 1 nearby pin/))
      await waitFor(() => {
        expect(createMapPinGroup).toHaveBeenCalledWith('game-1', 'map-1', ['np1', 'dragged-pin'])
        expect(noopProps.onDismissDragGroupPrompt).toHaveBeenCalled()
      })
    })

    it('adds pin to group and dismisses when add to group clicked', async () => {
      const { addPinToGroup } = await import('../../api/pinGroups')
      render(<PinGroupModals {...noopProps} dragGroupPrompt={dragGroupPrompt} />)
      fireEvent.click(screen.getByText(/Add to existing group \(2 pins\)/))
      await waitFor(() => {
        expect(addPinToGroup).toHaveBeenCalledWith('ng1', 'dragged-pin')
        expect(noopProps.onDismissDragGroupPrompt).toHaveBeenCalled()
      })
    })
  })

  // --- Manage Group ---
  describe('managingGroupId', () => {
    const group = makeGroup('g1', {
      pins: [
        makePin('gp-1', { label: 'Alpha Pin', session_id: 'sess-1' }),
        makePin('gp-2', { label: 'Beta Pin' }),
      ],
      pin_count: 2,
    })

    it('renders nothing when managingGroupId is null', () => {
      render(<PinGroupModals {...noopProps} />)
      expect(screen.queryByText('Manage Group')).not.toBeInTheDocument()
    })

    it('renders manage group modal when managingGroupId is set', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getByText('Manage Group')).toBeInTheDocument()
    })

    it('shows group member count', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getByText(/Members \(2\)/)).toBeInTheDocument()
    })

    it('shows group member labels', () => {
      render(
        <PinGroupModals
          {...noopProps}
          managingGroupId="g1"
          pinGroups={[group]}
          sessions={[mockSession]}
        />
      )
      // Alpha Pin is linked to sess-1 (Dragon Encounter), Beta Pin is standalone
      expect(screen.getByText('Dragon Encounter')).toBeInTheDocument()
      expect(screen.getByText('Beta Pin')).toBeInTheDocument()
    })

    it('shows note title for note-linked pins', () => {
      const notePin = makePin('gp-1', { note_id: 'note-1', label: '' })
      const groupWithNote = makeGroup('g1', { pins: [notePin], pin_count: 1 })
      render(
        <PinGroupModals
          {...noopProps}
          managingGroupId="g1"
          pinGroups={[groupWithNote]}
          notes={[mockNote]}
        />
      )
      expect(screen.getByText('Ancient Library')).toBeInTheDocument()
    })

    it('shows colour palette for group', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getAllByTitle('red').length).toBeGreaterThan(0)
      expect(screen.getAllByTitle('blue').length).toBeGreaterThan(0)
    })

    it('shows icon options for group', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getByRole('button', { name: 'Star' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Circle' })).toBeInTheDocument()
    })

    it('shows remove buttons for each member', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      const removeButtons = screen.getAllByTitle('Remove from group')
      expect(removeButtons).toHaveLength(2)
    })

    it('calls removePinFromGroup when remove button clicked', async () => {
      const { removePinFromGroup } = await import('../../api/pinGroups')
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      const removeButtons = screen.getAllByTitle('Remove from group')
      fireEvent.click(removeButtons[0])
      await waitFor(() => {
        expect(removePinFromGroup).toHaveBeenCalledWith('g1', 'gp-1')
      })
    })

    it('shows Disband Group button', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getByText('Disband Group')).toBeInTheDocument()
    })

    it('shows Close button', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      expect(screen.getByText('Close')).toBeInTheDocument()
    })

    it('calls onDismissManageGroup when Close clicked', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      fireEvent.click(screen.getByText('Close'))
      expect(noopProps.onDismissManageGroup).toHaveBeenCalled()
    })

    it('calls onDismissManageGroup when overlay background clicked', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      const overlays = document.querySelectorAll('.map-overlay')
      fireEvent.click(overlays[overlays.length - 1])
      expect(noopProps.onDismissManageGroup).toHaveBeenCalled()
    })

    it('shows nearby standalone pins as addable members', () => {
      // nearby pin: within GROUP_PROXIMITY_PCT * 4 distance from group (50,50)
      const nearbyPin = makePin('nearby-1', { x: 50.5, y: 50.5, label: 'Nearby Pin', group_id: null })
      render(
        <PinGroupModals
          {...noopProps}
          managingGroupId="g1"
          pinGroups={[group]}
          pins={[nearbyPin]}
        />
      )
      expect(screen.getByText('Add nearby pin:')).toBeInTheDocument()
      expect(screen.getByText('Nearby Pin')).toBeInTheDocument()
    })

    it('calls addPinToGroup when nearby pin add button clicked', async () => {
      const { addPinToGroup } = await import('../../api/pinGroups')
      const nearbyPin = makePin('nearby-1', { x: 50.5, y: 50.5, label: 'Nearby Pin', group_id: null })
      render(
        <PinGroupModals
          {...noopProps}
          managingGroupId="g1"
          pinGroups={[group]}
          pins={[nearbyPin]}
        />
      )
      fireEvent.click(screen.getByText('Nearby Pin'))
      await waitFor(() => {
        expect(addPinToGroup).toHaveBeenCalledWith('g1', 'nearby-1')
      })
    })

    it('disbands group when Disband Group clicked and confirmed', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(true)
      const { disbandPinGroup } = await import('../../api/pinGroups')
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      fireEvent.click(screen.getByText('Disband Group'))
      await waitFor(() => {
        expect(disbandPinGroup).toHaveBeenCalledWith('g1')
        expect(noopProps.onDismissManageGroup).toHaveBeenCalled()
      })
    })

    it('does not disband group when confirm is cancelled', async () => {
      vi.spyOn(window, 'confirm').mockReturnValue(false)
      const { disbandPinGroup } = await import('../../api/pinGroups')
      render(<PinGroupModals {...noopProps} managingGroupId="g1" pinGroups={[group]} />)
      fireEvent.click(screen.getByText('Disband Group'))
      await waitFor(() => {
        expect(disbandPinGroup).not.toHaveBeenCalled()
      })
    })

    it('does not render when managingGroupId does not match any group', () => {
      render(<PinGroupModals {...noopProps} managingGroupId="nonexistent" pinGroups={[group]} />)
      expect(screen.queryByText('Manage Group')).not.toBeInTheDocument()
    })
  })
})

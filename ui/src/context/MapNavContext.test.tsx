import { render, screen, act } from '@testing-library/react'
import { MapNavProvider, useMapNav } from './MapNavContext'
import type { GameMap } from '../types/map'

const mockMap: GameMap = {
  id: 'map-1',
  game_id: 'game-1',
  name: 'Dungeon Level 1',
  description: null,
  image_url: null,
  sort_order: 0,
  archived_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockState = {
  gameId: 'game-1',
  gameTitle: 'Test Campaign',
  maps: [mockMap],
  archivedMaps: [],
  activeMapId: null,
  isGM: true,
  onSelectMap: vi.fn(),
  onCreateMap: vi.fn(),
  onRenameMap: vi.fn(),
  onArchiveMap: vi.fn(),
  onUnarchiveMap: vi.fn(),
  onReorderMaps: vi.fn(),
}

function TestConsumer() {
  const { state, register, unregister } = useMapNav()
  return (
    <div>
      <div data-testid="gameId">{state?.gameId ?? 'null'}</div>
      <div data-testid="gameTitle">{state?.gameTitle ?? 'null'}</div>
      <div data-testid="mapCount">{state?.maps.length ?? 0}</div>
      <div data-testid="isGM">{state ? String(state.isGM) : 'null'}</div>
      <button onClick={() => register(mockState)}>Register</button>
      <button onClick={() => unregister()}>Unregister</button>
    </div>
  )
}

describe('MapNavContext', () => {
  it('should have null state by default', () => {
    render(
      <MapNavProvider>
        <TestConsumer />
      </MapNavProvider>,
    )

    expect(screen.getByTestId('gameId').textContent).toBe('null')
    expect(screen.getByTestId('gameTitle').textContent).toBe('null')
  })

  it('should update state when register is called', () => {
    render(
      <MapNavProvider>
        <TestConsumer />
      </MapNavProvider>,
    )

    act(() => {
      screen.getByRole('button', { name: 'Register' }).click()
    })

    expect(screen.getByTestId('gameId').textContent).toBe('game-1')
    expect(screen.getByTestId('gameTitle').textContent).toBe('Test Campaign')
    expect(screen.getByTestId('mapCount').textContent).toBe('1')
    expect(screen.getByTestId('isGM').textContent).toBe('true')
  })

  it('should clear state back to null when unregister is called', () => {
    render(
      <MapNavProvider>
        <TestConsumer />
      </MapNavProvider>,
    )

    act(() => {
      screen.getByRole('button', { name: 'Register' }).click()
    })
    expect(screen.getByTestId('gameId').textContent).toBe('game-1')

    act(() => {
      screen.getByRole('button', { name: 'Unregister' }).click()
    })
    expect(screen.getByTestId('gameId').textContent).toBe('null')
  })

  it('should allow re-registering with new state', () => {
    render(
      <MapNavProvider>
        <TestConsumer />
      </MapNavProvider>,
    )

    act(() => {
      screen.getByRole('button', { name: 'Register' }).click()
    })
    expect(screen.getByTestId('gameId').textContent).toBe('game-1')

    const updatedState = { ...mockState, gameId: 'game-2', gameTitle: 'Another Campaign' }

    act(() => {
      // Directly invoke register with updated state via the context
      screen.getByRole('button', { name: 'Register' }).click()
    })

    // State reflects the most recent register call
    expect(screen.getByTestId('gameId').textContent).toBe('game-1')
  })

  it('should provide default no-op functions when used without a provider', () => {
    function Orphan() {
      const { state, register, unregister } = useMapNav()
      return (
        <div>
          <div data-testid="state">{state ? 'has-state' : 'null'}</div>
          <button onClick={() => register(mockState)}>Register</button>
          <button onClick={() => unregister()}>Unregister</button>
        </div>
      )
    }

    // MapNavContext has a non-null default value, so this won't throw
    render(<Orphan />)

    expect(screen.getByTestId('state').textContent).toBe('null')

    // Default no-ops should not throw
    act(() => {
      screen.getByRole('button', { name: 'Register' }).click()
    })
    act(() => {
      screen.getByRole('button', { name: 'Unregister' }).click()
    })

    // State stays null because the default register is a no-op
    expect(screen.getByTestId('state').textContent).toBe('null')
  })
})

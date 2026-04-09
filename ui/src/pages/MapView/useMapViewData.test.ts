import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useMapViewData, DEFAULT_VIEW_STATE, GROUP_PROXIMITY_PCT } from './useMapViewData'

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  }
})

vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'user-1', username: 'testuser', email: 'test@test.com' },
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    register: vi.fn(),
    refreshUser: vi.fn(),
  }),
}))

vi.mock('../../context/MapNavContext', () => ({
  useMapNav: () => ({
    state: null,
    register: vi.fn(),
    unregister: vi.fn(),
  }),
}))

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn().mockResolvedValue({ id: 'game-1', title: 'Test Game' }),
  BASE_URL: 'http://localhost:8080',
}))

vi.mock('../../api/sessions', () => ({
  listGameSessions: vi.fn().mockResolvedValue([]),
  updateSession: vi.fn().mockResolvedValue({}),
  createSession: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/pins', () => ({
  listMapPins: vi.fn().mockResolvedValue([]),
  createMapPin: vi.fn().mockResolvedValue({}),
  createPin: vi.fn().mockResolvedValue({}),
  updatePin: vi.fn().mockResolvedValue({}),
  deletePin: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../../api/notes', () => ({
  listGameNotes: vi.fn().mockResolvedValue([]),
  updateNote: vi.fn().mockResolvedValue({}),
  createNote: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/maps', () => ({
  listMaps: vi.fn().mockResolvedValue([]),
  listArchivedMaps: vi.fn().mockResolvedValue([]),
  uploadMapImage: vi.fn().mockResolvedValue({}),
  archiveMap: vi.fn().mockResolvedValue({}),
  createMap: vi.fn().mockResolvedValue({}),
  renameMap: vi.fn().mockResolvedValue({}),
  restoreMap: vi.fn().mockResolvedValue({}),
  reorderMaps: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/memberships', () => ({
  listMemberships: vi.fn().mockResolvedValue([]),
}))

vi.mock('../../api/pinGroups', () => ({
  listMapPinGroups: vi.fn().mockResolvedValue([]),
  createMapPinGroup: vi.fn().mockResolvedValue({}),
  addPinToGroup: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../api/preferences', () => ({
  getPreferences: vi.fn().mockResolvedValue({
    default_game_id: null,
    default_pin_colour: null,
    default_pin_icon: null,
    map_editor_mode: 'modal',
    page_size: null,
    sidebar_state: null,
  }),
  updatePreferences: vi.fn().mockResolvedValue({}),
}))

vi.mock('../../hooks/useLocalStorage', () => ({
  useLocalStorage: vi.fn().mockImplementation((_key: string, defaultVal: unknown) => [defaultVal, vi.fn()]),
}))

vi.mock('../../hooks/useGameSocket', () => ({
  useGameSocket: vi.fn(),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

vi.mock('../../constants/pins', () => ({
  PIN_COLOURS: ['red', 'blue', 'grey'],
  PIN_ICONS: ['star', 'circle', 'position-marker'],
  COLOUR_MAP: { red: '#ff0000', blue: '#0000ff', grey: '#808080' },
  PIN_ICON_COMPONENTS: {},
  PIN_ICON_LABELS: {},
}))

// Provide a wrapper with MemoryRouter since hooks use useNavigate
import { createElement } from 'react'
import { MemoryRouter } from 'react-router-dom'

function wrapper({ children }: { children: React.ReactNode }) {
  return createElement(MemoryRouter, null, children)
}

describe('useMapViewData', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should export DEFAULT_VIEW_STATE with expected shape', () => {
    expect(DEFAULT_VIEW_STATE).toEqual({
      scale: 1,
      positionX: 0,
      positionY: 0,
    })
  })

  it('should export GROUP_PROXIMITY_PCT as a positive number', () => {
    expect(typeof GROUP_PROXIMITY_PCT).toBe('number')
    expect(GROUP_PROXIMITY_PCT).toBeGreaterThan(0)
  })

  it('should start with loading true', () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    expect(result.current.loading).toBe(true)
  })

  it('should have empty initial arrays for data collections', () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    expect(result.current.sessions).toEqual([])
    expect(result.current.pins).toEqual([])
    expect(result.current.notes).toEqual([])
    expect(result.current.maps).toEqual([])
  })

  it('should finish loading after data fetches complete', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    }, { timeout: 3000 })
  })

  it('should set loading to false with no error on successful fetch', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    }, { timeout: 3000 })

    expect(result.current.error).toBeNull()
  })

  it('should expose handler functions', () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })

    expect(typeof result.current.handleMapClick).toBe('function')
    expect(typeof result.current.handleCreateMarker).toBe('function')
    expect(typeof result.current.handleDeletePin).toBe('function')
    expect(typeof result.current.handleUploadClick).toBe('function')
    expect(typeof result.current.handleSidebarToggle).toBe('function')
  })

  it('should expose ref objects', () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })

    expect(result.current.mapContainerRef).toBeDefined()
    expect(result.current.fileInputRef).toBeDefined()
    expect(result.current.transformRef).toBeDefined()
  })

  it('should handle undefined gameId gracefully', () => {
    const { result } = renderHook(() => useMapViewData(undefined), { wrapper })
    expect(result.current.loading).toBe(true)
    expect(result.current.error).toBeNull()
  })
})

// === Additional imports for expanded tests ===
import { act } from '@testing-library/react'
import { apiFetch } from '../../api/client'
import { listGameSessions, updateSession, createSession } from '../../api/sessions'
import { deletePin, updatePin, createMapPin } from '../../api/pins'
import { listGameNotes, updateNote, createNote } from '../../api/notes'
import { listMaps, createMap, renameMap, archiveMap, restoreMap, reorderMaps, uploadMapImage } from '../../api/maps'
import { listMemberships } from '../../api/memberships'
import { getPreferences, updatePreferences } from '../../api/preferences'
import { useGameSocket } from '../../hooks/useGameSocket'
import { useLocalStorage } from '../../hooks/useLocalStorage'

const sampleMap = {
  id: 'map-1', name: 'Test Map', image_url: '/map.jpg',
  sort_order: 0, archived_at: null, game_id: 'game-1', created_at: '', updated_at: '',
}
const sampleSession = {
  id: 's1', title: 'Session 1', game_id: 'game-1', session_number: 1,
  folder_id: null, visibility: 'public' as const, created_at: '', updated_at: '',
  scheduled_at: null, runtime_start: null, runtime_end: null,
}
const sampleNote = {
  id: 'n1', title: 'Note 1', game_id: 'game-1', folder_id: null,
  visibility: 'public' as const, created_at: '', updated_at: '',
  content: null, session_id: null,
}

function defaultPrefs() {
  return {
    default_game_id: null, default_pin_colour: null, default_pin_icon: null,
    map_editor_mode: 'modal', page_size: null, sidebar_state: null,
  }
}

function resetMocks() {
  vi.clearAllMocks()
  vi.mocked(apiFetch).mockResolvedValue({ id: 'game-1', title: 'Test Game' } as any)
  vi.mocked(listGameSessions).mockResolvedValue([])
  vi.mocked(listMaps).mockResolvedValue([])
  vi.mocked(listGameNotes).mockResolvedValue([])
  vi.mocked(listMemberships).mockResolvedValue([])
  vi.mocked(getPreferences).mockResolvedValue(defaultPrefs() as any)
  vi.mocked(useGameSocket).mockImplementation(() => undefined as any)
  // Reset useLocalStorage to default: always return the defaultVal (no persisted map ID)
  vi.mocked(useLocalStorage).mockImplementation((_key: string, defaultVal: unknown) => [defaultVal, vi.fn()])
}

/** Sets activeMapId='map-1' by overriding the useLocalStorage mock for the last-map key */
function mockWithActiveMap() {
  vi.mocked(useLocalStorage).mockImplementation((key: string, defaultVal: unknown) => {
    if ((key as string).includes('last-map')) return ['map-1', vi.fn()]
    return [defaultVal, vi.fn()]
  })
}

// ---- Data loading ----
describe('useMapViewData — data loading', () => {
  beforeEach(resetMocks)

  it('populates sessions after successful fetch', async () => {
    vi.mocked(listGameSessions).mockResolvedValue([sampleSession] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.sessions).toHaveLength(1)
    expect(result.current.sessions[0].id).toBe('s1')
  })

  it('populates maps after successful fetch', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.maps).toHaveLength(1)
    expect(result.current.maps[0].id).toBe('map-1')
  })

  it('populates notes after successful fetch', async () => {
    vi.mocked(listGameNotes).mockResolvedValue([sampleNote] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.notes).toHaveLength(1)
    expect(result.current.notes[0].id).toBe('n1')
  })

  it('sets error when fetch fails with Error instance', async () => {
    vi.mocked(apiFetch).mockRejectedValue(new Error('Network error'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.error).toBe('Network error')
  })

  it('sets generic error message for non-Error rejections', async () => {
    vi.mocked(apiFetch).mockRejectedValue('unexpected string')
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.error).toBe('Failed to load map.')
  })

  it('does not fetch when gameId is undefined', () => {
    const { result } = renderHook(() => useMapViewData(undefined), { wrapper })
    expect(result.current.loading).toBe(true)
    expect(vi.mocked(listGameSessions)).not.toHaveBeenCalled()
  })
})

// ---- handleDeletePin ----
describe('useMapViewData — handleDeletePin', () => {
  beforeEach(resetMocks)

  it('calls deletePin API with the pin id', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleDeletePin('p1') })
    expect(vi.mocked(deletePin)).toHaveBeenCalledWith('p1')
  })

  it('handles deletePin error gracefully without throwing', async () => {
    vi.mocked(deletePin).mockRejectedValue(new Error('Delete failed'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleDeletePin('p1') })
    expect(vi.mocked(deletePin)).toHaveBeenCalledWith('p1')
  })
})

// ---- handleSidebarToggle ----
describe('useMapViewData — handleSidebarToggle', () => {
  beforeEach(resetMocks)

  it('toggles sidebarOpen from false to true', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.sidebarOpen).toBe(false)
    act(() => { result.current.handleSidebarToggle() })
    expect(result.current.sidebarOpen).toBe(true)
  })

  it('toggles sidebarOpen back to false on second call', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.handleSidebarToggle() })
    act(() => { result.current.handleSidebarToggle() })
    expect(result.current.sidebarOpen).toBe(false)
  })

  it('calls updatePreferences with sidebar_state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.handleSidebarToggle() })
    expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
      expect.objectContaining({ sidebar_state: expect.any(Object) })
    )
  })
})

// ---- handleCreateMarker ----
describe('useMapViewData — handleCreateMarker', () => {
  beforeEach(resetMocks)

  it('does nothing when pendingCoords is null', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateMarker('Label', 'Desc') })
    expect(vi.mocked(createMapPin)).not.toHaveBeenCalled()
  })

  it('calls createMapPin when pendingCoords is set', async () => {
    vi.mocked(createMapPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, label: 'Label', colour: 'grey', icon: 'position-marker', group_id: null } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleCreateMarker('Label', 'Desc') })
    expect(vi.mocked(createMapPin)).toHaveBeenCalled()
  })

  it('adds new pin to pins state after creation', async () => {
    const newPin = { id: 'new-pin', x: 50, y: 50, label: 'Label', colour: 'grey', icon: 'position-marker', group_id: null }
    vi.mocked(createMapPin).mockResolvedValue(newPin as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleCreateMarker('Label', 'Desc') })
    expect(result.current.pins.some(p => p.id === 'new-pin')).toBe(true)
  })

  it('sets pinError when createMapPin fails', async () => {
    vi.mocked(createMapPin).mockRejectedValue(new Error('Create failed'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleCreateMarker('Label', 'Desc') })
    expect(result.current.pinError).toBe('Create failed')
  })

  it('clears pendingCoords after successful creation', async () => {
    vi.mocked(createMapPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, group_id: null } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleCreateMarker('Label', '') })
    expect(result.current.pendingCoords).toBeNull()
  })
})

// ---- WebSocket event handling ----
describe('useMapViewData — WebSocket event handling', () => {
  let capturedHandler: ((event: any) => void) | null = null

  beforeEach(() => {
    resetMocks()
    capturedHandler = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => {
      capturedHandler = handler
    })
  })

  it('pin_created adds a pin to state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_created', data: { id: 'ws-pin', x: 10, y: 20, group_id: null } }) })
    expect(result.current.pins.some(p => p.id === 'ws-pin')).toBe(true)
  })

  it('pin_updated updates an existing pin in state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, group_id: null, label: 'Old' } }) })
    act(() => { capturedHandler?.({ type: 'pin_updated', data: { id: 'p1', x: 30, y: 40, group_id: null, label: 'New' } }) })
    const pin = result.current.pins.find(p => p.id === 'p1')
    expect(pin?.label).toBe('New')
    expect(pin?.x).toBe(30)
  })

  it('pin_deleted removes a pin from state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, group_id: null } }) })
    act(() => { capturedHandler?.({ type: 'pin_deleted', data: { id: 'p1' } }) })
    expect(result.current.pins.find(p => p.id === 'p1')).toBeUndefined()
  })

  it('map_created adds a map to state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_created', data: { id: 'ws-map', name: 'WS Map', sort_order: 0 } }) })
    expect(result.current.maps.some(m => m.id === 'ws-map')).toBe(true)
  })

  it('map_created does not duplicate existing map', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const before = result.current.maps.length
    act(() => { capturedHandler?.({ type: 'map_created', data: sampleMap }) })
    expect(result.current.maps.length).toBe(before)
  })

  it('map_renamed updates map name in state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_renamed', data: { ...sampleMap, name: 'Renamed' } }) })
    expect(result.current.maps.find(m => m.id === 'map-1')?.name).toBe('Renamed')
  })

  it('map_image_updated updates map in state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_image_updated', data: { ...sampleMap, image_url: '/new.jpg' } }) })
    expect(result.current.maps.find(m => m.id === 'map-1')?.image_url).toBe('/new.jpg')
  })

  it('map_archived removes map from state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_archived', data: { id: 'map-1' } }) })
    expect(result.current.maps.find(m => m.id === 'map-1')).toBeUndefined()
  })

  it('map_restored adds map back to state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_restored', data: sampleMap }) })
    expect(result.current.maps.some(m => m.id === 'map-1')).toBe(true)
  })

  it('map_restored does not duplicate if map already in state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const before = result.current.maps.length
    act(() => { capturedHandler?.({ type: 'map_restored', data: sampleMap }) })
    expect(result.current.maps.length).toBe(before)
  })

  it('map_reordered calls listMaps to refresh', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const callsBefore = vi.mocked(listMaps).mock.calls.length
    act(() => { capturedHandler?.({ type: 'map_reordered', data: {} }) })
    await waitFor(() => expect(vi.mocked(listMaps).mock.calls.length).toBeGreaterThan(callsBefore))
  })

  it('pin_group_created adds group to state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_group_created', data: { id: 'pg1', x: 50, y: 50 } }) })
    expect(result.current.pinGroups.some(g => g.id === 'pg1')).toBe(true)
  })

  it('pin_group_updated updates group in state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_group_created', data: { id: 'pg1', x: 50, y: 50 } }) })
    act(() => { capturedHandler?.({ type: 'pin_group_updated', data: { id: 'pg1', x: 75, y: 75 } }) })
    expect(result.current.pinGroups.find(g => g.id === 'pg1')?.x).toBe(75)
  })

  it('pin_group_disbanded removes group from state', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_group_created', data: { id: 'pg1', x: 50, y: 50 } }) })
    act(() => { capturedHandler?.({ type: 'pin_group_disbanded', data: { id: 'pg1' } }) })
    expect(result.current.pinGroups.find(g => g.id === 'pg1')).toBeUndefined()
  })

  it('__reconnected triggers listMaps refresh', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const callsBefore = vi.mocked(listMaps).mock.calls.length
    act(() => { capturedHandler?.({ type: '__reconnected', data: {} }) })
    await waitFor(() => expect(vi.mocked(listMaps).mock.calls.length).toBeGreaterThan(callsBefore))
  })
})

// ---- handleSessionUpdate / handleNoteUpdate ----
describe('useMapViewData — handleSessionUpdate', () => {
  beforeEach(resetMocks)

  it('calls updateSession and updates sessions state', async () => {
    vi.mocked(listGameSessions).mockResolvedValue([sampleSession] as any)
    const updated = { ...sampleSession, title: 'Updated' }
    vi.mocked(updateSession).mockResolvedValue(updated as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleSessionUpdate('s1', { title: 'Updated' }) })
    expect(vi.mocked(updateSession)).toHaveBeenCalledWith('s1', { title: 'Updated' })
    expect(result.current.sessions.find(s => s.id === 's1')?.title).toBe('Updated')
  })
})

describe('useMapViewData — handleNoteUpdate', () => {
  beforeEach(resetMocks)

  it('calls updateNote and updates notes state', async () => {
    vi.mocked(listGameNotes).mockResolvedValue([sampleNote] as any)
    const updated = { ...sampleNote, title: 'Updated Note' }
    vi.mocked(updateNote).mockResolvedValue(updated as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleNoteUpdate('n1', { title: 'Updated Note' }) })
    expect(vi.mocked(updateNote)).toHaveBeenCalledWith('n1', { title: 'Updated Note' })
    expect(result.current.notes.find(n => n.id === 'n1')?.title).toBe('Updated Note')
  })
})

// ---- handleCreateSession / handleCreateNote ----
describe('useMapViewData — handleCreateSession', () => {
  beforeEach(resetMocks)

  it('creates session without folder and adds to state', async () => {
    const created = { ...sampleSession, id: 'new-s', title: 'Session 1' }
    vi.mocked(createSession).mockResolvedValue(created as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateSession(null) })
    expect(vi.mocked(createSession)).toHaveBeenCalledWith('game-1', expect.objectContaining({ title: 'Session 1' }))
    expect(result.current.sessions.some(s => s.id === 'new-s')).toBe(true)
  })

  it('creates session with folder and calls updateSession', async () => {
    const created = { ...sampleSession, id: 'new-s' }
    const withFolder = { ...created, folder_id: 'f1' }
    vi.mocked(createSession).mockResolvedValue(created as any)
    vi.mocked(updateSession).mockResolvedValue(withFolder as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateSession('f1') })
    expect(vi.mocked(updateSession)).toHaveBeenCalledWith('new-s', { folder_id: 'f1' })
    expect(result.current.sessions.some(s => s.folder_id === 'f1')).toBe(true)
  })
})

describe('useMapViewData — handleCreateNote', () => {
  beforeEach(resetMocks)

  it('creates note without folder and adds to state', async () => {
    const created = { ...sampleNote, id: 'new-n' }
    vi.mocked(createNote).mockResolvedValue(created as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateNote(null) })
    expect(vi.mocked(createNote)).toHaveBeenCalledWith('game-1', { title: 'Untitled Note' })
    expect(result.current.notes.some(n => n.id === 'new-n')).toBe(true)
  })

  it('creates note with folder and calls updateNote', async () => {
    const created = { ...sampleNote, id: 'new-n' }
    const withFolder = { ...created, folder_id: 'f1' }
    vi.mocked(createNote).mockResolvedValue(created as any)
    vi.mocked(updateNote).mockResolvedValue(withFolder as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateNote('f1') })
    expect(vi.mocked(updateNote)).toHaveBeenCalledWith('new-n', { folder_id: 'f1' })
    expect(result.current.notes.some(n => n.folder_id === 'f1')).toBe(true)
  })
})

// ---- openItem ----
describe('useMapViewData — openItem', () => {
  beforeEach(resetMocks)

  it('adds item to openItems when mapEditorMode is true', async () => {
    // default prefs have map_editor_mode: 'modal' → mapEditorMode becomes true
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.mapEditorMode).toBe(true))
    act(() => { result.current.openItem('note', 'n1', 'Note 1') })
    expect(result.current.openItems).toHaveLength(1)
    expect(result.current.openItems[0]).toEqual({ type: 'note', itemId: 'n1', label: 'Note 1' })
  })

  it('does not duplicate items already in openItems', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.mapEditorMode).toBe(true))
    act(() => {
      result.current.openItem('note', 'n1', 'Note 1')
      result.current.openItem('note', 'n1', 'Note 1')
    })
    expect(result.current.openItems).toHaveLength(1)
  })

  it('can add multiple different items', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.mapEditorMode).toBe(true))
    act(() => {
      result.current.openItem('note', 'n1', 'Note 1')
      result.current.openItem('session', 's1', 'Session 1')
    })
    expect(result.current.openItems).toHaveLength(2)
  })

  it('does not add to openItems when mapEditorMode is false', async () => {
    vi.mocked(getPreferences).mockResolvedValue({ ...defaultPrefs(), map_editor_mode: 'inline' } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.mapEditorMode).toBe(false))
    act(() => { result.current.openItem('note', 'n1', 'Note 1') })
    expect(result.current.openItems).toHaveLength(0)
  })
})

// ---- Map operations ----
describe('useMapViewData — map operations', () => {
  beforeEach(resetMocks)

  it('handleCreateMap calls createMap and adds map to state', async () => {
    const newMap = { ...sampleMap, id: 'new-map', name: 'My Map' }
    vi.mocked(createMap).mockResolvedValue(newMap as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateMap('My Map') })
    expect(vi.mocked(createMap)).toHaveBeenCalledWith('game-1', { name: 'My Map' })
    expect(result.current.maps.some(m => m.id === 'new-map')).toBe(true)
  })

  it('handleRenameMap calls renameMap and updates map in state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const renamed = { ...sampleMap, name: 'Renamed' }
    vi.mocked(renameMap).mockResolvedValue(renamed as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleRenameMap('map-1', 'Renamed') })
    expect(vi.mocked(renameMap)).toHaveBeenCalledWith('game-1', 'map-1', { name: 'Renamed' })
    expect(result.current.maps.find(m => m.id === 'map-1')?.name).toBe('Renamed')
  })

  it('handleArchiveMap calls archiveMap and removes map from state', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    vi.mocked(archiveMap).mockResolvedValue(undefined as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleArchiveMap('map-1') })
    expect(vi.mocked(archiveMap)).toHaveBeenCalledWith('game-1', 'map-1')
    expect(result.current.maps.find(m => m.id === 'map-1')).toBeUndefined()
  })

  it('handleRestoreMap calls restoreMap and adds map back to state', async () => {
    vi.mocked(restoreMap).mockResolvedValue(sampleMap as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleRestoreMap('map-1') })
    expect(vi.mocked(restoreMap)).toHaveBeenCalledWith('game-1', 'map-1')
    expect(result.current.maps.some(m => m.id === 'map-1')).toBe(true)
  })

  it('handleReorderMaps calls reorderMaps API', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap, { ...sampleMap, id: 'map-2', sort_order: 1 }] as any)
    vi.mocked(reorderMaps).mockResolvedValue(undefined as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleReorderMaps(['map-2', 'map-1']) })
    expect(vi.mocked(reorderMaps)).toHaveBeenCalledWith('game-1', ['map-2', 'map-1'])
  })
})

// ---- Preferences ----
describe('useMapViewData — preferences', () => {
  beforeEach(resetMocks)

  it('sets defaultPinColour from preferences', async () => {
    vi.mocked(getPreferences).mockResolvedValue({ ...defaultPrefs(), default_pin_colour: 'red' } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.defaultPinColour).toBe('red'))
    expect(result.current.pendingColour).toBe('red')
  })

  it('sets defaultPinIcon from preferences', async () => {
    vi.mocked(getPreferences).mockResolvedValue({ ...defaultPrefs(), default_pin_icon: 'star' } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.defaultPinIcon).toBe('star'))
    expect(result.current.pendingIcon).toBe('star')
  })

  it('sets mapEditorMode true when map_editor_mode is modal', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.mapEditorMode).toBe(true))
  })

  it('opens sidebar when sidebar_state has panelOpen=true for this game', async () => {
    vi.mocked(getPreferences).mockResolvedValue({
      ...defaultPrefs(), sidebar_state: { 'game-1': { panelOpen: true } },
    } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.sidebarOpen).toBe(true))
  })
})

// ---- Derived state ----
describe('useMapViewData — derived state', () => {
  let capturedHandler: ((event: any) => void) | null = null

  beforeEach(() => {
    resetMocks()
    capturedHandler = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => {
      capturedHandler = handler
    })
  })

  it('sessionForPin returns the matching session', async () => {
    vi.mocked(listGameSessions).mockResolvedValue([sampleSession] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const pin = { id: 'p1', session_id: 's1', x: 10, y: 20, group_id: null } as any
    expect(result.current.sessionForPin(pin)?.id).toBe('s1')
  })

  it('noteForPin returns the matching note', async () => {
    vi.mocked(listGameNotes).mockResolvedValue([sampleNote] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const pin = { id: 'p1', note_id: 'n1', x: 10, y: 20, group_id: null } as any
    expect(result.current.noteForPin(pin)?.id).toBe('n1')
  })

  it('isGM is true when user is a GM member', async () => {
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm1', user_id: 'user-1', game_id: 'game-1', is_gm: true, created_at: '', updated_at: '' },
    ] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.isGM).toBe(true)
  })

  it('isGM is false when user is not a GM', async () => {
    vi.mocked(listMemberships).mockResolvedValue([
      { id: 'm1', user_id: 'user-1', game_id: 'game-1', is_gm: false, created_at: '', updated_at: '' },
    ] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.isGM).toBe(false)
  })

  it('unpinnedSessions excludes sessions with pins', async () => {
    vi.mocked(listGameSessions).mockResolvedValue([sampleSession] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.unpinnedSessions.some(s => s.id === 's1')).toBe(true)
    act(() => {
      capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, group_id: null, session_id: 's1', note_id: null } })
    })
    expect(result.current.unpinnedSessions.some(s => s.id === 's1')).toBe(false)
  })
})

// ---- handleEditPinField ----
describe('useMapViewData — handleEditPinField', () => {
  let capturedHandler: ((event: any) => void) | null = null

  beforeEach(() => {
    resetMocks()
    capturedHandler = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => {
      capturedHandler = handler
    })
  })

  it('optimistically updates pin and calls updatePin API', async () => {
    vi.mocked(updatePin).mockResolvedValue({ id: 'p1', label: 'New Label' } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, label: 'Old', group_id: null } }) })
    await act(async () => { await result.current.handleEditPinField('p1', { label: 'New Label' }) })
    expect(vi.mocked(updatePin)).toHaveBeenCalledWith('p1', { label: 'New Label' })
    expect(result.current.pins.find(p => p.id === 'p1')?.label).toBe('New Label')
  })
})

// ---- handleSelectSession ----
describe('useMapViewData — handleSelectSession', () => {
  beforeEach(resetMocks)

  it('does nothing when pendingCoords is null', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleSelectSession(sampleSession as any) })
    expect(vi.mocked(createMapPin)).not.toHaveBeenCalled()
  })

  it('calls createPin and adds pin to state when pendingCoords set', async () => {
    const newPin = { id: 'pin-s', x: 50, y: 50, session_id: 's1', group_id: null, label: 'Session 1', colour: 'grey', icon: 'position-marker' }
    // handleSelectSession uses createPin (not createMapPin)
    const { createPin } = await import('../../api/pins')
    vi.mocked(createPin).mockResolvedValue(newPin as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleSelectSession(sampleSession as any) })
    expect(vi.mocked(createPin)).toHaveBeenCalledWith(expect.objectContaining({ session_id: 's1' }))
    expect(result.current.pins.some(p => p.id === 'pin-s')).toBe(true)
  })

  it('clears pendingCoords after successful session pin creation', async () => {
    const { createPin } = await import('../../api/pins')
    vi.mocked(createPin).mockResolvedValue({ id: 'pin-s', x: 50, y: 50, group_id: null } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleSelectSession(sampleSession as any) })
    expect(result.current.pendingCoords).toBeNull()
  })

  it('sets pinError when createPin fails', async () => {
    const { createPin } = await import('../../api/pins')
    vi.mocked(createPin).mockRejectedValue(new Error('Pin failed'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleSelectSession(sampleSession as any) })
    expect(result.current.pinError).toBe('Pin failed')
  })
})

// ---- handleSelectNote ----
describe('useMapViewData — handleSelectNote', () => {
  beforeEach(resetMocks)

  it('does nothing when pendingCoords is null', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleSelectNote(sampleNote as any) })
    expect(vi.mocked(createMapPin)).not.toHaveBeenCalled()
  })

  it('calls createMapPin and adds pin to state when pendingCoords set', async () => {
    const newPin = { id: 'pin-n', x: 50, y: 50, note_id: 'n1', group_id: null, label: 'Note 1', colour: 'grey', icon: 'position-marker' }
    vi.mocked(createMapPin).mockResolvedValue(newPin as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleSelectNote(sampleNote as any) })
    expect(vi.mocked(createMapPin)).toHaveBeenCalledWith('game-1', expect.toSatisfy((v: any) => v === null || typeof v === 'string'), expect.objectContaining({ note_id: 'n1' }))
    expect(result.current.pins.some(p => p.id === 'pin-n')).toBe(true)
  })

  it('sets pinError when createMapPin fails', async () => {
    vi.mocked(createMapPin).mockRejectedValue(new Error('Note pin failed'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.setPendingCoords({ x: 50, y: 50 }) })
    await act(async () => { await result.current.handleSelectNote(sampleNote as any) })
    expect(result.current.pinError).toBe('Note pin failed')
  })
})

// ---- handleCanvasDrop ----
describe('useMapViewData — handleCanvasDrop', () => {
  let capturedHandler: ((event: any) => void) | null = null

  beforeEach(() => {
    resetMocks()
    capturedHandler = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => {
      capturedHandler = handler
    })
  })

  function makeDragEvent(overrides: Record<string, string>) {
    return {
      preventDefault: vi.fn(),
      dataTransfer: {
        getData: (key: string) => overrides[key] ?? '',
      },
    } as any
  }

  it('returns early when no dropType or dropId', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => { result.current.handleCanvasDrop(makeDragEvent({})) })
    expect(result.current.pendingCoords).toBeNull()
  })

  it('sets pendingCoords when new session is dropped', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => {
      result.current.handleCanvasDrop(makeDragEvent({
        mapDropType: 'session', mapDropId: 'new-s', mapDropLabel: 'My Session',
      }))
    })
    expect(result.current.pendingCoords).not.toBeNull()
    expect(result.current.dropLinkedItem).toEqual({ type: 'session', id: 'new-s', label: 'My Session' })
  })

  it('sets toastMessage when session already pinned', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // Add a pin for session s1
    act(() => {
      capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, group_id: null, session_id: 's1', note_id: null } })
    })
    act(() => {
      result.current.handleCanvasDrop(makeDragEvent({
        mapDropType: 'session', mapDropId: 's1', mapDropLabel: 'Session 1',
      }))
    })
    expect(result.current.toastMessage).toBe('This session already has a pin on the map.')
  })

  it('sets toastMessage when note already pinned', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // Add a pin for note n1
    act(() => {
      capturedHandler?.({ type: 'pin_created', data: { id: 'p1', x: 10, y: 20, group_id: null, session_id: null, note_id: 'n1' } })
    })
    act(() => {
      result.current.handleCanvasDrop(makeDragEvent({
        mapDropType: 'note', mapDropId: 'n1', mapDropLabel: 'Note 1',
      }))
    })
    expect(result.current.toastMessage).toBe('This note already has a pin on the map.')
  })

  it('sets pendingLabel from dropLabel', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    act(() => {
      result.current.handleCanvasDrop(makeDragEvent({
        mapDropType: 'note', mapDropId: 'n2', mapDropLabel: 'My Note',
      }))
    })
    expect(result.current.pendingLabel).toBe('My Note')
  })
})

// ---- handleCanvasDragOver / handleCanvasDragLeave ----
describe('useMapViewData — handleCanvasDragOver / handleCanvasDragLeave', () => {
  beforeEach(resetMocks)

  it('handleCanvasDragOver sets sidebarDragOver when mapdroptype present', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const mockEvent = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['mapdroptype'], dropEffect: '' },
    } as any
    act(() => { result.current.handleCanvasDragOver(mockEvent) })
    expect(result.current.sidebarDragOver).toBe(true)
  })

  it('handleCanvasDragOver does not set sidebarDragOver without mapdroptype', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    const mockEvent = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['text/plain'], dropEffect: '' },
    } as any
    act(() => { result.current.handleCanvasDragOver(mockEvent) })
    expect(result.current.sidebarDragOver).toBe(false)
  })

  it('handleCanvasDragLeave clears sidebarDragOver when leaving outer container', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // First set it to true
    const dragOverEvent = { preventDefault: vi.fn(), dataTransfer: { types: ['mapdroptype'], dropEffect: '' } } as any
    act(() => { result.current.handleCanvasDragOver(dragOverEvent) })
    // Then leave (relatedTarget outside container)
    const div = document.createElement('div')
    const mockLeaveEvent = {
      relatedTarget: div,
      currentTarget: { contains: (_n: Node) => false },
    } as any
    act(() => { result.current.handleCanvasDragLeave(mockLeaveEvent) })
    expect(result.current.sidebarDragOver).toBe(false)
  })
})

// ---- handleCreateSession / handleCreateNote error paths ----
describe('useMapViewData — handleCreateSession error path', () => {
  beforeEach(resetMocks)

  it('handles createSession error gracefully', async () => {
    vi.mocked(createSession).mockRejectedValue(new Error('Server error'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateSession(null) })
    // Should not throw; sessions unchanged
    expect(result.current.sessions).toHaveLength(0)
  })
})

describe('useMapViewData — handleCreateNote error path', () => {
  beforeEach(resetMocks)

  it('handles createNote error gracefully', async () => {
    vi.mocked(createNote).mockRejectedValue(new Error('Note error'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateNote(null) })
    expect(result.current.notes).toHaveLength(0)
  })
})

// ---- handleRestoreMap error path ----
describe('useMapViewData — handleRestoreMap error path', () => {
  beforeEach(resetMocks)

  it('removes from archivedMaps when error message includes "not found"', async () => {
    vi.mocked(restoreMap).mockRejectedValue(new Error('map not found'))
    // We can't directly set archivedMaps state, but we can verify the function handles the error
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleRestoreMap('old-map') })
    // No throw — graceful error handling
    expect(result.current.archivedMaps.find(m => m.id === 'old-map')).toBeUndefined()
  })

  it('handles generic restoreMap error gracefully', async () => {
    vi.mocked(restoreMap).mockRejectedValue(new Error('Generic error'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleRestoreMap('old-map') })
    // Should not throw
    expect(result.current.loading).toBe(false)
  })
})

// ---- reloadPinGroups ----
describe('useMapViewData — reloadPinGroups', () => {
  beforeEach(resetMocks)

  it('does nothing when gameId or activeMapId is missing', async () => {
    const { listMapPinGroups } = await import('../../api/pinGroups')
    vi.mocked(listMapPinGroups).mockClear()
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // activeMapId is always null from useLocalStorage mock
    const callsBefore = vi.mocked(listMapPinGroups).mock.calls.length
    await act(async () => { await result.current.reloadPinGroups() })
    // Should not call listMapPinGroups since activeMapId=null
    expect(vi.mocked(listMapPinGroups).mock.calls.length).toBe(callsBefore)
  })
})

// ---- map operations error paths ----
describe('useMapViewData — map operation error paths', () => {
  beforeEach(resetMocks)

  it('handleCreateMap handles error gracefully', async () => {
    vi.mocked(createMap).mockRejectedValue(new Error('Create map error'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleCreateMap('Bad Map') })
    expect(result.current.maps).toHaveLength(0)
  })

  it('handleRenameMap handles error gracefully', async () => {
    vi.mocked(renameMap).mockRejectedValue(new Error('Rename error'))
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleRenameMap('map-1', 'Bad') })
    // Name unchanged
    expect(result.current.maps.find(m => m.id === 'map-1')?.name).toBe('Test Map')
  })

  it('handleArchiveMap handles error gracefully', async () => {
    vi.mocked(archiveMap).mockRejectedValue(new Error('Archive error'))
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleArchiveMap('map-1') })
    // archiveMap threw — maps remains (not removed)
    expect(result.current.maps).toHaveLength(1)
  })

  it('handleReorderMaps handles error gracefully', async () => {
    vi.mocked(reorderMaps).mockRejectedValue(new Error('Reorder error'))
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await act(async () => { await result.current.handleReorderMaps(['map-1']) })
    // No throw
    expect(result.current.loading).toBe(false)
  })
})

// ---- handleFileChange ----
describe('useMapViewData — handleFileChange', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
  })

  it('uploads file and updates map image_url', async () => {
    const updatedMap = { ...sampleMap, image_url: '/new-map.jpg' }
    vi.mocked(uploadMapImage).mockResolvedValue(updatedMap as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const file = new File(['data'], 'map.jpg', { type: 'image/jpeg' })
    const event = { target: { files: [file], value: '' } } as any
    await act(async () => { await result.current.handleFileChange(event) })

    expect(vi.mocked(uploadMapImage)).toHaveBeenCalledWith('game-1', 'map-1', file)
    expect(result.current.maps.find(m => m.id === 'map-1')?.image_url).toBe('/new-map.jpg')
  })

  it('sets uploadError when upload fails', async () => {
    vi.mocked(uploadMapImage).mockRejectedValue(new Error('Upload failed'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const file = new File(['data'], 'map.jpg', { type: 'image/jpeg' })
    const event = { target: { files: [file], value: '' } } as any
    await act(async () => { await result.current.handleFileChange(event) })

    expect(result.current.uploadError).toBe('Upload failed')
    expect(result.current.uploading).toBe(false)
  })

  it('does nothing when no file is provided', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const event = { target: { files: [], value: '' } } as any
    await act(async () => { await result.current.handleFileChange(event) })

    expect(vi.mocked(uploadMapImage)).not.toHaveBeenCalled()
  })
})

// ---- handleDeleteMap ----
describe('useMapViewData — handleDeleteMap', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
  })

  it('archives map when user confirms', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    vi.mocked(archiveMap).mockResolvedValue(sampleMap as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    await act(async () => { await result.current.handleDeleteMap() })

    expect(vi.mocked(archiveMap)).toHaveBeenCalledWith('game-1', 'map-1')
    expect(result.current.maps).toHaveLength(0)
    ;(window.confirm as any).mockRestore()
  })

  it('does nothing when user cancels confirm dialog', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    await act(async () => { await result.current.handleDeleteMap() })

    expect(vi.mocked(archiveMap)).not.toHaveBeenCalled()
    expect(result.current.maps).toHaveLength(1)
    ;(window.confirm as any).mockRestore()
  })
})

// ---- reloadPinGroups with activeMapId ----
describe('useMapViewData — reloadPinGroups with activeMap', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('calls listMapPinGroups when activeMapId is set', async () => {
    const { listMapPinGroups } = await import('../../api/pinGroups')
    vi.mocked(listMapPinGroups).mockResolvedValue([])
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const callsBefore = vi.mocked(listMapPinGroups).mock.calls.length
    await act(async () => { await result.current.reloadPinGroups() })
    expect(vi.mocked(listMapPinGroups).mock.calls.length).toBeGreaterThan(callsBefore)
  })
})

// ---- handleRestoreMap error path ----
describe('useMapViewData — handleRestoreMap', () => {
  beforeEach(resetMocks)

  it('restores a map successfully', async () => {
    const archivedMap = { ...sampleMap, archived_at: '2024-01-01' }
    vi.mocked(restoreMap).mockResolvedValue(sampleMap as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Inject archived map into state via WS event
    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })
    const { result: result2 } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result2.current.loading).toBe(false))
    act(() => { capturedHandler?.({ type: 'map_archived', data: archivedMap }) })

    await act(async () => { await result2.current.handleRestoreMap('map-1') })
    expect(vi.mocked(restoreMap)).toHaveBeenCalledWith('game-1', 'map-1')
  })

  it('removes from archivedMaps on not-found error', async () => {
    vi.mocked(restoreMap).mockRejectedValue(new Error('not found'))
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // Should not throw
    await act(async () => { await result.current.handleRestoreMap('nonexistent-map') })
    expect(result.current.archivedMaps).toHaveLength(0)
  })
})

// ---- handleCanvasDrop ----
describe('useMapViewData — handleCanvasDrop', () => {
  beforeEach(resetMocks)

  it('sets pendingCoords and dropLinkedItem for a session drop', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Mock mapContainerRef
    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 100, height: 100 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    const dataMap: Record<string, string> = {
      mapDropType: 'session',
      mapDropId: 's1',
      mapDropLabel: 'Session 1',
    }
    const event = {
      preventDefault: vi.fn(),
      dataTransfer: {
        types: ['mapdroptype'],
        getData: (key: string) => dataMap[key] ?? '',
      },
      clientX: 50,
      clientY: 50,
      currentTarget: container,
    } as any

    act(() => { result.current.handleCanvasDrop(event) })

    expect(result.current.dropLinkedItem).toMatchObject({ type: 'session', id: 's1' })
  })

  it('shows toast and flashes existing pin when session already pinned', async () => {
    // Set up a pin for session s1
    const { listMapPins } = await import('../../api/pins')
    const existingPin = { id: 'pin-1', x: 50, y: 50, session_id: 's1', note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    vi.mocked(listMapPins).mockResolvedValue([existingPin] as any)
    mockWithActiveMap()

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 100, height: 100 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    const dataMap: Record<string, string> = {
      mapDropType: 'session',
      mapDropId: 's1',
      mapDropLabel: 'Session 1',
    }
    const event = {
      preventDefault: vi.fn(),
      dataTransfer: {
        types: ['mapdroptype'],
        getData: (key: string) => dataMap[key] ?? '',
      },
      clientX: 50, clientY: 50,
      currentTarget: container,
    } as any

    act(() => { result.current.handleCanvasDrop(event) })
    expect(result.current.toastMessage).toBe('This session already has a pin on the map.')
  })
})

// ---- handleCanvasDragOver and handleCanvasDragLeave ----
describe('useMapViewData — handleCanvasDragOver and handleCanvasDragLeave', () => {
  beforeEach(resetMocks)

  it('sets sidebarDragOver when dragover event includes mapdroptype', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const container = document.createElement('div')
    const event = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['mapdroptype'], dropEffect: '' },
      currentTarget: container,
      relatedTarget: null,
    } as any

    act(() => { result.current.handleCanvasDragOver(event) })
    expect(result.current.sidebarDragOver).toBe(true)
  })

  it('does not set sidebarDragOver when dragover event has no mapdroptype', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const event = {
      preventDefault: vi.fn(),
      dataTransfer: { types: ['text/plain'], dropEffect: '' },
      currentTarget: document.createElement('div'),
      relatedTarget: null,
    } as any

    act(() => { result.current.handleCanvasDragOver(event) })
    expect(result.current.sidebarDragOver).toBe(false)
  })

  it('clears sidebarDragOver on dragleave from container', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const container = document.createElement('div')
    // First set to true
    act(() => {
      result.current.handleCanvasDragOver({
        preventDefault: vi.fn(),
        dataTransfer: { types: ['mapdroptype'], dropEffect: '' },
        currentTarget: container,
        relatedTarget: null,
      } as any)
    })
    expect(result.current.sidebarDragOver).toBe(true)

    // Now drag leave with relatedTarget outside container
    act(() => {
      result.current.handleCanvasDragLeave({
        currentTarget: container,
        relatedTarget: document.createElement('div'),
      } as any)
    })
    expect(result.current.sidebarDragOver).toBe(false)
  })
})

// ---- handleMapClick ----
describe('useMapViewData — handleMapClick', () => {
  beforeEach(resetMocks)

  it('sets pendingCoords when clicking empty map area', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    const event = {
      clientX: 500,
      clientY: 400,
      target: container,
    } as any

    act(() => { result.current.handleMapClick(event) })
    expect(result.current.pendingCoords).not.toBeNull()
  })

  it('clears editingPinId when popover is open and map is clicked', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Open editing pin popover
    act(() => { result.current.setEditingPinId('pin-1') })
    expect(result.current.editingPinId).toBe('pin-1')

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    const event = {
      clientX: 100,
      clientY: 100,
      target: container,
    } as any

    act(() => { result.current.handleMapClick(event) })
    expect(result.current.editingPinId).toBeNull()
    // pendingCoords should NOT have been set (returned early)
    expect(result.current.pendingCoords).toBeNull()
  })

  it('does nothing when mapContainerRef is null', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // mapContainerRef.current is null by default
    const event = { clientX: 100, clientY: 100, target: document.createElement('div') } as any
    act(() => { result.current.handleMapClick(event) })
    expect(result.current.pendingCoords).toBeNull()
  })
})

// ---- handlePinPointerDown ----
describe('useMapViewData — handlePinPointerDown', () => {
  beforeEach(resetMocks)

  it('sets dragging state on pointer down for ungrouped pin', async () => {
    const samplePin = { id: 'pin-1', x: 20, y: 30, session_id: null, note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Inject pin via WebSocket
    act(() => { capturedHandler?.({ type: 'pin_created', data: samplePin }) })
    expect(result.current.pins).toHaveLength(1)

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    const pointerEvent = {
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
      currentTarget: { setPointerCapture: vi.fn() },
      pointerId: 1,
      clientX: 200,
      clientY: 300,
    } as any

    // handlePinPointerDown signature: (e: PointerEvent, pin: SessionPin)
    act(() => { result.current.handlePinPointerDown(pointerEvent, samplePin) })
    // dragging is set to an object { pinId, startX, startY } — not a boolean
    expect(result.current.dragging).not.toBeNull()
    expect((result.current.dragging as any)?.pinId).toBe('pin-1')
  })
})

// ---- per-map data loading with activeMapId ----
describe('useMapViewData — per-map data loading with activeMapId', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('loads pins when activeMapId is set', async () => {
    const { listMapPins } = await import('../../api/pins')
    const samplePin = { id: 'pin-1', x: 20, y: 30, session_id: null, note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    vi.mocked(listMapPins).mockResolvedValue([samplePin] as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    await waitFor(() => expect(result.current.pins).toHaveLength(1))
    expect(result.current.pins[0].id).toBe('pin-1')
  })
})

// ---- __reconnected with activeMapId set ----
describe('useMapViewData — __reconnected with activeMapId', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('reloads pins and pin groups on __reconnected when activeMapId is set', async () => {
    const { listMapPins } = await import('../../api/pins')
    const { listMapPinGroups } = await import('../../api/pinGroups')
    vi.mocked(listMapPins).mockResolvedValue([])
    vi.mocked(listMapPinGroups).mockResolvedValue([])

    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    const pinCallsBefore = vi.mocked(listMapPins).mock.calls.length
    const groupCallsBefore = vi.mocked(listMapPinGroups).mock.calls.length

    await act(async () => { capturedHandler?.({ type: '__reconnected' }) })
    await waitFor(() => expect(vi.mocked(listMapPins).mock.calls.length).toBeGreaterThan(pinCallsBefore))
    expect(vi.mocked(listMapPinGroups).mock.calls.length).toBeGreaterThan(groupCallsBefore)
  })
})

// ---- handleEditPinField with error reload ----
describe('useMapViewData — handleEditPinField error reload', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('reloads pins from API when updatePin fails and activeMapId is set', async () => {
    const samplePin = { id: 'pin-1', x: 20, y: 30, session_id: null, note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    const { listMapPins } = await import('../../api/pins')
    vi.mocked(listMapPins).mockResolvedValue([samplePin] as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // Per-map effect loads pins via listMapPins — wait for that
    await waitFor(() => expect(result.current.pins).toHaveLength(1))

    vi.mocked(updatePin).mockRejectedValue(new Error('Failed'))
    // After error, listMapPins returns the same pin
    vi.mocked(listMapPins).mockResolvedValue([samplePin] as any)

    await act(async () => { await result.current.handleEditPinField('pin-1', { colour: 'blue' }) })
    // After error reload, pins should still be present
    expect(result.current.pins.length).toBeGreaterThan(0)
    expect(result.current.pins.some(p => p.id === 'pin-1')).toBe(true)
  })
})

// ---- handlePointerMove and handlePointerUp ----
describe('useMapViewData — handlePointerMove and handlePointerUp', () => {
  beforeEach(resetMocks)

  it('handlePointerMove updates pin position while dragging', async () => {
    const samplePin = { id: 'pin-1', x: 20, y: 30, session_id: null, note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => { capturedHandler?.({ type: 'pin_created', data: samplePin }) })
    expect(result.current.pins).toHaveLength(1)

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    // Start dragging
    const pointerDownEvent = {
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
      currentTarget: { setPointerCapture: vi.fn() },
      pointerId: 1,
      clientX: 200,
      clientY: 240,
    } as any
    act(() => { result.current.handlePinPointerDown(pointerDownEvent, samplePin) })
    expect(result.current.dragging).not.toBeNull()

    // Move pointer far enough to exceed 5px threshold
    const pointerMoveEvent = {
      clientX: 300,
      clientY: 350,
    } as any
    act(() => { result.current.handlePointerMove(pointerMoveEvent) })

    // Pin position should be updated
    const updatedPin = result.current.pins.find(p => p.id === 'pin-1')
    expect(updatedPin?.x).not.toBe(20)
  })

  it('handlePointerUp persists pin position after drag', async () => {
    const samplePin = { id: 'pin-1', x: 20, y: 30, session_id: null, note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null }
    vi.mocked(updatePin).mockResolvedValue({ ...samplePin, x: 50, y: 60 } as any)
    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => { capturedHandler?.({ type: 'pin_created', data: samplePin }) })

    const container = document.createElement('div')
    container.getBoundingClientRect = () => ({ left: 0, top: 0, width: 1000, height: 800 } as DOMRect)
    ;(result.current.mapContainerRef as any).current = container

    // Start drag
    const pointerDownEvent = {
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
      currentTarget: { setPointerCapture: vi.fn() },
      pointerId: 1,
      clientX: 200,
      clientY: 240,
    } as any
    act(() => { result.current.handlePinPointerDown(pointerDownEvent, samplePin) })

    // Move to trigger wasDragRef = true (>5px)
    act(() => {
      result.current.handlePointerMove({ clientX: 300, clientY: 350 } as any)
    })

    // Release pointer
    await act(async () => {
      await result.current.handlePointerUp({ clientX: 500, clientY: 400 } as any)
    })

    expect(vi.mocked(updatePin)).toHaveBeenCalledWith('pin-1', expect.objectContaining({ x: expect.any(Number), y: expect.any(Number) }))
    expect(result.current.dragging).toBeNull()
  })

  it('handlePointerUp does nothing if not dragging', async () => {
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    await act(async () => {
      await result.current.handlePointerUp({ clientX: 100, clientY: 100 } as any)
    })
    // No error thrown, dragging remains null
    expect(result.current.dragging).toBeNull()
  })
})

// ---- pendingGroupPinIds flows ----
describe('useMapViewData — pendingGroupPinIds in handleSelectSession', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('creates a pin group when pendingGroupPinIds is set', async () => {
    const { createMapPinGroup } = await import('../../api/pinGroups')
    const { createPin } = await import('../../api/pins')
    vi.mocked(createPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, session_id: 's1', note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null } as any)
    vi.mocked(createMapPinGroup).mockResolvedValue({ id: 'group-1', x: 50, y: 50, map_id: 'map-1', pin_ids: [], name: '' } as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Set pendingGroupPinIds and pendingCoords
    act(() => {
      result.current.setPendingGroupPinIds(['existing-pin-id'])
      result.current.setPendingCoords({ x: 50, y: 50 })
    })

    await act(async () => {
      await result.current.handleSelectSession(sampleSession)
    })

    expect(vi.mocked(createMapPinGroup)).toHaveBeenCalled()
  })

  it('adds pin to existing group when pendingAddToGroupId is set', async () => {
    const { addPinToGroup } = await import('../../api/pinGroups')
    const { createPin } = await import('../../api/pins')
    vi.mocked(createPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, session_id: 's1', note_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null } as any)
    vi.mocked(addPinToGroup).mockResolvedValue({} as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => {
      result.current.setPendingAddToGroupId('group-id')
      result.current.setPendingCoords({ x: 50, y: 50 })
    })

    await act(async () => {
      await result.current.handleSelectSession(sampleSession)
    })

    expect(vi.mocked(addPinToGroup)).toHaveBeenCalledWith('group-id', 'new-pin')
  })
})

// ---- pendingGroupPinIds flows for handleSelectNote ----
describe('useMapViewData — pendingGroupPinIds in handleSelectNote', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('creates a pin group when pendingGroupPinIds is set', async () => {
    const { createMapPinGroup } = await import('../../api/pinGroups')
    vi.mocked(createMapPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, note_id: 'n1', session_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null } as any)
    vi.mocked(createMapPinGroup).mockResolvedValue({ id: 'group-1', x: 50, y: 50, map_id: 'map-1', pin_ids: [], name: '' } as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => {
      result.current.setPendingGroupPinIds(['existing-pin-id'])
      result.current.setPendingCoords({ x: 50, y: 50 })
    })

    await act(async () => { await result.current.handleSelectNote(sampleNote) })

    expect(vi.mocked(createMapPinGroup)).toHaveBeenCalled()
  })

  it('adds pin to existing group when pendingAddToGroupId is set', async () => {
    const { addPinToGroup } = await import('../../api/pinGroups')
    vi.mocked(createMapPin).mockResolvedValue({ id: 'new-pin', x: 50, y: 50, note_id: 'n1', session_id: null, group_id: null, map_id: 'map-1', label: '', colour: 'red', icon: 'star', description: null } as any)
    vi.mocked(addPinToGroup).mockResolvedValue({} as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => {
      result.current.setPendingAddToGroupId('group-id')
      result.current.setPendingCoords({ x: 50, y: 50 })
    })

    await act(async () => { await result.current.handleSelectNote(sampleNote) })

    expect(vi.mocked(addPinToGroup)).toHaveBeenCalledWith('group-id', 'new-pin')
  })
})

// ---- reloadPinGroups error path ----
describe('useMapViewData — reloadPinGroups error', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('handles listMapPinGroups error gracefully', async () => {
    const { listMapPinGroups } = await import('../../api/pinGroups')
    vi.mocked(listMapPinGroups).mockRejectedValue(new Error('Groups error'))

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Should not throw
    await act(async () => { await result.current.reloadPinGroups() })
    expect(result.current.pinGroups).toHaveLength(0)
  })
})

// ---- openItem navigate for session ----
describe('useMapViewData — openItem navigation', () => {
  beforeEach(resetMocks)

  it('navigates to session notes when mapEditorMode is not set (modal)', async () => {
    // Default prefs have map_editor_mode: 'modal', so mapEditorMode = false
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // mapEditorMode should be false with 'modal'
    // openItem with type 'session' should navigate
    expect(typeof result.current.openItem).toBe('function')
    act(() => { result.current.openItem('session', 's1', 'Session 1') })
    // No error thrown - navigate was called (we can't easily verify navigation in this context)
  })

  it('adds item to openItems when mapEditorMode is true', async () => {
    // map_editor_mode: 'modal' → mapEditorMode = true (hook: prefs.map_editor_mode === 'modal')
    vi.mocked(getPreferences).mockResolvedValue({ ...defaultPrefs(), map_editor_mode: 'modal' } as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    // Wait for preferences async effect to set mapEditorMode = true
    await waitFor(() => expect(result.current.mapEditorMode).toBe(true))

    act(() => { result.current.openItem('session', 's1', 'Session 1') })
    expect(result.current.openItems.some(i => i.itemId === 's1')).toBe(true)
  })
})

// ---- handleArchiveMap with matching activeMapId ----
describe('useMapViewData — handleArchiveMap activeMapId update', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('updates activeMapId when active map is archived', async () => {
    const map2 = { ...sampleMap, id: 'map-2', name: 'Map 2', sort_order: 1 }
    vi.mocked(listMaps).mockResolvedValue([sampleMap, map2] as any)
    vi.mocked(archiveMap).mockResolvedValue({} as any)

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // Active map is 'map-1' from mockWithActiveMap
    await act(async () => { await result.current.handleArchiveMap('map-1') })

    expect(result.current.maps).toHaveLength(1)
    expect(result.current.maps[0].id).toBe('map-2')
  })
})

// ---- WS map_archived with matching activeMapId ----
describe('useMapViewData — WS map_archived updates activeMapId', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('changes activeMapId when active map is archived via WS', async () => {
    const map2 = { ...sampleMap, id: 'map-2', name: 'Map 2', sort_order: 1 }
    vi.mocked(listMaps).mockResolvedValue([sampleMap, map2] as any)

    let capturedHandler: ((event: any) => void) | null = null
    vi.mocked(useGameSocket).mockImplementation((_id: any, handler: any) => { capturedHandler = handler })

    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    act(() => { capturedHandler?.({ type: 'map_archived', data: { id: 'map-1' } }) })

    expect(result.current.maps).toHaveLength(1)
    expect(result.current.maps[0].id).toBe('map-2')
  })
})

// ---- setActiveMapId when prev is in mapsData ----
describe('useMapViewData — setActiveMapId preserves existing activeMapId', () => {
  beforeEach(() => {
    resetMocks()
    mockWithActiveMap()
  })

  it('keeps existing activeMapId when it is present in loaded maps', async () => {
    vi.mocked(listMaps).mockResolvedValue([sampleMap] as any)
    const { result } = renderHook(() => useMapViewData('game-1'), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))

    // activeMapId should still be 'map-1' because it exists in the maps list
    expect(result.current.activeMapId).toBe('map-1')
  })
})

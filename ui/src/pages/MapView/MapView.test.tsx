import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import MapView from './MapView'

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useParams: () => ({ gameId: 'game-1' }),
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

// Mock the entire useMapViewData hook to control loading/error/data states
const mockUseMapViewData = vi.fn()

vi.mock('./useMapViewData', () => ({
  useMapViewData: (...args: unknown[]) => mockUseMapViewData(...args),
  GROUP_PROXIMITY_PCT: 1.5,
  DEFAULT_VIEW_STATE: { scale: 1, positionX: 0, positionY: 0 },
}))

vi.mock('../../api/client', () => ({
  BASE_URL: 'http://localhost:8080',
  apiFetch: vi.fn().mockResolvedValue({}),
}))

vi.mock('@tiptap/react', () => ({
  useEditor: () => null,
  EditorContent: () => <div data-testid="editor" />,
}))

vi.mock('motion/react', () => ({
  motion: {
    div: 'div',
    ul: 'ul',
    li: 'li',
    span: 'span',
    p: 'p',
    button: 'button',
    aside: 'aside',
    nav: 'nav',
    section: 'section',
  },
  AnimatePresence: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// Smart mock — exposes buttons to trigger all inline callbacks from MapView
vi.mock('./MapCanvas', () => ({
  default: (props: any) => (
    <div data-testid="map-canvas">
      <button data-testid="cb-zoom" onClick={() => props.onZoom?.({ state: { scale: 2 } })}>Zoom</button>
      <button data-testid="cb-edit-pin" onClick={() => props.onEditPin?.('pin-1')}>EditPin</button>
      <button data-testid="cb-pin-click-note" onClick={() => props.onPinClick?.({ id: 'p1', note_id: 'n1', session_id: null, label: 'Note' })}>PinClickNote</button>
      <button data-testid="cb-pin-click-session" onClick={() => props.onPinClick?.({ id: 'p2', note_id: null, session_id: 's1', label: 'Session' })}>PinClickSession</button>
      <button data-testid="cb-pin-click-marker" onClick={() => props.onPinClick?.({ id: 'p3', note_id: null, session_id: null, label: 'Marker' })}>PinClickMarker</button>
      <button data-testid="cb-group-click-same" onClick={() => props.onGroupClick?.('g1')}>GroupClickSame</button>
      <button data-testid="cb-group-click-other" onClick={() => props.onGroupClick?.('g2')}>GroupClickOther</button>
      <button data-testid="cb-manage-group" onClick={() => props.onManageGroup?.('g1')}>ManageGroup</button>
      <button data-testid="cb-error-dismiss" onClick={() => props.onPinErrorDismiss?.()}>ErrorDismiss</button>
    </div>
  ),
}))

vi.mock('./MapOverlayPanel', () => ({
  default: (props: any) => (
    <div data-testid="map-overlay-panel">
      <button data-testid="cb-panel-close" onClick={() => props.onClose?.()}>PanelClose</button>
      <button data-testid="cb-panel-open" onClick={() => props.onOpen?.()}>PanelOpen</button>
    </div>
  ),
}))

// PinPickerModal: conditionally rendered when pendingCoords is set.
// Exposes trigger buttons for all its inline callbacks defined in MapView.tsx
vi.mock('./PinPickerModal', () => ({
  default: (props: any) => (
    <div data-testid="pin-picker-modal">
      <button data-testid="cb-picker-close" onClick={() => props.onClose?.()}>PickerClose</button>
    </div>
  ),
}))

// PinGroupModals: always rendered (unconditional).
// Exposes trigger buttons for all its inline callbacks defined in MapView.tsx
vi.mock('./PinGroupModals', () => ({
  default: (props: any) => (
    <div data-testid="pin-group-modals">
      <button data-testid="cb-dismiss-grouping" onClick={() => props.onDismissGroupingPrompt?.()}>DismissGrouping</button>
      <button data-testid="cb-place-standalone" onClick={() => props.onPlaceStandalone?.({ x: 10, y: 20 })}>PlaceStandalone</button>
      <button data-testid="cb-create-group-from-prompt" onClick={() => props.onCreateGroupFromPrompt?.({ x: 10, y: 20 }, ['p1', 'p2'])}>CreateGroup</button>
      <button data-testid="cb-add-to-group-from-prompt" onClick={() => props.onAddToGroupFromPrompt?.({ x: 10, y: 20 }, 'g1')}>AddToGroup</button>
      <button data-testid="cb-dismiss-drag-group" onClick={() => props.onDismissDragGroupPrompt?.()}>DismissDragGroup</button>
      <button data-testid="cb-dismiss-manage-group" onClick={() => props.onDismissManageGroup?.()}>DismissManageGroup</button>
      <button data-testid="cb-update-pins" onClick={() => props.onUpdatePins?.((prev: any) => prev)}>UpdatePins</button>
    </div>
  ),
}))

// FolderSidebar mock exposes buttons for session/note item clicks
vi.mock('../../components/FolderSidebar/FolderSidebar', () => ({
  default: (props: any) => (
    <div data-testid="folder-sidebar">
      <button data-testid="cb-session-click" onClick={() => props.onSessionClick?.('s1')}>SessionClick</button>
      <button data-testid="cb-note-click" onClick={() => props.onNoteClick?.('n1')}>NoteClick</button>
    </div>
  ),
}))

// EditorModalManager mock exposes close/closeAll buttons
vi.mock('../../components/EditorModalManager/EditorModalManager', () => ({
  default: (props: any) => (
    <div data-testid="editor-modal-manager">
      <button data-testid="cb-editor-close" onClick={() => props.onClose?.('n1')}>EditorClose</button>
      <button data-testid="cb-editor-close-all" onClick={() => props.onCloseAll?.()}>EditorCloseAll</button>
    </div>
  ),
}))

// ─── fixtures ─────────────────────────────────────────────────────────────────

const mapNoImage = {
  id: 'map-1', name: 'World Map', image_url: '',
  sort_order: 0, archived_at: null, game_id: 'game-1', created_at: '', updated_at: '',
}

const mapWithImage = {
  id: 'map-1', name: 'World Map', image_url: '/uploads/map.jpg',
  sort_order: 0, archived_at: null, game_id: 'game-1', created_at: '', updated_at: '',
}

const baseMapViewData = {
  game: null,
  sessions: [],
  pins: [],
  notes: [],
  maps: [],
  loading: false,
  error: null,
  isGM: false,
  unpinnedSessions: [],
  activeMapId: null,
  viewState: { scale: 1, positionX: 0, positionY: 0 },
  displayScale: 1,
  setDisplayScale: vi.fn(),
  pendingCoords: null,
  setPendingCoords: vi.fn(),
  pendingLabel: '',
  setPendingLabel: vi.fn(),
  pendingDescription: '',
  setPendingDescription: vi.fn(),
  pendingColour: 'red',
  setPendingColour: vi.fn(),
  pendingIcon: 'star',
  setPendingIcon: vi.fn(),
  pinError: null,
  setPinError: vi.fn(),
  editingPinId: null,
  setEditingPinId: vi.fn(),
  pickerSearch: '',
  setPickerSearch: vi.fn(),
  editLinkSearch: '',
  setEditLinkSearch: vi.fn(),
  hoveredPinId: null,
  setHoveredPinId: vi.fn(),
  dragging: null,
  dropTargetIds: [],
  pinGroups: [],
  activeGroupId: null,
  setActiveGroupId: vi.fn(),
  managingGroupId: null,
  setManagingGroupId: vi.fn(),
  groupingPrompt: null,
  setGroupingPrompt: vi.fn(),
  setPendingGroupPinIds: vi.fn(),
  setPendingAddToGroupId: vi.fn(),
  dragGroupPrompt: null,
  setDragGroupPrompt: vi.fn(),
  panelOpen: false,
  setPanelOpen: vi.fn(),
  sidebarOpen: false,
  uploading: false,
  uploadError: null,
  openItems: [],
  setOpenItems: vi.fn(),
  setPins: vi.fn(),
  mapContainerRef: { current: null },
  viewportContainerRef: { current: null },
  fileInputRef: { current: null },
  transformRef: { current: null },
  wasDragRef: { current: false },
  handleImageLoad: vi.fn(),
  handleTransformed: vi.fn(),
  handleTransformEnd: vi.fn(),
  handleMapClick: vi.fn(),
  handlePointerMove: vi.fn(),
  handlePointerUp: vi.fn(),
  handlePinPointerDown: vi.fn(),
  handleSelectSession: vi.fn(),
  handleSelectNote: vi.fn(),
  handleCreateMarker: vi.fn(),
  handleDeletePin: vi.fn(),
  handleEditPinField: vi.fn(),
  handleUploadClick: vi.fn(),
  handleFileChange: vi.fn(),
  handleDeleteMap: vi.fn(),
  handleSidebarToggle: vi.fn(),
  handleSessionUpdate: vi.fn(),
  handleNoteUpdate: vi.fn(),
  handleCreateSession: vi.fn(),
  handleCreateNote: vi.fn(),
  openItem: vi.fn(),
  sessionForPin: vi.fn().mockReturnValue(null),
  noteForPin: vi.fn().mockReturnValue(null),
  reloadPinGroups: vi.fn(),
  handleCreateMap: vi.fn(),
  sidebarDragOver: false,
  toastMessage: null,
  dropLinkedItem: null,
  setDropLinkedItem: vi.fn(),
  flashPinId: null,
  handleCanvasDragOver: vi.fn(),
  handleCanvasDragLeave: vi.fn(),
  handleCanvasDrop: vi.fn(),
}

// Base state with a map that has an image (activates the main viewport branch)
const withImageMap = {
  ...baseMapViewData,
  maps: [mapWithImage],
  activeMapId: 'map-1',
  isGM: true,
}

function renderMapView() {
  return render(
    <MemoryRouter>
      <MapView />
    </MemoryRouter>,
  )
}

// ─── tests ────────────────────────────────────────────────────────────────────

describe('MapView', () => {
  beforeEach(() => {
    mockUseMapViewData.mockReset()
  })

  // ── original tests ──────────────────────────────────────────────────────────

  it('should show loading spinner when loading', () => {
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, loading: true })
    renderMapView()
    expect(screen.getByText(/unrolling the map/i)).toBeInTheDocument()
  })

  it('should show error message on error', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData,
      loading: false,
      error: 'Failed to load map data',
    })
    renderMapView()
    expect(screen.getByText('Failed to load map data')).toBeInTheDocument()
  })

  it('should show empty state when no maps and not GM', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData,
      loading: false,
      error: null,
      maps: [],
      isGM: false,
    })
    renderMapView()
    expect(screen.getByText('The Map Awaits')).toBeInTheDocument()
  })

  it('should render map canvas when maps exist', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData,
      loading: false,
      error: null,
      maps: [mapWithImage],
      activeMapId: 'map-1',
    })
    renderMapView()
    expect(screen.getByTestId('map-canvas')).toBeInTheDocument()
  })

  it('should call useMapViewData with the gameId param', () => {
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, loading: true })
    renderMapView()
    expect(mockUseMapViewData).toHaveBeenCalledWith('game-1')
  })

  // ── GM / no-maps state ─────────────────────────────────────────────────────

  it('shows "No Maps Yet" when GM has no maps', () => {
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, maps: [], isGM: true })
    renderMapView()
    expect(screen.getByText('No Maps Yet')).toBeInTheDocument()
    expect(screen.getByText(/Name your first map/i)).toBeInTheDocument()
  })

  it('shows create-map input when GM has no maps', () => {
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, maps: [], isGM: true })
    renderMapView()
    expect(screen.getByPlaceholderText(/Otari Region/i)).toBeInTheDocument()
  })

  it('calls handleCreateMap when GM submits the first-map form', () => {
    const handleCreateMap = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, maps: [], isGM: true, handleCreateMap })
    renderMapView()
    const input = screen.getByPlaceholderText(/Otari Region/i)
    fireEvent.change(input, { target: { value: 'The Keep' } })
    fireEvent.submit(input.closest('form')!)
    expect(handleCreateMap).toHaveBeenCalledWith('The Keep')
  })

  it('does not call handleCreateMap when form submitted with empty value', () => {
    const handleCreateMap = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, maps: [], isGM: true, handleCreateMap })
    renderMapView()
    const input = screen.getByPlaceholderText(/Otari Region/i)
    fireEvent.submit(input.closest('form')!)
    expect(handleCreateMap).not.toHaveBeenCalled()
  })

  // ── GM with active map but no image ───────────────────────────────────────

  it('shows "No Map Image" when GM has a map without image', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true,
    })
    renderMapView()
    expect(screen.getByText('No Map Image')).toBeInTheDocument()
  })

  it('shows upload button when GM map has no image', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true,
    })
    renderMapView()
    expect(screen.getByText('+ Upload Map Image')).toBeInTheDocument()
  })

  it('shows "Uploading…" text while uploading', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true, uploading: true,
    })
    renderMapView()
    expect(screen.getByText('Uploading…')).toBeInTheDocument()
  })

  it('displays uploadError when present (GM, no image)', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true, uploadError: 'File too large',
    })
    renderMapView()
    expect(screen.getByText('File too large')).toBeInTheDocument()
  })

  it('calls handleUploadClick when upload button is clicked', () => {
    const handleUploadClick = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true, handleUploadClick,
    })
    renderMapView()
    fireEvent.click(screen.getByText('+ Upload Map Image'))
    expect(handleUploadClick).toHaveBeenCalled()
  })

  it('renders file input for image upload (GM, no image)', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: true,
    })
    const { container } = renderMapView()
    expect(container.querySelector('input[type="file"]')).toBeInTheDocument()
  })

  // ── non-GM with active map but no image ───────────────────────────────────

  it('shows "not yet uploaded an image" when non-GM and map has no image', () => {
    mockUseMapViewData.mockReturnValue({
      ...baseMapViewData, maps: [mapNoImage], activeMapId: 'map-1', isGM: false,
    })
    renderMapView()
    expect(screen.getByText(/not yet uploaded an image/i)).toBeInTheDocument()
  })

  // ── PinPickerModal conditional rendering ──────────────────────────────────

  it('renders PinPickerModal when pendingCoords is set', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, pendingCoords: { x: 10, y: 20 } })
    renderMapView()
    expect(screen.getByTestId('pin-picker-modal')).toBeInTheDocument()
  })

  it('does not render PinPickerModal when pendingCoords is null', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, pendingCoords: null })
    renderMapView()
    expect(screen.queryByTestId('pin-picker-modal')).not.toBeInTheDocument()
  })

  // ── PinGroupModals (always rendered) ──────────────────────────────────────

  it('renders PinGroupModals unconditionally', () => {
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData })
    renderMapView()
    expect(screen.getByTestId('pin-group-modals')).toBeInTheDocument()
  })

  // ── EditorModalManager conditional rendering ──────────────────────────────

  it('renders EditorModalManager when openItems has items', () => {
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      openItems: [{ type: 'note', itemId: 'n1', label: 'My Note' }],
    })
    renderMapView()
    expect(screen.getByTestId('editor-modal-manager')).toBeInTheDocument()
  })

  it('does not render EditorModalManager when openItems is empty', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, openItems: [] })
    renderMapView()
    expect(screen.queryByTestId('editor-modal-manager')).not.toBeInTheDocument()
  })

  // ── toast / drag ──────────────────────────────────────────────────────────

  it('shows toast message when toastMessage is set', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, toastMessage: 'Session already pinned' })
    renderMapView()
    expect(screen.getByText('Session already pinned')).toBeInTheDocument()
  })

  it('adds drop-active class to container when sidebarDragOver is true', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, sidebarDragOver: true })
    const { container } = renderMapView()
    expect(container.querySelector('.map-viewport-container--drop-active')).toBeInTheDocument()
  })

  it('does not have drop-active class when sidebarDragOver is false', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, sidebarDragOver: false })
    const { container } = renderMapView()
    expect(container.querySelector('.map-viewport-container--drop-active')).not.toBeInTheDocument()
  })

  // ── sidebar toggle ────────────────────────────────────────────────────────

  it('renders sidebar toggle button when map has image', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap })
    renderMapView()
    expect(screen.getByTitle(/show folders|hide folders/i)).toBeInTheDocument()
  })

  it('calls handleSidebarToggle when sidebar toggle button is clicked', () => {
    const handleSidebarToggle = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, handleSidebarToggle })
    renderMapView()
    fireEvent.click(screen.getByTitle(/show folders|hide folders/i))
    expect(handleSidebarToggle).toHaveBeenCalled()
  })

  it('shows "Hide folders" title when sidebarOpen=true', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, sidebarOpen: true })
    renderMapView()
    expect(screen.getByTitle('Hide folders')).toBeInTheDocument()
  })

  // ── FolderSidebar conditional rendering ───────────────────────────────────

  it('renders FolderSidebar when sidebarOpen is true', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, sidebarOpen: true })
    renderMapView()
    expect(screen.getByTestId('folder-sidebar')).toBeInTheDocument()
  })

  it('does not render FolderSidebar when sidebarOpen is false', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap, sidebarOpen: false })
    renderMapView()
    expect(screen.queryByTestId('folder-sidebar')).not.toBeInTheDocument()
  })

  it('renders MapOverlayPanel when map has image', () => {
    mockUseMapViewData.mockReturnValue({ ...withImageMap })
    renderMapView()
    expect(screen.getByTestId('map-overlay-panel')).toBeInTheDocument()
  })

  // ── inline callbacks from MapCanvas ───────────────────────────────────────

  it('onZoom callback calls setDisplayScale with new scale', () => {
    const setDisplayScale = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setDisplayScale })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-zoom'))
    expect(setDisplayScale).toHaveBeenCalledWith(2)
  })

  it('onEditPin callback calls setEditingPinId and clears editLinkSearch', () => {
    const setEditingPinId = vi.fn()
    const setEditLinkSearch = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setEditingPinId, setEditLinkSearch })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-edit-pin'))
    expect(setEditingPinId).toHaveBeenCalledWith('pin-1')
    expect(setEditLinkSearch).toHaveBeenCalledWith('')
  })

  it('onPinClick with note_id calls openItem("note", ...)', () => {
    const openItem = vi.fn()
    const noteForPin = vi.fn().mockReturnValue({ title: 'My Note' })
    mockUseMapViewData.mockReturnValue({ ...withImageMap, openItem, noteForPin })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-pin-click-note'))
    expect(openItem).toHaveBeenCalledWith('note', 'n1', 'My Note')
  })

  it('onPinClick with session_id calls openItem("session", ...)', () => {
    const openItem = vi.fn()
    const sessionForPin = vi.fn().mockReturnValue({ title: 'Session 1' })
    mockUseMapViewData.mockReturnValue({ ...withImageMap, openItem, sessionForPin })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-pin-click-session'))
    expect(openItem).toHaveBeenCalledWith('session', 's1', 'Session 1')
  })

  it('onPinClick with no links toggles editingPinId', () => {
    const setEditingPinId = vi.fn()
    const setEditLinkSearch = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      setEditingPinId,
      setEditLinkSearch,
      editingPinId: null,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-pin-click-marker'))
    expect(setEditingPinId).toHaveBeenCalledWith('p3')
  })

  it('onPinClick with no links and same pin id clears editingPinId', () => {
    const setEditingPinId = vi.fn()
    const setEditLinkSearch = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      setEditingPinId,
      setEditLinkSearch,
      editingPinId: 'p3', // same as the clicked pin
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-pin-click-marker'))
    expect(setEditingPinId).toHaveBeenCalledWith(null)
  })

  it('onGroupClick toggles activeGroupId', () => {
    const setActiveGroupId = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setActiveGroupId, activeGroupId: 'g1' })
    renderMapView()
    // clicking the same group id → sets to null (toggle off)
    fireEvent.click(screen.getByTestId('cb-group-click-same'))
    expect(setActiveGroupId).toHaveBeenCalledWith(null)
  })

  it('onGroupClick sets activeGroupId to new group id', () => {
    const setActiveGroupId = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setActiveGroupId, activeGroupId: 'g1' })
    renderMapView()
    // clicking a different group id → sets to g2
    fireEvent.click(screen.getByTestId('cb-group-click-other'))
    expect(setActiveGroupId).toHaveBeenCalledWith('g2')
  })

  it('onManageGroup calls setManagingGroupId and clears activeGroupId', () => {
    const setManagingGroupId = vi.fn()
    const setActiveGroupId = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setManagingGroupId, setActiveGroupId })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-manage-group'))
    expect(setManagingGroupId).toHaveBeenCalledWith('g1')
    expect(setActiveGroupId).toHaveBeenCalledWith(null)
  })

  it('onPinErrorDismiss clears pinError', () => {
    const setPinError = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setPinError })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-error-dismiss'))
    expect(setPinError).toHaveBeenCalledWith(null)
  })

  // ── inline callbacks from MapOverlayPanel ──────────────────────────────────

  it('onClose from MapOverlayPanel calls setPanelOpen(false)', () => {
    const setPanelOpen = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setPanelOpen })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-panel-close'))
    expect(setPanelOpen).toHaveBeenCalledWith(false)
  })

  it('onOpen from MapOverlayPanel calls setPanelOpen(true)', () => {
    const setPanelOpen = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...withImageMap, setPanelOpen })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-panel-open'))
    expect(setPanelOpen).toHaveBeenCalledWith(true)
  })

  // ── inline callbacks from FolderSidebar ───────────────────────────────────

  it('onSessionClick in FolderSidebar calls openItem("session", ...)', () => {
    const openItem = vi.fn()
    const session = { id: 's1', title: 'Session One' }
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      sidebarOpen: true,
      sessions: [session],
      openItem,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-session-click'))
    expect(openItem).toHaveBeenCalledWith('session', 's1', 'Session One')
  })

  it('onNoteClick in FolderSidebar calls openItem("note", ...)', () => {
    const openItem = vi.fn()
    const note = { id: 'n1', title: 'Note One' }
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      sidebarOpen: true,
      notes: [note],
      openItem,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-note-click'))
    expect(openItem).toHaveBeenCalledWith('note', 'n1', 'Note One')
  })

  // ── inline callbacks from EditorModalManager ───────────────────────────────

  it('onClose from EditorModalManager removes that item from openItems', () => {
    const setOpenItems = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      openItems: [{ type: 'note', itemId: 'n1', label: 'Note' }],
      setOpenItems,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-editor-close'))
    expect(setOpenItems).toHaveBeenCalled()
    // Verify the updater function filters out the closed item
    const updater = setOpenItems.mock.calls[0][0]
    const result = updater([{ type: 'note', itemId: 'n1', label: 'Note' }, { type: 'note', itemId: 'n2', label: 'Note2' }])
    expect(result).toEqual([{ type: 'note', itemId: 'n2', label: 'Note2' }])
  })

  it('onCloseAll from EditorModalManager clears all open items', () => {
    const setOpenItems = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      openItems: [{ type: 'note', itemId: 'n1', label: 'Note' }],
      setOpenItems,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-editor-close-all'))
    expect(setOpenItems).toHaveBeenCalledWith([])
  })

  // ── PinPickerModal inline onClose callback ────────────────────────────────

  it('PinPickerModal onClose clears pendingCoords, label, description, search, and dropLinkedItem', () => {
    const setPendingCoords = vi.fn()
    const setPendingLabel = vi.fn()
    const setPendingDescription = vi.fn()
    const setPickerSearch = vi.fn()
    const setDropLinkedItem = vi.fn()
    mockUseMapViewData.mockReturnValue({
      ...withImageMap,
      pendingCoords: { x: 10, y: 20 },
      setPendingCoords,
      setPendingLabel,
      setPendingDescription,
      setPickerSearch,
      setDropLinkedItem,
    })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-picker-close'))
    expect(setPendingCoords).toHaveBeenCalledWith(null)
    expect(setPendingLabel).toHaveBeenCalledWith('')
    expect(setPendingDescription).toHaveBeenCalledWith('')
    expect(setPickerSearch).toHaveBeenCalledWith('')
    expect(setDropLinkedItem).toHaveBeenCalledWith(null)
  })

  // ── PinGroupModals inline callbacks ───────────────────────────────────────

  it('onDismissGroupingPrompt clears groupingPrompt', () => {
    const setGroupingPrompt = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setGroupingPrompt })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-dismiss-grouping'))
    expect(setGroupingPrompt).toHaveBeenCalledWith(null)
  })

  it('onPlaceStandalone sets pendingCoords and clears groupingPrompt', () => {
    const setPendingCoords = vi.fn()
    const setGroupingPrompt = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setPendingCoords, setGroupingPrompt })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-place-standalone'))
    expect(setPendingCoords).toHaveBeenCalledWith({ x: 10, y: 20 })
    expect(setGroupingPrompt).toHaveBeenCalledWith(null)
  })

  it('onCreateGroupFromPrompt sets pendingCoords, pendingGroupPinIds, and clears groupingPrompt', () => {
    const setPendingCoords = vi.fn()
    const setPendingGroupPinIds = vi.fn()
    const setGroupingPrompt = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setPendingCoords, setPendingGroupPinIds, setGroupingPrompt })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-create-group-from-prompt'))
    expect(setPendingCoords).toHaveBeenCalledWith({ x: 10, y: 20 })
    expect(setPendingGroupPinIds).toHaveBeenCalledWith(['p1', 'p2'])
    expect(setGroupingPrompt).toHaveBeenCalledWith(null)
  })

  it('onAddToGroupFromPrompt sets pendingCoords, pendingAddToGroupId, and clears groupingPrompt', () => {
    const setPendingCoords = vi.fn()
    const setPendingAddToGroupId = vi.fn()
    const setGroupingPrompt = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setPendingCoords, setPendingAddToGroupId, setGroupingPrompt })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-add-to-group-from-prompt'))
    expect(setPendingCoords).toHaveBeenCalledWith({ x: 10, y: 20 })
    expect(setPendingAddToGroupId).toHaveBeenCalledWith('g1')
    expect(setGroupingPrompt).toHaveBeenCalledWith(null)
  })

  it('onDismissDragGroupPrompt clears dragGroupPrompt', () => {
    const setDragGroupPrompt = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setDragGroupPrompt })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-dismiss-drag-group'))
    expect(setDragGroupPrompt).toHaveBeenCalledWith(null)
  })

  it('onDismissManageGroup clears managingGroupId', () => {
    const setManagingGroupId = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setManagingGroupId })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-dismiss-manage-group'))
    expect(setManagingGroupId).toHaveBeenCalledWith(null)
  })

  it('onUpdatePins calls setPins with updater function', () => {
    const setPins = vi.fn()
    mockUseMapViewData.mockReturnValue({ ...baseMapViewData, setPins })
    renderMapView()
    fireEvent.click(screen.getByTestId('cb-update-pins'))
    expect(setPins).toHaveBeenCalled()
  })
})

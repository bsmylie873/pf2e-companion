import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import TopBar from './TopBar'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockLogout = vi.fn()
vi.mock('../../hooks/useDarkMode', () => ({
  useDarkMode: vi.fn().mockReturnValue([false, vi.fn()]),
}))

vi.mock('../../context/AuthContext', () => ({
  useAuth: vi.fn(),
}))

vi.mock('../../context/MapNavContext', () => ({
  useMapNav: vi.fn(),
}))

vi.mock('../MapSelector/MapSelector', () => ({
  default: () => <div data-testid="map-selector" />,
}))

vi.mock('../Modal/Modal', () => ({
  default: ({ children, onClose }: { children: React.ReactNode; onClose: () => void }) => (
    <div data-testid="modal">
      <button onClick={onClose}>Close Modal</button>
      {children}
    </div>
  ),
}))

vi.mock('../../pages/Settings/Settings', () => ({
  default: () => <div data-testid="settings-page">Settings Content</div>,
}))

import { useAuth } from '../../context/AuthContext'
import { useMapNav } from '../../context/MapNavContext'
import { useDarkMode } from '../../hooks/useDarkMode'

const mockUseAuth = useAuth as ReturnType<typeof vi.fn>
const mockUseMapNav = useMapNav as ReturnType<typeof vi.fn>

describe('TopBar', () => {
  beforeEach(() => {
    mockUseAuth.mockReturnValue({
      user: { id: 'user-1', username: 'testuser', email: 'test@example.com', avatar_url: null },
      isAuthenticated: true,
      isLoading: false,
      logout: mockLogout,
    })
    mockUseMapNav.mockReturnValue({ state: null })
  })

  it('should render the brand title when not on map view', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByText('PF2E Companion')).toBeInTheDocument()
  })

  it('should render dark mode toggle button', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByRole('button', { name: /Switch to dark mode/i })).toBeInTheDocument()
  })

  it('should render settings button', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByRole('button', { name: 'Settings' })).toBeInTheDocument()
  })

  it('should open settings modal when settings button is clicked', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(screen.getByTestId('modal')).toBeInTheDocument()
    expect(screen.getByTestId('settings-page')).toBeInTheDocument()
  })

  it('should render logout button when authenticated', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByRole('button', { name: 'Logout' })).toBeInTheDocument()
  })

  it('should call logout when logout button is clicked', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Logout' }))
    expect(mockLogout).toHaveBeenCalledTimes(1)
  })

  it('should render profile button when authenticated', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByRole('button', { name: 'Profile' })).toBeInTheDocument()
  })

  it('should navigate to profile when profile button is clicked', () => {
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByRole('button', { name: 'Profile' }))
    expect(mockNavigate).toHaveBeenCalledWith('/profile')
  })

  it('should not show logout/profile buttons when not authenticated', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      logout: mockLogout,
    })
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.queryByRole('button', { name: 'Logout' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Profile' })).not.toBeInTheDocument()
  })

  it('should render game title and map dropdown when mapNav state is active', () => {
    mockUseMapNav.mockReturnValue({
      state: {
        gameId: 'game-1',
        gameTitle: 'The Shattered Throne',
        maps: [{ id: 'map-1', name: 'World Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '', updated_at: '' }],
        archivedMaps: [],
        activeMapId: 'map-1',
        isGM: true,
        onSelectMap: vi.fn(),
        onCreateMap: vi.fn(),
        onRenameMap: vi.fn(),
        onArchiveMap: vi.fn(),
        onUnarchiveMap: vi.fn(),
        onReorderMaps: vi.fn(),
      },
    })
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    expect(screen.getByText('The Shattered Throne')).toBeInTheDocument()
    expect(screen.getByText('World Map')).toBeInTheDocument()
  })

  it('should navigate to game page when game title breadcrumb is clicked', () => {
    mockUseMapNav.mockReturnValue({
      state: {
        gameId: 'game-1',
        gameTitle: 'The Shattered Throne',
        maps: [],
        archivedMaps: [],
        activeMapId: null,
        isGM: false,
        onSelectMap: vi.fn(),
        onCreateMap: vi.fn(),
        onRenameMap: vi.fn(),
        onArchiveMap: vi.fn(),
        onUnarchiveMap: vi.fn(),
        onReorderMaps: vi.fn(),
      },
    })
    render(
      <MemoryRouter>
        <TopBar />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByText('The Shattered Throne'))
    expect(mockNavigate).toHaveBeenCalledWith('/games/game-1')
  })

  // ── Map dropdown ─────────────────────────────────────────────────
  it('should open map dropdown when map toggle is clicked', () => {
    mockUseMapNav.mockReturnValue({
      state: {
        gameId: 'game-1',
        gameTitle: 'The Realm',
        maps: [{ id: 'map-1', name: 'World Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '', updated_at: '' }],
        archivedMaps: [],
        activeMapId: 'map-1',
        isGM: true,
        onSelectMap: vi.fn(),
        onCreateMap: vi.fn(),
        onRenameMap: vi.fn(),
        onArchiveMap: vi.fn(),
        onUnarchiveMap: vi.fn(),
        onReorderMaps: vi.fn(),
      },
    })
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    fireEvent.click(screen.getByText('World Map'))
    expect(screen.getByTestId('map-selector')).toBeInTheDocument()
  })

  it('should close map dropdown when clicked outside', () => {
    mockUseMapNav.mockReturnValue({
      state: {
        gameId: 'game-1',
        gameTitle: 'The Realm',
        maps: [{ id: 'map-1', name: 'World Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '', updated_at: '' }],
        archivedMaps: [],
        activeMapId: 'map-1',
        isGM: true,
        onSelectMap: vi.fn(),
        onCreateMap: vi.fn(),
        onRenameMap: vi.fn(),
        onArchiveMap: vi.fn(),
        onUnarchiveMap: vi.fn(),
        onReorderMaps: vi.fn(),
      },
    })
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    // Open the dropdown
    fireEvent.click(screen.getByText('World Map'))
    expect(screen.getByTestId('map-selector')).toBeInTheDocument()
    // Click outside
    fireEvent.mouseDown(document.body)
    expect(screen.queryByTestId('map-selector')).not.toBeInTheDocument()
  })

  it('should show "Maps" as label when activeMapId not in maps list', () => {
    mockUseMapNav.mockReturnValue({
      state: {
        gameId: 'game-1',
        gameTitle: 'The Realm',
        maps: [{ id: 'map-1', name: 'World Map', game_id: 'game-1', archived_at: null, image_url: null, position: 0, created_at: '', updated_at: '' }],
        archivedMaps: [],
        activeMapId: 'non-existent-map',
        isGM: true,
        onSelectMap: vi.fn(),
        onCreateMap: vi.fn(),
        onRenameMap: vi.fn(),
        onArchiveMap: vi.fn(),
        onUnarchiveMap: vi.fn(),
        onReorderMaps: vi.fn(),
      },
    })
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    expect(screen.getByText('Maps')).toBeInTheDocument()
  })

  // ── Settings modal ───────────────────────────────────────────────
  it('should close settings modal when modal close button is clicked', () => {
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    fireEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(screen.getByTestId('modal')).toBeInTheDocument()
    fireEvent.click(screen.getByText('Close Modal'))
    expect(screen.queryByTestId('modal')).not.toBeInTheDocument()
  })

  // ── Dark mode ────────────────────────────────────────────────────
  it('should show sun icon and "Switch to light mode" when isDark is true', () => {
    vi.mocked(useDarkMode).mockReturnValueOnce([true, vi.fn()])
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    expect(screen.getByRole('button', { name: /switch to light mode/i })).toBeInTheDocument()
  })

  it('should render avatar image when user has avatar_url', () => {
    mockUseAuth.mockReturnValue({
      user: { id: 'user-1', username: 'testuser', email: 'test@example.com', avatar_url: 'https://example.com/avatar.png' },
      isAuthenticated: true,
      isLoading: false,
      logout: mockLogout,
    })
    render(<MemoryRouter><TopBar /></MemoryRouter>)
    const img = screen.getByRole('img', { name: 'testuser' })
    expect(img).toBeInTheDocument()
    expect(img).toHaveAttribute('src', 'https://example.com/avatar.png')
  })
})

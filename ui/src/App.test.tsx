import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, afterEach } from 'vitest'
import App from './App'

// Mock all pages as simple divs to avoid their complex dependencies
vi.mock('./pages/Login/Login', () => ({ default: () => <div data-testid="page-login">Login</div> }))
vi.mock('./pages/GamesList/GamesList', () => ({ default: () => <div data-testid="page-games">GamesList</div> }))
vi.mock('./pages/Editor/Editor', () => ({ default: () => <div data-testid="page-editor">Editor</div> }))
vi.mock('./pages/SessionNotes/SessionNotes', () => ({ default: () => <div data-testid="page-session-notes">SessionNotes</div> }))
vi.mock('./pages/NoteEditor/NoteEditor', () => ({ default: () => <div data-testid="page-note-editor">NoteEditor</div> }))
vi.mock('./pages/Profile/Profile', () => ({ default: () => <div data-testid="page-profile">Profile</div> }))
vi.mock('./pages/MapView/MapView', () => ({ default: () => <div data-testid="page-map">MapView</div> }))
vi.mock('./pages/ForgotPassword/ForgotPassword', () => ({ default: () => <div data-testid="page-forgot">ForgotPassword</div> }))
vi.mock('./pages/ResetPassword/ResetPassword', () => ({ default: () => <div data-testid="page-reset">ResetPassword</div> }))
vi.mock('./pages/JoinGame/JoinGame', () => ({ default: () => <div data-testid="page-join">JoinGame</div> }))
vi.mock('./pages/GameSettings/GameSettings', () => ({ default: () => <div data-testid="page-settings">GameSettings</div> }))
vi.mock('./components/TopBar/TopBar', () => ({ default: () => <div data-testid="topbar">TopBar</div> }))
vi.mock('./components/ProtectedRoute/ProtectedRoute', () => ({ default: ({ children }: { children: React.ReactNode }) => <>{children}</> }))
vi.mock('./context/AuthContext', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useAuth: vi.fn(() => ({ user: null, loading: false, login: vi.fn(), logout: vi.fn(), register: vi.fn() })),
}))
vi.mock('./context/MapNavContext', () => ({
  MapNavProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  useMapNav: vi.fn(() => ({ state: null, register: vi.fn(), unregister: vi.fn() })),
}))

// Override BrowserRouter with MemoryRouter so we can control the URL in tests
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    BrowserRouter: ({ children }: { children: React.ReactNode }) => {
      return (
        <actual.MemoryRouter initialEntries={[window.location.pathname]}>
          {children}
        </actual.MemoryRouter>
      )
    },
  }
})

afterEach(() => {
  // Reset URL to / after each test
  window.history.replaceState(null, '', '/')
})

describe('App', () => {
  it('renders without crashing and shows TopBar', () => {
    render(<App />)
    expect(screen.getByTestId('topbar')).toBeInTheDocument()
  })

  it('renders the .app wrapper div', () => {
    render(<App />)
    expect(document.querySelector('.app')).not.toBeNull()
  })

  it('renders the main.app-content element', () => {
    render(<App />)
    const main = document.querySelector('main.app-content')
    expect(main).not.toBeNull()
  })

  it('renders Login page on / route', () => {
    window.history.replaceState(null, '', '/')
    render(<App />)
    expect(screen.getByTestId('page-login')).toBeInTheDocument()
  })

  it('renders ForgotPassword page on /forgot-password route', () => {
    window.history.replaceState(null, '', '/forgot-password')
    render(<App />)
    expect(screen.getByTestId('page-forgot')).toBeInTheDocument()
  })

  it('renders ResetPassword page on /reset-password route', () => {
    window.history.replaceState(null, '', '/reset-password')
    render(<App />)
    expect(screen.getByTestId('page-reset')).toBeInTheDocument()
  })

  it('renders JoinGame page on /join/:token route', () => {
    window.history.replaceState(null, '', '/join/abc123')
    render(<App />)
    expect(screen.getByTestId('page-join')).toBeInTheDocument()
  })

  it('renders GamesList page on /games route (ProtectedRoute is transparent)', () => {
    window.history.replaceState(null, '', '/games')
    render(<App />)
    expect(screen.getByTestId('page-games')).toBeInTheDocument()
  })

  it('wildcard route redirects to / and renders Login', () => {
    window.history.replaceState(null, '', '/this-does-not-exist')
    render(<App />)
    expect(screen.getByTestId('page-login')).toBeInTheDocument()
  })

  it('renders Profile page on /profile route', () => {
    window.history.replaceState(null, '', '/profile')
    render(<App />)
    expect(screen.getByTestId('page-profile')).toBeInTheDocument()
  })
})

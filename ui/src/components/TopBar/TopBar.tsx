import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useDarkMode } from '../../hooks/useDarkMode'
import { useAuth } from '../../context/AuthContext'
import { useMapNav } from '../../context/MapNavContext'
import MapSelector from '../MapSelector/MapSelector'
import Modal from '../Modal/Modal'
import Settings from '../../pages/Settings/Settings'
import './TopBar.css'

function SunIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="4" />
      <line x1="12" y1="2" x2="12" y2="4" />
      <line x1="12" y1="20" x2="12" y2="22" />
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
      <line x1="2" y1="12" x2="4" y2="12" />
      <line x1="20" y1="12" x2="22" y2="12" />
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
  )
}

function MoonIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
  )
}

function GearIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
    </svg>
  )
}

function UserIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </svg>
  )
}

function LogoutIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
      <polyline points="16 17 21 12 16 7" />
      <line x1="21" y1="12" x2="9" y2="12" />
    </svg>
  )
}

function ChevronDownIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" width="10" height="10">
      <polyline points="6 9 12 15 18 9" />
    </svg>
  )
}

export default function TopBar() {
  const [isDark, setIsDark] = useDarkMode()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const { user, isAuthenticated, logout } = useAuth()
  const navigate = useNavigate()

  const { state: mapNav } = useMapNav()
  const [mapDropdownOpen, setMapDropdownOpen] = useState(false)
  const mapDropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown on outside click
  useEffect(() => {
    if (!mapDropdownOpen) return
    const handler = (e: MouseEvent) => {
      if (mapDropdownRef.current && !mapDropdownRef.current.contains(e.target as Node)) {
        setMapDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [mapDropdownOpen])

  // Close dropdown when navigating away from map view
  useEffect(() => {
    if (!mapNav) setMapDropdownOpen(false)
  }, [mapNav])

  const activeMapName = mapNav?.maps.find(m => m.id === mapNav.activeMapId)?.name ?? 'Maps'

  return (
    <>
      <header className="topbar">
        {mapNav ? (
          <div className="topbar-brand">
            <span className="topbar-ornament">✦</span>
            <button
              className="topbar-breadcrumb"
              onClick={() => navigate(`/games/${mapNav.gameId}`)}
            >
              {mapNav.gameTitle}
            </button>
            <span className="topbar-ornament">✦</span>
          </div>
        ) : (
          <div className="topbar-brand">
            <span className="topbar-ornament">✦</span>
            <span className="topbar-title">PF2E Companion</span>
            <span className="topbar-ornament">✦</span>
          </div>
        )}

        {mapNav && (
          <div className="topbar-map-selector" ref={mapDropdownRef}>
            <button
              className="topbar-map-toggle"
              onClick={() => setMapDropdownOpen(o => !o)}
            >
              {activeMapName}
              <ChevronDownIcon />
            </button>
            {mapDropdownOpen && (
              <div className="topbar-map-dropdown">
                <MapSelector
                  maps={mapNav.maps}
                  activeMapId={mapNav.activeMapId}
                  onSelect={(id) => { mapNav.onSelectMap(id); setMapDropdownOpen(false) }}
                  isGM={mapNav.isGM}
                  onCreateMap={mapNav.onCreateMap}
                  onRenameMap={mapNav.onRenameMap}
                  onArchiveMap={mapNav.onArchiveMap}
                  onUnarchiveMap={mapNav.onUnarchiveMap}
                  onReorderMaps={mapNav.onReorderMaps}
                  archivedMaps={mapNav.archivedMaps}
                />
              </div>
            )}
          </div>
        )}

        <nav className="topbar-actions">
          <button
            className="topbar-icon-btn"
            onClick={() => setIsDark(!isDark)}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            title={isDark ? 'Light Mode' : 'Dark Mode'}
          >
            {isDark ? <SunIcon /> : <MoonIcon />}
          </button>

          <button
            className="topbar-icon-btn"
            onClick={() => setSettingsOpen(true)}
            aria-label="Settings"
            title="Settings"
          >
            <GearIcon />
          </button>

          {isAuthenticated && (
            <>
              <button
                className="topbar-icon-btn"
                onClick={() => navigate('/profile')}
                aria-label="Profile"
                title="Profile"
              >
                {user?.avatar_url ? (
                  <img
                    className="topbar-avatar"
                    src={user.avatar_url}
                    alt={user.username}
                  />
                ) : (
                  <UserIcon />
                )}
              </button>

              <button
                className="topbar-icon-btn"
                onClick={() => logout()}
                aria-label="Logout"
                title="Logout"
              >
                <LogoutIcon />
              </button>
            </>
          )}
        </nav>
      </header>

      {settingsOpen && (
        <Modal title="Sanctum" onClose={() => setSettingsOpen(false)}>
          <Settings />
        </Modal>
      )}
    </>
  )
}

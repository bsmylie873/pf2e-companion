import { useState, useEffect } from 'react'
import { apiFetch } from '../../api/client'
import { getPreferences, updatePreferences } from '../../api/preferences'
import type { UserPreferences } from '../../api/preferences'
import { useDarkMode } from '../../hooks/useDarkMode'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import { COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS, PIN_COLOURS, PIN_ICONS } from '../../constants/pins'
import type { Game } from '../../types/game'
import './Settings.css'

export default function Settings() {
  const [isDark, setIsDark] = useDarkMode()
  const [layout, setLayout] = useLocalStorage<'grid' | 'list'>('pf2e-layout-pref', 'grid')
  const [games, setGames] = useState<Game[]>([])
  const [prefs, setPrefs] = useState<UserPreferences>({
    default_game_id: null,
    default_pin_colour: null,
    default_pin_icon: null,
    sidebar_state: null,
  })
  const [prefsError, setPrefsError] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      getPreferences().catch(() => {
        setPrefsError(true)
        return { default_game_id: null, default_pin_colour: null, default_pin_icon: null } as UserPreferences
      }),
      apiFetch<Game[]>('/games').catch(() => [] as Game[]),
    ]).then(([fetchedPrefs, fetchedGames]) => {
      setPrefs(fetchedPrefs)
      setGames(fetchedGames)
      setLoading(false)
    })
  }, [])

  const savePrefs = async (updates: Partial<UserPreferences>) => {
    setPrefs(prev => ({ ...prev, ...updates }))
    try {
      const updated = await updatePreferences(updates)
      setPrefs(updated)
    } catch {
      // optimistic update left in place on failure
    }
  }

  return (
    <div className="settings-content">
      {/* ── Section 1: Device Preferences ────────────────────── */}
      <section className="settings-section settings-section--device">
        <div className="settings-section-header">
          <h2 className="settings-section-title">This Device</h2>
          <span className="settings-badge settings-badge--device">On this device only</span>
        </div>

        {/* Dark mode toggle */}
        <div className="settings-row">
          <div className="settings-row-info">
            <span className="settings-row-label">
              {isDark ? 'Veil of Night' : 'Light of Day'}
            </span>
            <span className="settings-row-hint">Toggle the ambient illumination</span>
          </div>
          <button
            className={`settings-toggle${isDark ? ' settings-toggle--on' : ''}`}
            onClick={() => setIsDark(!isDark)}
            aria-pressed={isDark}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            <span className="settings-toggle-track">
              <span className="settings-toggle-icon settings-toggle-icon--sun" aria-hidden="true">☀</span>
              <span className="settings-toggle-icon settings-toggle-icon--moon" aria-hidden="true">☽</span>
            </span>
            <span className="settings-toggle-thumb" />
          </button>
        </div>

        {/* Layout preference */}
        <div className="settings-row settings-row--last">
          <div className="settings-row-info">
            <span className="settings-row-label">Chronicle Layout</span>
            <span className="settings-row-hint">How entries are arranged in your journal</span>
          </div>
          <div className="settings-layout-toggle">
            <button
              className={`settings-layout-btn${layout === 'grid' ? ' active' : ''}`}
              onClick={() => setLayout('grid')}
              aria-pressed={layout === 'grid'}
              aria-label="Grid layout"
            >
              <svg viewBox="0 0 16 16" fill="currentColor" width="13" height="13">
                <rect x="1" y="1" width="6" height="6" rx="1"/>
                <rect x="9" y="1" width="6" height="6" rx="1"/>
                <rect x="1" y="9" width="6" height="6" rx="1"/>
                <rect x="9" y="9" width="6" height="6" rx="1"/>
              </svg>
              Grid
            </button>
            <button
              className={`settings-layout-btn${layout === 'list' ? ' active' : ''}`}
              onClick={() => setLayout('list')}
              aria-pressed={layout === 'list'}
              aria-label="List layout"
            >
              <svg viewBox="0 0 16 16" fill="currentColor" width="13" height="13">
                <rect x="1" y="2" width="14" height="2" rx="1"/>
                <rect x="1" y="7" width="14" height="2" rx="1"/>
                <rect x="1" y="12" width="14" height="2" rx="1"/>
              </svg>
              List
            </button>
          </div>
        </div>
      </section>

      {/* ── Section 2: Account Preferences ───────────────────── */}
      <section className="settings-section settings-section--account">
        <div className="settings-section-header">
          <h2 className="settings-section-title">Account</h2>
          <span className="settings-badge settings-badge--account">Synced to your account</span>
        </div>

        {prefsError && (
          <div className="settings-error-banner" role="alert">
            <svg viewBox="0 0 20 20" fill="currentColor" width="15" height="15" aria-hidden="true">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            The arcane servers could not be reached. Account preferences may not save.
          </div>
        )}

        {loading ? (
          <div className="settings-loading">
            <div className="settings-loading-rune" aria-hidden="true">
              <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="20" cy="20" r="16" stroke="currentColor" strokeWidth="1.5" strokeDasharray="8 4" />
                <circle cx="20" cy="20" r="8" stroke="currentColor" strokeWidth="1" opacity="0.6" />
                <circle cx="20" cy="20" r="3" fill="currentColor" />
              </svg>
            </div>
            <span>Consulting the oracle…</span>
          </div>
        ) : (
          <>
            {/* Default Campaign */}
            <div className="settings-row settings-row--stacked">
              <div className="settings-row-info">
                <span className="settings-row-label">Default Campaign</span>
                <span className="settings-row-hint">Venture here upon entering the realm</span>
              </div>
              <select
                className="settings-select"
                value={prefs.default_game_id ?? ''}
                onChange={e => savePrefs({ default_game_id: e.target.value || null })}
              >
                <option value="">— No default campaign —</option>
                {games.map(g => (
                  <option key={g.id} value={g.id}>{g.title}</option>
                ))}
              </select>
            </div>

            {/* Default Pin Colour */}
            <div className="settings-row settings-row--stacked">
              <div className="settings-row-info">
                <span className="settings-row-label">Default Pin Colour</span>
                <span className="settings-row-hint">The hue of your placed markers</span>
              </div>
              <div className="settings-colour-grid">
                {PIN_COLOURS.map(colour => (
                  <button
                    key={colour}
                    className={`settings-colour-swatch${prefs.default_pin_colour === colour ? ' active' : ''}`}
                    onClick={() => savePrefs({ default_pin_colour: colour })}
                    title={colour.charAt(0).toUpperCase() + colour.slice(1)}
                    aria-label={`${colour} pin colour`}
                    aria-pressed={prefs.default_pin_colour === colour}
                  >
                    <span
                      className="settings-colour-swatch-fill"
                      style={{ backgroundColor: COLOUR_MAP[colour] }}
                    />
                    {prefs.default_pin_colour === colour && (
                      <svg className="settings-colour-check" viewBox="0 0 16 16" fill="none">
                        <path d="M3 8.5l3.5 3.5 6.5-7" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    )}
                  </button>
                ))}
              </div>
            </div>

            {/* Default Pin Icon */}
            <div className="settings-row settings-row--stacked settings-row--last">
              <div className="settings-row-info">
                <span className="settings-row-label">Default Pin Icon</span>
                <span className="settings-row-hint">The sigil that marks your path</span>
              </div>
              <div className="settings-icon-grid">
                {PIN_ICONS.map(icon => {
                  const IconComponent = PIN_ICON_COMPONENTS[icon]
                  return (
                    <button
                      key={icon}
                      className={`settings-icon-btn${prefs.default_pin_icon === icon ? ' active' : ''}`}
                      onClick={() => savePrefs({ default_pin_icon: icon })}
                      title={PIN_ICON_LABELS[icon]}
                      aria-label={PIN_ICON_LABELS[icon]}
                      aria-pressed={prefs.default_pin_icon === icon}
                    >
                      <IconComponent size={16} />
                    </button>
                  )
                })}
              </div>
            </div>
          </>
        )}
      </section>
    </div>
  )
}

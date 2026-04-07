import { useState, useEffect, useRef } from 'react'
import { apiFetch } from '../../api/client'
import { exportGameBackup, importGameBackup } from '../../api/backup'
import type { ImportSummary } from '../../api/backup'
import { getPreferences, updatePreferences } from '../../api/preferences'
import type { UserPreferences, PageSizePreferences } from '../../api/preferences'
import { useDarkMode } from '../../hooks/useDarkMode'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import { COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS, PIN_COLOURS, PIN_ICONS } from '../../constants/pins'
import type { Game } from '../../types/game'
import './Settings.css'

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10 MB

type SettingsTab = 'preferences' | 'backup'

export default function Settings() {
  useDocumentTitle('Settings')
  const [activeTab, setActiveTab] = useState<SettingsTab>('preferences')
  const [isDark, setIsDark] = useDarkMode()
  const [layout, setLayout] = useLocalStorage<'grid' | 'list'>('pf2e-layout-pref', 'grid')
  const [games, setGames] = useState<Game[]>([])
  const [prefs, setPrefs] = useState<UserPreferences>({
    default_game_id: null,
    default_pin_colour: null,
    default_pin_icon: null,
    sidebar_state: null,
    default_view_mode: null,
    map_editor_mode: 'modal',
    page_size: null,
  })
  const [prefsError, setPrefsError] = useState(false)
  const [loading, setLoading] = useState(true)

  // Backup state
  const [backupGameId, setBackupGameId] = useState('')
  const [importFile, setImportFile] = useState<File | null>(null)
  const [importMode, setImportMode] = useState<'merge' | 'overwrite' | null>(null)
  const [exporting, setExporting] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<ImportSummary | null>(null)
  const [backupError, setBackupError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    Promise.all([
      getPreferences().catch(() => {
        setPrefsError(true)
        return { default_game_id: null, default_pin_colour: null, default_pin_icon: null, sidebar_state: null, default_view_mode: null, map_editor_mode: 'modal', page_size: null } as UserPreferences
      }),
      apiFetch<Game[]>('/games').catch(() => [] as Game[]),
    ]).then(([fetchedPrefs, fetchedGames]) => {
      setPrefs(fetchedPrefs)
      setGames(fetchedGames)
      setLoading(false)
    })
  }, [])

  const handleExport = async () => {
    setBackupError(null)
    setImportResult(null)
    setExporting(true)
    try {
      await exportGameBackup(backupGameId)
    } catch (e) {
      setBackupError(e instanceof Error ? e.message : 'Export failed.')
    } finally {
      setExporting(false)
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setBackupError(null)
    setImportResult(null)
    if (file.size > MAX_FILE_SIZE) {
      setBackupError('File exceeds 10 MB limit.')
      setImportFile(null)
      e.target.value = ''
      return
    }
    setImportFile(file)
  }

  const handleImport = async () => {
    if (!importFile || !importMode || !backupGameId) return
    setBackupError(null)
    setImportResult(null)
    setImporting(true)
    try {
      const result = await importGameBackup(backupGameId, importFile, importMode)
      setImportResult(result)
      setImportFile(null)
      setImportMode(null)
      if (fileInputRef.current) fileInputRef.current.value = ''
    } catch (e) {
      setBackupError(e instanceof Error ? e.message : 'Import failed.')
    } finally {
      setImporting(false)
    }
  }

  const savePrefs = async (updates: Partial<UserPreferences>) => {
    setPrefs(prev => ({ ...prev, ...updates }))
    try {
      const updated = await updatePreferences(updates)
      setPrefs(updated)
    } catch {
      // optimistic update left in place on failure
    }
  }

  const isBusy = exporting || importing

  return (
    <div className="settings-content">
      {/* ── Tab bar ─────────────────────────────────────────── */}
      <nav className="settings-tabs" role="tablist">
        <button
          className={`settings-tab${activeTab === 'preferences' ? ' settings-tab--active' : ''}`}
          onClick={() => setActiveTab('preferences')}
          role="tab"
          aria-selected={activeTab === 'preferences'}
          aria-controls="settings-panel-preferences"
        >
          <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <circle cx="8" cy="8" r="3"/>
            <path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.05 3.05l1.41 1.41M11.54 11.54l1.41 1.41M3.05 12.95l1.41-1.41M11.54 4.46l1.41-1.41"/>
          </svg>
          Preferences
        </button>
        <button
          className={`settings-tab${activeTab === 'backup' ? ' settings-tab--active' : ''}`}
          onClick={() => setActiveTab('backup')}
          role="tab"
          aria-selected={activeTab === 'backup'}
          aria-controls="settings-panel-backup"
        >
          <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M2 2h8l4 4v8a1 1 0 01-1 1H2a1 1 0 01-1-1V3a1 1 0 011-1z"/>
            <path d="M10 2v4h4M5 9h6M5 12h4"/>
          </svg>
          Data Backup
        </button>
      </nav>

      {/* ── Preferences panel ──────────────────────────────── */}
      {activeTab === 'preferences' && (
        <div id="settings-panel-preferences" role="tabpanel">
          {/* Section 1: Device Preferences */}
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

          {/* Section 2: Account Preferences */}
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
                <div className="settings-row settings-row--stacked">
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

                {/* Map Editor Mode */}
                <div className="settings-row">
                  <div className="settings-row-info">
                    <span className="settings-row-label">Map Editor Mode</span>
                    <span className="settings-row-hint">How sessions and notes open from the map view</span>
                  </div>
                  <div className="settings-layout-toggle">
                    <button
                      className={`settings-layout-btn${prefs.map_editor_mode === 'modal' ? ' active' : ''}`}
                      onClick={() => savePrefs({ map_editor_mode: 'modal' })}
                      aria-pressed={prefs.map_editor_mode === 'modal'}
                    >
                      Modal Overlay
                    </button>
                    <button
                      className={`settings-layout-btn${prefs.map_editor_mode === 'navigate' ? ' active' : ''}`}
                      onClick={() => savePrefs({ map_editor_mode: 'navigate' })}
                      aria-pressed={prefs.map_editor_mode === 'navigate'}
                    >
                      Full Page
                    </button>
                  </div>
                </div>

                {/* Pagination Preferences */}
                <div className="settings-row settings-row--stacked settings-row--last">
                  <div className="settings-row-info">
                    <span className="settings-row-label">Items Per Page</span>
                    <span className="settings-row-hint">How many entries to show in each list. Per-resource overrides take precedence.</span>
                  </div>
                  <div className="settings-page-size-grid">
                    {(['default', 'campaigns', 'sessions', 'notes'] as const).map(key => {
                      const ps = prefs.page_size ?? { default: 10 }
                      const value = key === 'default' ? (ps.default ?? 10) : (ps[key] ?? '')
                      const label = key === 'default' ? 'Default' : key.charAt(0).toUpperCase() + key.slice(1)
                      return (
                        <div key={key} className="settings-page-size-field">
                          <label className="settings-page-size-label">{label}</label>
                          <select
                            className="settings-select settings-select--compact"
                            value={value}
                            onChange={e => {
                              const raw = e.target.value
                              const newPs: PageSizePreferences = { ...(prefs.page_size ?? { default: 10 }) }
                              if (key === 'default') {
                                newPs.default = Number(raw)
                              } else {
                                newPs[key] = raw === '' ? null : Number(raw)
                              }
                              savePrefs({ page_size: newPs })
                            }}
                          >
                            {key !== 'default' && <option value="">Use default</option>}
                            {[5, 10, 20, 50, 100].map(n => (
                              <option key={n} value={n}>{n}</option>
                            ))}
                          </select>
                        </div>
                      )
                    })}
                  </div>
                </div>
              </>
            )}
          </section>
        </div>
      )}

      {/* ── Backup panel ───────────────────────────────────── */}
      {activeTab === 'backup' && (
        <div id="settings-panel-backup" role="tabpanel">
          <section className="settings-section settings-section--backup">
            <div className="settings-section-header">
              <h2 className="settings-section-title">Data Backup</h2>
              <span className="settings-badge settings-badge--backup">Per-game export &amp; import</span>
            </div>

            {/* Game selector */}
            <div className="settings-row settings-row--stacked">
              <div className="settings-row-info">
                <span className="settings-row-label">Campaign</span>
                <span className="settings-row-hint">Select which campaign to export or import into</span>
              </div>
              <select
                className="settings-select"
                value={backupGameId}
                onChange={e => { setBackupGameId(e.target.value); setBackupError(null); setImportResult(null) }}
              >
                <option value="">— Select a campaign —</option>
                {games.map(g => (
                  <option key={g.id} value={g.id}>{g.title}</option>
                ))}
              </select>
            </div>

            {/* Export */}
            <div className="settings-row">
              <div className="settings-row-info">
                <span className="settings-row-label">Export</span>
                <span className="settings-row-hint">Download a JSON backup of all sessions and notes</span>
              </div>
              <button
                className="settings-backup-btn settings-backup-btn--export"
                onClick={handleExport}
                disabled={!backupGameId || isBusy}
              >
                <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M8 1v9M5 7l3 3 3-3M2 12v2a1 1 0 001 1h10a1 1 0 001-1v-2"/>
                </svg>
                {exporting ? 'Sealing…' : 'Export Chronicle'}
              </button>
            </div>

            {/* Import: file input */}
            <div className="settings-row settings-row--stacked">
              <div className="settings-row-info">
                <span className="settings-row-label">Import</span>
                <span className="settings-row-hint">Restore sessions and notes from a backup file (max 10 MB)</span>
              </div>
              <div className="settings-backup-import-controls">
                <label className="settings-backup-file-label">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept=".json,application/json"
                    className="settings-backup-file-input"
                    onChange={handleFileSelect}
                    disabled={isBusy}
                  />
                  <span className="settings-backup-file-btn">
                    <svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M13 10v3a1 1 0 01-1 1H4a1 1 0 01-1-1v-3M8 2v8M5 5l3-3 3 3"/>
                    </svg>
                    {importFile ? importFile.name : 'Choose File'}
                  </span>
                </label>
              </div>
            </div>

            {/* Import: conflict mode */}
            <div className="settings-row settings-row--stacked">
              <div className="settings-row-info">
                <span className="settings-row-label">Conflict Resolution</span>
                <span className="settings-row-hint">How to handle records that already exist in the campaign</span>
              </div>
              <div className="settings-layout-toggle">
                <button
                  className={`settings-layout-btn${importMode === 'merge' ? ' active' : ''}`}
                  onClick={() => setImportMode('merge')}
                  aria-pressed={importMode === 'merge'}
                  disabled={isBusy}
                >
                  Merge (skip existing)
                </button>
                <button
                  className={`settings-layout-btn${importMode === 'overwrite' ? ' active' : ''}`}
                  onClick={() => setImportMode('overwrite')}
                  aria-pressed={importMode === 'overwrite'}
                  disabled={isBusy}
                >
                  Overwrite (replace)
                </button>
              </div>
            </div>

            {/* Import: submit */}
            <div className="settings-row settings-row--last">
              <div className="settings-row-info">
                <span className="settings-row-hint">
                  {!backupGameId ? 'Select a campaign first' :
                   !importFile ? 'Choose a backup file' :
                   !importMode ? 'Select a conflict resolution mode' :
                   `Ready to import into selected campaign (${importMode} mode)`}
                </span>
              </div>
              <button
                className="settings-backup-btn settings-backup-btn--import"
                onClick={handleImport}
                disabled={!backupGameId || !importFile || !importMode || isBusy}
              >
                <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M8 15V6M5 9l3-3 3 3M2 4V2a1 1 0 011-1h10a1 1 0 011 1v2"/>
                </svg>
                {importing ? 'Restoring…' : 'Import Chronicle'}
              </button>
            </div>

            {/* Error display */}
            {backupError && (
              <div className="settings-backup-message settings-backup-message--error" role="alert">
                <svg viewBox="0 0 20 20" fill="currentColor" width="15" height="15" aria-hidden="true">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                {backupError}
              </div>
            )}

            {/* Import result summary */}
            {importResult && (
              <div className="settings-backup-message settings-backup-message--success" role="status">
                <div className="settings-backup-summary">
                  <span className="settings-backup-summary-title">Import Complete</span>
                  <div className="settings-backup-summary-grid">
                    <div className="settings-backup-summary-col">
                      <span className="settings-backup-summary-heading">Sessions</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.sessions_created}</strong> created</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.sessions_skipped}</strong> skipped</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.sessions_overwritten}</strong> overwritten</span>
                    </div>
                    <div className="settings-backup-summary-col">
                      <span className="settings-backup-summary-heading">Notes</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.notes_created}</strong> created</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.notes_skipped}</strong> skipped</span>
                      <span className="settings-backup-summary-stat"><strong>{importResult.notes_overwritten}</strong> overwritten</span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </section>
        </div>
      )}
    </div>
  )
}

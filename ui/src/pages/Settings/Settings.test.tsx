import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import Settings from './Settings'
import { getPreferences, updatePreferences } from '../../api/preferences'
import { exportGameBackup, importGameBackup } from '../../api/backup'
import { apiFetch } from '../../api/client'
import { useDarkMode } from '../../hooks/useDarkMode'

vi.mock('../../api/preferences', () => ({
  getPreferences: vi.fn().mockResolvedValue({
    default_game_id: null,
    default_pin_colour: null,
    default_pin_icon: null,
    sidebar_state: null,
    default_view_mode: null,
    map_editor_mode: 'modal',
    page_size: null,
  }),
  updatePreferences: vi.fn().mockResolvedValue({
    default_game_id: null,
    default_pin_colour: null,
    default_pin_icon: null,
    sidebar_state: null,
    default_view_mode: null,
    map_editor_mode: 'modal',
    page_size: null,
  }),
}))

vi.mock('../../api/client', () => ({
  apiFetch: vi.fn().mockResolvedValue([]),
}))

vi.mock('../../api/backup', () => ({
  exportGameBackup: vi.fn().mockResolvedValue(undefined),
  importGameBackup: vi.fn().mockResolvedValue({
    sessions_created: 0,
    sessions_skipped: 0,
    sessions_overwritten: 0,
    notes_created: 0,
    notes_skipped: 0,
    notes_overwritten: 0,
  }),
}))

vi.mock('../../hooks/useDarkMode', () => ({
  useDarkMode: vi.fn().mockReturnValue([false, vi.fn()]),
}))

vi.mock('../../hooks/useLocalStorage', () => ({
  useLocalStorage: vi.fn().mockReturnValue(['grid', vi.fn()]),
}))

vi.mock('../../hooks/useDocumentTitle', () => ({
  useDocumentTitle: vi.fn(),
}))

vi.mock('../../constants/pins', () => ({
  COLOUR_MAP: { red: '#ff0000', blue: '#0000ff' },
  PIN_ICON_COMPONENTS: { star: () => <span>★</span>, circle: () => <span>●</span> },
  PIN_ICON_LABELS: { star: 'Star', circle: 'Circle' },
  PIN_COLOURS: ['red', 'blue'],
  PIN_ICONS: ['star', 'circle'],
}))

function renderSettings() {
  return render(
    <MemoryRouter>
      <Settings />
    </MemoryRouter>,
  )
}

describe('Settings', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render tab bar with Preferences and Data Backup tabs', () => {
    renderSettings()
    expect(screen.getByRole('tab', { name: /preferences/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /data backup/i })).toBeInTheDocument()
  })

  it('should show preferences panel by default', () => {
    renderSettings()
    expect(screen.getByRole('tabpanel')).toBeInTheDocument()
    expect(screen.getByText('This Device')).toBeInTheDocument()
  })

  it('should show loading state while preferences load', async () => {
    renderSettings()
    // Initially loading before Promise resolves
    await waitFor(() => {
      // After loading resolves, should show account preferences
      expect(screen.queryByText(/consulting the oracle/i)).not.toBeInTheDocument()
    })
  })

  it('should switch to backup tab when clicked', async () => {
    const user = userEvent.setup()
    renderSettings()

    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    // The panel heading 'Data Backup' + the tab label both appear - use role=tabpanel
    expect(screen.getByRole('tabpanel')).toBeInTheDocument()
    expect(screen.getAllByText('Data Backup').length).toBeGreaterThan(0)
  })

  it('should show dark mode toggle in preferences', async () => {
    renderSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /switch to dark mode|switch to light mode/i })).toBeInTheDocument()
    })
  })

  it('should show layout toggle buttons in preferences', async () => {
    renderSettings()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /grid layout/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /list layout/i })).toBeInTheDocument()
    })
  })

  it('should show account section after preferences load', async () => {
    renderSettings()

    await waitFor(() => {
      expect(screen.getByText('Account')).toBeInTheDocument()
    })
  })

  it('should show export and import buttons in backup tab', async () => {
    const user = userEvent.setup()
    renderSettings()

    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    expect(screen.getByRole('button', { name: /export chronicle/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /import chronicle/i })).toBeInTheDocument()
  })

  it('should show map editor mode toggle in preferences', async () => {
    renderSettings()

    await waitFor(() => {
      expect(screen.getByText('Map Editor Mode')).toBeInTheDocument()
    })
  })

  // ── Loading & error states ──────────────────────────────────────
  it('should show loading indicator initially before prefs resolve', () => {
    // Make getPreferences never resolve so loading stays true
    vi.mocked(getPreferences).mockReturnValueOnce(new Promise(() => {}))
    renderSettings()
    expect(screen.getByText(/consulting the oracle/i)).toBeInTheDocument()
  })

  it('should show error banner when preferences API fails', async () => {
    vi.mocked(getPreferences).mockRejectedValueOnce(new Error('Server error'))
    renderSettings()
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  // ── Device preference interactions ─────────────────────────────
  it('should handle dark mode toggle click', async () => {
    renderSettings()
    const toggle = screen.getByRole('button', { name: /switch to dark mode|switch to light mode/i })
    fireEvent.click(toggle)
    expect(toggle).toBeInTheDocument()
  })

  it('should show "Veil of Night" label when isDark is true', () => {
    vi.mocked(useDarkMode).mockReturnValueOnce([true, vi.fn()])
    renderSettings()
    expect(screen.getByText('Veil of Night')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /switch to light mode/i })).toBeInTheDocument()
  })

  it('should handle list layout button click', async () => {
    renderSettings()
    const listBtn = screen.getByRole('button', { name: /list layout/i })
    fireEvent.click(listBtn)
    expect(listBtn).toBeInTheDocument()
  })

  it('should handle grid layout button click', async () => {
    renderSettings()
    const gridBtn = screen.getByRole('button', { name: /grid layout/i })
    fireEvent.click(gridBtn)
    expect(gridBtn).toBeInTheDocument()
  })

  // ── Account preference interactions ────────────────────────────
  it('should render pin colour buttons after preferences load', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /red pin colour/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /blue pin colour/i })).toBeInTheDocument()
    })
  })

  it('should call updatePreferences when pin colour is clicked', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /red pin colour/i })).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /red pin colour/i }))
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ default_pin_colour: 'red' })
      )
    })
  })

  it('should render pin icon buttons after preferences load', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Star' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Circle' })).toBeInTheDocument()
    })
  })

  it('should call updatePreferences when pin icon is clicked', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Star' })).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: 'Star' }))
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ default_pin_icon: 'star' })
      )
    })
  })

  it('should call updatePreferences when map editor mode changed to Full Page', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Map Editor Mode')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /full page/i }))
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ map_editor_mode: 'navigate' })
      )
    })
  })

  it('should call updatePreferences when map editor mode set back to Modal Overlay', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Map Editor Mode')).toBeInTheDocument()
    })
    fireEvent.click(screen.getByRole('button', { name: /modal overlay/i }))
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ map_editor_mode: 'modal' })
      )
    })
  })

  it('should call updatePreferences when default campaign is changed', async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'g1', title: 'My Campaign', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Account')).toBeInTheDocument()
    })
    // Change the first combobox (Default Campaign)
    const selects = screen.getAllByRole('combobox')
    fireEvent.change(selects[0], { target: { value: 'g1' } })
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ default_game_id: 'g1' })
      )
    })
  })

  it('should call updatePreferences when default campaign cleared', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Account')).toBeInTheDocument()
    })
    const selects = screen.getAllByRole('combobox')
    fireEvent.change(selects[0], { target: { value: '' } })
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ default_game_id: null })
      )
    })
  })

  it('should call updatePreferences when page size changed', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Items Per Page')).toBeInTheDocument()
    })
    const compactSelects = document.querySelectorAll('.settings-select--compact')
    expect(compactSelects.length).toBeGreaterThan(0)
    fireEvent.change(compactSelects[0], { target: { value: '20' } })
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ page_size: expect.objectContaining({ default: 20 }) })
      )
    })
  })

  it('should call updatePreferences when per-resource page size changed', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Items Per Page')).toBeInTheDocument()
    })
    const compactSelects = document.querySelectorAll('.settings-select--compact')
    // Second compact select is 'campaigns'
    fireEvent.change(compactSelects[1], { target: { value: '20' } })
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ page_size: expect.objectContaining({ campaigns: 20 }) })
      )
    })
  })

  it('should call updatePreferences with null when per-resource page size cleared', async () => {
    renderSettings()
    await waitFor(() => {
      expect(screen.getByText('Items Per Page')).toBeInTheDocument()
    })
    const compactSelects = document.querySelectorAll('.settings-select--compact')
    fireEvent.change(compactSelects[1], { target: { value: '' } })
    await waitFor(() => {
      expect(vi.mocked(updatePreferences)).toHaveBeenCalledWith(
        expect.objectContaining({ page_size: expect.objectContaining({ campaigns: null }) })
      )
    })
  })

  // ── Backup tab interactions ─────────────────────────────────────
  it('should call exportGameBackup when Export Chronicle is clicked with a game selected', async () => {
    const user = userEvent.setup()
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Test Game', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()

    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    // Wait for the game option to appear (loaded async via apiFetch)
    await waitFor(() => expect(screen.getByRole('option', { name: 'Test Game' })).toBeInTheDocument())
    await user.selectOptions(screen.getByRole('combobox'), 'game-1')

    // Export button should now be enabled
    await waitFor(() => expect(screen.getByRole('button', { name: /export chronicle/i })).not.toBeDisabled())
    await user.click(screen.getByRole('button', { name: /export chronicle/i }))

    expect(vi.mocked(exportGameBackup)).toHaveBeenCalledWith('game-1')
  })

  it('should show error message when export fails', async () => {
    const user = userEvent.setup()
    vi.mocked(exportGameBackup).mockRejectedValueOnce(new Error('Export failed'))
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Test Game', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()

    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    await waitFor(() => expect(screen.getByRole('option', { name: 'Test Game' })).toBeInTheDocument())
    await user.selectOptions(screen.getByRole('combobox'), 'game-1')
    await waitFor(() => expect(screen.getByRole('button', { name: /export chronicle/i })).not.toBeDisabled())

    await user.click(screen.getByRole('button', { name: /export chronicle/i }))

    await waitFor(() => {
      expect(screen.getByText('Export failed')).toBeInTheDocument()
    })
  })

  it('should show Merge and Overwrite conflict resolution buttons in backup tab', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    expect(screen.getByRole('button', { name: /merge \(skip existing\)/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /overwrite \(replace\)/i })).toBeInTheDocument()
  })

  it('should toggle conflict mode when Merge button clicked', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    await user.click(screen.getByRole('button', { name: /merge \(skip existing\)/i }))
    // Merge is now selected (aria-pressed=true)
    expect(screen.getByRole('button', { name: /merge \(skip existing\)/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('should toggle conflict mode when Overwrite button clicked', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    await user.click(screen.getByRole('button', { name: /overwrite \(replace\)/i }))
    expect(screen.getByRole('button', { name: /overwrite \(replace\)/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('should show "Select a campaign first" hint when no game selected', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    expect(screen.getByText('Select a campaign first')).toBeInTheDocument()
  })

  it('should show backup error when file exceeds 10 MB', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    const largeFile = new File(['content'], 'backup.json', { type: 'application/json' })
    Object.defineProperty(largeFile, 'size', { value: 11 * 1024 * 1024 })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(fileInput, { target: { files: [largeFile] } })

    await waitFor(() => {
      expect(screen.getByText(/file exceeds 10 mb/i)).toBeInTheDocument()
    })
  })

  it('should show "Choose File" by default and filename after valid file selected', async () => {
    const user = userEvent.setup()
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    expect(screen.getByText('Choose File')).toBeInTheDocument()

    const validFile = new File(['{}'], 'backup.json', { type: 'application/json' })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(fileInput, { target: { files: [validFile] } })

    await waitFor(() => {
      expect(screen.getByText('backup.json')).toBeInTheDocument()
    })
  })

  it('should show import result summary after successful import', async () => {
    const user = userEvent.setup()
    vi.mocked(importGameBackup).mockResolvedValueOnce({
      sessions_created: 3,
      sessions_skipped: 1,
      sessions_overwritten: 0,
      notes_created: 5,
      notes_skipped: 0,
      notes_overwritten: 2,
    })
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Test Game', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    // Wait for game option and select it
    await waitFor(() => expect(screen.getByRole('option', { name: 'Test Game' })).toBeInTheDocument())
    await user.selectOptions(screen.getByRole('combobox'), 'game-1')

    // Select conflict mode
    await user.click(screen.getByRole('button', { name: /merge/i }))

    // Select a valid file
    const validFile = new File(['{}'], 'backup.json', { type: 'application/json' })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(fileInput, { target: { files: [validFile] } })

    // Import button should now be enabled
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /import chronicle/i })).not.toBeDisabled()
    })
    await user.click(screen.getByRole('button', { name: /import chronicle/i }))

    await waitFor(() => {
      expect(screen.getByText('Import Complete')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument() // sessions_created
      expect(screen.getByText('5')).toBeInTheDocument() // notes_created
    })
  })

  it('should show backup error when import fails', async () => {
    const user = userEvent.setup()
    vi.mocked(importGameBackup).mockRejectedValueOnce(new Error('Import failed'))
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Test Game', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    await waitFor(() => expect(screen.getByRole('option', { name: 'Test Game' })).toBeInTheDocument())
    await user.selectOptions(screen.getByRole('combobox'), 'game-1')
    await user.click(screen.getByRole('button', { name: /merge/i }))

    const validFile = new File(['{}'], 'backup.json', { type: 'application/json' })
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
    fireEvent.change(fileInput, { target: { files: [validFile] } })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /import chronicle/i })).not.toBeDisabled()
    })
    await user.click(screen.getByRole('button', { name: /import chronicle/i }))

    await waitFor(() => {
      expect(screen.getByText('Import failed')).toBeInTheDocument()
    })
  })

  it('should clear backupError when game select changes', async () => {
    const user = userEvent.setup()
    vi.mocked(exportGameBackup).mockRejectedValueOnce(new Error('Export failed'))
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Game One', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
      { id: 'game-2', title: 'Game Two', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))

    await waitFor(() => expect(screen.getByRole('option', { name: 'Game One' })).toBeInTheDocument())
    const gameSelect = screen.getByRole('combobox')
    await user.selectOptions(gameSelect, 'game-1')
    await waitFor(() => expect(screen.getByRole('button', { name: /export chronicle/i })).not.toBeDisabled())
    await user.click(screen.getByRole('button', { name: /export chronicle/i }))
    await waitFor(() => expect(screen.getByText('Export failed')).toBeInTheDocument())

    // Changing game select clears the error
    await user.selectOptions(gameSelect, 'game-2')
    await waitFor(() => {
      expect(screen.queryByText('Export failed')).not.toBeInTheDocument()
    })
  })

  it('should keep Import Chronicle button disabled when game but no file or mode', async () => {
    const user = userEvent.setup()
    vi.mocked(apiFetch).mockResolvedValueOnce([
      { id: 'game-1', title: 'Test Game', description: '', owner_id: 'u1', created_at: '', updated_at: '' },
    ])
    renderSettings()
    await user.click(screen.getByRole('tab', { name: /data backup/i }))
    await waitFor(() => expect(screen.getByRole('option', { name: 'Test Game' })).toBeInTheDocument())
    await user.selectOptions(screen.getByRole('combobox'), 'game-1')
    await waitFor(() => expect(screen.getByRole('button', { name: /import chronicle/i })).toBeDisabled())
  })
})

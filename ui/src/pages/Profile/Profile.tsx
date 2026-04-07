import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { apiFetch } from '../../api/client'
import { useDocumentTitle } from '../../hooks/useDocumentTitle'
import './Profile.css'

export default function Profile() {
  useDocumentTitle('Profile')
  const navigate = useNavigate()
  const { user, logout, refreshUser } = useAuth()

  const [editMode, setEditMode] = useState(false)
  const [username, setUsername] = useState(user?.username ?? '')
  const [avatarUrl, setAvatarUrl] = useState(user?.avatar_url ?? '')
  const [description, setDescription] = useState(user?.description ?? '')
  const [location, setLocation] = useState(user?.location ?? '')
  const [profileError, setProfileError] = useState<string | null>(null)
  const [profileSaving, setProfileSaving] = useState(false)

  const [pwSection, setPwSection] = useState(false)
  const [currentPw, setCurrentPw] = useState('')
  const [newPw, setNewPw] = useState('')
  const [confirmPw, setConfirmPw] = useState('')
  const [pwError, setPwError] = useState<string | null>(null)
  const [pwSaving, setPwSaving] = useState(false)

  if (!user) return null

  const handleProfileSave = async (e: React.FormEvent) => {
    e.preventDefault()
    setProfileError(null)
    setProfileSaving(true)
    try {
      await apiFetch(`/users/${user.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          username: username || undefined,
          avatar_url: avatarUrl || undefined,
          description: description || undefined,
          location: location || undefined,
        }),
      })
      await refreshUser()
      setEditMode(false)
    } catch (err: unknown) {
      setProfileError(err instanceof Error ? err.message : 'Failed to update profile.')
    } finally {
      setProfileSaving(false)
    }
  }

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault()
    setPwError(null)
    if (newPw !== confirmPw) {
      setPwError('New passphrases do not match.')
      return
    }
    setPwSaving(true)
    try {
      await apiFetch(`/users/${user.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ current_password: currentPw, password: newPw }),
      })
      await logout()
    } catch (err: unknown) {
      setPwError(err instanceof Error ? err.message : 'Failed to change passphrase.')
    } finally {
      setPwSaving(false)
    }
  }

  return (
    <div className="profile-page">
      <div className="profile-inner">
        <button className="profile-back-btn" onClick={() => navigate('/games')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Campaigns
        </button>

        <header className="profile-header">
          <div className="profile-title-rule" aria-hidden="true">
            <span /><span className="profile-ornament">✦</span><span />
          </div>
          <h1 className="profile-title">{user.username}</h1>
          <p className="profile-subtitle">Adventurer's Chronicle</p>
          <div className="profile-title-rule" aria-hidden="true">
            <span /><span className="profile-ornament">✦</span><span />
          </div>
        </header>

        <div className="profile-card">
          <div className="profile-section">
            <div className="profile-section-header">
              <h2 className="profile-section-title">Identity</h2>
              {!editMode && (
                <button className="profile-edit-btn" onClick={() => setEditMode(true)}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                  </svg>
                  Edit
                </button>
              )}
            </div>

            {!editMode ? (
              <>
                <div className="profile-readonly-field">
                  <span className="profile-field-label">Adventurer</span>
                  <span className="profile-field-value">{user.username}</span>
                </div>
                <div className="profile-readonly-field">
                  <span className="profile-field-label">Sending Stone</span>
                  <span className="profile-field-value">{user.email}</span>
                </div>
                {user.avatar_url && (
                  <div className="profile-readonly-field">
                    <span className="profile-field-label">Sigil</span>
                    <img className="profile-avatar" src={user.avatar_url} alt={`${user.username}'s avatar`} />
                  </div>
                )}
                {user.location && (
                  <div className="profile-readonly-field">
                    <span className="profile-field-label">Realm</span>
                    <span className="profile-field-value">{user.location}</span>
                  </div>
                )}
                {user.description && (
                  <div className="profile-readonly-field">
                    <span className="profile-field-label">Chronicle</span>
                    <p className="profile-field-value profile-description">{user.description}</p>
                  </div>
                )}
              </>
            ) : (
              <form className="profile-edit-form" onSubmit={handleProfileSave}>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="username">Adventurer</label>
                  <input
                    id="username"
                    type="text"
                    className="profile-input"
                    placeholder="Your name, adventurer..."
                    value={username}
                    onChange={e => setUsername(e.target.value)}
                    required
                  />
                </div>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="avatarUrl">Sigil URL</label>
                  <input
                    id="avatarUrl"
                    type="text"
                    className="profile-input"
                    placeholder="https://..."
                    value={avatarUrl}
                    onChange={e => setAvatarUrl(e.target.value)}
                  />
                </div>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="location">Realm</label>
                  <input
                    id="location"
                    type="text"
                    className="profile-input"
                    placeholder="Where do you hail from?"
                    value={location}
                    onChange={e => setLocation(e.target.value)}
                  />
                </div>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="description">Chronicle</label>
                  <textarea
                    id="description"
                    className="profile-input profile-textarea"
                    placeholder="Tell your tale..."
                    rows={4}
                    value={description}
                    onChange={e => setDescription(e.target.value)}
                  />
                </div>
                {profileError && <div className="profile-error" role="alert">{profileError}</div>}
                <div className="profile-edit-actions">
                  <button type="button" className="profile-cancel-btn" onClick={() => { setEditMode(false); setUsername(user.username); setAvatarUrl(user.avatar_url ?? ''); setLocation(user.location ?? ''); setDescription(user.description ?? ''); setProfileError(null) }} disabled={profileSaving}>
                    Cancel
                  </button>
                  <button type="submit" className="profile-save-btn" disabled={profileSaving}>
                    {profileSaving ? 'Inscribing…' : 'Inscribe Changes'}
                  </button>
                </div>
              </form>
            )}
          </div>

          <div className="profile-divider" aria-hidden="true">
            <span /><span>✦</span><span />
          </div>

          <div className="profile-section">
            <div className="profile-section-header">
              <h2 className="profile-section-title">Change Passphrase</h2>
              <button
                className="profile-edit-btn"
                onClick={() => setPwSection(p => !p)}
                type="button"
              >
                {pwSection ? 'Cancel' : 'Change'}
              </button>
            </div>

            {pwSection && (
              <form className="profile-edit-form" onSubmit={handlePasswordChange}>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="currentPw">Current Passphrase</label>
                  <input
                    id="currentPw"
                    type="password"
                    className="profile-input"
                    placeholder="Your current passphrase..."
                    autoComplete="current-password"
                    value={currentPw}
                    onChange={e => setCurrentPw(e.target.value)}
                    required
                  />
                </div>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="newPw">New Passphrase</label>
                  <input
                    id="newPw"
                    type="password"
                    className="profile-input"
                    placeholder="The new arcane word..."
                    autoComplete="new-password"
                    value={newPw}
                    onChange={e => setNewPw(e.target.value)}
                    required
                  />
                </div>
                <div className="profile-edit-field">
                  <label className="profile-field-label" htmlFor="confirmPw">Confirm Passphrase</label>
                  <input
                    id="confirmPw"
                    type="password"
                    className="profile-input"
                    placeholder="Once more..."
                    autoComplete="new-password"
                    value={confirmPw}
                    onChange={e => setConfirmPw(e.target.value)}
                    required
                  />
                </div>
                {pwError && <div className="profile-error" role="alert">{pwError}</div>}
                <p className="profile-pw-warning">⚠ Changing your passphrase will end your current session.</p>
                <div className="profile-edit-actions">
                  <button type="submit" className="profile-save-btn" disabled={pwSaving}>
                    {pwSaving ? 'Sealing the scroll…' : 'Seal the Scroll'}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

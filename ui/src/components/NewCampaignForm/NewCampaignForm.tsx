import { useState, useEffect } from 'react'
import { apiFetch } from '../../api/client'
import type { Game } from '../../types/game'
import type { User } from '../../types/user'
import { useAuth } from '../../context/AuthContext'
import UserSearch from '../UserSearch/UserSearch'
import './NewCampaignForm.css'

interface NewCampaignFormProps {
  onSuccess: (gameId: string, title: string) => void
  onDirtyChange: (isDirty: boolean) => void
}

export default function NewCampaignForm({ onSuccess, onDirtyChange }: NewCampaignFormProps) {
  const { user: currentUser } = useAuth()

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [splashImageUrl, setSplashImageUrl] = useState('')
  const [members, setMembers] = useState<Array<{ user: User; isGm: boolean }>>(() =>
    currentUser ? [{ user: currentUser, isGm: true }] : []
  )
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [apiError, setApiError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    onDirtyChange(title !== '' || description !== '' || splashImageUrl !== '' || members.length > 1)
  }, [title, description, splashImageUrl, members.length, onDirtyChange])

  function handleAddMember(user: User) {
    setMembers((prev) => {
      if (prev.some((m) => m.user.id === user.id)) return prev
      return [...prev, { user, isGm: false }]
    })
  }

  function handleRemoveMember(id: string) {
    if (id === currentUser?.id) return
    setMembers((prev) => prev.filter((m) => m.user.id !== id))
  }

  function handleToggleRole(id: string) {
    if (id === currentUser?.id) return
    setMembers((prev) =>
      prev.map((m) => (m.user.id === id ? { ...m, isGm: !m.isGm } : m))
    )
  }

  function validate(): Record<string, string> {
    const errs: Record<string, string> = {}
    if (!title.trim()) {
      errs.title = 'Title is required'
    } else if (title.length > 100) {
      errs.title = 'Title must be 100 characters or fewer'
    }
    if (splashImageUrl && !/^https?:\/\//.test(splashImageUrl)) {
      errs.splashImageUrl = 'Must be a valid URL starting with http:// or https://'
    }
    return errs
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const errs = validate()
    if (Object.keys(errs).length > 0) {
      setErrors(errs)
      return
    }
    setErrors({})
    setApiError(null)
    setSubmitting(true)
    try {
      const game = await apiFetch<Game>('/games', {
        method: 'POST',
        body: JSON.stringify({
          title,
          description: description || null,
          splash_image_url: splashImageUrl || null,
          members: members.map((m) => ({ user_id: m.user.id, is_gm: m.isGm })),
        }),
      })
      onSuccess(game.id, game.title)
    } catch (err: unknown) {
      setApiError(err instanceof Error ? err.message : 'Failed to create campaign.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="ncf" onSubmit={handleSubmit} noValidate>
      {apiError && (
        <div className="ncf-api-error" role="alert">
          {apiError}
        </div>
      )}

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ncf-title">Title</label>
        <input
          id="ncf-title"
          className={`ncf-input ${errors.title ? 'ncf-input--error' : ''}`}
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          aria-describedby={errors.title ? 'ncf-title-error' : undefined}
          autoFocus
          placeholder="The Shattered Throne"
          maxLength={120}
        />
        {errors.title && (
          <p id="ncf-title-error" className="ncf-field-error" role="alert">{errors.title}</p>
        )}
      </div>

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ncf-description">Description <span className="ncf-optional">(optional)</span></label>
        <textarea
          id="ncf-description"
          className="ncf-textarea"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="A tale of heroes, shadows, and ancient power..."
          rows={3}
        />
      </div>

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ncf-splash">Splash Image URL <span className="ncf-optional">(optional)</span></label>
        <input
          id="ncf-splash"
          className={`ncf-input ${errors.splashImageUrl ? 'ncf-input--error' : ''}`}
          type="url"
          value={splashImageUrl}
          onChange={(e) => setSplashImageUrl(e.target.value)}
          aria-describedby={errors.splashImageUrl ? 'ncf-splash-error' : undefined}
          placeholder="https://example.com/banner.jpg"
        />
        {errors.splashImageUrl && (
          <p id="ncf-splash-error" className="ncf-field-error" role="alert">{errors.splashImageUrl}</p>
        )}
      </div>

      <div className="ncf-field">
        <span className="ncf-members-label">Members</span>
        <UserSearch
          excludeIds={members.map((m) => m.user.id)}
          onSelect={handleAddMember}
        />
        {members.length > 0 && (
          <ul className="ncf-member-list">
            {members.map(({ user, isGm }) => {
              const isCreator = user.id === currentUser?.id
              return (
                <li key={user.id} className="ncf-member-row">
                  <span className="ncf-member-name">
                    {user.username}
                    {isCreator && <span className="ncf-creator-badge">Creator</span>}
                  </span>
                  <div className={`ncf-role-toggle${isCreator ? ' ncf-role-toggle--locked' : ''}`}
                       title={isCreator ? 'The game creator must remain a GM' : undefined}>
                    <button
                      type="button"
                      className={`ncf-role-btn${!isGm ? ' ncf-role-btn--active' : ''}`}
                      onClick={() => handleToggleRole(user.id)}
                      disabled={isCreator}
                    >
                      Player
                    </button>
                    <button
                      type="button"
                      className={`ncf-role-btn${isGm ? ' ncf-role-btn--active' : ''}`}
                      onClick={() => handleToggleRole(user.id)}
                      disabled={isCreator}
                    >
                      GM
                    </button>
                  </div>
                  <button
                    type="button"
                    className="ncf-member-remove"
                    onClick={() => handleRemoveMember(user.id)}
                    aria-label={isCreator ? 'The game creator must remain a GM' : `Remove ${user.username}`}
                    disabled={isCreator}
                    title={isCreator ? 'The game creator must remain a GM' : undefined}
                  >
                    &times;
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      <div className="ncf-actions">
        <button
          type="submit"
          className="ncf-submit"
          disabled={submitting}
        >
          {submitting ? (
            <><span className="ncf-spinner" aria-hidden="true" />Creating&hellip;</>
          ) : (
            'Create Campaign'
          )}
        </button>
      </div>
    </form>
  )
}

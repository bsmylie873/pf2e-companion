import { useState, useEffect, useCallback } from 'react'
import { apiFetch } from '../../api/client'
import type { Game } from '../../types/game'
import type { User } from '../../types/user'
import type { GameMembership } from '../../types/membership'
import UserSearch from '../UserSearch/UserSearch'
import '../NewCampaignForm/NewCampaignForm.css'
import './EditCampaignForm.css'

interface EditCampaignFormProps {
  gameId: string
  onSuccess: (title: string) => void
  onDirtyChange: (isDirty: boolean) => void
}

interface MemberEntry {
  user: User
  isGm: boolean
  membershipId?: string
}

export default function EditCampaignForm({ gameId, onSuccess, onDirtyChange }: EditCampaignFormProps) {
  const [loading, setLoading] = useState(true)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [splashImageUrl, setSplashImageUrl] = useState('')
  const [members, setMembers] = useState<MemberEntry[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [apiError, setApiError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [initialTitle, setInitialTitle] = useState('')
  const [initialDescription, setInitialDescription] = useState('')
  const [initialSplashImageUrl, setInitialSplashImageUrl] = useState('')
  const [initialMembers, setInitialMembers] = useState<MemberEntry[]>([])

  useEffect(() => {
    async function fetchData() {
      try {
        const [game, memberships, users] = await Promise.all([
          apiFetch<Game>(`/games/${gameId}`),
          apiFetch<GameMembership[]>(`/memberships?game_id=${gameId}`),
          apiFetch<User[]>('/users'),
        ])

        setTitle(game.title)
        setDescription(game.description ?? '')
        setSplashImageUrl(game.splash_image_url ?? '')
        setInitialTitle(game.title)
        setInitialDescription(game.description ?? '')
        setInitialSplashImageUrl(game.splash_image_url ?? '')

        const usersMap = new Map(users.map((u) => [u.id, u]))
        const resolvedMembers: MemberEntry[] = memberships.reduce<MemberEntry[]>((acc, m) => {
          const user = usersMap.get(m.user_id)
          if (user) acc.push({ user, isGm: m.is_gm, membershipId: m.id })
          return acc
        }, [])

        setMembers(resolvedMembers)
        setInitialMembers(resolvedMembers)
      } catch (err: unknown) {
        setApiError(err instanceof Error ? err.message : 'Failed to load campaign.')
      } finally {
        setLoading(false)
      }
    }
    fetchData()
  }, [gameId])

  const isDirty = useCallback((): boolean => {
    if (title !== initialTitle) return true
    if (description !== initialDescription) return true
    if (splashImageUrl !== initialSplashImageUrl) return true
    if (members.length !== initialMembers.length) return true
    const initialIds = new Set(initialMembers.map((m) => m.user.id))
    const currentIds = new Set(members.map((m) => m.user.id))
    for (const id of Array.from(currentIds)) {
      if (!initialIds.has(id)) return true
    }
    for (const id of Array.from(initialIds)) {
      if (!currentIds.has(id)) return true
    }
    for (const m of members) {
      const init = initialMembers.find((im) => im.user.id === m.user.id)
      if (init && init.isGm !== m.isGm) return true
    }
    return false
  }, [title, description, splashImageUrl, members, initialTitle, initialDescription, initialSplashImageUrl, initialMembers])

  useEffect(() => {
    if (!loading) onDirtyChange(isDirty())
  }, [loading, isDirty, onDirtyChange])

  function handleAddMember(user: User) {
    setMembers((prev) => {
      if (prev.some((m) => m.user.id === user.id)) return prev
      return [...prev, { user, isGm: false }]
    })
  }

  function handleRemoveMember(id: string) {
    setMembers((prev) => prev.filter((m) => m.user.id !== id))
  }

  function handleToggleRole(id: string) {
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
      const gameChanged =
        title !== initialTitle ||
        description !== initialDescription ||
        splashImageUrl !== initialSplashImageUrl
      if (gameChanged) {
        const body: Record<string, unknown> = {}
        if (title !== initialTitle) body.title = title
        if (description !== initialDescription) body.description = description || null
        if (splashImageUrl !== initialSplashImageUrl) body.splash_image_url = splashImageUrl || null
        await apiFetch(`/games/${gameId}`, { method: 'PATCH', body: JSON.stringify(body) })
      }

      const initialMap = new Map(initialMembers.map((m) => [m.user.id, m]))
      const currentMap = new Map(members.map((m) => [m.user.id, m]))

      const memberOps: Promise<unknown>[] = []

      for (const m of members) {
        if (!initialMap.has(m.user.id)) {
          memberOps.push(
            apiFetch('/memberships', {
              method: 'POST',
              body: JSON.stringify({ game_id: gameId, user_id: m.user.id, is_gm: m.isGm }),
            })
          )
        }
      }

      for (const im of initialMembers) {
        const current = currentMap.get(im.user.id)
        if (!current) {
          if (im.membershipId) {
            memberOps.push(
              apiFetch(`/memberships/${im.membershipId}`, { method: 'DELETE' })
            )
          }
        } else if (current.isGm !== im.isGm && im.membershipId) {
          memberOps.push(
            apiFetch(`/memberships/${im.membershipId}`, {
              method: 'PATCH',
              body: JSON.stringify({ is_gm: current.isGm }),
            })
          )
        }
      }

      await Promise.all(memberOps)
      onSuccess(title)
    } catch (err: unknown) {
      setApiError(err instanceof Error ? err.message : 'Failed to save campaign.')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="ecf-loading">
        <span className="ncf-spinner" aria-hidden="true" />
        Loading campaign...
      </div>
    )
  }

  return (
    <form className="ncf" onSubmit={handleSubmit} noValidate>
      {apiError && (
        <div className="ncf-api-error" role="alert">
          {apiError}
        </div>
      )}

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ecf-title">Title</label>
        <input
          id="ecf-title"
          className={`ncf-input ${errors.title ? 'ncf-input--error' : ''}`}
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          aria-describedby={errors.title ? 'ecf-title-error' : undefined}
          autoFocus
          placeholder="The Shattered Throne"
          maxLength={120}
        />
        {errors.title && (
          <p id="ecf-title-error" className="ncf-field-error" role="alert">{errors.title}</p>
        )}
      </div>

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ecf-description">
          Description <span className="ncf-optional">(optional)</span>
        </label>
        <textarea
          id="ecf-description"
          className="ncf-textarea"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="A tale of heroes, shadows, and ancient power..."
          rows={3}
        />
      </div>

      <div className="ncf-field">
        <label className="ncf-label" htmlFor="ecf-splash">
          Splash Image URL <span className="ncf-optional">(optional)</span>
        </label>
        <input
          id="ecf-splash"
          className={`ncf-input ${errors.splashImageUrl ? 'ncf-input--error' : ''}`}
          type="url"
          value={splashImageUrl}
          onChange={(e) => setSplashImageUrl(e.target.value)}
          aria-describedby={errors.splashImageUrl ? 'ecf-splash-error' : undefined}
          placeholder="https://example.com/banner.jpg"
        />
        {errors.splashImageUrl && (
          <p id="ecf-splash-error" className="ncf-field-error" role="alert">{errors.splashImageUrl}</p>
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
            {members.map(({ user, isGm }) => (
              <li key={user.id} className="ncf-member-row">
                <span className="ncf-member-name">{user.username}</span>
                <div className="ncf-role-toggle">
                  <button
                    type="button"
                    className={`ncf-role-btn${!isGm ? ' ncf-role-btn--active' : ''}`}
                    onClick={() => handleToggleRole(user.id)}
                  >
                    Player
                  </button>
                  <button
                    type="button"
                    className={`ncf-role-btn${isGm ? ' ncf-role-btn--active' : ''}`}
                    onClick={() => handleToggleRole(user.id)}
                  >
                    GM
                  </button>
                </div>
                <button
                  type="button"
                  className="ncf-member-remove"
                  onClick={() => handleRemoveMember(user.id)}
                  aria-label={`Remove ${user.username}`}
                >
                  &times;
                </button>
              </li>
            ))}
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
            <><span className="ncf-spinner" aria-hidden="true" />Saving&hellip;</>
          ) : (
            'Save Changes'
          )}
        </button>
      </div>
    </form>
  )
}

import { useEffect, useState } from 'react'
import { apiFetch } from '../../api/client'
import type { Game } from '../../types/game'
import GameCard from '../../components/GameCard/GameCard'
import { useLocalStorage } from '../../hooks/useLocalStorage'
import './GamesList.css'

type Layout = 'grid' | 'list'

function GridIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  )
}

function ListIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  )
}

export default function GamesList() {
  const [games, setGames] = useState<Game[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [layout, setLayout] = useLocalStorage<Layout>('pf2e-layout-pref', 'grid')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    apiFetch<Game[]>('/games')
      .then((data) => {
        if (!cancelled) {
          setGames(data)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load campaigns.')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => { cancelled = true }
  }, [])

  return (
    <div className="gameslist-page">
      <div className="gameslist-header">
        <div className="gameslist-title-group">
          <span className="gameslist-ornament" aria-hidden="true">⚔</span>
          <h1 className="gameslist-title">Your Campaigns</h1>
          <span className="gameslist-ornament" aria-hidden="true">⚔</span>
        </div>

        <div className="gameslist-controls">
          <button
            className={`layout-btn ${layout === 'grid' ? 'layout-btn--active' : ''}`}
            onClick={() => setLayout('grid')}
            aria-label="Grid view"
            title="Grid view"
          >
            <GridIcon />
          </button>
          <button
            className={`layout-btn ${layout === 'list' ? 'layout-btn--active' : ''}`}
            onClick={() => setLayout('list')}
            aria-label="List view"
            title="List view"
          >
            <ListIcon />
          </button>
        </div>
      </div>

      <div className="gameslist-content">
        {loading && (
          <div className="gameslist-loading">
            <div className="spinner-ring" />
            <p className="spinner-label">Consulting the chronicles…</p>
          </div>
        )}

        {!loading && error && (
          <div className="gameslist-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <p>{error}</p>
          </div>
        )}

        {!loading && !error && games.length === 0 && (
          <div className="gameslist-empty">
            <div className="empty-sigil" aria-hidden="true">✦</div>
            <p className="empty-message">No campaigns found.</p>
            <p className="empty-sub">The realm awaits your stories.</p>
          </div>
        )}

        {!loading && !error && games.length > 0 && (
          <div className={`gameslist-grid gameslist-grid--${layout}`}>
            {games.map((game) => (
              <GameCard key={game.id} game={game} mode={layout} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

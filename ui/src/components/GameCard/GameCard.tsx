import { useNavigate } from 'react-router-dom'
import type { Game } from '../../types/game'
import './GameCard.css'

interface GameCardProps {
  game: Game
  mode: 'grid' | 'list'
}

function FallbackArt({ size }: { size: 'large' | 'small' }) {
  return (
    <div className={`gamecard-fallback gamecard-fallback--${size}`}>
      <svg
        className="gamecard-fallback-sigil"
        viewBox="0 0 100 100"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Ornamental compass rose / sigil */}
        <circle cx="50" cy="50" r="30" stroke="currentColor" strokeWidth="0.75" strokeDasharray="4 3" opacity="0.5" />
        <circle cx="50" cy="50" r="18" stroke="currentColor" strokeWidth="0.75" opacity="0.6" />
        <path d="M50 20 L54 46 L50 50 L46 46 Z" fill="currentColor" opacity="0.7" />
        <path d="M80 50 L54 54 L50 50 L54 46 Z" fill="currentColor" opacity="0.5" />
        <path d="M50 80 L46 54 L50 50 L54 54 Z" fill="currentColor" opacity="0.7" />
        <path d="M20 50 L46 46 L50 50 L46 54 Z" fill="currentColor" opacity="0.5" />
        <circle cx="50" cy="50" r="4" fill="currentColor" opacity="0.8" />
        <path d="M50 8 L50 16 M50 84 L50 92 M8 50 L16 50 M84 50 L92 50" stroke="currentColor" strokeWidth="0.75" opacity="0.4" />
      </svg>
    </div>
  )
}

export default function GameCard({ game, mode }: GameCardProps) {
  const navigate = useNavigate()

  const handleClick = () => {
    navigate(`/games/${game.id}`, { state: { title: game.title } })
  }

  if (mode === 'list') {
    return (
      <article className="gamecard gamecard--list" onClick={handleClick} role="button" tabIndex={0}>
        <div className="gamecard-list-thumb">
          {game.splash_image_url ? (
            <img src={game.splash_image_url} alt={game.title} className="gamecard-list-img" />
          ) : (
            <FallbackArt size="small" />
          )}
        </div>
        <div className="gamecard-list-body">
          <h3 className="gamecard-list-title">{game.title}</h3>
          {game.description && (
            <p className="gamecard-list-desc">{game.description}</p>
          )}
        </div>
        <div className="gamecard-list-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M9 18l6-6-6-6" />
          </svg>
        </div>
      </article>
    )
  }

  return (
    <article className="gamecard gamecard--grid" onClick={handleClick} role="button" tabIndex={0}>
      <div className="gamecard-image-wrap">
        {game.splash_image_url ? (
          <img src={game.splash_image_url} alt={game.title} className="gamecard-img" />
        ) : (
          <FallbackArt size="large" />
        )}
      </div>
      <div className="gamecard-content">
        <h3 className="gamecard-title">{game.title}</h3>
        {game.description && (
          <p className="gamecard-desc">{game.description}</p>
        )}
      </div>
    </article>
  )
}

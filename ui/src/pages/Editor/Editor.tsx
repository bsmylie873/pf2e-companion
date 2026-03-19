import { useParams, useLocation, useNavigate } from 'react-router-dom'
import './Editor.css'

interface LocationState {
  title?: string
}

export default function Editor() {
  const { gameId } = useParams<{ gameId: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const state = location.state as LocationState | null

  const title = state?.title ?? `Game #${gameId}`

  return (
    <div className="editor-page">
      <div className="editor-inner">
        <button className="editor-back-btn" onClick={() => navigate('/games')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
          Back to Campaigns
        </button>

        <header className="editor-header">
          <div className="editor-title-rule" aria-hidden="true">
            <span /><span className="editor-title-ornament">✦</span><span />
          </div>
          <h1 className="editor-title">{title}</h1>
          <div className="editor-title-rule" aria-hidden="true">
            <span /><span className="editor-title-ornament">✦</span><span />
          </div>
        </header>

        <div className="editor-coming-soon">
          <div className="editor-scroll-icon" aria-hidden="true">
            <svg viewBox="0 0 80 100" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="8" y="8" width="64" height="84" rx="4" stroke="currentColor" strokeWidth="1.5" />
              <line x1="20" y1="28" x2="60" y2="28" stroke="currentColor" strokeWidth="1" strokeDasharray="3 3" opacity="0.5" />
              <line x1="20" y1="40" x2="60" y2="40" stroke="currentColor" strokeWidth="1" strokeDasharray="3 3" opacity="0.5" />
              <line x1="20" y1="52" x2="50" y2="52" stroke="currentColor" strokeWidth="1" strokeDasharray="3 3" opacity="0.5" />
              <line x1="20" y1="64" x2="55" y2="64" stroke="currentColor" strokeWidth="1" strokeDasharray="3 3" opacity="0.5" />
              <path d="M35 80 L40 72 L45 80" stroke="currentColor" strokeWidth="1" opacity="0.6" strokeLinejoin="round" />
            </svg>
          </div>

          <h2 className="editor-coming-title">The Chapter Awaits</h2>
          <p className="editor-coming-text">
            The campaign editor is being inscribed by our finest scribes.
            <br />
            Return when the ink has dried.
          </p>

          <div className="editor-coming-badge">
            <span>Coming Soon</span>
          </div>
        </div>
      </div>
    </div>
  )
}

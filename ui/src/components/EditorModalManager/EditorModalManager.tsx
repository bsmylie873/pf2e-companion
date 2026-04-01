import { useState } from 'react'
import EditorModal from '../EditorModal/EditorModal'
import './EditorModalManager.css'

interface OpenItem {
  type: 'session' | 'note'
  itemId: string
  label: string
}

interface EditorModalManagerProps {
  items: OpenItem[]
  gameId: string
  onClose: (itemId: string) => void
  onCloseAll: () => void
}

export default function EditorModalManager({ items, gameId, onClose, onCloseAll }: EditorModalManagerProps) {
  const [activeItemId, setActiveItemId] = useState<string>(() => items[0]?.itemId ?? '')

  // Keep activeItemId valid when items change
  const activeId = items.some(i => i.itemId === activeItemId)
    ? activeItemId
    : (items[0]?.itemId ?? '')

  const activeItem = items.find(i => i.itemId === activeId)

  if (items.length === 0 || !activeItem) return null

  const handleClose = (itemId: string) => {
    if (itemId === activeId) {
      const idx = items.findIndex(i => i.itemId === itemId)
      const next = items[idx + 1] ?? items[idx - 1]
      if (next) setActiveItemId(next.itemId)
    }
    onClose(itemId)
  }

  return (
    <div className="emm-root">
      {items.length > 1 && (
        <div className="emm-tab-strip" role="tablist">
          {items.map(item => (
            <div
              key={item.itemId}
              className={`emm-tab${item.itemId === activeId ? ' emm-tab--active' : ''}`}
              role="tab"
              aria-selected={item.itemId === activeId}
            >
              <button
                className="emm-tab-btn"
                onClick={() => setActiveItemId(item.itemId)}
              >
                <span className="emm-tab-type">{item.type === 'session' ? '⚔' : '📜'}</span>
                <span className="emm-tab-label">{item.label}</span>
              </button>
              <button
                className="emm-tab-close"
                onClick={() => handleClose(item.itemId)}
                aria-label={`Close ${item.label}`}
              >
                <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                  <line x1="9" y1="3" x2="3" y2="9" />
                  <line x1="3" y1="3" x2="9" y2="9" />
                </svg>
              </button>
            </div>
          ))}
          <button className="emm-close-all" onClick={onCloseAll} title="Close all">
            Close all
          </button>
        </div>
      )}
      {items.map(item => (
        <div
          key={item.itemId}
          style={{ display: item.itemId === activeId ? 'contents' : 'none' }}
        >
          <EditorModal
            type={item.type}
            itemId={item.itemId}
            gameId={gameId}
            onClose={() => handleClose(item.itemId)}
          />
        </div>
      ))}
    </div>
  )
}

import './Modal.css'

interface ModalProps {
  title: string
  onClose: () => void
}

export default function Modal({ title, onClose }: ModalProps) {
  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-card" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">{title}</h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div className="modal-body">
          <div className="modal-coming-soon">
            <div className="modal-sigil">✦</div>
            <p className="modal-message">
              The scribes are still writing this chapter.
            </p>
            <p className="modal-sub">Coming Soon</p>
          </div>
        </div>
      </div>
    </div>
  )
}

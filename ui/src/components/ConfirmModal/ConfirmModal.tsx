import '../Modal/Modal.css'
import './ConfirmModal.css'

interface ConfirmModalProps {
  title: string
  message: string
  confirmLabel: string
  error: string | null
  loading: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function ConfirmModal({ title, message, confirmLabel, error, loading, onConfirm, onCancel }: ConfirmModalProps) {
  return (
    <div className="modal-backdrop" onClick={onCancel}>
      <div className="modal-card confirm-modal-card" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">{title}</h2>
          <button className="modal-close" onClick={onCancel} aria-label="Close">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className="confirm-modal-body">
          <p className="confirm-modal-message">{message}</p>

          {error && (
            <div className="confirm-modal-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <p>{error}</p>
            </div>
          )}

          <div className="confirm-modal-actions">
            <button className="confirm-modal-btn confirm-modal-btn--cancel" onClick={onCancel} type="button">
              Cancel
            </button>
            <button
              className="confirm-modal-btn confirm-modal-btn--confirm"
              onClick={onConfirm}
              disabled={loading}
              type="button"
            >
              {loading ? 'Deleting…' : confirmLabel}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

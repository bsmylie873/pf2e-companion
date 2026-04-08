import type { ReactZoomPanPinchRef } from 'react-zoom-pan-pinch'
import './MapOverlayPanel.css'

interface MapOverlayPanelProps {
  gameTitle: string
  isGM: boolean
  uploading: boolean
  uploadError: string | null
  displayScale: number
  panelOpen: boolean
  unpinnedSessionsCount: number
  pinsCount: number
  transformRef: React.RefObject<ReactZoomPanPinchRef | null>
  fileInputRef: React.RefObject<HTMLInputElement | null>
  onClose: () => void
  onOpen: () => void
  onUploadClick: () => void
  onDeleteMap: () => void
  onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}

export default function MapOverlayPanel({
  gameTitle,
  isGM,
  uploading,
  uploadError,
  displayScale,
  panelOpen,
  unpinnedSessionsCount,
  pinsCount,
  transformRef,
  fileInputRef,
  onClose,
  onOpen,
  onUploadClick,
  onDeleteMap,
  onFileChange,
}: MapOverlayPanelProps) {
  return (
    <>
      {!panelOpen && (
        <button
          className="map-panel-toggle"
          onClick={onOpen}
          title="Show controls"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M9 18l6-6-6-6" />
          </svg>
        </button>
      )}

      {panelOpen && (
        <div className="map-overlay-panel">
          <button
            className="map-panel-close"
            onClick={onClose}
            title="Hide controls"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M15 18l-6-6 6-6" />
            </svg>
          </button>

          <h2 className="map-panel-title">{gameTitle}</h2>

          {isGM && (
            <div className="map-toolbar">
              {uploadError && <span className="map-upload-error">{uploadError}</span>}
              <button className="map-upload-btn" onClick={onUploadClick} disabled={uploading}>
                {uploading ? 'Uploading…' : 'Replace Map'}
              </button>
              <button className="map-delete-btn" onClick={onDeleteMap}>
                Remove Map
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="map-file-input"
                onChange={onFileChange}
              />
            </div>
          )}

          <p className="map-gm-hint">
            Click anywhere on the map to place a pin. Drag pins to reposition.
            Right-click and drag to pan. Scroll to zoom.
          </p>

          <div className="map-zoom-controls">
            <button
              className="map-zoom-btn"
              onClick={() => transformRef.current?.zoomIn()}
              disabled={displayScale >= 5}
              title="Zoom in"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <line x1="12" y1="5" x2="12" y2="19" />
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
            </button>
            <button
              className="map-zoom-level"
              onClick={() => transformRef.current?.resetTransform()}
              title="Reset zoom"
            >
              {Math.round(displayScale * 100)}%
            </button>
            <button
              className="map-zoom-btn"
              onClick={() => transformRef.current?.zoomOut()}
              disabled={displayScale <= 0.75}
              title="Zoom out"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <line x1="5" y1="12" x2="19" y2="12" />
              </svg>
            </button>
          </div>

          {unpinnedSessionsCount === 0 && pinsCount > 0 && (
            <p className="map-all-pinned">✦ All sessions are pinned on the map.</p>
          )}
        </div>
      )}
    </>
  )
}

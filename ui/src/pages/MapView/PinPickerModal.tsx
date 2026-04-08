import { createPortal } from 'react-dom'
import { PIN_COLOURS, PIN_ICONS, COLOUR_MAP, PIN_ICON_COMPONENTS, PIN_ICON_LABELS } from '../../constants/pins'
import type { PinColour, PinIcon } from '../../constants/pins'
import type { Session } from '../../types/session'
import type { Note } from '../../types/note'
import './PinPickerModal.css'

interface PinPickerModalProps {
  pendingColour: PinColour
  pendingIcon: PinIcon
  pendingLabel: string
  pendingDescription: string
  pickerSearch: string
  unpinnedSessions: Session[]
  notes: Note[]
  dropLinkedItem?: { type: 'session' | 'note'; id: string; label: string } | null
  onClose: () => void
  onColourChange: (c: PinColour) => void
  onIconChange: (i: PinIcon) => void
  onLabelChange: (v: string) => void
  onDescriptionChange: (v: string) => void
  onSearchChange: (v: string) => void
  onCreateMarker: (label: string, description: string) => void
  onSelectSession: (session: Session) => void
  onSelectNote: (note: Note) => void
}

export default function PinPickerModal({
  pendingColour,
  pendingIcon,
  pendingLabel,
  pendingDescription,
  pickerSearch,
  unpinnedSessions,
  notes,
  dropLinkedItem,
  onClose,
  onColourChange,
  onIconChange,
  onLabelChange,
  onDescriptionChange,
  onSearchChange,
  onCreateMarker,
  onSelectSession,
  onSelectNote,
}: PinPickerModalProps) {
  return createPortal(
    <div className='map-overlay' onClick={onClose}>
      <div className='map-session-picker' onClick={e => e.stopPropagation()}>
        <div className='map-picker-header'>
          <span className='map-picker-rune' aria-hidden='true'>⬡</span>
          <h3 className='map-picker-title'>Mark This Location</h3>
        </div>

        <div className='map-picker-customise'>
          <span className='map-picker-customise-label'>Pin colour:</span>
          <div className='pin-colour-palette'>
            {PIN_COLOURS.map(c => (
              <button
                key={c}
                className={`pin-colour-swatch${pendingColour === c ? ' pin-colour-swatch--selected' : ''}`}
                style={{ '--swatch-colour': COLOUR_MAP[c] } as React.CSSProperties}
                onClick={() => onColourChange(c)}
                title={c}
              />
            ))}
          </div>
          <span className='map-picker-customise-label'>Pin icon:</span>
          <div className='pin-colour-palette'>
            {PIN_ICONS.map(i => {
              const IconComp = PIN_ICON_COMPONENTS[i]
              return (
                <button
                  key={i}
                  className={`pin-icon-option${pendingIcon === i ? ' pin-icon-option--selected' : ''}`}
                  onClick={() => onIconChange(i)}
                  title={PIN_ICON_LABELS[i]}
                  aria-label={PIN_ICON_LABELS[i]}
                >
                  <IconComp size={16} />
                </button>
              )
            })}
          </div>
        </div>

        <div className='map-picker-customise'>
          <span className='map-picker-customise-label'>Label (optional):</span>
          <input
            className='map-marker-input'
            type='text'
            placeholder='Pin label…'
            value={pendingLabel}
            onChange={e => onLabelChange(e.target.value)}
            maxLength={100}
          />
          <span className='map-picker-customise-label'>Description (optional):</span>
          <textarea
            className='map-marker-textarea'
            placeholder='Description…'
            rows={2}
            value={pendingDescription}
            onChange={e => onDescriptionChange(e.target.value)}
            maxLength={1000}
          />
        </div>

        {dropLinkedItem && (
          <div className="map-picker-prelinked">
            <span className="map-picker-prelinked-label">
              Linking to {dropLinkedItem.type}: <strong>{dropLinkedItem.label}</strong>
            </span>
            <button
              className="map-marker-submit map-marker-submit--full map-marker-submit--linked"
              onClick={() => {
                if (dropLinkedItem.type === 'session') {
                  const session = unpinnedSessions.find(s => s.id === dropLinkedItem.id)
                  if (session) onSelectSession(session)
                } else {
                  const note = notes.find(n => n.id === dropLinkedItem.id)
                  if (note) onSelectNote(note)
                }
              }}
            >
              Place Pin &amp; Link {dropLinkedItem.type === 'session' ? 'Session' : 'Note'}
            </button>
          </div>
        )}

        <button
          className='map-marker-submit map-marker-submit--full'
          onClick={() => onCreateMarker(pendingLabel, pendingDescription)}
        >
          Place as Standalone Marker
        </button>

        {(unpinnedSessions.length > 0 || notes.length > 0) && (
          <div className='map-picker-search-section'>
            <span className='map-picker-customise-label'>Link to a session or note:</span>
            <input
              className='map-marker-input'
              type='text'
              placeholder='Search sessions &amp; notes…'
              value={pickerSearch}
              onChange={e => onSearchChange(e.target.value)}
            />
            {pickerSearch.trim() !== '' && (() => {
              const q = pickerSearch.trim().toLowerCase()
              const matchedSessions = unpinnedSessions.filter(s =>
                s.title.toLowerCase().includes(q) ||
                (s.session_number != null && `#${s.session_number}`.includes(q))
              )
              const matchedNotes = notes.filter(n => n.title.toLowerCase().includes(q))
              if (matchedSessions.length === 0 && matchedNotes.length === 0) {
                return <p className='map-picker-empty'>No matching sessions or notes.</p>
              }
              return (
                <ul className='map-picker-list'>
                  {matchedSessions.slice(0, 4).map(session => (
                    <li key={session.id}>
                      <button
                        className='map-picker-item'
                        onClick={() => onSelectSession(session)}
                      >
                        <span className='map-picker-item-type'>Session</span>
                        {session.session_number != null && (
                          <span className='map-picker-num'>#{session.session_number}</span>
                        )}
                        <span className='map-picker-name'>{session.title}</span>
                      </button>
                    </li>
                  ))}
                  {matchedNotes.slice(0, 4).map(note => (
                    <li key={note.id}>
                      <button
                        className='map-picker-item'
                        onClick={() => onSelectNote(note)}
                      >
                        <span className='map-picker-item-type'>Note</span>
                        <span className={`map-picker-vis map-picker-vis--${note.visibility}`}>
                          {note.visibility === 'private' ? '🔒' : note.visibility === 'visible' ? '👁' : '✏️'}
                        </span>
                        <span className='map-picker-name'>{note.title}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )
            })()}
          </div>
        )}
      </div>
    </div>,
    document.body,
  )
}

import { useEffect, useRef, useState } from 'react'

interface Props {
  value: string
  onCommit: (value: string) => void
  onCancel: () => void
  error?: string | null
  placeholder?: string
  autoFocus?: boolean
}

export default function InlineNameInput({ value, onCommit, onCancel, error, placeholder, autoFocus }: Props) {
  const [draft, setDraft] = useState(value)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (autoFocus) {
      inputRef.current?.focus()
      inputRef.current?.select()
    }
  }, [autoFocus])

  const commit = () => {
    const trimmed = draft.trim()
    if (trimmed && trimmed !== value) onCommit(trimmed)
    else if (!trimmed) onCancel()
  }

  return (
    <div className="folder-inline-wrap">
      <input
        ref={inputRef}
        className="folder-inline-input"
        value={draft}
        placeholder={placeholder}
        maxLength={100}
        onChange={e => setDraft(e.target.value)}
        onKeyDown={e => {
          if (e.key === 'Enter') { e.preventDefault(); commit() }
          if (e.key === 'Escape') { e.preventDefault(); onCancel() }
        }}
        onBlur={commit}
      />
      {error && <span className="folder-inline-error">{error}</span>}
    </div>
  )
}

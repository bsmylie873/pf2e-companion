import { useRef, useState, useCallback } from 'react'
import type { JSONContent } from '@tiptap/react'

export type AutosaveStatus = 'idle' | 'saving' | 'saved' | 'error'

export function useAutosave<T>(
  onSave: (content: JSONContent) => Promise<T>,
) {
  const [status, setStatus] = useState<AutosaveStatus>('idle')
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const scheduleAutosave = useCallback((content: JSONContent) => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(async () => {
      setStatus('saving')
      try {
        await onSave(content)
        setStatus('saved')
      } catch {
        setStatus('error')
      }
    }, 2000)
  }, [onSave])

  return { scheduleAutosave, status }
}

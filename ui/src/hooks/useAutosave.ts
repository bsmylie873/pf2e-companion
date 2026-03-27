import { useRef, useState, useCallback } from 'react'
import type { JSONContent } from '@tiptap/react'

export type AutosaveStatus = 'idle' | 'saving' | 'saved' | 'conflict' | 'error'

export function useAutosave<T extends { version: number }>(
  onSave: (content: JSONContent, version: number) => Promise<T>,
  initialVersion: number,
) {
  const [status, setStatus] = useState<AutosaveStatus>('idle')
  const currentVersionRef = useRef<number>(initialVersion)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const scheduleAutosave = useCallback((content: JSONContent) => {
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(async () => {
      setStatus('saving')
      try {
        const updated = await onSave(content, currentVersionRef.current)
        currentVersionRef.current = updated.version
        setStatus('saved')
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message.toLowerCase() : ''
        if (msg.includes('409') || msg.includes('conflict')) {
          setStatus('conflict')
        } else {
          setStatus('error')
        }
      }
    }, 2000)
  }, [onSave])

  return { scheduleAutosave, status }
}

import { useEffect, useRef, useCallback } from 'react'
import { WS_BASE_URL } from '../api/client'

export interface GameSocketEvent {
  type: string
  game_id: string
  entity_id?: string
  data: unknown
}

type EventHandler = (event: GameSocketEvent) => void

export function useGameSocket(gameId: string | undefined, onEvent: EventHandler) {
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef(onEvent)
  handlersRef.current = onEvent

  useEffect(() => {
    if (!gameId) return
    const wsUrl = WS_BASE_URL + `/games/${gameId}/ws`
    let retryDelay = 1000
    let closed = false
    let hasOpenedBefore = false

    function connect() {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        retryDelay = 1000
        if (hasOpenedBefore) {
          handlersRef.current({ type: '__reconnected', game_id: gameId!, data: null })
        }
        hasOpenedBefore = true
      }

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg = JSON.parse(event.data) as GameSocketEvent
          handlersRef.current(msg)
        } catch { /* ignore */ }
      }

      ws.onclose = () => {
        if (!closed) {
          setTimeout(() => { connect(); retryDelay = Math.min(retryDelay * 2, 30000) }, retryDelay)
        }
      }

      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; wsRef.current?.close() }
  }, [gameId])

  const send = useCallback((msg: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg))
    }
  }, [])

  return { send }
}

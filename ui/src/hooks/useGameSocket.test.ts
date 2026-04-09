import { renderHook, act } from '@testing-library/react'
import { useGameSocket } from './useGameSocket'

vi.mock('../api/client', () => ({
  WS_BASE_URL: 'ws://localhost:8080',
  BASE_URL: 'http://localhost:8080',
  apiFetch: vi.fn(),
  apiFetchRaw: vi.fn(),
}))

class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  static OPEN = 1
  static CONNECTING = 0
  static CLOSING = 2
  static CLOSED = 3

  url: string
  onmessage: ((e: MessageEvent) => void) | null = null
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onerror: ((e: Event) => void) | null = null
  readyState = FakeWebSocket.CONNECTING
  sentMessages: string[] = []

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sentMessages.push(data)
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.()
  }
}

describe('useGameSocket', () => {
  beforeEach(() => {
    FakeWebSocket.instances = []
    globalThis.WebSocket = FakeWebSocket as unknown as typeof WebSocket
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should not create a WebSocket when gameId is undefined', () => {
    renderHook(() => useGameSocket(undefined, vi.fn()))
    expect(FakeWebSocket.instances).toHaveLength(0)
  })

  it('should connect to the correct WebSocket URL when gameId is provided', () => {
    renderHook(() => useGameSocket('game-abc', vi.fn()))
    expect(FakeWebSocket.instances).toHaveLength(1)
    expect(FakeWebSocket.instances[0].url).toBe('ws://localhost:8080/games/game-abc/ws')
  })

  it('should fire a __reconnected event when the WebSocket opens', () => {
    const onEvent = vi.fn()
    renderHook(() => useGameSocket('game-abc', onEvent))

    act(() => {
      FakeWebSocket.instances[0].onopen?.()
    })

    expect(onEvent).toHaveBeenCalledWith({
      type: '__reconnected',
      game_id: 'game-abc',
      data: null,
    })
  })

  it('should parse and forward a valid JSON message to the event handler', () => {
    const onEvent = vi.fn()
    renderHook(() => useGameSocket('game-abc', onEvent))

    const msg = { type: 'ot_steps', game_id: 'game-abc', data: { steps: [1, 2] } }
    act(() => {
      FakeWebSocket.instances[0].onmessage?.(
        new MessageEvent('message', { data: JSON.stringify(msg) }),
      )
    })

    expect(onEvent).toHaveBeenCalledWith(msg)
  })

  it('should silently ignore non-JSON messages', () => {
    const onEvent = vi.fn()
    renderHook(() => useGameSocket('game-abc', onEvent))

    act(() => {
      FakeWebSocket.instances[0].onmessage?.(
        new MessageEvent('message', { data: 'not-valid-json{{' }),
      )
    })

    // onEvent not called (the __reconnected hasn't been triggered either)
    expect(onEvent).not.toHaveBeenCalled()
  })

  it('should close the WebSocket on unmount and prevent reconnection', () => {
    const { unmount } = renderHook(() => useGameSocket('game-abc', vi.fn()))
    const ws = FakeWebSocket.instances[0]

    unmount()

    expect(ws.readyState).toBe(FakeWebSocket.CLOSED)
    // No reconnection timer should fire
    vi.advanceTimersByTime(5000)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })

  it('should send JSON-serialised messages when readyState is OPEN', () => {
    const { result } = renderHook(() => useGameSocket('game-abc', vi.fn()))
    const ws = FakeWebSocket.instances[0]
    ws.readyState = FakeWebSocket.OPEN

    act(() => {
      result.current.send({ type: 'ping', payload: 42 })
    })

    expect(ws.sentMessages).toHaveLength(1)
    expect(ws.sentMessages[0]).toBe(JSON.stringify({ type: 'ping', payload: 42 }))
  })

  it('should not send a message when readyState is not OPEN', () => {
    const { result } = renderHook(() => useGameSocket('game-abc', vi.fn()))
    const ws = FakeWebSocket.instances[0]
    // readyState starts at CONNECTING (0)

    act(() => {
      result.current.send({ type: 'ping' })
    })

    expect(ws.sentMessages).toHaveLength(0)
  })

  it('should trigger a reconnect when the socket closes unexpectedly', () => {
    renderHook(() => useGameSocket('game-abc', vi.fn()))
    const ws = FakeWebSocket.instances[0]

    // Simulate unexpected close (don't unmount — closed flag stays false)
    // We manually call onclose without going through close() to avoid setting readyState
    ws.onclose?.()
    vi.advanceTimersByTime(1000)

    expect(FakeWebSocket.instances).toHaveLength(2)
    expect(FakeWebSocket.instances[1].url).toBe('ws://localhost:8080/games/game-abc/ws')
  })
})

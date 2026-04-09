import { renderHook, act } from '@testing-library/react'
import { useAutosave } from './useAutosave'

describe('useAutosave', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should start with "idle" status', () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    const { result } = renderHook(() => useAutosave(onSave))

    expect(result.current.status).toBe('idle')
  })

  it('should not call onSave before the 2000ms debounce delay elapses', () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    const { result } = renderHook(() => useAutosave(onSave))
    const content = { type: 'doc', content: [] }

    act(() => {
      result.current.scheduleAutosave(content)
    })
    vi.advanceTimersByTime(1999)

    expect(onSave).not.toHaveBeenCalled()
  })

  it('should call onSave with the scheduled content after the debounce delay', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    const { result } = renderHook(() => useAutosave(onSave))
    const content = { type: 'doc', content: [{ type: 'text', text: 'hello' }] }

    act(() => {
      result.current.scheduleAutosave(content)
    })
    await act(async () => {
      vi.advanceTimersByTime(2000)
    })

    expect(onSave).toHaveBeenCalledOnce()
    expect(onSave).toHaveBeenCalledWith(content)
  })

  it('should set status to "saved" after a successful save', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    const { result } = renderHook(() => useAutosave(onSave))

    act(() => {
      result.current.scheduleAutosave({ type: 'doc', content: [] })
    })
    await act(async () => {
      vi.advanceTimersByTime(2000)
    })

    expect(result.current.status).toBe('saved')
  })

  it('should set status to "error" when onSave rejects', async () => {
    const onSave = vi.fn().mockRejectedValue(new Error('Network failure'))
    const { result } = renderHook(() => useAutosave(onSave))

    act(() => {
      result.current.scheduleAutosave({ type: 'doc', content: [] })
    })
    await act(async () => {
      vi.advanceTimersByTime(2000)
    })

    expect(result.current.status).toBe('error')
  })

  it('should debounce rapid calls and only save with the latest content', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined)
    const { result } = renderHook(() => useAutosave(onSave))
    const first = { type: 'doc', content: [{ type: 'text', text: 'first' }] }
    const second = { type: 'doc', content: [{ type: 'text', text: 'second' }] }
    const third = { type: 'doc', content: [{ type: 'text', text: 'third' }] }

    act(() => {
      result.current.scheduleAutosave(first)
    })
    vi.advanceTimersByTime(500)
    act(() => {
      result.current.scheduleAutosave(second)
    })
    vi.advanceTimersByTime(500)
    act(() => {
      result.current.scheduleAutosave(third)
    })
    await act(async () => {
      vi.advanceTimersByTime(2000)
    })

    expect(onSave).toHaveBeenCalledOnce()
    expect(onSave).toHaveBeenCalledWith(third)
  })
})

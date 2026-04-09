import { renderHook, act } from '@testing-library/react'
import { useLocalStorage } from './useLocalStorage'

describe('useLocalStorage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should return the initial value when localStorage is empty', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'))
    expect(result.current[0]).toBe('default')
  })

  it('should read and return an existing value from localStorage', () => {
    localStorage.setItem('test-key', JSON.stringify('stored'))
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'))
    expect(result.current[0]).toBe('stored')
  })

  it('should update state when setValue is called with a new value', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'initial'))

    act(() => {
      result.current[1]('updated')
    })

    expect(result.current[0]).toBe('updated')
  })

  it('should persist the updated value to localStorage', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'initial'))

    act(() => {
      result.current[1]('persisted')
    })

    expect(localStorage.getItem('test-key')).toBe(JSON.stringify('persisted'))
  })

  it('should support a function updater (like setState)', () => {
    const { result } = renderHook(() => useLocalStorage('count', 0))

    act(() => {
      result.current[1]((prev) => prev + 1)
    })

    expect(result.current[0]).toBe(1)
    expect(localStorage.getItem('count')).toBe('1')
  })

  it('should work with boolean values', () => {
    const { result } = renderHook(() => useLocalStorage('flag', false))

    act(() => {
      result.current[1](true)
    })

    expect(result.current[0]).toBe(true)
    expect(localStorage.getItem('flag')).toBe('true')
  })

  it('should work with object values', () => {
    const { result } = renderHook(() =>
      useLocalStorage<{ name: string }>('obj', { name: 'default' }),
    )

    act(() => {
      result.current[1]({ name: 'updated' })
    })

    expect(result.current[0]).toEqual({ name: 'updated' })
    expect(JSON.parse(localStorage.getItem('obj')!)).toEqual({ name: 'updated' })
  })

  it('should fall back to the initial value when localStorage contains invalid JSON', () => {
    localStorage.setItem('bad-key', 'not{{valid-json')
    const { result } = renderHook(() => useLocalStorage('bad-key', 'fallback'))
    expect(result.current[0]).toBe('fallback')
  })

  it('should isolate values by key', () => {
    const { result: a } = renderHook(() => useLocalStorage('key-a', 'alpha'))
    const { result: b } = renderHook(() => useLocalStorage('key-b', 'beta'))

    act(() => {
      a.current[1]('updated-a')
    })

    expect(a.current[0]).toBe('updated-a')
    expect(b.current[0]).toBe('beta')
  })
})

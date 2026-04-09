import { renderHook, act } from '@testing-library/react'
import { useDarkMode } from './useDarkMode'

describe('useDarkMode', () => {
  beforeEach(() => {
    localStorage.clear()
    delete document.documentElement.dataset.theme
  })

  it('should default to light mode when localStorage has no value', () => {
    const { result } = renderHook(() => useDarkMode())
    const [isDark] = result.current
    expect(isDark).toBe(false)
  })

  it('should apply "light" theme to the document element by default', () => {
    renderHook(() => useDarkMode())
    expect(document.documentElement.dataset.theme).toBe('light')
  })

  it('should read initial dark mode value from localStorage', () => {
    localStorage.setItem('pf2e-dark-mode', JSON.stringify(true))
    const { result } = renderHook(() => useDarkMode())
    const [isDark] = result.current
    expect(isDark).toBe(true)
  })

  it('should apply "dark" theme when localStorage has true', () => {
    localStorage.setItem('pf2e-dark-mode', JSON.stringify(true))
    renderHook(() => useDarkMode())
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('should toggle dark mode on', () => {
    const { result } = renderHook(() => useDarkMode())

    act(() => {
      const [, setIsDark] = result.current
      setIsDark(true)
    })

    expect(result.current[0]).toBe(true)
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('should toggle dark mode off', () => {
    localStorage.setItem('pf2e-dark-mode', JSON.stringify(true))
    const { result } = renderHook(() => useDarkMode())

    act(() => {
      const [, setIsDark] = result.current
      setIsDark(false)
    })

    expect(result.current[0]).toBe(false)
    expect(document.documentElement.dataset.theme).toBe('light')
  })

  it('should persist the dark mode preference to localStorage', () => {
    const { result } = renderHook(() => useDarkMode())

    act(() => {
      const [, setIsDark] = result.current
      setIsDark(true)
    })

    expect(localStorage.getItem('pf2e-dark-mode')).toBe('true')
  })
})

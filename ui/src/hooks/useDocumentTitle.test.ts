import { renderHook } from '@testing-library/react'
import { useDocumentTitle } from './useDocumentTitle'

const APP_NAME = 'PF2e Companion'

describe('useDocumentTitle', () => {
  afterEach(() => {
    document.title = ''
  })

  it('should set document.title to the app name when no page title is provided', () => {
    renderHook(() => useDocumentTitle())
    expect(document.title).toBe(APP_NAME)
  })

  it('should set document.title with page title in "Page | App" format', () => {
    renderHook(() => useDocumentTitle('Dashboard'))
    expect(document.title).toBe(`Dashboard | ${APP_NAME}`)
  })

  it('should update document.title when the page title changes', () => {
    const { rerender } = renderHook(
      ({ title }: { title?: string }) => useDocumentTitle(title),
      { initialProps: { title: 'First Page' } },
    )
    expect(document.title).toBe(`First Page | ${APP_NAME}`)

    rerender({ title: 'Second Page' })
    expect(document.title).toBe(`Second Page | ${APP_NAME}`)
  })

  it('should fall back to app name when title changes to undefined', () => {
    const { rerender } = renderHook(
      ({ title }: { title?: string }) => useDocumentTitle(title),
      { initialProps: { title: 'Some Page' } },
    )
    expect(document.title).toBe(`Some Page | ${APP_NAME}`)

    rerender({ title: undefined })
    expect(document.title).toBe(APP_NAME)
  })

  it('should reset document.title to the app name on unmount', () => {
    const { unmount } = renderHook(() => useDocumentTitle('Temporary Page'))
    expect(document.title).toBe(`Temporary Page | ${APP_NAME}`)

    unmount()
    expect(document.title).toBe(APP_NAME)
  })
})

import { useEffect } from 'react'
import { useLocalStorage } from './useLocalStorage'

export function useDarkMode() {
  const [isDark, setIsDark] = useLocalStorage('pf2e-dark-mode', false)

  useEffect(() => {
    document.documentElement.dataset.theme = isDark ? 'dark' : 'light'
  }, [isDark])

  return [isDark, setIsDark] as const
}

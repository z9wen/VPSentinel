import { useEffect, useState } from 'react'

export type Theme = 'light' | 'dark'

const storageKey = 'vpsentinel-theme'

function initialTheme(): Theme {
  try {
    const stored = localStorage.getItem(storageKey)
    if (stored === 'light' || stored === 'dark') return stored
  } catch {
    // Storage can be unavailable in private windows; fall back to the system theme.
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(initialTheme)

  useEffect(() => {
    document.documentElement.dataset.theme = `sentinel-${theme}`
  }, [theme])

  function toggle() {
    const next = theme === 'dark' ? 'light' : 'dark'
    try {
      localStorage.setItem(storageKey, next)
    } catch {
      // The choice still applies for this page view.
    }
    setTheme(next)
  }

  return { theme, toggle }
}

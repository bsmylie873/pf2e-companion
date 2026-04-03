import { useState, useEffect } from 'react'
import { getPreferences } from '../api/preferences'

const FALLBACK_PAGE_SIZE = 10

type Resource = 'campaigns' | 'sessions' | 'notes'

/**
 * Returns the user's preferred page size for a given resource.
 * Falls back to the user's global default, then to 10.
 */
export function usePageSize(resource: Resource): number {
  const [pageSize, setPageSize] = useState(FALLBACK_PAGE_SIZE)

  useEffect(() => {
    getPreferences()
      .then(prefs => {
        if (!prefs.page_size) return
        const override = prefs.page_size[resource]
        if (override != null && override > 0) {
          setPageSize(override)
        } else if (prefs.page_size.default > 0) {
          setPageSize(prefs.page_size.default)
        }
      })
      .catch(() => {})
  }, [resource])

  return pageSize
}

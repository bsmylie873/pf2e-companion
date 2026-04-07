export const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL ?? BASE_URL.replace(/^http/, 'ws')

// --- Silent token refresh machinery ---
let isRefreshing = false
let pendingQueue: Array<{ resolve: () => void; reject: (err: Error) => void }> = []

function enqueueRetry(): Promise<void> {
  return new Promise((resolve, reject) => {
    pendingQueue.push({ resolve, reject })
  })
}

function drainQueue(error?: Error) {
  const queue = pendingQueue
  pendingQueue = []
  for (const { resolve, reject } of queue) {
    error ? reject(error) : resolve()
  }
}

async function attemptRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    return res.ok
  } catch {
    return false
  }
}

async function handleUnauthorized<T>(
  path: string,
  options: RequestInit | undefined,
  parseResponse: (res: Response) => Promise<T>,
): Promise<T> {
  // If we're already refreshing, queue this request
  if (isRefreshing) {
    await enqueueRetry()
    // Retry the original request after refresh completes
    const retryRes = await fetch(`${BASE_URL}${path}`, {
      credentials: 'include',
      headers: { 'Content-Type': 'application/json', ...(options?.headers as Record<string, string>) },
      ...options,
    })
    if (!retryRes.ok) {
      const json = await retryRes.json().catch(() => ({}))
      throw new Error(json.message ?? `Request failed: ${retryRes.status}`)
    }
    return parseResponse(retryRes)
  }

  // First 401 — attempt refresh
  isRefreshing = true
  const refreshed = await attemptRefresh()
  isRefreshing = false

  if (!refreshed) {
    drainQueue(new Error('Session expired'))
    window.location.href = '/?expired=true'
    throw new Error('Session expired')
  }

  // Refresh succeeded — drain the queue so waiting requests retry
  drainQueue()

  // Retry the original request
  const retryRes = await fetch(`${BASE_URL}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(options?.headers as Record<string, string>) },
    ...options,
  })
  if (!retryRes.ok) {
    const json = await retryRes.json().catch(() => ({}))
    throw new Error(json.message ?? `Request failed: ${retryRes.status}`)
  }
  return parseResponse(retryRes)
}

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: 'include',
    headers,
    ...options,
  })

  if (res.status === 401 && !path.startsWith('/auth/')) {
    return handleUnauthorized(path, options, async (r) => {
      const json = await r.json()
      return json.data as T
    })
  }

  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.message ?? `Request failed: ${res.status}`)
  }
  return json.data as T
}

export async function apiFetchRaw<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: 'include',
    headers,
    ...options,
  })

  if (res.status === 401 && !path.startsWith('/auth/')) {
    return handleUnauthorized(path, options, async (r) => {
      const json = await r.json()
      return json as T
    })
  }

  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.message ?? `Request failed: ${res.status}`)
  }
  return json as T
}

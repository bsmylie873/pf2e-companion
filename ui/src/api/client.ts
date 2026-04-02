export const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
export const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL ?? BASE_URL.replace(/^http/, 'ws')

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
    window.location.href = '/'
    throw new Error('Unauthorized')
  }

  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.message ?? `Request failed: ${res.status}`)
  }
  return json.data as T
}

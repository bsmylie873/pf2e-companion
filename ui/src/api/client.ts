export const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : ''
}

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const method = (options?.method ?? 'GET').toUpperCase()
  const isMutating = ['POST', 'PATCH', 'DELETE', 'PUT'].includes(method)

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }
  if (isMutating) {
    const csrf = getCsrfToken()
    if (csrf) headers['X-CSRF-Token'] = csrf
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

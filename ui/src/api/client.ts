const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  const json = await res.json()
  if (!res.ok) {
    throw new Error(json.message ?? `Request failed: ${res.status}`)
  }
  return json.data as T
}

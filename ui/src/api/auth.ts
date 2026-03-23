import { apiFetch } from './client'
import type { User } from '../types/user'

export interface LoginPayload {
  username: string
  password: string
}

export interface RegisterPayload {
  username: string
  email: string
  password: string
}

export function login(payload: LoginPayload): Promise<User> {
  return apiFetch<User>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function register(payload: RegisterPayload): Promise<User> {
  return apiFetch<User>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function logout(): Promise<{ message: string }> {
  return apiFetch<{ message: string }>('/auth/logout', { method: 'POST' })
}

export function getMe(): Promise<User> {
  return apiFetch<User>('/auth/me')
}

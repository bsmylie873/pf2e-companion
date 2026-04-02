import { apiFetch } from './client'
import type { Note, NoteFormData } from '../types/note'
import type { JSONContent } from '@tiptap/react'

export function listGameNotes(gameId: string, params?: { sort?: string; session_id?: string; unlinked?: boolean }): Promise<Note[]> {
  const query = new URLSearchParams()
  if (params?.sort) query.set('sort', params.sort)
  if (params?.session_id) query.set('session_id', params.session_id)
  if (params?.unlinked) query.set('unlinked', 'true')
  const qs = query.toString()
  return apiFetch<Note[]>(`/games/${gameId}/notes${qs ? '?' + qs : ''}`)
}

export function getNote(noteId: string): Promise<Note> {
  return apiFetch<Note>(`/notes/${noteId}`)
}

export function createNote(gameId: string, data: NoteFormData): Promise<Note> {
  return apiFetch<Note>(`/games/${gameId}/notes`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateNote(noteId: string, data: Record<string, unknown>): Promise<Note> {
  return apiFetch<Note>(`/notes/${noteId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

export function deleteNote(noteId: string): Promise<void> {
  return apiFetch<void>(`/notes/${noteId}`, { method: 'DELETE' })
}

export function updateNoteContent(noteId: string, data: { content: JSONContent; version: number }): Promise<Note> {
  return apiFetch<Note>(`/notes/${noteId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  })
}

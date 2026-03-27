import type { JSONContent } from '@tiptap/react'

export interface Note {
  id: string
  game_id: string
  user_id: string
  session_id: string | null
  title: string
  content: JSONContent | null
  visibility: 'private' | 'visible' | 'editable'
  version: number
  foundry_data: unknown
  created_at: string
  updated_at: string
}

export interface NoteFormData {
  title: string
  session_id?: string | null
  visibility?: 'private' | 'visible' | 'editable'
}

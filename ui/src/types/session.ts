import type { JSONContent } from '@tiptap/react'

export interface Session {
  id: string
  game_id: string
  title: string
  session_number: number | null
  scheduled_at: string | null
  runtime_start: string | null
  runtime_end: string | null
  folder_id: string | null
  notes: JSONContent | null
  version: number
  foundry_data: unknown
  created_at: string
  updated_at: string
}

export interface SessionFormData {
  title: string
  session_number: number | null
  scheduled_at: string | null
  runtime_start: string | null
  runtime_end: string | null
}

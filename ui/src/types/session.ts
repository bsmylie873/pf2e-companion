export interface Session {
  id: string
  game_id: string
  title: string
  session_number: number | null
  scheduled_at: string | null
  notes: unknown
  version: number
  foundry_data: unknown
  created_at: string
  updated_at: string
}

export interface SessionFormData {
  title: string
  session_number: number | null
  scheduled_at: string | null
}

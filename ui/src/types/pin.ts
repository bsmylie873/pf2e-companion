export interface SessionPin {
  id: string
  game_id: string
  session_id: string | null
  note_id: string | null
  group_id: string | null
  label: string
  x: number
  y: number
  colour: string
  icon: string
  description: string | null
  created_at: string
  updated_at: string
}

export interface SessionPinFormData {
  session_id?: string
  x: number
  y: number
  label?: string
  colour?: string
  icon?: string
  description?: string
}

export interface PinGroup {
  id: string
  game_id: string
  x: number
  y: number
  colour: string
  icon: string
  pin_count: number
  pins: SessionPin[]
  created_at: string
  updated_at: string
}

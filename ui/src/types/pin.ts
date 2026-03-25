export interface SessionPin {
  id: string
  session_id: string
  label: string
  x: number
  y: number
  pin_type: string
  description: string | null
  created_at: string
  updated_at: string
}

export interface SessionPinFormData {
  session_id: string
  x: number
  y: number
  label?: string
}

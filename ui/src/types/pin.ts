export interface PinType {
  id: number
  name: string
}

export interface SessionPin {
  id: string
  session_id: string
  label: string
  x: number
  y: number
  pin_type_id: number
  pin_type: PinType
  description: string | null
  created_at: string
  updated_at: string
}

export interface SessionPinFormData {
  session_id: string
  x: number
  y: number
  label?: string
  pin_type_id?: number
}

export interface GameMembership {
  id: string
  game_id: string
  user_id: string
  is_gm: boolean
  foundry_data: unknown
  created_at: string
  updated_at: string
}

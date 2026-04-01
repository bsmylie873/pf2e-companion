export interface GameMap {
  id: string
  game_id: string
  name: string
  description: string | null
  image_url: string | null
  sort_order: number
  archived_at: string | null
  created_at: string
  updated_at: string
}

export interface Game {
  id: string
  title: string
  description: string | null
  splash_image_url: string | null
  map_image_url: string | null
  foundry_data: unknown
  created_at: string
  updated_at: string
}

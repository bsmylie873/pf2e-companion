export interface Game {
  id: string
  title: string
  description: string | null
  splash_image_url: string | null
  foundry_data: unknown
  is_gm: boolean
  created_at: string
  updated_at: string
}

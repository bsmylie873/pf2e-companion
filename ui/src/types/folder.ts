export interface Folder {
  id: string
  game_id: string
  user_id: string | null
  name: string
  folder_type: 'session' | 'note'
  visibility: 'private' | 'game-wide'
  position: number
  created_at: string
  updated_at: string
}

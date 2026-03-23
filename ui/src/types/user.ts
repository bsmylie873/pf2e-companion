export interface User {
  id: string
  username: string
  email: string
  avatar_url?: string
  description?: string
  location?: string
}

export interface UserPublic {
  id: string
  username: string
  avatar_url?: string
}

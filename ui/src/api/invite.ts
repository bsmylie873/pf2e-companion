import { apiFetch } from './client'

export interface InviteTokenResponse {
  token: string
  expires_at: string | null
  created_at: string
}

export interface InviteStatusResponse {
  has_active_invite: boolean
  expires_at?: string
  created_at?: string
  token?: string
}

export interface InviteValidationResponse {
  game_id: string
  game_title: string
}

export interface InviteRedeemResponse {
  game_id: string
  membership_id: string
  already_member: boolean
}

export function generateInvite(gameId: string, expiresIn: string): Promise<InviteTokenResponse> {
  return apiFetch<InviteTokenResponse>(`/games/${gameId}/invite`, {
    method: 'POST',
    body: JSON.stringify({ expires_in: expiresIn }),
  })
}

export function getInviteStatus(gameId: string): Promise<InviteStatusResponse> {
  return apiFetch<InviteStatusResponse>(`/games/${gameId}/invite`)
}

export function revokeInvite(gameId: string): Promise<void> {
  return apiFetch<void>(`/games/${gameId}/invite`, { method: 'DELETE' })
}

export function validateInvite(token: string): Promise<InviteValidationResponse> {
  return apiFetch<InviteValidationResponse>(`/invite/${token}`)
}

export function redeemInvite(token: string): Promise<InviteRedeemResponse> {
  return apiFetch<InviteRedeemResponse>(`/invite/${token}/redeem`, { method: 'POST' })
}

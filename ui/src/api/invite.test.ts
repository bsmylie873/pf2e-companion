import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  generateInvite,
  getInviteStatus,
  revokeInvite,
  validateInvite,
  redeemInvite,
} from './invite'

vi.mock('./client', () => ({
  apiFetch: vi.fn(),
}))

import { apiFetch } from './client'

const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('generateInvite', () => {
  it('should call apiFetch POST /games/:gameId/invite with expires_in', async () => {
    const mockToken = {
      token: 'invite-token-abc',
      expires_at: '2024-12-31T00:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(mockToken)

    const result = await generateInvite('game-1', '7d')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/invite', {
      method: 'POST',
      body: JSON.stringify({ expires_in: '7d' }),
    })
    expect(result).toEqual(mockToken)
  })

  it('should support no-expiry invites', async () => {
    const mockToken = { token: 'perm-token', expires_at: null, created_at: '2024-01-01T00:00:00Z' }
    mockApiFetch.mockResolvedValueOnce(mockToken)

    const result = await generateInvite('game-2', 'never')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-2/invite', {
      method: 'POST',
      body: JSON.stringify({ expires_in: 'never' }),
    })
    expect(result.expires_at).toBeNull()
  })
})

describe('getInviteStatus', () => {
  it('should call apiFetch GET /games/:gameId/invite', async () => {
    const mockStatus = {
      has_active_invite: true,
      token: 'some-token',
      expires_at: '2024-12-31T00:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(mockStatus)

    const result = await getInviteStatus('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/invite')
    expect(result).toEqual(mockStatus)
  })

  it('should return has_active_invite false when no active invite', async () => {
    mockApiFetch.mockResolvedValueOnce({ has_active_invite: false })

    const result = await getInviteStatus('game-no-invite')

    expect(result.has_active_invite).toBe(false)
  })
})

describe('revokeInvite', () => {
  it('should call apiFetch DELETE /games/:gameId/invite', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)

    await revokeInvite('game-1')

    expect(mockApiFetch).toHaveBeenCalledWith('/games/game-1/invite', { method: 'DELETE' })
  })
})

describe('validateInvite', () => {
  it('should call apiFetch GET /invite/:token', async () => {
    const mockValidation = { game_id: 'game-1', game_title: 'My Campaign' }
    mockApiFetch.mockResolvedValueOnce(mockValidation)

    const result = await validateInvite('token-xyz')

    expect(mockApiFetch).toHaveBeenCalledWith('/invite/token-xyz')
    expect(result).toEqual(mockValidation)
  })
})

describe('redeemInvite', () => {
  it('should call apiFetch POST /invite/:token/redeem', async () => {
    const mockRedeem = {
      game_id: 'game-1',
      membership_id: 'mem-123',
      already_member: false,
    }
    mockApiFetch.mockResolvedValueOnce(mockRedeem)

    const result = await redeemInvite('token-xyz')

    expect(mockApiFetch).toHaveBeenCalledWith('/invite/token-xyz/redeem', { method: 'POST' })
    expect(result).toEqual(mockRedeem)
  })

  it('should indicate already_member when user is already in the game', async () => {
    mockApiFetch.mockResolvedValueOnce({
      game_id: 'game-1',
      membership_id: 'existing-mem',
      already_member: true,
    })

    const result = await redeemInvite('token-xyz')

    expect(result.already_member).toBe(true)
  })
})

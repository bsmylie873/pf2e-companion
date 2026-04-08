-- =============================================================================
-- V1.002 — Invite Tokens
-- =============================================================================

CREATE TABLE invite_tokens (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id     UUID        NOT NULL,
    created_by  UUID        NOT NULL,
    token_hash  TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_invite_tokens_hash UNIQUE (token_hash),
    CONSTRAINT fk_invite_tokens_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_invite_tokens_creator FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_invite_tokens_game_id ON invite_tokens (game_id);

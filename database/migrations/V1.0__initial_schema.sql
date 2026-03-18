-- =============================================================================
-- V1 — Initial Schema
-- Pathfinder 2E Companion Application
--
-- Creates all v1 tables: users, games, game_memberships, sessions, notes,
-- characters, items.
--
-- Target: PostgreSQL 14+
-- =============================================================================

-- ---------------------------------------------------------------------------
-- 1. users
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    username        VARCHAR(100)    NOT NULL,
    email           VARCHAR(255)    NOT NULL,
    password_hash   TEXT            NOT NULL,
    avatar_url      TEXT,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email    UNIQUE (email)
);

-- ---------------------------------------------------------------------------
-- 2. games
-- ---------------------------------------------------------------------------
CREATE TABLE games (
    id                  UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    title               VARCHAR(255)    NOT NULL,
    description         TEXT,
    splash_image_url    TEXT,
    foundry_data        JSONB,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 3. game_memberships (join table: users <-> games)
-- ---------------------------------------------------------------------------
CREATE TABLE game_memberships (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    user_id         UUID            NOT NULL,
    is_gm           BOOLEAN         NOT NULL DEFAULT FALSE,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_game_memberships_game FOREIGN KEY (game_id) REFERENCES games (id),
    CONSTRAINT fk_game_memberships_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT uq_game_memberships_game_user UNIQUE (game_id, user_id)
);

-- ---------------------------------------------------------------------------
-- 4. sessions
-- ---------------------------------------------------------------------------
CREATE TABLE sessions (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    title           VARCHAR(255)    NOT NULL,
    session_number  INTEGER,
    scheduled_at    TIMESTAMPTZ,
    notes           JSONB,
    version         INTEGER         NOT NULL DEFAULT 1,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_sessions_game FOREIGN KEY (game_id) REFERENCES games (id)
);

-- ---------------------------------------------------------------------------
-- 5. notes (dual-ownership: exactly one of game_id / user_id must be set)
-- ---------------------------------------------------------------------------
CREATE TABLE notes (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID,
    user_id         UUID,
    title           VARCHAR(255)    NOT NULL,
    content         JSONB,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_notes_game FOREIGN KEY (game_id) REFERENCES games (id),
    CONSTRAINT fk_notes_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT chk_notes_ownership CHECK (
        (game_id IS NOT NULL AND user_id IS NULL)
        OR (game_id IS NULL AND user_id IS NOT NULL)
    )
);

-- ---------------------------------------------------------------------------
-- 6. characters
-- ---------------------------------------------------------------------------
CREATE TABLE characters (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    user_id         UUID,
    name            VARCHAR(255)    NOT NULL,
    is_npc          BOOLEAN         NOT NULL DEFAULT FALSE,
    ancestry        VARCHAR(100),
    heritage        VARCHAR(100),
    class           VARCHAR(100),
    background      VARCHAR(100),
    level           INTEGER         NOT NULL DEFAULT 1,
    hp_max          INTEGER,
    hp_current      INTEGER,
    ac              INTEGER,
    strength        INTEGER,
    dexterity       INTEGER,
    constitution    INTEGER,
    intelligence    INTEGER,
    wisdom          INTEGER,
    charisma        INTEGER,
    fortitude       INTEGER,
    reflex          INTEGER,
    will            INTEGER,
    skills          JSONB,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_characters_game FOREIGN KEY (game_id) REFERENCES games (id),
    CONSTRAINT fk_characters_user FOREIGN KEY (user_id) REFERENCES users (id)
);

-- ---------------------------------------------------------------------------
-- 7. items
-- ---------------------------------------------------------------------------
CREATE TABLE items (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    character_id    UUID,
    name            VARCHAR(255)    NOT NULL,
    description     TEXT,
    level           INTEGER         NOT NULL DEFAULT 0,
    price_gp        NUMERIC(10, 2),
    bulk            VARCHAR(10),
    traits          TEXT[],
    quantity        INTEGER         NOT NULL DEFAULT 1,
    foundry_data    JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_items_game      FOREIGN KEY (game_id)      REFERENCES games (id),
    CONSTRAINT fk_items_character  FOREIGN KEY (character_id)  REFERENCES characters (id)
);

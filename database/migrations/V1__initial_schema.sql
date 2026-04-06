-- =============================================================================
-- V1 — Initial Schema
-- Pathfinder 2E Companion Application
--
-- Creates all tables: users, games, game_memberships, sessions, notes,
-- characters, items, refresh_tokens, session_pins, pin_groups,
-- user_preferences, folders, maps.
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
    description     TEXT,
    location        TEXT,
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

    CONSTRAINT fk_game_memberships_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_game_memberships_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT uq_game_memberships_game_user UNIQUE (game_id, user_id)
);

-- ---------------------------------------------------------------------------
-- 4. folders
-- ---------------------------------------------------------------------------
CREATE TABLE folders (
    id          UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id     UUID            NOT NULL,
    user_id     UUID,
    name        VARCHAR(100)    NOT NULL,
    folder_type VARCHAR(10)     NOT NULL,
    visibility  VARCHAR(10)     NOT NULL DEFAULT 'game-wide',
    position    INTEGER         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_folders_game FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
    CONSTRAINT fk_folders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_folder_type CHECK (folder_type IN ('session', 'note')),
    CONSTRAINT chk_folder_visibility CHECK (visibility IN ('private', 'game-wide'))
);

-- Session folders: unique name per game
CREATE UNIQUE INDEX uq_folders_session_name ON folders (game_id, LOWER(name))
    WHERE folder_type = 'session';

-- Note folders: unique name per user per game
CREATE UNIQUE INDEX uq_folders_note_name ON folders (game_id, user_id, LOWER(name))
    WHERE folder_type = 'note';

-- ---------------------------------------------------------------------------
-- 5. sessions
-- ---------------------------------------------------------------------------
CREATE TABLE sessions (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    title           VARCHAR(255)    NOT NULL,
    session_number  INTEGER,
    scheduled_at    TIMESTAMPTZ,
    runtime_start   TIMESTAMPTZ,
    runtime_end     TIMESTAMPTZ,
    notes           JSONB,
    version         INTEGER         NOT NULL DEFAULT 1,
    foundry_data    JSONB,
    folder_id       UUID,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_sessions_game   FOREIGN KEY (game_id)   REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_sessions_folder FOREIGN KEY (folder_id)  REFERENCES folders(id) ON DELETE SET NULL
);

COMMENT ON COLUMN sessions.version IS 'OT sequence counter; incremented on every update';

-- ---------------------------------------------------------------------------
-- 6. notes
-- ---------------------------------------------------------------------------
CREATE TABLE notes (
    id              UUID            NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id         UUID            NOT NULL,
    user_id         UUID            NOT NULL,
    session_id      UUID,
    title           VARCHAR(255)    NOT NULL,
    content         JSONB,
    visibility      VARCHAR(10)     NOT NULL DEFAULT 'private',
    version         INTEGER         NOT NULL DEFAULT 1,
    foundry_data    JSONB,
    folder_id       UUID,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT fk_notes_game    FOREIGN KEY (game_id)    REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_notes_user    FOREIGN KEY (user_id)    REFERENCES users (id),
    CONSTRAINT fk_notes_session FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE SET NULL,
    CONSTRAINT fk_notes_folder  FOREIGN KEY (folder_id)  REFERENCES folders(id) ON DELETE SET NULL
);

COMMENT ON COLUMN notes.version IS 'OT sequence counter; incremented on every update';

-- ---------------------------------------------------------------------------
-- 7. characters
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

    CONSTRAINT fk_characters_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_characters_user FOREIGN KEY (user_id) REFERENCES users (id)
);

-- ---------------------------------------------------------------------------
-- 8. items
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

    CONSTRAINT fk_items_game      FOREIGN KEY (game_id)      REFERENCES games (id) ON DELETE CASCADE,
    CONSTRAINT fk_items_character  FOREIGN KEY (character_id)  REFERENCES characters (id) ON DELETE CASCADE
);

-- ---------------------------------------------------------------------------
-- 9. refresh_tokens
-- ---------------------------------------------------------------------------
CREATE TABLE refresh_tokens (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- ---------------------------------------------------------------------------
-- 10. maps
-- ---------------------------------------------------------------------------
CREATE TABLE maps (
    id          UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id     UUID          NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    name        VARCHAR(255)  NOT NULL,
    description TEXT,
    image_url   TEXT,
    sort_order  INTEGER       NOT NULL DEFAULT 0,
    archived_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

-- Partial unique index: name must be unique per game among non-archived maps
CREATE UNIQUE INDEX uq_maps_active_game_name ON maps(game_id, name) WHERE archived_at IS NULL;
CREATE INDEX idx_maps_game_id ON maps(game_id);

-- ---------------------------------------------------------------------------
-- 11. pin_groups
-- ---------------------------------------------------------------------------
CREATE TABLE pin_groups (
    id         UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id    UUID          NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    map_id     UUID          REFERENCES maps(id) ON DELETE CASCADE,
    x          NUMERIC(6,4)  NOT NULL,
    y          NUMERIC(6,4)  NOT NULL,
    colour     VARCHAR(20)   NOT NULL DEFAULT 'grey',
    icon       VARCHAR(50)   NOT NULL DEFAULT 'position-marker',
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_pin_groups_map_id ON pin_groups(map_id);

-- ---------------------------------------------------------------------------
-- 12. session_pins
-- ---------------------------------------------------------------------------
CREATE TABLE session_pins (
    id          UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    session_id  UUID          REFERENCES sessions(id) ON DELETE CASCADE,
    game_id     UUID          NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    map_id      UUID          REFERENCES maps(id) ON DELETE CASCADE,
    note_id     UUID          REFERENCES notes(id) ON DELETE SET NULL,
    group_id    UUID          REFERENCES pin_groups(id) ON DELETE SET NULL,
    label       TEXT          NOT NULL DEFAULT '',
    x           NUMERIC(6,4)  NOT NULL,
    y           NUMERIC(6,4)  NOT NULL,
    description TEXT,
    colour      VARCHAR(20)   NOT NULL DEFAULT 'grey',
    icon        VARCHAR(50)   NOT NULL DEFAULT 'position-marker',
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_session_pins_group_id ON session_pins(group_id);
CREATE INDEX idx_session_pins_map_id ON session_pins(map_id);

-- ---------------------------------------------------------------------------
-- 13. user_preferences
-- ---------------------------------------------------------------------------
CREATE TABLE user_preferences (
    user_id            UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE PRIMARY KEY,
    default_game_id    UUID         REFERENCES games(id) ON DELETE SET NULL,
    default_pin_colour VARCHAR(20),
    default_pin_icon   VARCHAR(50),
    sidebar_state      JSONB        DEFAULT '{}'::jsonb,
    default_view_mode  JSONB        DEFAULT '{}'::jsonb,
    map_editor_mode    VARCHAR(10)  NOT NULL DEFAULT 'modal',
    page_size          JSONB        DEFAULT '{"default":10}'::jsonb,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
);

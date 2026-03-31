-- V2.2: Add folders table and folder_id to sessions and notes

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

-- Add nullable folder_id to sessions
ALTER TABLE sessions ADD COLUMN folder_id UUID;
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_folder
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL;

-- Add nullable folder_id to notes
ALTER TABLE notes ADD COLUMN folder_id UUID;
ALTER TABLE notes ADD CONSTRAINT fk_notes_folder
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL;

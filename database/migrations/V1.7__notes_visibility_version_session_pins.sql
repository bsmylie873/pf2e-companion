-- 1. Add visibility, version, and session_id to notes
ALTER TABLE notes
    ADD COLUMN visibility VARCHAR(10) NOT NULL DEFAULT 'private',
    ADD COLUMN version    INTEGER     NOT NULL DEFAULT 1,
    ADD COLUMN session_id UUID;

ALTER TABLE notes
    ADD CONSTRAINT fk_notes_session FOREIGN KEY (session_id) REFERENCES sessions (id);

-- 2. Drop the old ownership check constraint (game_id XOR user_id).
ALTER TABLE notes DROP CONSTRAINT chk_notes_ownership;

-- Make game_id NOT NULL and user_id NOT NULL
ALTER TABLE notes ALTER COLUMN game_id SET NOT NULL;
ALTER TABLE notes ALTER COLUMN user_id SET NOT NULL;

-- 3. Add nullable note_id and game_id to session_pins
ALTER TABLE session_pins
    ADD COLUMN note_id UUID,
    ADD COLUMN game_id UUID;

ALTER TABLE session_pins
    ADD CONSTRAINT fk_session_pins_note FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE SET NULL;

ALTER TABLE session_pins
    ADD CONSTRAINT fk_session_pins_game FOREIGN KEY (game_id) REFERENCES games(id);

-- Backfill game_id from sessions
UPDATE session_pins SET game_id = sessions.game_id FROM sessions WHERE session_pins.session_id = sessions.id;

ALTER TABLE session_pins ALTER COLUMN game_id SET NOT NULL;

-- Make session_id nullable
ALTER TABLE session_pins ALTER COLUMN session_id DROP NOT NULL;

-- Drop the unique constraint on session_id
ALTER TABLE session_pins DROP CONSTRAINT uq_session_pins_session;

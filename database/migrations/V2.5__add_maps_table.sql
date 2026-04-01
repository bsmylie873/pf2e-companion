-- 1. Create the maps table
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

-- 2. Migrate existing map_image_url data into maps table
INSERT INTO maps (game_id, name, image_url, sort_order)
SELECT id, 'Default Map', map_image_url, 0
FROM games
WHERE map_image_url IS NOT NULL;

-- 3. Add map_id column to session_pins (nullable initially)
ALTER TABLE session_pins ADD COLUMN map_id UUID REFERENCES maps(id) ON DELETE CASCADE;

-- 4. Backfill map_id for existing session_pins
UPDATE session_pins sp
SET map_id = m.id
FROM maps m
WHERE sp.game_id = m.game_id AND sp.map_id IS NULL;

-- 5. Add map_id column to pin_groups (nullable initially)
ALTER TABLE pin_groups ADD COLUMN map_id UUID REFERENCES maps(id) ON DELETE CASCADE;

-- 6. Backfill map_id for existing pin_groups
UPDATE pin_groups pg
SET map_id = m.id
FROM maps m
WHERE pg.game_id = m.game_id AND pg.map_id IS NULL;

-- 7. Create indexes
CREATE INDEX idx_session_pins_map_id ON session_pins(map_id);
CREATE INDEX idx_pin_groups_map_id ON pin_groups(map_id);

-- 8. Drop legacy column from games
ALTER TABLE games DROP COLUMN map_image_url;

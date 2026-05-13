CREATE TABLE party_markers (
    id         UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id    UUID          NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    map_id     UUID          NOT NULL REFERENCES maps(id) ON DELETE CASCADE,
    x          NUMERIC(6,4)  NOT NULL,
    y          NUMERIC(6,4)  NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_party_markers_game_id ON party_markers(game_id);
CREATE INDEX idx_party_markers_map_id ON party_markers(map_id);

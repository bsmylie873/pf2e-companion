-- V1.9: Add pin_groups table and link session_pins to groups

CREATE TABLE pin_groups (
    id         UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    game_id    UUID          NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    x          NUMERIC(6,4)  NOT NULL,
    y          NUMERIC(6,4)  NOT NULL,
    colour     VARCHAR(20)   NOT NULL DEFAULT 'grey',
    icon       VARCHAR(50)   NOT NULL DEFAULT 'position-marker',
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

ALTER TABLE session_pins
    ADD COLUMN group_id UUID REFERENCES pin_groups(id) ON DELETE SET NULL;

CREATE INDEX idx_session_pins_group_id ON session_pins(group_id);

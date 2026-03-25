-- Add map image URL to games table
ALTER TABLE games ADD COLUMN map_image_url TEXT;

-- Create session_pins table
CREATE TABLE session_pins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    label TEXT NOT NULL DEFAULT '',
    x NUMERIC(6,4) NOT NULL,
    y NUMERIC(6,4) NOT NULL,
    pin_type TEXT NOT NULL DEFAULT 'default',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_session_pins_session UNIQUE (session_id)
);

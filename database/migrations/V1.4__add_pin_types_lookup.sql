-- Create pin_types lookup table
CREATE TABLE pin_types (
    id   SERIAL      PRIMARY KEY,
    name TEXT        NOT NULL UNIQUE
);

-- Seed with initial values
INSERT INTO pin_types (name) VALUES ('up'), ('down');

-- Replace the text column with a foreign key reference
ALTER TABLE session_pins ADD COLUMN pin_type_id INTEGER;

-- Backfill existing rows: map text values to the new FK
UPDATE session_pins SET pin_type_id = pt.id
FROM pin_types pt
WHERE session_pins.pin_type = pt.name;

-- Any remaining rows (pin_type was 'default' or unmatched) get 'down'
UPDATE session_pins SET pin_type_id = (SELECT id FROM pin_types WHERE name = 'down')
WHERE pin_type_id IS NULL;

-- Now enforce NOT NULL and FK constraint
ALTER TABLE session_pins ALTER COLUMN pin_type_id SET NOT NULL;
ALTER TABLE session_pins ALTER COLUMN pin_type_id SET DEFAULT 2; -- 'down'
ALTER TABLE session_pins
    ADD CONSTRAINT fk_session_pins_pin_type
    FOREIGN KEY (pin_type_id) REFERENCES pin_types (id);

-- Drop the old text column
ALTER TABLE session_pins DROP COLUMN pin_type;

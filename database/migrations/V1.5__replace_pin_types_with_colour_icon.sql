-- Add colour and icon columns to session_pins with sensible defaults
ALTER TABLE session_pins ADD COLUMN colour VARCHAR(20) NOT NULL DEFAULT 'grey';
ALTER TABLE session_pins ADD COLUMN icon   VARCHAR(50) NOT NULL DEFAULT 'map-marker';

-- Drop the foreign key constraint referencing pin_types
ALTER TABLE session_pins DROP CONSTRAINT IF EXISTS fk_session_pins_pin_type;

-- Drop the now-unused pin_type_id column
ALTER TABLE session_pins DROP COLUMN pin_type_id;

-- Drop the pin_types lookup table
DROP TABLE IF EXISTS pin_types;

ALTER TABLE session_pins ALTER COLUMN icon SET DEFAULT 'position-marker';
UPDATE session_pins SET icon = 'position-marker' WHERE icon = 'map-marker';

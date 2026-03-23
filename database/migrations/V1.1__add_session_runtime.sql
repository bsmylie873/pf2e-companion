-- Add runtime start/end timestamps to sessions
ALTER TABLE sessions ADD COLUMN runtime_start TIMESTAMPTZ;
ALTER TABLE sessions ADD COLUMN runtime_end   TIMESTAMPTZ;

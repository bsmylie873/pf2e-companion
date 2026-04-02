-- Repurpose the version column as an OT (operational transform) sequence counter.
-- Previously used for optimistic locking; now incremented on every update to
-- serve as a monotonically increasing sequence for operational transform operations.
COMMENT ON COLUMN notes.version IS 'OT sequence counter; incremented on every update';
COMMENT ON COLUMN sessions.version IS 'OT sequence counter; incremented on every update';

-- ---------------------------------------------------------------------------
-- V2 — Add page_size preference to user_preferences
-- ---------------------------------------------------------------------------
-- Stores per-user pagination preferences as JSONB.
-- Shape: { "default": 10, "campaigns": null, "sessions": null, "notes": null }
-- null values mean "use the default".
-- ---------------------------------------------------------------------------

ALTER TABLE user_preferences
    ADD COLUMN page_size JSONB DEFAULT '{"default":10}'::jsonb;

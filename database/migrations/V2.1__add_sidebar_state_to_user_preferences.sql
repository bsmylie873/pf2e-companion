-- V2.1: Add sidebar_state JSONB column for per-game sidebar UI state
ALTER TABLE user_preferences
    ADD COLUMN sidebar_state JSONB DEFAULT '{}'::jsonb;

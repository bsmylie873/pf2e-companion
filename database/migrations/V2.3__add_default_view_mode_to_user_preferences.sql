-- V2.3: Add default_view_mode JSONB column for per-user default view mode preference
ALTER TABLE user_preferences
    ADD COLUMN default_view_mode JSONB DEFAULT '{}'::jsonb;

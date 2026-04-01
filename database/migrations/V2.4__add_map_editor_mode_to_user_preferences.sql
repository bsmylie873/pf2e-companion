-- V2.3: Add map_editor_mode column to user_preferences
ALTER TABLE user_preferences
    ADD COLUMN map_editor_mode VARCHAR(10) NOT NULL DEFAULT 'modal';

-- V2.0: Add user_preferences table for per-user defaults

CREATE TABLE user_preferences (
    user_id           UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    default_game_id   UUID         REFERENCES games(id) ON DELETE SET NULL,
    default_pin_colour VARCHAR(20),
    default_pin_icon  VARCHAR(50),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id)
);

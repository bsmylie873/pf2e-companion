-- =============================================================================
-- V1__seed.sql — Local development seed data
-- All inserts are idempotent via ON CONFLICT DO NOTHING.
-- Password for all users: password123
-- =============================================================================

-- -----------------------------------------------------------------------------
-- users
-- -----------------------------------------------------------------------------
INSERT INTO users (id, username, email, password_hash, avatar_url, foundry_data, created_at, updated_at)
VALUES
    ('a1b2c3d4-0001-4000-8000-000000000001', 'gm_valen',     'valen@example.com', '$2y$10$4ZtZMk0tY40kYxuxOC2UuS0xdr5qAC7ya0vqRV201Uc56YsqY9di', NULL, NULL, NOW(), NOW()),
    ('a1b2c3d4-0002-4000-8000-000000000002', 'player_elara', 'elara@example.com', '$2y$10$4ZtZMk0tY40kYxuxOC2UuS0xdr5qAC7ya0vqRV201Uc56YsqY9di', NULL, NULL, NOW(), NOW()),
    ('a1b2c3d4-0003-4000-8000-000000000003', 'player_dorn',  'dorn@example.com',  '$2y$10$4ZtZMk0tY40kYxuxOC2UuS0xdr5qAC7ya0vqRV201Uc56YsqY9di', NULL, NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- games
-- -----------------------------------------------------------------------------
INSERT INTO games (id, title, description, splash_image_url, foundry_data, created_at, updated_at)
VALUES
    ('b1b2c3d4-0001-4000-8000-000000000001', 'Rise of the Runewardens', 'A homebrew PF2e campaign', NULL, NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- game_memberships
-- -----------------------------------------------------------------------------
INSERT INTO game_memberships (id, game_id, user_id, is_gm, foundry_data, created_at, updated_at)
VALUES
    ('c1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', true,  NULL, NOW(), NOW()),
    ('c1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0002-4000-8000-000000000002', false, NULL, NOW(), NOW()),
    ('c1b2c3d4-0003-4000-8000-000000000003', 'b1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0003-4000-8000-000000000003', false, NULL, NOW(), NOW())
ON CONFLICT (game_id, user_id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- sessions
-- -----------------------------------------------------------------------------
INSERT INTO sessions (id, game_id, title, session_number, scheduled_at, notes, version, foundry_data, created_at, updated_at)
VALUES
    ('d1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'The Goblin Warrens', 1, '2025-01-15 18:00:00+00', '{}'::jsonb, 1, NULL, NOW(), NOW()),
    ('d1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0001-4000-8000-000000000001', 'Into the Ruins',     2, '2025-02-15 18:00:00+00', '{}'::jsonb, 1, NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- notes
-- -----------------------------------------------------------------------------
INSERT INTO notes (id, game_id, user_id, title, content, foundry_data, created_at, updated_at)
VALUES
    (
        'e1b2c3d4-0001-4000-8000-000000000001',
        NULL,
        'a1b2c3d4-0002-4000-8000-000000000002',
        'Elara''s Notes',
        '{"text": "Personal notes about the campaign."}'::jsonb,
        NULL, NOW(), NOW()
    ),
    (
        'e1b2c3d4-0002-4000-8000-000000000002',
        'b1b2c3d4-0001-4000-8000-000000000001',
        NULL,
        'Campaign Lore',
        '{"text": "Key lore and world details."}'::jsonb,
        NULL, NOW(), NOW()
    )
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- characters
-- -----------------------------------------------------------------------------
INSERT INTO characters (
    id, game_id, user_id, name, is_npc,
    ancestry, heritage, class, background,
    level, hp_max, hp_current, ac,
    strength, dexterity, constitution, intelligence, wisdom, charisma,
    fortitude, reflex, will,
    skills, foundry_data, created_at, updated_at
)
VALUES
    (
        'f1b2c3d4-0001-4000-8000-000000000001',
        'b1b2c3d4-0001-4000-8000-000000000001',
        'a1b2c3d4-0002-4000-8000-000000000002',
        'Elara Brightleaf', false,
        'Elf', 'Skilled Heritage', 'Ranger', 'Bounty Hunter',
        3, 36, 36, 18,
        10, 18, 14, 12, 16, 10,
        7, 9, 8,
        '{"Acrobatics": 9, "Athletics": 5, "Nature": 8, "Stealth": 9, "Survival": 8, "Perception": 8}'::jsonb,
        NULL, NOW(), NOW()
    ),
    (
        'f1b2c3d4-0002-4000-8000-000000000002',
        'b1b2c3d4-0001-4000-8000-000000000001',
        'a1b2c3d4-0003-4000-8000-000000000003',
        'Dorn Ironfist', false,
        'Dwarf', 'Anvil Dwarf Heritage', 'Fighter', 'Guard',
        3, 46, 46, 20,
        18, 12, 16, 10, 14, 8,
        9, 6, 7,
        '{"Athletics": 9, "Intimidation": 5, "Medicine": 7, "Perception": 7, "Society": 5}'::jsonb,
        NULL, NOW(), NOW()
    )
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- items
-- -----------------------------------------------------------------------------
INSERT INTO items (id, game_id, character_id, name, description, level, price_gp, bulk, traits, quantity, foundry_data, created_at, updated_at)
VALUES
    (
        '01b2c3d4-0001-4000-8000-000000000001',
        'b1b2c3d4-0001-4000-8000-000000000001',
        'f1b2c3d4-0001-4000-8000-000000000001',
        '+1 Composite Longbow',
        'A finely crafted longbow with a +1 potency rune.',
        3, 50, '1',
        ARRAY['Deadly d10', 'Propulsive', 'Volley 30 ft.'],
        1, NULL, NOW(), NOW()
    ),
    (
        '01b2c3d4-0002-4000-8000-000000000002',
        'b1b2c3d4-0001-4000-8000-000000000001',
        'f1b2c3d4-0002-4000-8000-000000000002',
        'Dwarven Waraxe',
        'A sturdy waraxe of dwarven make.',
        1, 3, '1',
        ARRAY['Dwarf', 'Sweep', 'Two-Hand d12'],
        1, NULL, NOW(), NOW()
    )
ON CONFLICT (id) DO NOTHING;

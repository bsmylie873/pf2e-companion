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
    ('a1b2c3d4-0001-4000-8000-000000000001', 'gm_valen',     'valen@example.com', '$2a$10$4q/Kh9TstbqRMtajn7Fb6e.uy8YKIZfGIj8qW5Scpr/ML2wpLqwMS', NULL, NULL, NOW(), NOW()),
    ('a1b2c3d4-0002-4000-8000-000000000002', 'player_elara', 'elara@example.com', '$2a$10$4q/Kh9TstbqRMtajn7Fb6e.uy8YKIZfGIj8qW5Scpr/ML2wpLqwMS', NULL, NULL, NOW(), NOW()),
    ('a1b2c3d4-0003-4000-8000-000000000003', 'player_dorn',  'dorn@example.com',  '$2a$10$4q/Kh9TstbqRMtajn7Fb6e.uy8YKIZfGIj8qW5Scpr/ML2wpLqwMS', NULL, NULL, NOW(), NOW())
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
INSERT INTO sessions (id, game_id, title, session_number, scheduled_at, runtime_start, runtime_end, notes, version, foundry_data, created_at, updated_at)
VALUES
    ('d1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'The Goblin Warrens', 1, '2025-01-15 18:00:00+00', '2025-01-15 18:00:00+00', '2025-01-15 22:30:00+00', '{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"The Goblin Warrens"}]},{"type":"paragraph","content":[{"type":"text","text":"The party ventured into the goblin warrens beneath "},{"type":"text","marks":[{"type":"bold"}],"text":"Otari"},{"type":"text","text":". What they found was far worse than expected."}]},{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"italic"}],"text":"\"You tread where even goblins fear to wander, adventurers. Turn back, or join the bones.\" — Skraggle, Goblin Chieftain"}]}]},{"type":"horizontalRule"},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Loot Recovered"}]},{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Item"}]}]},{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Qty"}]}]},{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Value (gp)"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Healing Potion (Minor)"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"3"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"12"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Goblin-forged Dagger"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"1"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"2"}]}]}]}]},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Action Items"}]},{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":true},"content":[{"type":"paragraph","content":[{"type":"text","text":"Report findings to Captain Longsaddle"}]}]},{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"Identify the strange rune found in the lower cavern"}]}]},{"type":"taskItem","attrs":{"checked":false},"content":[{"type":"paragraph","content":[{"type":"text","text":"Buy healing potions before next session"}]}]}]},{"type":"paragraph","content":[{"type":"text","text":"Reference: "},{"type":"text","marks":[{"type":"link","attrs":{"href":"https://2e.aonprd.com/","target":"_blank"}}],"text":"Archives of Nethys"},{"type":"text","text":" for goblin stat blocks."}]},{"type":"paragraph","content":[{"type":"text","text":"Key moment: Elara "},{"type":"text","marks":[{"type":"strike"}],"text":"attacked"},{"type":"text","text":" negotiated with the goblin scouts, earning passage through the first gate. "},{"type":"text","marks":[{"type":"highlight"}],"text":"This alliance may prove crucial later."}]}]}'::jsonb, 1, NULL, NOW(), NOW()),
    ('d1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0001-4000-8000-000000000001', 'Into the Ruins',     2, '2025-02-15 18:00:00+00', '2025-02-15 18:15:00+00', '2025-02-15 21:45:00+00', '{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Into the Ruins"}]},{"type":"paragraph","content":[{"type":"text","text":"Session two took the party deeper into the ancient Thassilonian ruins beneath the warrens."}]},{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","marks":[{"type":"italic"}],"text":"The walls here are carved with runes older than any living language. Dorn felt the hairs on his neck rise."}]}]},{"type":"horizontalRule"},{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Key Events"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Discovered a hidden chamber with "},{"type":"text","marks":[{"type":"highlight"}],"text":"ancient Thassilonian inscriptions"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Fought a "},{"type":"text","marks":[{"type":"bold"}],"text":"giant spider"},{"type":"text","text":" guarding the entrance to the lower level"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Found a "},{"type":"text","marks":[{"type":"link","attrs":{"href":"https://2e.aonprd.com/Equipment.aspx?ID=244","target":"_blank"}}],"text":"Wand of Heal"},{"type":"text","text":" hidden in a collapsed alcove"}]}]}]},{"type":"paragraph","content":[{"type":"text","marks":[{"type":"underline"}],"text":"Next session"},{"type":"text","text":": The party must decide whether to press deeper or return to Otari to resupply."}]}]}'::jsonb, 1, NULL, NOW(), NOW()),
    ('d1b2c3d4-0003-4000-8000-000000000003', 'b1b2c3d4-0001-4000-8000-000000000001', 'The Crimson Pact',     3, '2025-03-15 18:00:00+00', '2025-03-15 18:00:00+00', '2025-03-15 23:00:00+00', '{}'::jsonb, 1, NULL, NOW(), NOW()),
    ('d1b2c3d4-0004-4000-8000-000000000004', 'b1b2c3d4-0001-4000-8000-000000000001', 'Downtime in Otari',   NULL, NULL, NULL, NULL, '{}'::jsonb, 1, NULL, NOW(), NOW()),
    ('d1b2c3d4-0005-4000-8000-000000000005', 'b1b2c3d4-0001-4000-8000-000000000001', 'The Dragon''s Demand', 4, '2025-04-12 18:00:00+00', NULL, NULL, '{}'::jsonb, 1, NULL, NOW(), NOW())
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

-- +goose Up
INSERT INTO roles (id, code, name) VALUES
    (1, 'student', 'Student'),
    (2, 'employer', 'Employer'),
    (3, 'curator', 'Curator'),
    (4, 'admin', 'Administrator')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (
    id,
    email,
    password_hash,
    display_name,
    email_verified,
    status,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@tramplin.local',
    '$2a$10$7sQxY8x0f3wV8Q6C9mT4fOq0G4dM4V5l7UQW9jI7o7g2M1A5w4n5K',
    'System Administrator',
    TRUE,
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
VALUES
    ('00000000-0000-0000-0000-000000000001', 3),
    ('00000000-0000-0000-0000-000000000001', 4)
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO curator_profiles (
    user_id,
    curator_type,
    created_by_user_id,
    notes,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'administrator',
    NULL,
    'Bootstrap administrator created by migration. Replace credentials after first deploy.',
    NOW(),
    NOW()
)
ON CONFLICT (user_id) DO NOTHING;

-- +goose Down
DELETE FROM curator_profiles
WHERE user_id = '00000000-0000-0000-0000-000000000001';

DELETE FROM user_roles
WHERE user_id = '00000000-0000-0000-0000-000000000001'
  AND role_id IN (3, 4);

DELETE FROM users
WHERE id = '00000000-0000-0000-0000-000000000001';

DELETE FROM roles
WHERE id IN (1, 2, 3, 4);

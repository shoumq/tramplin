-- +goose Up
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
    '00000000-0000-0000-0000-000000000008',
    'admin@mail.ru',
    '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9',
    'Admin',
    TRUE,
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    display_name = EXCLUDED.display_name,
    email_verified = EXCLUDED.email_verified,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.id IN (3, 4)
WHERE u.email = 'admin@mail.ru'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO curator_profiles (
    user_id,
    curator_type,
    created_by_user_id,
    notes,
    created_at,
    updated_at
)
SELECT
    u.id,
    'administrator',
    NULL,
    'Bootstrap admin@mail.ru created by migration.',
    NOW(),
    NOW()
FROM users u
WHERE u.email = 'admin@mail.ru'
ON CONFLICT (user_id) DO NOTHING;

-- +goose Down
DELETE FROM curator_profiles
WHERE user_id IN (
    SELECT id FROM users WHERE email = 'admin@mail.ru'
);

DELETE FROM user_roles
WHERE user_id IN (
    SELECT id FROM users WHERE email = 'admin@mail.ru'
)
AND role_id IN (3, 4);

DELETE FROM users
WHERE email = 'admin@mail.ru';

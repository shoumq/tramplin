-- +goose Up
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_user_roles_role_id'
    ) THEN
        ALTER TABLE user_roles
            ADD CONSTRAINT fk_user_roles_role_id
            FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
    END IF;
END $$;

INSERT INTO cities (country, region, city_name, latitude, longitude) VALUES
    ('Russia', 'Moscow', 'Moscow', 55.7558, 37.6176),
    ('Russia', 'Saint Petersburg', 'Saint Petersburg', 59.9343, 30.3351)
ON CONFLICT (country, region, city_name) DO NOTHING;

INSERT INTO tags (name, tag_type, is_system, is_active) VALUES
    ('Go', 'technology', TRUE, TRUE),
    ('SQL', 'technology', TRUE, TRUE),
    ('Junior', 'level', TRUE, TRUE)
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM tags
WHERE name IN ('Go', 'SQL', 'Junior');

DELETE FROM cities
WHERE (country, region, city_name) IN (
    ('Russia', 'Moscow', 'Moscow'),
    ('Russia', 'Saint Petersburg', 'Saint Petersburg')
);

ALTER TABLE user_roles
    DROP CONSTRAINT IF EXISTS fk_user_roles_role_id;

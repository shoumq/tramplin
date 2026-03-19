-- +goose Up
CREATE TABLE chat_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant_a_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participant_b_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID NULL REFERENCES opportunities(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_chat_conversations_not_self CHECK (participant_a_user_id <> participant_b_user_id),
    CONSTRAINT uq_chat_conversations_pair UNIQUE (participant_a_user_id, participant_b_user_id, opportunity_id)
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES chat_conversations(id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_conversations_participant_a ON chat_conversations(participant_a_user_id);
CREATE INDEX idx_chat_conversations_participant_b ON chat_conversations(participant_b_user_id);
CREATE INDEX idx_chat_messages_conversation_created_at ON chat_messages(conversation_id, created_at DESC);

INSERT INTO users (
    id, email, password_hash, display_name, email_verified, status, created_at, updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000009',
    'lae345@mail.ru',
    '3a0acb9e242f986216a5ae6e697feb1d53e7a1afe03005ce5fc109f296e6c457',
    'lae345',
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
SELECT u.id, 1
FROM users u
WHERE u.email = 'lae345@mail.ru'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO student_profiles (
    user_id, last_name, first_name, middle_name, university_name, faculty, specialization, study_year, graduation_year, about, profile_visibility, show_resume, show_applications, show_career_interests, created_at, updated_at
)
SELECT
    u.id, 'Лясковский', 'Андрей', 'Евгеньевич', 'РЭУ им. Плеханова', 'ВШКМиС', 'МОИАИС', 3, 2027, 'Тестовый соискатель', 'authorized_only', TRUE, TRUE, TRUE, NOW(), NOW()
FROM users u
WHERE u.email = 'lae345@mail.ru'
ON CONFLICT (user_id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    middle_name = EXCLUDED.middle_name,
    university_name = EXCLUDED.university_name,
    faculty = EXCLUDED.faculty,
    specialization = EXCLUDED.specialization,
    study_year = EXCLUDED.study_year,
    graduation_year = EXCLUDED.graduation_year,
    about = EXCLUDED.about,
    profile_visibility = EXCLUDED.profile_visibility,
    show_resume = EXCLUDED.show_resume,
    show_applications = EXCLUDED.show_applications,
    show_career_interests = EXCLUDED.show_career_interests,
    updated_at = NOW();

INSERT INTO users (
    id, email, password_hash, display_name, email_verified, status, created_at, updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    'rea@mail.ru',
    '3a0acb9e242f986216a5ae6e697feb1d53e7a1afe03005ce5fc109f296e6c457',
    'rea',
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

INSERT INTO companies (
    id, legal_name, brand_name, inn, status, created_at, updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000110',
    'REA LLC',
    'REA',
    '7705043493',
    'verified',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    legal_name = EXCLUDED.legal_name,
    brand_name = EXCLUDED.brand_name,
    inn = EXCLUDED.inn,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, 2
FROM users u
WHERE u.email = 'rea@mail.ru'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO employer_profiles (
    user_id, company_id, position_title, is_company_owner, can_create_opportunities, can_edit_company_profile, created_at, updated_at
)
SELECT
    u.id, '00000000-0000-0000-0000-000000000110', 'HR Manager', TRUE, TRUE, TRUE, NOW(), NOW()
FROM users u
WHERE u.email = 'rea@mail.ru'
ON CONFLICT (user_id) DO UPDATE SET
    company_id = EXCLUDED.company_id,
    position_title = EXCLUDED.position_title,
    is_company_owner = EXCLUDED.is_company_owner,
    can_create_opportunities = EXCLUDED.can_create_opportunities,
    can_edit_company_profile = EXCLUDED.can_edit_company_profile,
    updated_at = NOW();

-- +goose Down
DELETE FROM employer_profiles
WHERE user_id IN (SELECT id FROM users WHERE email = 'rea@mail.ru');

DELETE FROM user_roles
WHERE user_id IN (
    SELECT id FROM users WHERE email IN ('rea@mail.ru', 'lae345@mail.ru')
)
AND role_id IN (1, 2);

DELETE FROM companies
WHERE id = '00000000-0000-0000-0000-000000000110';

DELETE FROM student_profiles
WHERE user_id IN (SELECT id FROM users WHERE email = 'lae345@mail.ru');

DELETE FROM users
WHERE email IN ('rea@mail.ru', 'lae345@mail.ru');

DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_conversations;

-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(120) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'blocked', 'pending_verification', 'deleted'))
);

CREATE TABLE roles (
    id SMALLINT PRIMARY KEY,
    code VARCHAR(30) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE cities (
    id BIGSERIAL PRIMARY KEY,
    country VARCHAR(100) NOT NULL,
    region VARCHAR(150) NULL,
    city_name VARCHAR(150) NOT NULL,
    latitude NUMERIC(9, 6) NULL,
    longitude NUMERIC(9, 6) NULL,
    CONSTRAINT uq_cities_country_region_name UNIQUE (country, region, city_name)
);

CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city_id BIGINT NULL REFERENCES cities(id) ON DELETE SET NULL,
    address_line VARCHAR(255) NULL,
    postal_code VARCHAR(20) NULL,
    latitude NUMERIC(9, 6) NULL,
    longitude NUMERIC(9, 6) NULL,
    location_type VARCHAR(30) NOT NULL,
    display_text VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_locations_location_type CHECK (location_type IN ('office', 'event_place', 'remote_city'))
);

CREATE TABLE media_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uploaded_by_user_id UUID NULL,
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    media_kind VARCHAR(30) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_media_files_media_kind CHECK (media_kind IN ('image', 'video', 'document'))
);

CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name VARCHAR(255) NOT NULL,
    brand_name VARCHAR(255) NULL,
    description TEXT NULL,
    industry VARCHAR(150) NULL,
    website_url TEXT NULL,
    email_domain VARCHAR(100) NULL,
    inn VARCHAR(20) NULL,
    ogrn VARCHAR(20) NULL,
    company_size VARCHAR(50) NULL,
    founded_year SMALLINT NULL,
    hq_city_id BIGINT NULL REFERENCES cities(id) ON DELETE SET NULL,
    logo_media_id UUID NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending_verification',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_companies_status CHECK (status IN ('pending_verification', 'verified', 'rejected', 'blocked'))
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id SMALLINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE student_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100) NULL,
    university_name VARCHAR(255) NOT NULL,
    faculty VARCHAR(255) NULL,
    specialization VARCHAR(255) NULL,
    study_year SMALLINT NULL,
    graduation_year SMALLINT NULL,
    about TEXT NULL,
    profile_visibility VARCHAR(30) NOT NULL DEFAULT 'authorized_only',
    show_resume BOOLEAN NOT NULL DEFAULT TRUE,
    show_applications BOOLEAN NOT NULL DEFAULT FALSE,
    show_career_interests BOOLEAN NOT NULL DEFAULT TRUE,
    telegram VARCHAR(100) NULL,
    github_url TEXT NULL,
    linkedin_url TEXT NULL,
    website_url TEXT NULL,
    city_id BIGINT NULL REFERENCES cities(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_student_profiles_visibility CHECK (profile_visibility IN ('private', 'contacts_only', 'authorized_only', 'public_inside_platform'))
);

CREATE TABLE employer_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    position_title VARCHAR(150) NULL,
    is_company_owner BOOLEAN NOT NULL DEFAULT FALSE,
    can_create_opportunities BOOLEAN NOT NULL DEFAULT FALSE,
    can_edit_company_profile BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE curator_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    curator_type VARCHAR(30) NOT NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_curator_profiles_type CHECK (curator_type IN ('administrator', 'moderator', 'university_curator'))
);

CREATE TABLE company_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    link_type VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_company_links_type CHECK (link_type IN ('website', 'telegram', 'vk', 'youtube', 'hh', 'github', 'other'))
);

CREATE TABLE company_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    verification_method VARCHAR(50) NOT NULL,
    submitted_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    corporate_email VARCHAR(255) NULL,
    inn_submitted VARCHAR(20) NULL,
    documents_comment TEXT NULL,
    status VARCHAR(30) NOT NULL,
    reviewed_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    review_comment TEXT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_company_verifications_method CHECK (verification_method IN ('corporate_email', 'inn_check', 'manual_documents', 'social_links_review', 'combined')),
    CONSTRAINT chk_company_verifications_status CHECK (status IN ('pending', 'approved', 'rejected', 'needs_revision'))
);

CREATE TABLE opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    short_description TEXT NOT NULL,
    full_description TEXT NOT NULL,
    opportunity_type VARCHAR(30) NOT NULL,
    vacancy_level VARCHAR(30) NULL,
    employment_type VARCHAR(30) NULL,
    work_format VARCHAR(30) NOT NULL,
    location_id UUID NULL REFERENCES locations(id) ON DELETE SET NULL,
    salary_min NUMERIC(12, 2) NULL,
    salary_max NUMERIC(12, 2) NULL,
    salary_currency CHAR(3) NULL,
    is_salary_visible BOOLEAN NOT NULL DEFAULT FALSE,
    application_deadline TIMESTAMPTZ NULL,
    event_start_at TIMESTAMPTZ NULL,
    event_end_at TIMESTAMPTZ NULL,
    published_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    contacts_info TEXT NULL,
    external_url TEXT NULL,
    views_count INTEGER NOT NULL DEFAULT 0,
    favorites_count INTEGER NOT NULL DEFAULT 0,
    applications_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_opportunities_opportunity_type CHECK (opportunity_type IN ('internship', 'vacancy', 'mentorship', 'event')),
    CONSTRAINT chk_opportunities_vacancy_level CHECK (vacancy_level IS NULL OR vacancy_level IN ('intern', 'junior', 'middle', 'senior')),
    CONSTRAINT chk_opportunities_employment_type CHECK (employment_type IS NULL OR employment_type IN ('full_time', 'part_time', 'project', 'temporary')),
    CONSTRAINT chk_opportunities_work_format CHECK (work_format IN ('office', 'hybrid', 'remote')),
    CONSTRAINT chk_opportunities_status CHECK (status IN ('draft', 'pending_moderation', 'published', 'rejected', 'archived', 'closed', 'scheduled')),
    CONSTRAINT chk_opportunities_salary_range CHECK (
        salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max
    ),
    CONSTRAINT chk_opportunities_event_dates CHECK (
        event_start_at IS NULL OR event_end_at IS NULL OR event_start_at <= event_end_at
    )
);

CREATE TABLE opportunity_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    tag_type VARCHAR(30) NOT NULL,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tags_tag_type CHECK (tag_type IN ('technology', 'level', 'employment', 'role', 'skill', 'format', 'custom'))
);

CREATE TABLE opportunity_tags (
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (opportunity_id, tag_id)
);

CREATE TABLE student_profile_tags (
    student_user_id UUID NOT NULL REFERENCES student_profiles(user_id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    proficiency_level VARCHAR(30) NULL,
    PRIMARY KEY (student_user_id, tag_id),
    CONSTRAINT chk_student_profile_tags_level CHECK (proficiency_level IS NULL OR proficiency_level IN ('beginner', 'junior', 'intermediate', 'advanced'))
);

CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_user_id UUID NOT NULL REFERENCES student_profiles(user_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    summary TEXT NULL,
    experience_text TEXT NULL,
    education_text TEXT NULL,
    resume_file_media_id UUID NULL REFERENCES media_files(id) ON DELETE SET NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE portfolio_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_user_id UUID NOT NULL REFERENCES student_profiles(user_id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    project_url TEXT NULL,
    repository_url TEXT NULL,
    demo_url TEXT NULL,
    started_at DATE NULL,
    finished_at DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_portfolio_projects_dates CHECK (
        started_at IS NULL OR finished_at IS NULL OR started_at <= finished_at
    )
);

CREATE TABLE profile_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
    media_purpose VARCHAR(30) NOT NULL,
    CONSTRAINT chk_profile_media_purpose CHECK (media_purpose IN ('avatar', 'cover', 'resume_attachment', 'portfolio_attachment'))
);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    student_user_id UUID NOT NULL REFERENCES student_profiles(user_id) ON DELETE CASCADE,
    resume_id UUID NULL REFERENCES resumes(id) ON DELETE SET NULL,
    cover_letter TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'submitted',
    status_changed_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    status_changed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_applications_opportunity_student UNIQUE (opportunity_id, student_user_id),
    CONSTRAINT chk_applications_status CHECK (status IN ('submitted', 'in_review', 'accepted', 'rejected', 'reserve', 'withdrawn'))
);

CREATE TABLE application_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    old_status VARCHAR(30) NULL,
    new_status VARCHAR(30) NOT NULL,
    changed_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_application_status_history_old_status CHECK (
        old_status IS NULL OR old_status IN ('submitted', 'in_review', 'accepted', 'rejected', 'reserve', 'withdrawn')
    ),
    CONSTRAINT chk_application_status_history_new_status CHECK (
        new_status IN ('submitted', 'in_review', 'accepted', 'rejected', 'reserve', 'withdrawn')
    )
);

CREATE TABLE favorite_opportunities (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, opportunity_id)
);

CREATE TABLE favorite_companies (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, company_id)
);

CREATE TABLE contact_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_contact_requests_status CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled')),
    CONSTRAINT chk_contact_requests_not_self CHECK (sender_user_id <> receiver_user_id)
);

CREATE TABLE contacts (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, contact_user_id),
    CONSTRAINT chk_contacts_not_self CHECK (user_id <> contact_user_id)
);

CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_recommendations_not_self CHECK (from_user_id <> to_user_id)
);

CREATE TABLE moderation_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(30) NOT NULL,
    entity_id UUID NOT NULL,
    submitted_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_to_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    moderator_comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_moderation_queue_entity_type CHECK (entity_type IN ('company', 'opportunity', 'student_profile', 'employer_profile', 'tag')),
    CONSTRAINT chk_moderation_queue_status CHECK (status IN ('pending', 'in_review', 'approved', 'rejected', 'needs_revision'))
);

CREATE TABLE moderation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL REFERENCES moderation_queue(id) ON DELETE CASCADE,
    moderator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action_type VARCHAR(30) NOT NULL,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_moderation_actions_type CHECK (action_type IN ('approve', 'reject', 'request_revision', 'edit', 'block', 'restore'))
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(30) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_data_json JSONB NULL,
    new_data_json JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_audit_logs_action CHECK (
        action IN ('create', 'update', 'delete', 'publish', 'close', 'status_change', 'verification_submit', 'moderation_approve', 'moderation_reject')
    )
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    related_entity_type VARCHAR(30) NULL,
    related_entity_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_notifications_type CHECK (
        type IN (
            'application_submitted',
            'application_status_changed',
            'contact_request_received',
            'contact_request_accepted',
            'recommendation_received',
            'moderation_result',
            'company_verification_result'
        )
    )
);

ALTER TABLE media_files
    ADD CONSTRAINT fk_media_files_uploaded_by_user
        FOREIGN KEY (uploaded_by_user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE companies
    ADD CONSTRAINT fk_companies_logo_media
        FOREIGN KEY (logo_media_id) REFERENCES media_files(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX uq_resumes_primary_per_student
    ON resumes(student_user_id)
    WHERE is_primary = TRUE;

CREATE INDEX idx_locations_city_id ON locations(city_id);
CREATE INDEX idx_locations_latitude_longitude ON locations(latitude, longitude);

CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_hq_city_id ON companies(hq_city_id);

CREATE INDEX idx_employer_profiles_company_id ON employer_profiles(company_id);

CREATE INDEX idx_company_links_company_id ON company_links(company_id);

CREATE INDEX idx_company_verifications_company_id ON company_verifications(company_id);
CREATE INDEX idx_company_verifications_status ON company_verifications(status);

CREATE INDEX idx_opportunities_status ON opportunities(status);
CREATE INDEX idx_opportunities_opportunity_type ON opportunities(opportunity_type);
CREATE INDEX idx_opportunities_work_format ON opportunities(work_format);
CREATE INDEX idx_opportunities_company_id ON opportunities(company_id);
CREATE INDEX idx_opportunities_published_at ON opportunities(published_at);
CREATE INDEX idx_opportunities_expires_at ON opportunities(expires_at);
CREATE INDEX idx_opportunities_location_id ON opportunities(location_id);

CREATE INDEX idx_opportunity_media_opportunity_id ON opportunity_media(opportunity_id);
CREATE INDEX idx_opportunity_media_media_id ON opportunity_media(media_id);

CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_tag_type ON tags(tag_type);

CREATE INDEX idx_opportunity_tags_tag_id ON opportunity_tags(tag_id);
CREATE INDEX idx_student_profile_tags_tag_id ON student_profile_tags(tag_id);

CREATE INDEX idx_resumes_student_user_id ON resumes(student_user_id);
CREATE INDEX idx_portfolio_projects_student_user_id ON portfolio_projects(student_user_id);

CREATE INDEX idx_profile_media_user_id ON profile_media(user_id);
CREATE INDEX idx_profile_media_media_id ON profile_media(media_id);

CREATE INDEX idx_applications_student_user_id ON applications(student_user_id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_opportunity_id ON applications(opportunity_id);

CREATE INDEX idx_application_status_history_application_id ON application_status_history(application_id);

CREATE INDEX idx_favorite_opportunities_user_id ON favorite_opportunities(user_id);
CREATE INDEX idx_favorite_companies_user_id ON favorite_companies(user_id);

CREATE INDEX idx_contact_requests_sender_user_id ON contact_requests(sender_user_id);
CREATE INDEX idx_contact_requests_receiver_user_id ON contact_requests(receiver_user_id);
CREATE INDEX idx_contact_requests_status ON contact_requests(status);

CREATE INDEX idx_recommendations_from_user_id ON recommendations(from_user_id);
CREATE INDEX idx_recommendations_to_user_id ON recommendations(to_user_id);
CREATE INDEX idx_recommendations_opportunity_id ON recommendations(opportunity_id);

CREATE INDEX idx_moderation_queue_status_entity_type ON moderation_queue(status, entity_type);
CREATE INDEX idx_moderation_queue_assigned_to_user_id ON moderation_queue(assigned_to_user_id);

CREATE INDEX idx_moderation_actions_queue_id ON moderation_actions(queue_id);
CREATE INDEX idx_audit_logs_actor_user_id ON audit_logs(actor_user_id);
CREATE INDEX idx_audit_logs_entity_type_entity_id ON audit_logs(entity_type, entity_id);

CREATE INDEX idx_notifications_user_id_is_read ON notifications(user_id, is_read);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS moderation_actions;
DROP TABLE IF EXISTS moderation_queue;
DROP TABLE IF EXISTS recommendations;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS contact_requests;
DROP TABLE IF EXISTS favorite_companies;
DROP TABLE IF EXISTS favorite_opportunities;
DROP TABLE IF EXISTS application_status_history;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS profile_media;
DROP TABLE IF EXISTS portfolio_projects;
DROP TABLE IF EXISTS resumes;
DROP TABLE IF EXISTS student_profile_tags;
DROP TABLE IF EXISTS opportunity_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS opportunity_media;
DROP TABLE IF EXISTS opportunities;
DROP TABLE IF EXISTS company_verifications;
DROP TABLE IF EXISTS company_links;
DROP TABLE IF EXISTS curator_profiles;
DROP TABLE IF EXISTS employer_profiles;
DROP TABLE IF EXISTS student_profiles;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS media_files;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

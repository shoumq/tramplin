-- +goose Up
CREATE TABLE internship_opportunities (
    opportunity_id UUID PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    vacancy_level VARCHAR(30) NULL,
    employment_type VARCHAR(30) NULL,
    salary_min NUMERIC(12, 2) NULL,
    salary_max NUMERIC(12, 2) NULL,
    salary_currency CHAR(3) NULL,
    is_salary_visible BOOLEAN NOT NULL DEFAULT FALSE,
    application_deadline TIMESTAMPTZ NULL,
    CONSTRAINT chk_internship_opportunities_vacancy_level CHECK (vacancy_level IS NULL OR vacancy_level IN ('intern', 'junior', 'middle', 'senior')),
    CONSTRAINT chk_internship_opportunities_employment_type CHECK (employment_type IS NULL OR employment_type IN ('full_time', 'part_time', 'project', 'temporary')),
    CONSTRAINT chk_internship_opportunities_salary_range CHECK (
        salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max
    )
);

CREATE TABLE vacancy_opportunities (
    opportunity_id UUID PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    vacancy_level VARCHAR(30) NULL,
    employment_type VARCHAR(30) NULL,
    salary_min NUMERIC(12, 2) NULL,
    salary_max NUMERIC(12, 2) NULL,
    salary_currency CHAR(3) NULL,
    is_salary_visible BOOLEAN NOT NULL DEFAULT FALSE,
    application_deadline TIMESTAMPTZ NULL,
    CONSTRAINT chk_vacancy_opportunities_vacancy_level CHECK (vacancy_level IS NULL OR vacancy_level IN ('intern', 'junior', 'middle', 'senior')),
    CONSTRAINT chk_vacancy_opportunities_employment_type CHECK (employment_type IS NULL OR employment_type IN ('full_time', 'part_time', 'project', 'temporary')),
    CONSTRAINT chk_vacancy_opportunities_salary_range CHECK (
        salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max
    )
);

CREATE TABLE mentorship_opportunities (
    opportunity_id UUID PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    application_deadline TIMESTAMPTZ NULL
);

CREATE TABLE event_opportunities (
    opportunity_id UUID PRIMARY KEY REFERENCES opportunities(id) ON DELETE CASCADE,
    application_deadline TIMESTAMPTZ NULL,
    event_start_at TIMESTAMPTZ NULL,
    event_end_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_event_opportunities_dates CHECK (
        event_start_at IS NULL OR event_end_at IS NULL OR event_start_at <= event_end_at
    )
);

INSERT INTO internship_opportunities (
    opportunity_id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
)
SELECT
    id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
FROM opportunities
WHERE opportunity_type = 'internship'
ON CONFLICT (opportunity_id) DO NOTHING;

INSERT INTO vacancy_opportunities (
    opportunity_id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
)
SELECT
    id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
FROM opportunities
WHERE opportunity_type = 'vacancy'
ON CONFLICT (opportunity_id) DO NOTHING;

INSERT INTO mentorship_opportunities (
    opportunity_id, application_deadline
)
SELECT
    id, application_deadline
FROM opportunities
WHERE opportunity_type = 'mentorship'
ON CONFLICT (opportunity_id) DO NOTHING;

INSERT INTO event_opportunities (
    opportunity_id, application_deadline, event_start_at, event_end_at
)
SELECT
    id, application_deadline, event_start_at, event_end_at
FROM opportunities
WHERE opportunity_type = 'event'
ON CONFLICT (opportunity_id) DO NOTHING;

ALTER TABLE opportunities DROP CONSTRAINT IF EXISTS chk_opportunities_vacancy_level;
ALTER TABLE opportunities DROP CONSTRAINT IF EXISTS chk_opportunities_employment_type;
ALTER TABLE opportunities DROP CONSTRAINT IF EXISTS chk_opportunities_salary_range;
ALTER TABLE opportunities DROP CONSTRAINT IF EXISTS chk_opportunities_event_dates;

ALTER TABLE opportunities
    DROP COLUMN IF EXISTS vacancy_level,
    DROP COLUMN IF EXISTS employment_type,
    DROP COLUMN IF EXISTS salary_min,
    DROP COLUMN IF EXISTS salary_max,
    DROP COLUMN IF EXISTS salary_currency,
    DROP COLUMN IF EXISTS is_salary_visible,
    DROP COLUMN IF EXISTS application_deadline,
    DROP COLUMN IF EXISTS event_start_at,
    DROP COLUMN IF EXISTS event_end_at;

-- +goose Down
ALTER TABLE opportunities
    ADD COLUMN IF NOT EXISTS vacancy_level VARCHAR(30) NULL,
    ADD COLUMN IF NOT EXISTS employment_type VARCHAR(30) NULL,
    ADD COLUMN IF NOT EXISTS salary_min NUMERIC(12, 2) NULL,
    ADD COLUMN IF NOT EXISTS salary_max NUMERIC(12, 2) NULL,
    ADD COLUMN IF NOT EXISTS salary_currency CHAR(3) NULL,
    ADD COLUMN IF NOT EXISTS is_salary_visible BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS application_deadline TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS event_start_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS event_end_at TIMESTAMPTZ NULL;

UPDATE opportunities o
SET
    vacancy_level = io.vacancy_level,
    employment_type = io.employment_type,
    salary_min = io.salary_min,
    salary_max = io.salary_max,
    salary_currency = io.salary_currency,
    is_salary_visible = io.is_salary_visible,
    application_deadline = io.application_deadline
FROM internship_opportunities io
WHERE io.opportunity_id = o.id AND o.opportunity_type = 'internship';

UPDATE opportunities o
SET
    vacancy_level = vo.vacancy_level,
    employment_type = vo.employment_type,
    salary_min = vo.salary_min,
    salary_max = vo.salary_max,
    salary_currency = vo.salary_currency,
    is_salary_visible = vo.is_salary_visible,
    application_deadline = vo.application_deadline
FROM vacancy_opportunities vo
WHERE vo.opportunity_id = o.id AND o.opportunity_type = 'vacancy';

UPDATE opportunities o
SET application_deadline = mo.application_deadline
FROM mentorship_opportunities mo
WHERE mo.opportunity_id = o.id AND o.opportunity_type = 'mentorship';

UPDATE opportunities o
SET
    application_deadline = eo.application_deadline,
    event_start_at = eo.event_start_at,
    event_end_at = eo.event_end_at
FROM event_opportunities eo
WHERE eo.opportunity_id = o.id AND o.opportunity_type = 'event';

ALTER TABLE opportunities
    ADD CONSTRAINT chk_opportunities_vacancy_level CHECK (vacancy_level IS NULL OR vacancy_level IN ('intern', 'junior', 'middle', 'senior')),
    ADD CONSTRAINT chk_opportunities_employment_type CHECK (employment_type IS NULL OR employment_type IN ('full_time', 'part_time', 'project', 'temporary')),
    ADD CONSTRAINT chk_opportunities_salary_range CHECK (
        salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max
    ),
    ADD CONSTRAINT chk_opportunities_event_dates CHECK (
        event_start_at IS NULL OR event_end_at IS NULL OR event_start_at <= event_end_at
    );

DROP TABLE IF EXISTS event_opportunities;
DROP TABLE IF EXISTS mentorship_opportunities;
DROP TABLE IF EXISTS vacancy_opportunities;
DROP TABLE IF EXISTS internship_opportunities;

-- +goose Up
CREATE TABLE resume_work_experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    company_id UUID NULL REFERENCES companies(id) ON DELETE SET NULL,
    company_name VARCHAR(255) NULL,
    position_title VARCHAR(255) NOT NULL,
    started_at DATE NOT NULL,
    finished_at DATE NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_resume_work_experiences_company CHECK (
        company_id IS NOT NULL OR NULLIF(company_name, '') IS NOT NULL
    ),
    CONSTRAINT chk_resume_work_experiences_dates CHECK (
        finished_at IS NULL OR started_at <= finished_at
    )
);

CREATE INDEX idx_resume_work_experiences_resume_id ON resume_work_experiences(resume_id);

-- +goose Down
DROP INDEX IF EXISTS idx_resume_work_experiences_resume_id;
DROP TABLE IF EXISTS resume_work_experiences;

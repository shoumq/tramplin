package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	. "tramplin/internal/models"
	"tramplin/internal/repository"
)

func (r *Repository) GetStudentProfile(userID string) (*StudentProfile, error) {
	var profile StudentProfile
	var middleName sql.NullString
	var faculty sql.NullString
	var specialization sql.NullString
	var about sql.NullString
	var telegram sql.NullString
	var githubURL sql.NullString
	var linkedinURL sql.NullString
	var websiteURL sql.NullString
	var avatarURL sql.NullString
	err := r.db.QueryRowContext(context.Background(), `
SELECT sp.user_id, COALESCE(u.avatar_url, ''), sp.last_name, sp.first_name, sp.middle_name, sp.university_name, sp.faculty, sp.specialization, COALESCE(sp.study_year, 0), COALESCE(sp.graduation_year, 0), sp.about, sp.profile_visibility, sp.show_resume, sp.show_applications, sp.show_career_interests, sp.telegram, sp.github_url, sp.linkedin_url, sp.website_url, COALESCE(sp.city_id, 0), sp.created_at, sp.updated_at
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE sp.user_id = $1
`, userID).Scan(&profile.UserID, &avatarURL, &profile.LastName, &profile.FirstName, &middleName, &profile.UniversityName, &faculty, &specialization, &profile.StudyYear, &profile.GraduationYear, &about, &profile.ProfileVisibility, &profile.ShowResume, &profile.ShowApplications, &profile.ShowCareerInterests, &telegram, &githubURL, &linkedinURL, &websiteURL, &profile.CityID, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("student profile not found")
		}
		return nil, fmt.Errorf("get student profile %s: %w", userID, err)
	}
	if avatarURL.Valid {
		profile.AvatarURL = avatarURL.String
	}
	if middleName.Valid {
		profile.MiddleName = middleName.String
	}
	if faculty.Valid {
		profile.Faculty = faculty.String
	}
	if specialization.Valid {
		profile.Specialization = specialization.String
	}
	if about.Valid {
		profile.About = about.String
	}
	if telegram.Valid {
		profile.Telegram = telegram.String
	}
	if githubURL.Valid {
		profile.GithubURL = githubURL.String
	}
	if linkedinURL.Valid {
		profile.LinkedinURL = linkedinURL.String
	}
	if websiteURL.Valid {
		profile.WebsiteURL = websiteURL.String
	}
	return &profile, nil
}

func (r *Repository) GetPublicStudentProfile(userID, viewerUserID string) (*PublicStudentProfile, error) {
	var profile PublicStudentProfile
	var middleName sql.NullString
	var faculty sql.NullString
	var specialization sql.NullString
	var about sql.NullString
	var telegram sql.NullString
	var githubURL sql.NullString
	var linkedinURL sql.NullString
	var websiteURL sql.NullString
	var avatarURL sql.NullString
	var contactRelation string
	var isContact bool
	var incomingPending bool
	var outgoingPending bool
	err := r.db.QueryRowContext(context.Background(), `
SELECT
	sp.user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	sp.last_name,
	sp.first_name,
	sp.middle_name,
	sp.university_name,
	sp.faculty,
	sp.specialization,
	COALESCE(sp.study_year, 0),
	COALESCE(sp.graduation_year, 0),
	sp.about,
	sp.profile_visibility,
	sp.show_resume,
	sp.show_applications,
	sp.show_career_interests,
	sp.telegram,
	sp.github_url,
	sp.linkedin_url,
	sp.website_url,
	(SELECT COUNT(1) FROM resumes r WHERE r.student_user_id = sp.user_id),
	(SELECT COUNT(1) FROM portfolio_projects p WHERE p.student_user_id = sp.user_id),
		EXISTS (SELECT 1 FROM contacts c WHERE c.user_id = NULLIF($2, '')::uuid AND c.contact_user_id = sp.user_id),
		EXISTS (SELECT 1 FROM contact_requests cr WHERE cr.sender_user_id = sp.user_id AND cr.receiver_user_id = NULLIF($2, '')::uuid AND cr.status = 'pending'),
		EXISTS (SELECT 1 FROM contact_requests cr WHERE cr.sender_user_id = NULLIF($2, '')::uuid AND cr.receiver_user_id = sp.user_id AND cr.status = 'pending'),
	sp.created_at,
	sp.updated_at
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE sp.user_id = $1
`, userID, viewerUserID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&avatarURL,
		&profile.LastName,
		&profile.FirstName,
		&middleName,
		&profile.UniversityName,
		&faculty,
		&specialization,
		&profile.StudyYear,
		&profile.GraduationYear,
		&about,
		&profile.ProfileVisibility,
		&profile.ShowResume,
		&profile.ShowApplications,
		&profile.ShowCareerInterests,
		&telegram,
		&githubURL,
		&linkedinURL,
		&websiteURL,
		&profile.ResumeCount,
		&profile.PortfolioCount,
		&isContact,
		&incomingPending,
		&outgoingPending,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("student profile not found")
		}
		return nil, fmt.Errorf("get public student profile %s: %w", userID, err)
	}
	switch profile.ProfileVisibility {
	case "public_inside_platform":
	case "authorized_only":
		if strings.TrimSpace(viewerUserID) == "" {
			return nil, errors.New("student profile is available only to authorized users")
		}
	case "contacts_only":
		if strings.TrimSpace(viewerUserID) == "" || (!isContact && strings.TrimSpace(viewerUserID) != profile.UserID) {
			return nil, errors.New("student profile is available only to contacts")
		}
	case "private":
		return nil, errors.New("student profile not found")
	default:
		return nil, errors.New("student profile not found")
	}
	if avatarURL.Valid {
		profile.AvatarURL = avatarURL.String
	}
	if middleName.Valid {
		profile.MiddleName = middleName.String
	}
	if faculty.Valid {
		profile.Faculty = faculty.String
	}
	if specialization.Valid {
		profile.Specialization = specialization.String
	}
	if about.Valid {
		profile.About = about.String
	}
	if telegram.Valid {
		profile.Telegram = telegram.String
	}
	if githubURL.Valid {
		profile.GithubURL = githubURL.String
	}
	if linkedinURL.Valid {
		profile.LinkedinURL = linkedinURL.String
	}
	if websiteURL.Valid {
		profile.WebsiteURL = websiteURL.String
	}
	profile.HasResume = profile.ResumeCount > 0
	profile.HasPortfolio = profile.PortfolioCount > 0
	contactRelation = "none"
	if strings.TrimSpace(viewerUserID) != "" {
		switch {
		case strings.TrimSpace(viewerUserID) == profile.UserID:
			contactRelation = "none"
		case isContact:
			contactRelation = "contact"
		case incomingPending:
			contactRelation = "incoming_pending"
		case outgoingPending:
			contactRelation = "outgoing_pending"
		}
	}
	profile.ContactRelation = contactRelation
	return &profile, nil
}

func (r *Repository) ListPublicStudentProfiles(filter repository.StudentFilter) ([]PublicStudentProfile, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	sp.user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	sp.last_name,
	sp.first_name,
	COALESCE(sp.middle_name, ''),
	sp.university_name,
	COALESCE(sp.faculty, ''),
	COALESCE(sp.specialization, ''),
	COALESCE(sp.study_year, 0),
	COALESCE(sp.graduation_year, 0),
	COALESCE(sp.about, ''),
	sp.profile_visibility,
	sp.show_resume,
	sp.show_applications,
	sp.show_career_interests,
	COALESCE(sp.telegram, ''),
	COALESCE(sp.github_url, ''),
	COALESCE(sp.linkedin_url, ''),
	COALESCE(sp.website_url, ''),
	(SELECT COUNT(1) FROM resumes r WHERE r.student_user_id = sp.user_id),
	(SELECT COUNT(1) FROM portfolio_projects p WHERE p.student_user_id = sp.user_id),
		EXISTS (SELECT 1 FROM contacts c WHERE c.user_id = NULLIF($1, '')::uuid AND c.contact_user_id = sp.user_id),
		EXISTS (SELECT 1 FROM contact_requests cr WHERE cr.sender_user_id = sp.user_id AND cr.receiver_user_id = NULLIF($1, '')::uuid AND cr.status = 'pending'),
		EXISTS (SELECT 1 FROM contact_requests cr WHERE cr.sender_user_id = NULLIF($1, '')::uuid AND cr.receiver_user_id = sp.user_id AND cr.status = 'pending'),
	sp.created_at,
	sp.updated_at
FROM student_profiles sp
JOIN users u ON u.id = sp.user_id
WHERE
	(
		sp.profile_visibility = 'public_inside_platform'
		OR (sp.profile_visibility = 'authorized_only' AND NULLIF($1, '') IS NOT NULL)
		OR (
			sp.profile_visibility = 'contacts_only'
			AND (
					sp.user_id = NULLIF($1, '')::uuid
					OR EXISTS (
						SELECT 1
						FROM contacts c
						WHERE c.user_id = NULLIF($1, '')::uuid AND c.contact_user_id = sp.user_id
					)
			)
		)
	)
	AND (
		NULLIF($2, '') IS NULL
		OR LOWER(u.display_name) LIKE '%' || LOWER($2) || '%'
		OR LOWER(sp.first_name) LIKE '%' || LOWER($2) || '%'
		OR LOWER(sp.last_name) LIKE '%' || LOWER($2) || '%'
		OR LOWER(COALESCE(sp.middle_name, '')) LIKE '%' || LOWER($2) || '%'
		OR LOWER(sp.university_name) LIKE '%' || LOWER($2) || '%'
		OR LOWER(COALESCE(sp.faculty, '')) LIKE '%' || LOWER($2) || '%'
		OR LOWER(COALESCE(sp.specialization, '')) LIKE '%' || LOWER($2) || '%'
	)
	AND (NULLIF($3, '') IS NULL OR LOWER(sp.university_name) = LOWER($3))
	AND (NULLIF($4, '') IS NULL OR LOWER(COALESCE(sp.faculty, '')) = LOWER($4))
	AND (NULLIF($5, '') IS NULL OR LOWER(COALESCE(sp.specialization, '')) = LOWER($5))
	AND (NULLIF($6, 0) IS NULL OR COALESCE(sp.study_year, 0) = $6)
ORDER BY sp.updated_at DESC, sp.created_at DESC
`, filter.ViewerUserID, filter.Search, filter.UniversityName, filter.Faculty, filter.Specialization, filter.StudyYear)
	if err != nil {
		return nil, fmt.Errorf("list public student profiles: %w", err)
	}
	defer rows.Close()

	var result []PublicStudentProfile
	for rows.Next() {
		var item PublicStudentProfile
		var middleName string
		var faculty string
		var specialization string
		var about string
		var telegram string
		var githubURL string
		var linkedinURL string
		var websiteURL string
		var avatarURL string
		var isContact bool
		var incomingPending bool
		var outgoingPending bool
		if err := rows.Scan(
			&item.UserID,
			&item.DisplayName,
			&avatarURL,
			&item.LastName,
			&item.FirstName,
			&middleName,
			&item.UniversityName,
			&faculty,
			&specialization,
			&item.StudyYear,
			&item.GraduationYear,
			&about,
			&item.ProfileVisibility,
			&item.ShowResume,
			&item.ShowApplications,
			&item.ShowCareerInterests,
			&telegram,
			&githubURL,
			&linkedinURL,
			&websiteURL,
			&item.ResumeCount,
			&item.PortfolioCount,
			&isContact,
			&incomingPending,
			&outgoingPending,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan public student profiles: %w", err)
		}
		item.AvatarURL = avatarURL
		item.MiddleName = middleName
		item.Faculty = faculty
		item.Specialization = specialization
		item.About = about
		item.Telegram = telegram
		item.GithubURL = githubURL
		item.LinkedinURL = linkedinURL
		item.WebsiteURL = websiteURL
		item.HasResume = item.ResumeCount > 0
		item.HasPortfolio = item.PortfolioCount > 0
		item.ContactRelation = "none"
		if strings.TrimSpace(filter.ViewerUserID) != "" {
			switch {
			case strings.TrimSpace(filter.ViewerUserID) == item.UserID:
				item.ContactRelation = "none"
			case isContact:
				item.ContactRelation = "contact"
			case incomingPending:
				item.ContactRelation = "incoming_pending"
			case outgoingPending:
				item.ContactRelation = "outgoing_pending"
			}
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) UpsertStudentProfile(profile StudentProfile, actorID string) (*StudentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.getUserByID(context.Background(), profile.UserID); err != nil {
		return nil, err
	}
	now := time.Now()
	existing, err := r.GetStudentProfile(profile.UserID)
	if err == nil {
		profile.CreatedAt = existing.CreatedAt
		profile.AvatarURL = existing.AvatarURL
	} else if err.Error() == "student profile not found" {
		profile.CreatedAt = now
	} else {
		return nil, err
	}
	profile.UpdatedAt = now
	if profile.ProfileVisibility == "" {
		profile.ProfileVisibility = "authorized_only"
	}
	if err := r.upsertStudentProfileTx(context.Background(), nil, &profile); err != nil {
		return nil, err
	}
	r.addAudit("update", "student_profiles", profile.UserID, actorID, "student profile upsert")
	return r.GetStudentProfile(profile.UserID)
}

func (r *Repository) ListResumes(studentUserID string) ([]Resume, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, student_user_id, title, COALESCE(summary, ''), COALESCE(experience_text, ''), COALESCE(education_text, ''), is_primary, created_at, updated_at
FROM resumes
WHERE student_user_id = $1
ORDER BY created_at DESC
`, studentUserID)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}
	defer rows.Close()
	result := []Resume{}
	for rows.Next() {
		var item Resume
		if err := rows.Scan(&item.ID, &item.StudentUserID, &item.Title, &item.Summary, &item.ExperienceText, &item.EducationText, &item.IsPrimary, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan resumes: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateResume(resume Resume) (*Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM student_profiles WHERE user_id = $1)`, resume.StudentUserID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check student profile: %w", err)
	}
	if !exists {
		return nil, errors.New("student profile not found")
	}
	now := time.Now()
	resume.ID = uuid.NewString()
	resume.CreatedAt = now
	resume.UpdatedAt = now
	var resumeCount int
	if err := r.db.QueryRowContext(context.Background(), `SELECT COUNT(1) FROM resumes WHERE student_user_id = $1`, resume.StudentUserID).Scan(&resumeCount); err != nil {
		return nil, fmt.Errorf("count resumes: %w", err)
	}
	if resumeCount == 0 {
		resume.IsPrimary = true
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO resumes (id, student_user_id, title, summary, experience_text, education_text, is_primary, created_at, updated_at)
VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9)
`, resume.ID, resume.StudentUserID, resume.Title, resume.Summary, resume.ExperienceText, resume.EducationText, resume.IsPrimary, resume.CreatedAt, resume.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create resume: %w", err)
	}
	r.addAudit("create", "resumes", resume.ID, resume.StudentUserID, resume.Title)
	return &resume, nil
}

func (r *Repository) SetPrimaryResume(studentUserID, resumeID string) (*Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var resume Resume
	err := r.db.QueryRowContext(context.Background(), `
SELECT id, student_user_id, title, COALESCE(summary, ''), COALESCE(experience_text, ''), COALESCE(education_text, ''), is_primary, created_at, updated_at
FROM resumes
WHERE id = $1 AND student_user_id = $2
`, resumeID, studentUserID).Scan(&resume.ID, &resume.StudentUserID, &resume.Title, &resume.Summary, &resume.ExperienceText, &resume.EducationText, &resume.IsPrimary, &resume.CreatedAt, &resume.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("resume not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get resume: %w", err)
	}
	now := time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE resumes SET is_primary = FALSE, updated_at = $2 WHERE student_user_id = $1`, studentUserID, now); err != nil {
		return nil, fmt.Errorf("reset primary resumes: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `UPDATE resumes SET is_primary = TRUE, updated_at = $3 WHERE id = $1 AND student_user_id = $2`, resumeID, studentUserID, now); err != nil {
		return nil, fmt.Errorf("set primary resume: %w", err)
	}
	resume.IsPrimary = true
	resume.UpdatedAt = now
	return &resume, nil
}

func (r *Repository) ListPortfolioProjects(studentUserID string) ([]PortfolioProject, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, student_user_id, title, COALESCE(description, ''), COALESCE(project_url, ''), COALESCE(repository_url, ''), COALESCE(demo_url, ''), started_at, finished_at, created_at, updated_at
FROM portfolio_projects
WHERE student_user_id = $1
ORDER BY created_at DESC
`, studentUserID)
	if err != nil {
		return nil, fmt.Errorf("list portfolio projects: %w", err)
	}
	defer rows.Close()
	result := []PortfolioProject{}
	for rows.Next() {
		var item PortfolioProject
		var startedAt sql.NullString
		var finishedAt sql.NullString
		if err := rows.Scan(&item.ID, &item.StudentUserID, &item.Title, &item.Description, &item.ProjectURL, &item.RepositoryURL, &item.DemoURL, &startedAt, &finishedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan portfolio projects: %w", err)
		}
		if startedAt.Valid {
			item.StartedAt = startedAt.String
		}
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.String
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreatePortfolioProject(project PortfolioProject) (*PortfolioProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project.ID = uuid.NewString()
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO portfolio_projects (id, student_user_id, title, description, project_url, repository_url, demo_url, started_at, finished_at, created_at, updated_at)
VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10, $11)
`, project.ID, project.StudentUserID, project.Title, project.Description, project.ProjectURL, project.RepositoryURL, project.DemoURL, nullableDate(project.StartedAt), nullableDate(project.FinishedAt), project.CreatedAt, project.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create portfolio project: %w", err)
	}
	r.addAudit("create", "portfolio_projects", project.ID, project.StudentUserID, project.Title)
	return &project, nil
}

func (r *Repository) ListStudentApplications(studentUserID string) ([]Application, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT a.id, a.opportunity_id, a.student_user_id, COALESCE(u.avatar_url, ''), COALESCE(a.resume_id::text, ''), COALESCE(a.cover_letter, ''), a.status, COALESCE(a.status_changed_by_user_id::text, ''), a.status_changed_at, a.created_at, a.updated_at
FROM applications a
JOIN users u ON u.id = a.student_user_id
WHERE a.student_user_id = $1
ORDER BY a.created_at DESC
`, studentUserID)
	if err != nil {
		return nil, fmt.Errorf("list student applications: %w", err)
	}
	defer rows.Close()
	result := []Application{}
	for rows.Next() {
		var item Application
		var statusChangedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OpportunityID, &item.StudentUserID, &item.StudentAvatarURL, &item.ResumeID, &item.CoverLetter, &item.Status, &item.StatusChangedByUserID, &statusChangedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan student applications: %w", err)
		}
		if statusChangedAt.Valid {
			item.StatusChangedAt = statusChangedAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListFavoriteOpportunities(userID string) ([]PublicOpportunity, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT opportunity_id
FROM favorite_opportunities
WHERE user_id = $1
ORDER BY created_at DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite opportunities ids: %w", err)
	}
	defer rows.Close()

	var result []PublicOpportunity
	for rows.Next() {
		var opportunityID string
		if err := rows.Scan(&opportunityID); err != nil {
			return nil, fmt.Errorf("scan favorite opportunities ids: %w", err)
		}
		opportunity, err := r.getPublicOpportunityByID(context.Background(), opportunityID)
		if err != nil {
			continue
		}
		result = append(result, *opportunity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorite opportunities ids: %w", err)
	}
	return result, nil
}

func (r *Repository) AddFavoriteOpportunity(userID, opportunityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM opportunities WHERE id = $1)`, opportunityID).Scan(&exists); err != nil {
		return fmt.Errorf("check opportunity: %w", err)
	}
	if !exists {
		return errors.New("opportunity not found")
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO favorite_opportunities (user_id, opportunity_id)
VALUES ($1, $2)
ON CONFLICT (user_id, opportunity_id) DO NOTHING
`, userID, opportunityID); err != nil {
		return fmt.Errorf("add favorite opportunity: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE opportunities
SET favorites_count = (SELECT COUNT(1) FROM favorite_opportunities WHERE opportunity_id = $1)
WHERE id = $1
`, opportunityID); err != nil {
		return fmt.Errorf("refresh opportunity favorites count: %w", err)
	}
	return nil
}

func (r *Repository) RemoveFavoriteOpportunity(userID, opportunityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if _, err := r.db.ExecContext(context.Background(), `DELETE FROM favorite_opportunities WHERE user_id = $1 AND opportunity_id = $2`, userID, opportunityID); err != nil {
		return fmt.Errorf("remove favorite opportunity: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE opportunities
SET favorites_count = (SELECT COUNT(1) FROM favorite_opportunities WHERE opportunity_id = $1)
WHERE id = $1
`, opportunityID); err != nil {
		return fmt.Errorf("refresh opportunity favorites count: %w", err)
	}
	return nil
}

func (r *Repository) ListFavoriteCompanies(userID string) ([]Company, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT c.id, c.legal_name, COALESCE(c.brand_name, ''), COALESCE(cu.avatar_url, c.avatar_url, ''), COALESCE(cu.avatar_object, c.avatar_object, ''), COALESCE(mf.storage_path, ''), COALESCE(c.description, ''), COALESCE(c.industry, ''), COALESCE(c.website_url, ''), COALESCE(c.email_domain, ''), COALESCE(c.inn, ''), COALESCE(c.ogrn, ''), COALESCE(c.company_size, ''), COALESCE(c.founded_year, 0), COALESCE(c.hq_city_id, 0), c.status, c.created_at, c.updated_at
FROM favorite_companies fc
JOIN companies c ON c.id = fc.company_id
LEFT JOIN LATERAL (
	SELECT u.avatar_url, u.avatar_object
	FROM employer_profiles ep
	JOIN users u ON u.id = ep.user_id
	WHERE ep.company_id = c.id
	  AND (COALESCE(u.avatar_url, '') <> '' OR COALESCE(u.avatar_object, '') <> '')
	ORDER BY ep.is_company_owner DESC, u.updated_at DESC, ep.created_at ASC
	LIMIT 1
) cu ON TRUE
LEFT JOIN media_files mf ON mf.id = c.logo_media_id
WHERE fc.user_id = $1
ORDER BY fc.created_at DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite companies: %w", err)
	}
	defer rows.Close()

	var result []Company
	for rows.Next() {
		var item Company
		var legacyAvatarPath string
		if err := rows.Scan(&item.ID, &item.LegalName, &item.BrandName, &item.AvatarURL, &item.AvatarObject, &legacyAvatarPath, &item.Description, &item.Industry, &item.WebsiteURL, &item.EmailDomain, &item.INN, &item.OGRN, &item.CompanySize, &item.FoundedYear, &item.HQCityID, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan favorite companies: %w", err)
		}
		if item.AvatarURL == "" {
			item.AvatarURL = r.legacyMediaURL(legacyAvatarPath)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorite companies: %w", err)
	}
	return result, nil
}

func (r *Repository) AddFavoriteCompany(userID, companyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM companies WHERE id = $1)`, companyID).Scan(&exists); err != nil {
		return fmt.Errorf("check company: %w", err)
	}
	if !exists {
		return errors.New("company not found")
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO favorite_companies (user_id, company_id)
VALUES ($1, $2)
ON CONFLICT (user_id, company_id) DO NOTHING
`, userID, companyID); err != nil {
		return fmt.Errorf("add favorite company: %w", err)
	}
	return nil
}

func (r *Repository) RemoveFavoriteCompany(userID, companyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if _, err := r.db.ExecContext(context.Background(), `DELETE FROM favorite_companies WHERE user_id = $1 AND company_id = $2`, userID, companyID); err != nil {
		return fmt.Errorf("remove favorite company: %w", err)
	}
	return nil
}

func (r *Repository) ListContacts(userID string) ([]User, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT u.id, u.email, u.password_hash, u.display_name, COALESCE(u.avatar_url, ''), COALESCE(u.avatar_object, ''), u.email_verified, u.status, u.last_login_at, u.created_at, u.updated_at
FROM contacts c
JOIN users u ON u.id = c.contact_user_id
WHERE c.user_id = $1
ORDER BY c.created_at DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var item User
		var lastLoginAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Email, &item.PasswordHash, &item.DisplayName, &item.AvatarURL, &item.AvatarObject, &item.EmailVerified, &item.Status, &lastLoginAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contacts: %w", err)
		}
		if lastLoginAt.Valid {
			item.LastLoginAt = lastLoginAt.Time
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contacts: %w", err)
	}
	return result, nil
}

func (r *Repository) ListContactRequests(userID string) ([]ContactRequest, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT cr.id, cr.sender_user_id, cr.receiver_user_id, s.display_name, r.display_name, COALESCE(s.avatar_url, ''), COALESCE(r.avatar_url, ''), COALESCE(cr.message, ''), cr.status, cr.created_at, cr.updated_at
FROM contact_requests cr
JOIN users s ON s.id = cr.sender_user_id
JOIN users r ON r.id = cr.receiver_user_id
WHERE cr.sender_user_id = $1 OR cr.receiver_user_id = $1
ORDER BY cr.created_at DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list contact requests: %w", err)
	}
	defer rows.Close()

	var result []ContactRequest
	for rows.Next() {
		var item ContactRequest
		if err := rows.Scan(&item.ID, &item.SenderUserID, &item.ReceiverUserID, &item.SenderName, &item.ReceiverName, &item.SenderAvatarURL, &item.ReceiverAvatarURL, &item.Message, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contact requests: %w", err)
		}
		if item.SenderUserID == userID {
			item.Direction = "outgoing"
		} else if item.ReceiverUserID == userID {
			item.Direction = "incoming"
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate contact requests: %w", err)
	}
	return result, nil
}

func (r *Repository) CreateContactRequest(senderUserID, receiverUserID, message string) (*ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	senderUserID = strings.TrimSpace(senderUserID)
	receiverUserID = strings.TrimSpace(receiverUserID)
	message = strings.TrimSpace(message)
	if receiverUserID == "" {
		return nil, errors.New("receiver_user_id is required")
	}
	if _, err := uuid.Parse(receiverUserID); err != nil {
		return nil, errors.New("receiver_user_id must be a valid UUID")
	}
	if senderUserID == receiverUserID {
		return nil, errors.New("cannot create contact request to yourself")
	}
	item := &ContactRequest{ID: uuid.NewString(), SenderUserID: senderUserID, ReceiverUserID: receiverUserID, Message: message, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_, err := r.db.ExecContext(context.Background(), `
INSERT INTO contact_requests (id, sender_user_id, receiver_user_id, message, status, created_at, updated_at)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
`, item.ID, item.SenderUserID, item.ReceiverUserID, item.Message, item.Status, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create contact request: %w", err)
	}
	r.notify(receiverUserID, "contact_request_received", "Новый запрос в контакты", "Вам пришёл новый запрос на профессиональный контакт.", "contact_request", item.ID)
	items, err := r.ListContactRequests(receiverUserID)
	if err == nil {
		for _, request := range items {
			if request.ID == item.ID {
				cp := request
				return &cp, nil
			}
		}
	}
	return item, nil
}

func (r *Repository) UpdateContactRequestStatus(requestID, userID, status string) (*ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	var item ContactRequest
	err := r.db.QueryRowContext(context.Background(), `
SELECT id, sender_user_id, receiver_user_id, COALESCE(message, ''), status, created_at, updated_at
FROM contact_requests
WHERE id = $1
`, requestID).Scan(&item.ID, &item.SenderUserID, &item.ReceiverUserID, &item.Message, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("contact request not found")
		}
		return nil, fmt.Errorf("get contact request: %w", err)
	}
	if item.ReceiverUserID != userID && item.SenderUserID != userID {
		return nil, errors.New("forbidden")
	}
	item.Status = status
	item.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE contact_requests SET status = $2, updated_at = $3 WHERE id = $1`, item.ID, item.Status, item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update contact request status: %w", err)
	}
	if status == "accepted" {
		r.addContact(item.SenderUserID, item.ReceiverUserID)
		r.addContact(item.ReceiverUserID, item.SenderUserID)
		r.notify(item.SenderUserID, "contact_request_accepted", "Запрос в контакты принят", "Ваш запрос на контакт был принят.", "contact_request", item.ID)
	}
	items, err := r.ListContactRequests(userID)
	if err == nil {
		for _, request := range items {
			if request.ID == item.ID {
				cp := request
				return &cp, nil
			}
		}
	}
	return &item, nil
}

func (r *Repository) CreateRecommendation(rec Recommendation) (*Recommendation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	rec.ID = uuid.NewString()
	rec.CreatedAt = time.Now()
	_, err := r.db.ExecContext(context.Background(), `
INSERT INTO recommendations (id, from_user_id, to_user_id, opportunity_id, message, created_at)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
`, rec.ID, rec.FromUserID, rec.ToUserID, rec.OpportunityID, rec.Message, rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create recommendation: %w", err)
	}
	title, body := r.recommendationNotificationText(rec.OpportunityID)
	r.notify(rec.ToUserID, "recommendation_received", title, body, "opportunity", rec.OpportunityID)
	return &rec, nil
}

func (r *Repository) ListNotifications(userID string) ([]Notification, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	n.id,
	n.user_id,
	n.type,
	n.title,
	n.body,
	n.is_read,
	COALESCE(n.related_entity_type, ''),
	COALESCE(n.related_entity_id::text, ''),
	COALESCE(o_direct.title, o_from_application.title, ''),
	COALESCE(c_direct.legal_name, c_from_application.legal_name, ''),
	COALESCE(o_direct.contacts_info, o_from_application.contacts_info, ''),
	COALESCE(o_direct.created_by_user_id::text, o_from_application.created_by_user_id::text, ''),
	n.created_at
FROM notifications n
LEFT JOIN opportunities o_direct
	ON n.related_entity_type = 'opportunity'
	AND o_direct.id = n.related_entity_id
LEFT JOIN companies c_direct ON c_direct.id = o_direct.company_id
LEFT JOIN applications a
	ON n.related_entity_type = 'application'
	AND a.id = n.related_entity_id
LEFT JOIN opportunities o_from_application ON o_from_application.id = a.opportunity_id
LEFT JOIN companies c_from_application ON c_from_application.id = o_from_application.company_id
WHERE n.user_id = $1
ORDER BY created_at DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var result []Notification
	for rows.Next() {
		var item Notification
		if err := rows.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Body, &item.IsRead, &item.RelatedEntityType, &item.RelatedEntityID, &item.OpportunityTitle, &item.CompanyLegalName, &item.EmployerContacts, &item.EmployerUserID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notifications: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return result, nil
}

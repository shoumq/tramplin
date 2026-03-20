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

func (r *Repository) getPublicOpportunityByID(ctx context.Context, id string) (*PublicOpportunity, error) {
	var item PublicOpportunity
	var companyName sql.NullString
	var companyAvatarURL sql.NullString
	var legacyCompanyAvatarPath sql.NullString
	var location sql.NullString
	row := r.db.QueryRowContext(ctx, `
SELECT
	o.id,
	o.company_id,
	o.created_by_user_id,
	o.title,
	o.short_description,
	o.full_description,
	o.opportunity_type,
	o.work_format,
	COALESCE(o.location_id::text, ''),
	o.published_at,
	o.expires_at,
	o.status,
	COALESCE(o.contacts_info, ''),
	COALESCE(o.external_url, ''),
	o.views_count,
	o.favorites_count,
	o.applications_count,
	o.created_at,
	o.updated_at,
	COALESCE(c.brand_name, c.legal_name),
	COALESCE(cu.avatar_url, c.avatar_url, ''),
	COALESCE(mf.storage_path, ''),
	COALESCE(l.display_text, '')
FROM opportunities o
JOIN companies c ON c.id = o.company_id
LEFT JOIN LATERAL (
	SELECT u.avatar_url
	FROM employer_profiles ep
	JOIN users u ON u.id = ep.user_id
	WHERE ep.company_id = c.id
	  AND COALESCE(u.avatar_url, '') <> ''
	ORDER BY ep.is_company_owner DESC, u.updated_at DESC, ep.created_at ASC
	LIMIT 1
) cu ON TRUE
LEFT JOIN media_files mf ON mf.id = c.logo_media_id
LEFT JOIN locations l ON l.id = o.location_id
WHERE o.id = $1
`, id)
	err := scanOpportunityBaseWithExtras(row, &item.Opportunity, &companyName, &companyAvatarURL, &legacyCompanyAvatarPath, &location)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("opportunity not found")
		}
		return nil, fmt.Errorf("get public opportunity %s: %w", id, err)
	}
	if err := r.loadOpportunityTypeDetails(ctx, &item.Opportunity); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT t.name
FROM opportunity_tags ot
JOIN tags t ON t.id = ot.tag_id
WHERE ot.opportunity_id = $1
ORDER BY t.name
`, id)
	if err != nil {
		return nil, fmt.Errorf("list opportunity tags %s: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var tagName string
		if err := rows.Scan(&tagName); err != nil {
			return nil, fmt.Errorf("scan opportunity tags %s: %w", id, err)
		}
		item.Tags = append(item.Tags, tagName)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate opportunity tags %s: %w", id, err)
	}

	if companyName.Valid {
		item.CompanyName = companyName.String
	}
	if companyAvatarURL.Valid {
		item.CompanyAvatarURL = companyAvatarURL.String
	}
	if item.CompanyAvatarURL == "" && legacyCompanyAvatarPath.Valid {
		item.CompanyAvatarURL = r.legacyMediaURL(legacyCompanyAvatarPath.String)
	}
	if location.Valid {
		item.Location = location.String
	}
	return &item, nil
}

func (r *Repository) ListOpportunities(filter repository.OpportunityFilter) ([]PublicOpportunity, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT o.id
FROM opportunities o
WHERE o.status IN ('published', 'scheduled')
ORDER BY o.created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list opportunities ids: %w", err)
	}
	defer rows.Close()

	var result []PublicOpportunity
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan opportunities ids: %w", err)
		}
		public, err := r.getPublicOpportunityByID(context.Background(), id)
		if err != nil {
			continue
		}
		if matchesFilter(*public, filter) {
			result = append(result, *public)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate opportunities ids: %w", err)
	}
	return result, nil
}

func (r *Repository) ListOpportunityMarkers(filter repository.OpportunityFilter) ([]OpportunityMarker, error) {
	targets, _ := r.ListOpportunities(filter)
	markers := make([]OpportunityMarker, 0, len(targets))
	for _, item := range targets {
		opp := item.Opportunity
		var latitude float64
		var longitude float64
		err := r.db.QueryRowContext(context.Background(), `
SELECT COALESCE(latitude, 0), COALESCE(longitude, 0)
FROM locations
WHERE id = $1
`, opp.LocationID).Scan(&latitude, &longitude)
		if err != nil || (latitude == 0 && longitude == 0) {
			continue
		}
		markers = append(markers, OpportunityMarker{
			ID:              opp.ID,
			Title:           opp.Title,
			CompanyName:     item.CompanyName,
			Latitude:        latitude,
			Longitude:       longitude,
			WorkFormat:      opp.WorkFormat,
			OpportunityType: opp.OpportunityType,
			SalaryLabel:     salaryLabel(opp),
		})
	}
	return markers, nil
}

func (r *Repository) GetOpportunity(id string) (*PublicOpportunity, error) {
	if _, err := r.db.ExecContext(context.Background(), `UPDATE opportunities SET views_count = views_count + 1 WHERE id = $1`, id); err != nil {
		return nil, fmt.Errorf("increment opportunity views: %w", err)
	}
	return r.getPublicOpportunityByID(context.Background(), id)
}

func (r *Repository) CreateApplication(application Application) (*Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM applications WHERE opportunity_id = $1 AND student_user_id = $2)`, application.OpportunityID, application.StudentUserID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check application duplicate: %w", err)
	}
	if exists {
		return nil, errors.New("application already exists")
	}
	var opportunity Opportunity
	row := r.db.QueryRowContext(context.Background(), `
SELECT id, company_id, created_by_user_id, title, short_description, full_description, opportunity_type, work_format, COALESCE(location_id::text, ''), published_at, expires_at, status, COALESCE(contacts_info, ''), COALESCE(external_url, ''), views_count, favorites_count, applications_count, created_at, updated_at
FROM opportunities
WHERE id = $1
`, application.OpportunityID)
	err := scanOpportunityBase(row, &opportunity)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("opportunity not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get opportunity for application: %w", err)
	}
	if err := r.loadOpportunityTypeDetails(context.Background(), &opportunity); err != nil {
		return nil, err
	}
	application.ID = uuid.NewString()
	application.Status = "submitted"
	application.CreatedAt = time.Now()
	application.UpdatedAt = application.CreatedAt
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO applications (id, opportunity_id, student_user_id, resume_id, cover_letter, status, created_at, updated_at)
VALUES ($1, $2, $3, NULLIF($4, '')::uuid, NULLIF($5, ''), $6, $7, $8)
`, application.ID, application.OpportunityID, application.StudentUserID, application.ResumeID, application.CoverLetter, application.Status, application.CreatedAt, application.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE opportunities
SET applications_count = (SELECT COUNT(1) FROM applications WHERE opportunity_id = $1), updated_at = $2
WHERE id = $1
`, application.OpportunityID, time.Now()); err != nil {
		return nil, fmt.Errorf("refresh opportunity applications count: %w", err)
	}
	user, err := r.getUserByID(context.Background(), application.StudentUserID)
	if err == nil {
		application.StudentAvatarURL = user.AvatarURL
	}
	r.notify(opportunity.CreatedByUserID, "application_submitted", "Новый отклик", "Студент откликнулся на вашу возможность.", "application", application.ID)
	return &application, nil
}

func (r *Repository) ListCompanies() ([]Company, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	c.id,
	c.legal_name,
	COALESCE(c.brand_name, ''),
	COALESCE(cu.avatar_url, c.avatar_url, ''),
	COALESCE(cu.avatar_object, c.avatar_object, ''),
	COALESCE(mf.storage_path, ''),
	COALESCE(c.description, ''),
	COALESCE(c.industry, ''),
	COALESCE(c.website_url, ''),
	COALESCE(c.email_domain, ''),
	COALESCE(c.inn, ''),
	COALESCE(c.ogrn, ''),
	COALESCE(c.company_size, ''),
	COALESCE(c.founded_year, 0),
	COALESCE(c.hq_city_id, 0),
	c.status,
	COALESCE(cp.is_online, FALSE),
	cp.last_seen_at,
	c.created_at,
	c.updated_at
FROM companies c
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
LEFT JOIN LATERAL (
	SELECT
		BOOL_OR(up.is_online) AS is_online,
		MAX(up.last_seen_at) AS last_seen_at
	FROM employer_profiles ep
	JOIN user_presence up ON up.user_id = ep.user_id
	WHERE ep.company_id = c.id
) cp ON TRUE
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()
	result := []Company{}
	for rows.Next() {
		var item Company
		var legacyAvatarPath string
		var lastSeenAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.LegalName, &item.BrandName, &item.AvatarURL, &item.AvatarObject, &legacyAvatarPath, &item.Description, &item.Industry, &item.WebsiteURL, &item.EmailDomain, &item.INN, &item.OGRN, &item.CompanySize, &item.FoundedYear, &item.HQCityID, &item.Status, &item.IsOnline, &lastSeenAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan companies: %w", err)
		}
		if item.AvatarURL == "" {
			item.AvatarURL = r.legacyMediaURL(legacyAvatarPath)
		}
		if lastSeenAt.Valid {
			item.LastSeenAt = &lastSeenAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) GetCompany(id string) (*Company, error) {
	var item Company
	var legacyAvatarPath string
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT
	c.id,
	c.legal_name,
	COALESCE(c.brand_name, ''),
	COALESCE(cu.avatar_url, c.avatar_url, ''),
	COALESCE(cu.avatar_object, c.avatar_object, ''),
	COALESCE(mf.storage_path, ''),
	COALESCE(c.description, ''),
	COALESCE(c.industry, ''),
	COALESCE(c.website_url, ''),
	COALESCE(c.email_domain, ''),
	COALESCE(c.inn, ''),
	COALESCE(c.ogrn, ''),
	COALESCE(c.company_size, ''),
	COALESCE(c.founded_year, 0),
	COALESCE(c.hq_city_id, 0),
	c.status,
	COALESCE(cp.is_online, FALSE),
	cp.last_seen_at,
	c.created_at,
	c.updated_at
FROM companies c
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
LEFT JOIN LATERAL (
	SELECT
		BOOL_OR(up.is_online) AS is_online,
		MAX(up.last_seen_at) AS last_seen_at
	FROM employer_profiles ep
	JOIN user_presence up ON up.user_id = ep.user_id
	WHERE ep.company_id = c.id
) cp ON TRUE
WHERE c.id = $1
	`, id).Scan(&item.ID, &item.LegalName, &item.BrandName, &item.AvatarURL, &item.AvatarObject, &legacyAvatarPath, &item.Description, &item.Industry, &item.WebsiteURL, &item.EmailDomain, &item.INN, &item.OGRN, &item.CompanySize, &item.FoundedYear, &item.HQCityID, &item.Status, &item.IsOnline, &lastSeenAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("company not found")
		}
		return nil, fmt.Errorf("get company: %w", err)
	}
	if item.AvatarURL == "" {
		item.AvatarURL = r.legacyMediaURL(legacyAvatarPath)
	}
	if lastSeenAt.Valid {
		item.LastSeenAt = &lastSeenAt.Time
	}
	return &item, nil
}

func (r *Repository) ListTags() ([]Tag, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, name, tag_type, COALESCE(created_by_user_id::text, ''), is_system, is_active, created_at
FROM tags
ORDER BY name
`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()
	result := []Tag{}
	for rows.Next() {
		var item Tag
		if err := rows.Scan(&item.ID, &item.Name, &item.TagType, &item.CreatedByUserID, &item.IsSystem, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tags: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListCities() ([]City, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, country, COALESCE(region, ''), city_name, COALESCE(latitude, 0), COALESCE(longitude, 0)
FROM cities
ORDER BY country, region, city_name
`)
	if err != nil {
		return nil, fmt.Errorf("list cities: %w", err)
	}
	defer rows.Close()
	result := []City{}
	for rows.Next() {
		var item City
		if err := rows.Scan(&item.ID, &item.Country, &item.Region, &item.CityName, &item.Latitude, &item.Longitude); err != nil {
			return nil, fmt.Errorf("scan cities: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListLocations() ([]Location, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, COALESCE(city_id, 0), COALESCE(address_line, ''), COALESCE(postal_code, ''), COALESCE(latitude, 0), COALESCE(longitude, 0), location_type, display_text, created_at
FROM locations
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer rows.Close()
	result := []Location{}
	for rows.Next() {
		var item Location
		if err := rows.Scan(&item.ID, &item.CityID, &item.AddressLine, &item.PostalCode, &item.Latitude, &item.Longitude, &item.LocationType, &item.DisplayText, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan locations: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) TouchUserPresence(userID string, isOnline bool) error {
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO user_presence (user_id, is_online, last_seen_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET
	is_online = EXCLUDED.is_online,
	last_seen_at = EXCLUDED.last_seen_at,
	updated_at = EXCLUDED.updated_at
`, userID, isOnline); err != nil {
		return fmt.Errorf("touch user presence: %w", err)
	}
	return nil
}

func (r *Repository) GetUserPresence(userID string) (*Presence, error) {
	var item Presence
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT u.id, COALESCE(up.is_online, FALSE), up.last_seen_at
FROM users u
LEFT JOIN user_presence up ON up.user_id = u.id
WHERE u.id = $1
`, userID).Scan(&item.UserID, &item.IsOnline, &lastSeenAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user presence: %w", err)
	}
	if lastSeenAt.Valid {
		item.LastSeenAt = &lastSeenAt.Time
	}
	return &item, nil
}

func (r *Repository) GetCompanyPresence(companyID string) (*Presence, error) {
	var item Presence
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT
	c.id,
	COALESCE(cp.is_online, FALSE),
	cp.last_seen_at
FROM companies c
LEFT JOIN LATERAL (
	SELECT
		BOOL_OR(up.is_online) AS is_online,
		MAX(up.last_seen_at) AS last_seen_at
	FROM employer_profiles ep
	JOIN user_presence up ON up.user_id = ep.user_id
	WHERE ep.company_id = c.id
) cp ON TRUE
WHERE c.id = $1
`, companyID).Scan(&item.CompanyID, &item.IsOnline, &lastSeenAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("company not found")
		}
		return nil, fmt.Errorf("get company presence: %w", err)
	}
	if lastSeenAt.Valid {
		item.LastSeenAt = &lastSeenAt.Time
	}
	return &item, nil
}

func matchesFilter(item PublicOpportunity, filter repository.OpportunityFilter) bool {
	if filter.WorkFormat != "" && item.WorkFormat != filter.WorkFormat {
		return false
	}
	if filter.Type != "" && item.OpportunityType != filter.Type {
		return false
	}
	if filter.SalaryFrom > 0 && item.SalaryMax < filter.SalaryFrom {
		return false
	}
	if filter.Tag != "" {
		matched := false
		for _, tag := range item.Tags {
			if strings.EqualFold(tag, filter.Tag) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if filter.Search != "" {
		needle := strings.ToLower(filter.Search)
		haystack := strings.ToLower(item.Title + " " + item.ShortDescription + " " + item.CompanyName)
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}

func salaryLabel(opp Opportunity) string {
	if !opp.IsSalaryVisible {
		return ""
	}
	if opp.SalaryMin == 0 && opp.SalaryMax == 0 {
		return ""
	}
	return fmt.Sprintf("%.0f-%.0f %s", opp.SalaryMin, opp.SalaryMax, opp.SalaryCurrency)
}

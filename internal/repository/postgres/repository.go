package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	. "tramplin/internal/models"
)

type Repository struct {
	mu            sync.RWMutex
	db            *sql.DB
	publicBaseURL string
	bucket        string
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewRepository(ctx context.Context, dsn, publicBaseURL, bucket string) (*Repository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Repository{
		db:            db,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		bucket:        strings.Trim(strings.TrimSpace(bucket), "/"),
	}, nil
}

func (r *Repository) persistLocked() {}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) loadRoleIDs(ctx context.Context, tx *sql.Tx) (map[string]int16, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, code FROM roles`)
	if err != nil {
		return nil, fmt.Errorf("load role ids: %w", err)
	}
	defer rows.Close()

	roleIDs := make(map[string]int16)
	for rows.Next() {
		var id int16
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, fmt.Errorf("scan role ids: %w", err)
		}
		roleIDs[code] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role ids: %w", err)
	}
	return roleIDs, nil
}

func (r *Repository) upsertUserTx(ctx context.Context, tx *sql.Tx, user *User) error {
	execTarget := sqlExecTarget(r.db, tx)
	_, err := execTarget.ExecContext(ctx, `
INSERT INTO users (
	id, email, password_hash, display_name, email_verified, status, last_login_at, created_at, updated_at, avatar_url, avatar_object
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), NULLIF($11, ''))
ON CONFLICT (id) DO UPDATE SET
	email = EXCLUDED.email,
	password_hash = EXCLUDED.password_hash,
	display_name = EXCLUDED.display_name,
	email_verified = EXCLUDED.email_verified,
	status = EXCLUDED.status,
	last_login_at = EXCLUDED.last_login_at,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at,
	avatar_url = EXCLUDED.avatar_url,
	avatar_object = EXCLUDED.avatar_object
`, user.ID, strings.ToLower(strings.TrimSpace(user.Email)), user.PasswordHash, user.DisplayName, user.EmailVerified, user.Status, nullableTime(user.LastLoginAt), user.CreatedAt, user.UpdatedAt, user.AvatarURL, user.AvatarObject)
	if err != nil {
		return fmt.Errorf("upsert user %s: %w", user.ID, err)
	}
	return nil
}

func (r *Repository) upsertCompanyTx(ctx context.Context, tx *sql.Tx, company *Company) error {
	execTarget := sqlExecTarget(r.db, tx)
	_, err := execTarget.ExecContext(ctx, `
INSERT INTO companies (
	id, legal_name, brand_name, avatar_url, avatar_object, description, industry, website_url, email_domain, inn, ogrn, company_size, founded_year, hq_city_id, status, created_at, updated_at
)
VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, 0), NULLIF($14, 0), $15, $16, $17)
ON CONFLICT (id) DO UPDATE SET
	legal_name = EXCLUDED.legal_name,
	brand_name = EXCLUDED.brand_name,
	avatar_url = EXCLUDED.avatar_url,
	avatar_object = EXCLUDED.avatar_object,
	description = EXCLUDED.description,
	industry = EXCLUDED.industry,
	website_url = EXCLUDED.website_url,
	email_domain = EXCLUDED.email_domain,
	inn = EXCLUDED.inn,
	ogrn = EXCLUDED.ogrn,
	company_size = EXCLUDED.company_size,
	founded_year = EXCLUDED.founded_year,
	hq_city_id = EXCLUDED.hq_city_id,
	status = EXCLUDED.status,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at
`, company.ID, company.LegalName, company.BrandName, company.AvatarURL, company.AvatarObject, company.Description, company.Industry, company.WebsiteURL, company.EmailDomain, company.INN, company.OGRN, company.CompanySize, zeroableInt(company.FoundedYear), zeroableInt64(company.HQCityID), company.Status, company.CreatedAt, company.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert company %s: %w", company.ID, err)
	}
	return nil
}

func (r *Repository) replaceUserRolesTx(ctx context.Context, tx *sql.Tx, userID string, roles []string, roleIDs map[string]int16) error {
	execTarget := sqlExecTarget(r.db, tx)
	if _, err := execTarget.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear user roles for %s: %w", userID, err)
	}
	for _, role := range roles {
		roleID, ok := roleIDs[role]
		if !ok {
			return fmt.Errorf("unknown role code %q", role)
		}
		if _, err := execTarget.ExecContext(ctx, `
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING
`, userID, roleID); err != nil {
			return fmt.Errorf("insert role %s for user %s: %w", role, userID, err)
		}
	}
	return nil
}

func (r *Repository) upsertStudentProfileTx(ctx context.Context, tx *sql.Tx, profile *StudentProfile) error {
	execTarget := sqlExecTarget(r.db, tx)
	_, err := execTarget.ExecContext(ctx, `
INSERT INTO student_profiles (
	user_id, last_name, first_name, middle_name, university_name, faculty, specialization, study_year, graduation_year, about, profile_visibility, show_resume, show_applications, show_career_interests, telegram, github_url, linkedin_url, website_url, city_id, created_at, updated_at
)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, 0), NULLIF($9, 0), NULLIF($10, ''), $11, $12, $13, $14, NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''), NULLIF($18, ''), NULLIF($19, 0), $20, $21)
ON CONFLICT (user_id) DO UPDATE SET
	last_name = EXCLUDED.last_name,
	first_name = EXCLUDED.first_name,
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
	telegram = EXCLUDED.telegram,
	github_url = EXCLUDED.github_url,
	linkedin_url = EXCLUDED.linkedin_url,
	website_url = EXCLUDED.website_url,
	city_id = EXCLUDED.city_id,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at
`, profile.UserID, profile.LastName, profile.FirstName, profile.MiddleName, profile.UniversityName, profile.Faculty, profile.Specialization, zeroableInt(profile.StudyYear), zeroableInt(profile.GraduationYear), profile.About, defaultString(profile.ProfileVisibility, "authorized_only"), profile.ShowResume, profile.ShowApplications, profile.ShowCareerInterests, profile.Telegram, profile.GithubURL, profile.LinkedinURL, profile.WebsiteURL, zeroableInt64(profile.CityID), profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert student profile %s: %w", profile.UserID, err)
	}
	return nil
}

func (r *Repository) upsertEmployerProfileTx(ctx context.Context, tx *sql.Tx, profile *EmployerProfile) error {
	execTarget := sqlExecTarget(r.db, tx)
	_, err := execTarget.ExecContext(ctx, `
INSERT INTO employer_profiles (
	user_id, company_id, position_title, is_company_owner, can_create_opportunities, can_edit_company_profile, created_at, updated_at
)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
	company_id = EXCLUDED.company_id,
	position_title = EXCLUDED.position_title,
	is_company_owner = EXCLUDED.is_company_owner,
	can_create_opportunities = EXCLUDED.can_create_opportunities,
	can_edit_company_profile = EXCLUDED.can_edit_company_profile,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at
`, profile.UserID, profile.CompanyID, profile.PositionTitle, profile.IsCompanyOwner, profile.CanCreateOpportunities, profile.CanEditCompanyProfile, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert employer profile %s: %w", profile.UserID, err)
	}
	return nil
}

func (r *Repository) upsertCuratorProfileTx(ctx context.Context, tx *sql.Tx, profile *CuratorProfile) error {
	execTarget := sqlExecTarget(r.db, tx)
	_, err := execTarget.ExecContext(ctx, `
INSERT INTO curator_profiles (
	user_id, curator_type, created_by_user_id, notes, created_at, updated_at
)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
ON CONFLICT (user_id) DO UPDATE SET
	curator_type = EXCLUDED.curator_type,
	created_by_user_id = EXCLUDED.created_by_user_id,
	notes = EXCLUDED.notes,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at
`, profile.UserID, profile.CuratorType, nullableUUID(profile.CreatedByUserID), profile.Notes, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert curator profile %s: %w", profile.UserID, err)
	}
	return nil
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func nullableUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableDate(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func zeroableInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func zeroableInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func sqlExecTarget(db *sql.DB, tx *sql.Tx) dbExecutor {
	if tx != nil {
		return tx
	}
	return db
}

func splitDisplayName(displayName string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(displayName))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func (r *Repository) legacyMediaURL(storagePath string) string {
	storagePath = strings.TrimLeft(strings.TrimSpace(storagePath), "/")
	if storagePath == "" {
		return ""
	}
	if strings.HasPrefix(storagePath, "http://") || strings.HasPrefix(storagePath, "https://") {
		return storagePath
	}
	if r.publicBaseURL == "" || r.bucket == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", r.publicBaseURL, r.bucket, path.Clean(storagePath))
}

func fallback(value, current string) string {
	if strings.TrimSpace(value) == "" {
		return current
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func cloneUser(user *User) *User {
	cp := *user
	return &cp
}

func cloneVerification(item *CompanyVerification) *CompanyVerification {
	cp := *item
	return &cp
}

func scanOpportunityBase(scanner rowScanner, item *Opportunity) error {
	var publishedAt sql.NullTime
	var expiresAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.CompanyID,
		&item.CreatedByUserID,
		&item.Title,
		&item.ShortDescription,
		&item.FullDescription,
		&item.OpportunityType,
		&item.WorkFormat,
		&item.LocationID,
		&publishedAt,
		&expiresAt,
		&item.Status,
		&item.ContactsInfo,
		&item.ExternalURL,
		&item.ViewsCount,
		&item.FavoritesCount,
		&item.ApplicationsCount,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return err
	}
	if publishedAt.Valid {
		item.PublishedAt = publishedAt.Time
	}
	if expiresAt.Valid {
		item.ExpiresAt = expiresAt.Time
	}
	return nil
}

func scanOpportunityBaseWithExtras(scanner rowScanner, item *Opportunity, extras ...any) error {
	var publishedAt sql.NullTime
	var expiresAt sql.NullTime
	dest := []any{
		&item.ID,
		&item.CompanyID,
		&item.CreatedByUserID,
		&item.Title,
		&item.ShortDescription,
		&item.FullDescription,
		&item.OpportunityType,
		&item.WorkFormat,
		&item.LocationID,
		&publishedAt,
		&expiresAt,
		&item.Status,
		&item.ContactsInfo,
		&item.ExternalURL,
		&item.ViewsCount,
		&item.FavoritesCount,
		&item.ApplicationsCount,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
	dest = append(dest, extras...)
	if err := scanner.Scan(dest...); err != nil {
		return err
	}
	if publishedAt.Valid {
		item.PublishedAt = publishedAt.Time
	}
	if expiresAt.Valid {
		item.ExpiresAt = expiresAt.Time
	}
	return nil
}

func zeroableFloat(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}

func applicationStatusLabel(status string) string {
	switch status {
	case "submitted":
		return "отправлен"
	case "in_review":
		return "на рассмотрении"
	case "accepted":
		return "принят"
	case "rejected":
		return "отклонён"
	case "reserve":
		return "в резерве"
	case "withdrawn":
		return "отозван"
	default:
		return status
	}
}

func orderedUserIDs(a, b string) (string, string) {
	if strings.Compare(a, b) <= 0 {
		return a, b
	}
	return b, a
}

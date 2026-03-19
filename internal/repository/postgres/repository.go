package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	. "tramplin/internal/models"
	"tramplin/internal/repository"
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

func (r *Repository) getUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	var avatarURL sql.NullString
	var avatarObject sql.NullString
	var lastLoginAt sql.NullTime
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT
	u.id,
	u.email,
	u.password_hash,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(u.avatar_object, ''),
	u.email_verified,
	u.status,
	u.last_login_at,
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	u.created_at,
	u.updated_at
FROM users
u
LEFT JOIN user_presence up ON up.user_id = u.id
WHERE u.id = $1
`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &avatarURL, &avatarObject, &user.EmailVerified, &user.Status, &lastLoginAt, &user.IsOnline, &lastSeenAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if avatarObject.Valid {
		user.AvatarObject = avatarObject.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return &user, nil
}

func (r *Repository) getUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	var avatarURL sql.NullString
	var avatarObject sql.NullString
	var lastLoginAt sql.NullTime
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT
	u.id,
	u.email,
	u.password_hash,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(u.avatar_object, ''),
	u.email_verified,
	u.status,
	u.last_login_at,
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	u.created_at,
	u.updated_at
FROM users
u
LEFT JOIN user_presence up ON up.user_id = u.id
WHERE u.email = $1
`, strings.ToLower(strings.TrimSpace(email))).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &avatarURL, &avatarObject, &user.EmailVerified, &user.Status, &lastLoginAt, &user.IsOnline, &lastSeenAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("get user by email %s: %w", email, err)
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if avatarObject.Valid {
		user.AvatarObject = avatarObject.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return &user, nil
}

func (r *Repository) getUserRolesFromDB(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT r.code
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.id
`, userID)
	if err != nil {
		return nil, fmt.Errorf("get roles for %s: %w", userID, err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan roles for %s: %w", userID, err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles for %s: %w", userID, err)
	}
	if len(roles) == 0 {
		return nil, errors.New("roles not found")
	}
	return roles, nil
}

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

func (r *Repository) RegisterUser(params repository.RegisterUserParams) (*User, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailKey := strings.ToLower(strings.TrimSpace(params.Email))
	if emailKey == "" || strings.TrimSpace(params.Password) == "" || strings.TrimSpace(params.DisplayName) == "" {
		return nil, "", errors.New("email, password and display_name are required")
	}
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, emailKey).Scan(&exists); err != nil {
		return nil, "", fmt.Errorf("check existing user: %w", err)
	}
	if exists {
		return nil, "", errors.New("user with this email already exists")
	}
	if params.Role != repository.RoleStudent && params.Role != repository.RoleEmployer {
		return nil, "", fmt.Errorf("role must be one of: %s, %s", repository.RoleStudent, repository.RoleEmployer)
	}

	now := time.Now()
	user := &User{ID: uuid.NewString(), Email: emailKey, PasswordHash: hashPassword(params.Password), DisplayName: params.DisplayName, Status: "active", EmailVerified: false, CreatedAt: now, UpdatedAt: now}
	var createdCompanyID string
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("begin register user: %w", err)
	}
	roleIDs, err := r.loadRoleIDs(context.Background(), tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, "", err
	}
	if err := r.upsertUserTx(context.Background(), tx, user); err != nil {
		_ = tx.Rollback()
		return nil, "", err
	}
	if err := r.replaceUserRolesTx(context.Background(), tx, user.ID, []string{params.Role}, roleIDs); err != nil {
		_ = tx.Rollback()
		return nil, "", err
	}
	if params.Role == repository.RoleStudent {
		firstName, lastName := splitDisplayName(params.DisplayName)
		profile := &StudentProfile{UserID: user.ID, FirstName: firstName, LastName: lastName, UniversityName: "", ProfileVisibility: "authorized_only", ShowResume: true, ShowApplications: false, ShowCareerInterests: true, CreatedAt: now, UpdatedAt: now}
		if err := r.upsertStudentProfileTx(context.Background(), tx, profile); err != nil {
			_ = tx.Rollback()
			return nil, "", err
		}
	}
	if params.Role == repository.RoleEmployer {
		companyName := strings.TrimSpace(params.CompanyName)
		if companyName == "" {
			companyName = params.DisplayName + " Company"
		}
		company := &Company{ID: uuid.NewString(), LegalName: companyName, BrandName: companyName, Status: "pending_verification", CreatedAt: now, UpdatedAt: now}
		if err := r.upsertCompanyTx(context.Background(), tx, company); err != nil {
			_ = tx.Rollback()
			return nil, "", err
		}
		profile := &EmployerProfile{UserID: user.ID, CompanyID: company.ID, IsCompanyOwner: true, CanCreateOpportunities: false, CanEditCompanyProfile: true, CreatedAt: now, UpdatedAt: now}
		if err := r.upsertEmployerProfileTx(context.Background(), tx, profile); err != nil {
			_ = tx.Rollback()
			return nil, "", err
		}
		createdCompanyID = company.ID
	}
	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("commit register user: %w", err)
	}
	if createdCompanyID != "" {
		r.addModerationItem("company", createdCompanyID, user.ID)
	}
	r.addAudit("create", "users", user.ID, user.ID, "user registered")
	return cloneUser(user), params.Role, nil
}

func (r *Repository) Login(email, password string) (*User, []string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, err := r.getUserByEmail(context.Background(), email)
	if err != nil {
		return nil, nil, err
	}
	if user.PasswordHash != hashPassword(password) {
		return nil, nil, errors.New("invalid credentials")
	}
	user.LastLoginAt = time.Now()
	user.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1`, user.ID, user.LastLoginAt, user.UpdatedAt); err != nil {
		return nil, nil, fmt.Errorf("update last login: %w", err)
	}
	roles, err := r.getUserRolesFromDB(context.Background(), user.ID)
	if err != nil {
		return nil, nil, err
	}
	return cloneUser(user), roles, nil
}

func (r *Repository) GetUser(userID string) (*User, error) {
	return r.getUserByID(context.Background(), userID)
}

func (r *Repository) GetUserRoles(userID string) ([]string, error) {
	return r.getUserRolesFromDB(context.Background(), userID)
}

func (r *Repository) CreateCurator(email, password, displayName, curatorType, createdBy string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.hasRole(createdBy, "admin") {
		return nil, errors.New("only administrator can create curator accounts")
	}
	if curatorType == "" {
		curatorType = "moderator"
	}
	key := strings.ToLower(strings.TrimSpace(email))
	var exists bool
	if err := r.db.QueryRowContext(context.Background(), `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, key).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check existing curator: %w", err)
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}
	now := time.Now()
	user := &User{ID: uuid.NewString(), Email: key, PasswordHash: hashPassword(password), DisplayName: displayName, Status: "active", EmailVerified: true, CreatedAt: now, UpdatedAt: now}
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("begin create curator: %w", err)
	}
	roleIDs, err := r.loadRoleIDs(context.Background(), tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := r.upsertUserTx(context.Background(), tx, user); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	roles := []string{"curator"}
	if curatorType == "administrator" {
		roles = append(roles, "admin")
	}
	if err := r.replaceUserRolesTx(context.Background(), tx, user.ID, roles, roleIDs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	profile := &CuratorProfile{UserID: user.ID, CuratorType: curatorType, CreatedByUserID: createdBy, CreatedAt: now, UpdatedAt: now}
	if err := r.upsertCuratorProfileTx(context.Background(), tx, profile); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create curator: %w", err)
	}
	r.addAudit("create", "curator_profiles", user.ID, createdBy, "curator created")
	return cloneUser(user), nil
}

func (r *Repository) UpdateUserStatus(userID, status, actorID string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, err := r.getUserByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	user.Status = status
	user.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE users SET status = $2, updated_at = $3 WHERE id = $1`, user.ID, user.Status, user.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update user status: %w", err)
	}
	r.addAudit("status_change", "users", user.ID, actorID, status)
	return cloneUser(user), nil
}

func (r *Repository) UpdateUserAvatar(userID, avatarObject, avatarURL string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, err := r.getUserByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	user.AvatarObject = avatarObject
	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE users SET avatar_object = NULLIF($2, ''), avatar_url = NULLIF($3, ''), updated_at = $4 WHERE id = $1`, userID, avatarObject, avatarURL, user.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update user avatar: %w", err)
	}
	r.addAudit("update", "users", userID, userID, "avatar updated")
	return cloneUser(user), nil
}

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
SELECT cr.id, cr.sender_user_id, cr.receiver_user_id, COALESCE(s.avatar_url, ''), COALESCE(r.avatar_url, ''), COALESCE(cr.message, ''), cr.status, cr.created_at, cr.updated_at
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
		if err := rows.Scan(&item.ID, &item.SenderUserID, &item.ReceiverUserID, &item.SenderAvatarURL, &item.ReceiverAvatarURL, &item.Message, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contact requests: %w", err)
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

func (r *Repository) CreateChatConversation(userID, participantUserID, opportunityID string) (*ChatConversation, error) {
	userID = strings.TrimSpace(userID)
	participantUserID = strings.TrimSpace(participantUserID)
	opportunityID = strings.TrimSpace(opportunityID)
	if userID == "" || participantUserID == "" {
		return nil, errors.New("participant_user_id is required")
	}
	if userID == participantUserID {
		return nil, errors.New("cannot create chat with yourself")
	}
	if _, err := r.getUserByID(context.Background(), participantUserID); err != nil {
		return nil, err
	}

	aID, bID := orderedUserIDs(userID, participantUserID)
	var existingID string
	err := r.db.QueryRowContext(context.Background(), `
SELECT id
FROM chat_conversations
WHERE participant_a_user_id = $1
  AND participant_b_user_id = $2
  AND (
      (opportunity_id = NULLIF($3, '')::uuid)
      OR (opportunity_id IS NULL AND NULLIF($3, '')::uuid IS NULL)
  )
`, aID, bID, opportunityID).Scan(&existingID)
	if err == nil {
		return r.GetChatConversation(userID, existingID)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing chat conversation: %w", err)
	}

	conversationID := uuid.NewString()
	now := time.Now()
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO chat_conversations (
	id, participant_a_user_id, participant_b_user_id, opportunity_id, created_at, updated_at
)
VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6)
`, conversationID, aID, bID, opportunityID, now, now); err != nil {
		return nil, fmt.Errorf("create chat conversation: %w", err)
	}
	return r.GetChatConversation(userID, conversationID)
}

func (r *Repository) GetChatConversation(userID, conversationID string) (*ChatConversation, error) {
	var item ChatConversation
	var participantName string
	var participantAvatarURL sql.NullString
	var lastMessage sql.NullString
	var lastMessageAt sql.NullTime
	var participantLastSeenAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT
	c.id,
	COALESCE(c.opportunity_id::text, ''),
	COALESCE(o.title, ''),
	COALESCE(comp.legal_name, ''),
	CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END AS participant_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	COALESCE(last_message.body, ''),
	last_message.created_at,
	COALESCE(unread.unread_count, 0),
	c.created_at,
	c.updated_at
FROM chat_conversations c
JOIN users u ON u.id = CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END
LEFT JOIN user_presence up ON up.user_id = u.id
LEFT JOIN opportunities o ON o.id = c.opportunity_id
LEFT JOIN companies comp ON comp.id = o.company_id
LEFT JOIN LATERAL (
	SELECT body, created_at
	FROM chat_messages
	WHERE conversation_id = c.id
	ORDER BY created_at DESC
	LIMIT 1
) AS last_message ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS unread_count
	FROM chat_messages m
	WHERE m.conversation_id = c.id
	  AND m.sender_user_id <> $1
	  AND m.read_at IS NULL
) AS unread ON TRUE
WHERE c.id = $2
  AND ($1 = c.participant_a_user_id OR $1 = c.participant_b_user_id)
`, userID, conversationID).Scan(&item.ID, &item.OpportunityID, &item.OpportunityTitle, &item.CompanyLegalName, &item.ParticipantUserID, &participantName, &participantAvatarURL, &item.ParticipantIsOnline, &participantLastSeenAt, &lastMessage, &lastMessageAt, &item.UnreadCount, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("chat conversation not found")
		}
		return nil, fmt.Errorf("get chat conversation: %w", err)
	}
	item.ParticipantName = participantName
	if participantAvatarURL.Valid {
		item.ParticipantAvatarURL = participantAvatarURL.String
	}
	if participantLastSeenAt.Valid {
		item.ParticipantLastSeenAt = &participantLastSeenAt.Time
	}
	if lastMessage.Valid {
		item.LastMessage = lastMessage.String
	}
	if lastMessageAt.Valid {
		item.LastMessageAt = lastMessageAt.Time
	}
	return &item, nil
}

func (r *Repository) ListChatConversations(userID string) ([]ChatConversation, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	c.id,
	COALESCE(c.opportunity_id::text, ''),
	COALESCE(o.title, ''),
	COALESCE(comp.legal_name, ''),
	CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END AS participant_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	up.last_seen_at,
	COALESCE(last_message.body, ''),
	last_message.created_at,
	COALESCE(unread.unread_count, 0),
	c.created_at,
	c.updated_at
FROM chat_conversations c
JOIN users u ON u.id = CASE WHEN c.participant_a_user_id = $1 THEN c.participant_b_user_id ELSE c.participant_a_user_id END
LEFT JOIN user_presence up ON up.user_id = u.id
LEFT JOIN opportunities o ON o.id = c.opportunity_id
LEFT JOIN companies comp ON comp.id = o.company_id
LEFT JOIN LATERAL (
	SELECT body, created_at
	FROM chat_messages
	WHERE conversation_id = c.id
	ORDER BY created_at DESC
	LIMIT 1
) AS last_message ON TRUE
LEFT JOIN LATERAL (
	SELECT COUNT(*) AS unread_count
	FROM chat_messages m
	WHERE m.conversation_id = c.id
	  AND m.sender_user_id <> $1
	  AND m.read_at IS NULL
) AS unread ON TRUE
WHERE c.participant_a_user_id = $1 OR c.participant_b_user_id = $1
ORDER BY COALESCE(last_message.created_at, c.updated_at) DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat conversations: %w", err)
	}
	defer rows.Close()

	var result []ChatConversation
	for rows.Next() {
		var item ChatConversation
		var participantName string
		var participantAvatarURL sql.NullString
		var participantLastSeenAt sql.NullTime
		var lastMessage sql.NullString
		var lastMessageAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OpportunityID, &item.OpportunityTitle, &item.CompanyLegalName, &item.ParticipantUserID, &participantName, &participantAvatarURL, &item.ParticipantIsOnline, &participantLastSeenAt, &lastMessage, &lastMessageAt, &item.UnreadCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat conversations: %w", err)
		}
		item.ParticipantName = participantName
		if participantAvatarURL.Valid {
			item.ParticipantAvatarURL = participantAvatarURL.String
		}
		if participantLastSeenAt.Valid {
			item.ParticipantLastSeenAt = &participantLastSeenAt.Time
		}
		if lastMessage.Valid {
			item.LastMessage = lastMessage.String
		}
		if lastMessageAt.Valid {
			item.LastMessageAt = lastMessageAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ListChatMessages(userID, conversationID string) ([]ChatMessage, error) {
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return nil, err
	}
	if _, err := r.MarkChatMessagesRead(userID, conversationID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(context.Background(), `
SELECT
	m.id,
	m.conversation_id,
	m.sender_user_id,
	u.display_name,
	COALESCE(u.avatar_url, ''),
	COALESCE(up.is_online, FALSE),
	m.body,
	(m.read_at IS NOT NULL),
	m.read_at,
	m.created_at
FROM chat_messages m
JOIN users u ON u.id = m.sender_user_id
LEFT JOIN user_presence up ON up.user_id = u.id
WHERE m.conversation_id = $1
ORDER BY m.created_at ASC
`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	var result []ChatMessage
	for rows.Next() {
		var item ChatMessage
		var readAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.SenderUserID, &item.SenderName, &item.SenderAvatarURL, &item.SenderIsOnline, &item.Body, &item.IsRead, &readAt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat messages: %w", err)
		}
		if readAt.Valid {
			item.ReadAt = &readAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateChatMessage(userID, conversationID, body string) (*ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errors.New("message body is required")
	}
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return nil, err
	}
	item := &ChatMessage{
		ID:             uuid.NewString(),
		ConversationID: conversationID,
		SenderUserID:   userID,
		Body:           body,
		CreatedAt:      time.Now(),
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO chat_messages (id, conversation_id, sender_user_id, body, created_at)
VALUES ($1, $2, $3, $4, $5)
`, item.ID, item.ConversationID, item.SenderUserID, item.Body, item.CreatedAt); err != nil {
		return nil, fmt.Errorf("create chat message: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE chat_conversations
SET updated_at = $2
WHERE id = $1
`, conversationID, item.CreatedAt); err != nil {
		return nil, fmt.Errorf("update chat conversation timestamp: %w", err)
	}
	sender, err := r.getUserByID(context.Background(), userID)
	if err != nil {
		return nil, err
	}
	item.SenderName = sender.DisplayName
	item.SenderAvatarURL = sender.AvatarURL
	item.SenderIsOnline = sender.IsOnline
	return item, nil
}

func (r *Repository) MarkChatMessagesRead(userID, conversationID string) (int64, error) {
	if _, err := r.GetChatConversation(userID, conversationID); err != nil {
		return 0, err
	}
	result, err := r.db.ExecContext(context.Background(), `
UPDATE chat_messages
SET read_at = NOW()
WHERE conversation_id = $1
  AND sender_user_id <> $2
  AND read_at IS NULL
`, conversationID, userID)
	if err != nil {
		return 0, fmt.Errorf("mark chat messages read: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count marked chat messages: %w", err)
	}
	return updated, nil
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

func (r *Repository) GetEmployerProfile(userID string) (*EmployerProfile, error) {
	var profile EmployerProfile
	var positionTitle sql.NullString
	var avatarURL sql.NullString
	err := r.db.QueryRowContext(context.Background(), `
SELECT ep.user_id, COALESCE(u.avatar_url, ''), ep.company_id, ep.position_title, ep.is_company_owner, ep.can_create_opportunities, ep.can_edit_company_profile, ep.created_at, ep.updated_at
FROM employer_profiles ep
JOIN users u ON u.id = ep.user_id
WHERE ep.user_id = $1
`, userID).Scan(&profile.UserID, &avatarURL, &profile.CompanyID, &positionTitle, &profile.IsCompanyOwner, &profile.CanCreateOpportunities, &profile.CanEditCompanyProfile, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("employer profile not found")
		}
		return nil, fmt.Errorf("get employer profile %s: %w", userID, err)
	}
	if avatarURL.Valid {
		profile.AvatarURL = avatarURL.String
	}
	if positionTitle.Valid {
		profile.PositionTitle = positionTitle.String
	}
	return &profile, nil
}

func (r *Repository) GetEmployerCompany(userID string) (*Company, error) {
	var company Company
	var brandName sql.NullString
	var avatarURL sql.NullString
	var avatarObject sql.NullString
	var legacyAvatarPath sql.NullString
	var description sql.NullString
	var industry sql.NullString
	var websiteURL sql.NullString
	var emailDomain sql.NullString
	var inn sql.NullString
	var ogrn sql.NullString
	var companySize sql.NullString
	err := r.db.QueryRowContext(context.Background(), `
SELECT c.id, c.legal_name, c.brand_name, COALESCE(cu.avatar_url, c.avatar_url), COALESCE(cu.avatar_object, c.avatar_object), COALESCE(mf.storage_path, ''), c.description, c.industry, c.website_url, c.email_domain, c.inn, c.ogrn, c.company_size, COALESCE(c.founded_year, 0), COALESCE(c.hq_city_id, 0), c.status, c.created_at, c.updated_at
FROM employer_profiles ep
JOIN companies c ON c.id = ep.company_id
LEFT JOIN LATERAL (
	SELECT u.avatar_url, u.avatar_object
	FROM employer_profiles ep2
	JOIN users u ON u.id = ep2.user_id
	WHERE ep2.company_id = c.id
	  AND (COALESCE(u.avatar_url, '') <> '' OR COALESCE(u.avatar_object, '') <> '')
	ORDER BY (ep2.user_id = ep.user_id) DESC, ep2.is_company_owner DESC, u.updated_at DESC, ep2.created_at ASC
	LIMIT 1
) cu ON TRUE
LEFT JOIN media_files mf ON mf.id = c.logo_media_id
WHERE ep.user_id = $1
`, userID).Scan(&company.ID, &company.LegalName, &brandName, &avatarURL, &avatarObject, &legacyAvatarPath, &description, &industry, &websiteURL, &emailDomain, &inn, &ogrn, &companySize, &company.FoundedYear, &company.HQCityID, &company.Status, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("company not found")
		}
		return nil, fmt.Errorf("get employer company for %s: %w", userID, err)
	}
	if brandName.Valid {
		company.BrandName = brandName.String
	}
	if avatarURL.Valid {
		company.AvatarURL = avatarURL.String
	}
	if avatarObject.Valid {
		company.AvatarObject = avatarObject.String
	}
	if company.AvatarURL == "" && legacyAvatarPath.Valid {
		company.AvatarURL = r.legacyMediaURL(legacyAvatarPath.String)
	}
	if description.Valid {
		company.Description = description.String
	}
	if industry.Valid {
		company.Industry = industry.String
	}
	if websiteURL.Valid {
		company.WebsiteURL = websiteURL.String
	}
	if emailDomain.Valid {
		company.EmailDomain = emailDomain.String
	}
	if inn.Valid {
		company.INN = inn.String
	}
	if ogrn.Valid {
		company.OGRN = ogrn.String
	}
	if companySize.Valid {
		company.CompanySize = companySize.String
	}
	return &company, nil
}

func (r *Repository) UpdateEmployerCompany(userID string, update repository.CompanyUpdate) (*Company, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	company, err := r.GetEmployerCompany(userID)
	if err != nil {
		return nil, err
	}
	company.LegalName = fallback(update.LegalName, company.LegalName)
	company.BrandName = fallback(update.BrandName, company.BrandName)
	company.Description = fallback(update.Description, company.Description)
	company.Industry = fallback(update.Industry, company.Industry)
	company.WebsiteURL = fallback(update.WebsiteURL, company.WebsiteURL)
	company.EmailDomain = fallback(update.EmailDomain, company.EmailDomain)
	company.INN = fallback(update.INN, company.INN)
	company.OGRN = fallback(update.OGRN, company.OGRN)
	company.CompanySize = fallback(update.CompanySize, company.CompanySize)
	if update.FoundedYear != 0 {
		company.FoundedYear = update.FoundedYear
	}
	if update.HQCityID != 0 {
		company.HQCityID = update.HQCityID
	}
	company.UpdatedAt = time.Now()
	if err := r.upsertCompanyTx(context.Background(), nil, company); err != nil {
		return nil, err
	}
	cp := *company
	r.addAudit("update", "companies", company.ID, userID, "company updated")
	return &cp, nil
}

func (r *Repository) UpdateCompanyAvatar(userID, avatarObject, avatarURL string) (*Company, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	company, err := r.GetEmployerCompany(userID)
	if err != nil {
		return nil, err
	}
	company.AvatarObject = avatarObject
	company.AvatarURL = avatarURL
	company.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `UPDATE companies SET avatar_object = NULLIF($2, ''), avatar_url = NULLIF($3, ''), updated_at = $4 WHERE id = $1`, company.ID, avatarObject, avatarURL, company.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update company avatar: %w", err)
	}
	r.addAudit("update", "companies", company.ID, userID, "company avatar updated")
	cp := *company
	return &cp, nil
}

func (r *Repository) CreateCompanyLink(userID, linkType, url string) (*CompanyLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, err := r.GetEmployerProfile(userID)
	if err != nil {
		return nil, err
	}
	link := CompanyLink{ID: uuid.NewString(), CompanyID: profile.CompanyID, LinkType: linkType, URL: url, CreatedAt: time.Now()}
	_, err = r.db.ExecContext(context.Background(), `
INSERT INTO company_links (id, company_id, link_type, url, created_at)
VALUES ($1, $2, $3, $4, $5)
`, link.ID, link.CompanyID, link.LinkType, link.URL, link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create company link: %w", err)
	}
	return &link, nil
}

func (r *Repository) SubmitCompanyVerification(userID, method, corporateEmail, inn, comment string) (*CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, err := r.GetEmployerProfile(userID)
	if err != nil {
		return nil, err
	}
	item := &CompanyVerification{ID: uuid.NewString(), CompanyID: profile.CompanyID, VerificationMethod: method, SubmittedByUserID: userID, CorporateEmail: corporateEmail, INNSubmitted: inn, DocumentsComment: comment, Status: "pending", SubmittedAt: time.Now()}
	_, err = r.db.ExecContext(context.Background(), `
INSERT INTO company_verifications (
	id, company_id, verification_method, submitted_by_user_id, corporate_email, inn_submitted, documents_comment, status, submitted_at
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, $9)
`, item.ID, item.CompanyID, item.VerificationMethod, item.SubmittedByUserID, item.CorporateEmail, item.INNSubmitted, item.DocumentsComment, item.Status, item.SubmittedAt)
	if err != nil {
		return nil, fmt.Errorf("submit company verification: %w", err)
	}
	r.addModerationItem("company", profile.CompanyID, userID)
	return cloneVerification(item), nil
}

func (r *Repository) ListEmployerOpportunities(userID string) ([]Opportunity, error) {
	profile, err := r.GetEmployerProfile(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, company_id, created_by_user_id, title, short_description, full_description, opportunity_type, work_format, COALESCE(location_id::text, ''), published_at, expires_at, status, COALESCE(contacts_info, ''), COALESCE(external_url, ''), views_count, favorites_count, applications_count, created_at, updated_at
FROM opportunities
WHERE company_id = $1
ORDER BY created_at DESC
`, profile.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("list employer opportunities: %w", err)
	}
	defer rows.Close()

	var result []Opportunity
	for rows.Next() {
		var item Opportunity
		if err := scanOpportunityBase(rows, &item); err != nil {
			return nil, fmt.Errorf("scan employer opportunities: %w", err)
		}
		if err := r.loadOpportunityTypeDetails(context.Background(), &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate employer opportunities: %w", err)
	}
	return result, nil
}

func (r *Repository) CreateOpportunity(opportunity Opportunity) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, err := r.GetEmployerProfile(opportunity.CreatedByUserID)
	if err != nil {
		return nil, err
	}
	if !profile.CanCreateOpportunities {
		return nil, errors.New("company verification is required before creating opportunities")
	}
	now := time.Now()
	opportunity.ID = uuid.NewString()
	opportunity.CompanyID = profile.CompanyID
	opportunity.CreatedAt = now
	opportunity.UpdatedAt = now
	if opportunity.Status == "" {
		opportunity.Status = "pending_moderation"
	}
	_, err = r.db.ExecContext(context.Background(), `
INSERT INTO opportunities (
	id, company_id, created_by_user_id, title, short_description, full_description, opportunity_type, work_format, location_id, published_at, expires_at, status, contacts_info, external_url, views_count, favorites_count, applications_count, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')::uuid, $10, $11, $12, NULLIF($13, ''), NULLIF($14, ''), $15, $16, $17, $18, $19)
`, opportunity.ID, opportunity.CompanyID, opportunity.CreatedByUserID, opportunity.Title, opportunity.ShortDescription, opportunity.FullDescription, opportunity.OpportunityType, opportunity.WorkFormat, opportunity.LocationID, nullableTime(opportunity.PublishedAt), nullableTime(opportunity.ExpiresAt), opportunity.Status, opportunity.ContactsInfo, opportunity.ExternalURL, opportunity.ViewsCount, opportunity.FavoritesCount, opportunity.ApplicationsCount, opportunity.CreatedAt, opportunity.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create opportunity: %w", err)
	}
	if err := r.replaceOpportunityTypeDetails(context.Background(), nil, &opportunity); err != nil {
		return nil, err
	}
	r.addModerationItem("opportunity", opportunity.ID, opportunity.CreatedByUserID)
	r.addAudit("create", "opportunities", opportunity.ID, opportunity.CreatedByUserID, opportunity.Title)
	return &opportunity, nil
}

func (r *Repository) GetEmployerOpportunity(userID, opportunityID string) (*Opportunity, error) {
	profile, err := r.GetEmployerProfile(userID)
	if err != nil {
		return nil, err
	}
	var item Opportunity
	row := r.db.QueryRowContext(context.Background(), `
SELECT id, company_id, created_by_user_id, title, short_description, full_description, opportunity_type, work_format, COALESCE(location_id::text, ''), published_at, expires_at, status, COALESCE(contacts_info, ''), COALESCE(external_url, ''), views_count, favorites_count, applications_count, created_at, updated_at
FROM opportunities
WHERE id = $1 AND company_id = $2
`, opportunityID, profile.CompanyID)
	err = scanOpportunityBase(row, &item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("opportunity not found")
		}
		return nil, fmt.Errorf("get employer opportunity: %w", err)
	}
	if err := r.loadOpportunityTypeDetails(context.Background(), &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) UpdateEmployerOpportunity(userID string, opportunity Opportunity) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	existing, err := r.GetEmployerOpportunity(userID, opportunity.ID)
	if err != nil {
		return nil, err
	}
	existing.Title = fallback(opportunity.Title, existing.Title)
	existing.ShortDescription = fallback(opportunity.ShortDescription, existing.ShortDescription)
	existing.FullDescription = fallback(opportunity.FullDescription, existing.FullDescription)
	existing.OpportunityType = fallback(opportunity.OpportunityType, existing.OpportunityType)
	existing.VacancyLevel = fallback(opportunity.VacancyLevel, existing.VacancyLevel)
	existing.EmploymentType = fallback(opportunity.EmploymentType, existing.EmploymentType)
	existing.WorkFormat = fallback(opportunity.WorkFormat, existing.WorkFormat)
	if opportunity.LocationID != "" {
		existing.LocationID = opportunity.LocationID
	}
	if opportunity.SalaryMin != 0 {
		existing.SalaryMin = opportunity.SalaryMin
	}
	if opportunity.SalaryMax != 0 {
		existing.SalaryMax = opportunity.SalaryMax
	}
	existing.SalaryCurrency = fallback(opportunity.SalaryCurrency, existing.SalaryCurrency)
	existing.IsSalaryVisible = opportunity.IsSalaryVisible || existing.IsSalaryVisible
	existing.Status = fallback(opportunity.Status, existing.Status)
	existing.ContactsInfo = fallback(opportunity.ContactsInfo, existing.ContactsInfo)
	existing.ExternalURL = fallback(opportunity.ExternalURL, existing.ExternalURL)
	if len(opportunity.TagIDs) > 0 {
		existing.TagIDs = opportunity.TagIDs
	}
	existing.UpdatedAt = time.Now()
	_, err = r.db.ExecContext(context.Background(), `
UPDATE opportunities
SET title = $2, short_description = $3, full_description = $4, opportunity_type = $5, work_format = $6, location_id = NULLIF($7, '')::uuid, status = $8, contacts_info = NULLIF($9, ''), external_url = NULLIF($10, ''), updated_at = $11, expires_at = $12
WHERE id = $1
`, existing.ID, existing.Title, existing.ShortDescription, existing.FullDescription, existing.OpportunityType, existing.WorkFormat, existing.LocationID, existing.Status, existing.ContactsInfo, existing.ExternalURL, existing.UpdatedAt, nullableTime(existing.ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("update employer opportunity: %w", err)
	}
	if err := r.replaceOpportunityTypeDetails(context.Background(), nil, existing); err != nil {
		return nil, err
	}
	cp := *existing
	return &cp, nil
}

func (r *Repository) ListOpportunityApplications(userID, opportunityID string) ([]Application, error) {
	if _, err := r.GetEmployerOpportunity(userID, opportunityID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(context.Background(), `
SELECT a.id, a.opportunity_id, a.student_user_id, COALESCE(u.avatar_url, ''), COALESCE(a.resume_id::text, ''), COALESCE(a.cover_letter, ''), a.status, COALESCE(a.status_changed_by_user_id::text, ''), a.status_changed_at, a.created_at, a.updated_at
FROM applications a
JOIN users u ON u.id = a.student_user_id
WHERE a.opportunity_id = $1
ORDER BY a.created_at DESC
`, opportunityID)
	if err != nil {
		return nil, fmt.Errorf("list opportunity applications: %w", err)
	}
	defer rows.Close()

	var result []Application
	for rows.Next() {
		var item Application
		var statusChangedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OpportunityID, &item.StudentUserID, &item.StudentAvatarURL, &item.ResumeID, &item.CoverLetter, &item.Status, &item.StatusChangedByUserID, &statusChangedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan opportunity applications: %w", err)
		}
		if statusChangedAt.Valid {
			item.StatusChangedAt = statusChangedAt.Time
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate opportunity applications: %w", err)
	}
	return result, nil
}

func (r *Repository) UpdateApplicationStatus(userID, applicationID, status string) (*Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	var applicationIDCheck string
	var opportunityID string
	err := r.db.QueryRowContext(context.Background(), `
SELECT a.id, a.opportunity_id
FROM applications a
JOIN opportunities o ON o.id = a.opportunity_id
JOIN employer_profiles ep ON ep.company_id = o.company_id
WHERE a.id = $1 AND ep.user_id = $2
`, applicationID, userID).Scan(&applicationIDCheck, &opportunityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("forbidden")
		}
		return nil, fmt.Errorf("check application access: %w", err)
	}
	var application Application
	var statusChangedAt sql.NullTime
	err = r.db.QueryRowContext(context.Background(), `
SELECT a.id, a.opportunity_id, a.student_user_id, COALESCE(u.avatar_url, ''), COALESCE(a.resume_id::text, ''), COALESCE(a.cover_letter, ''), a.status, COALESCE(a.status_changed_by_user_id::text, ''), a.status_changed_at, a.created_at, a.updated_at
FROM applications a
JOIN users u ON u.id = a.student_user_id
WHERE a.id = $1
`, applicationID).Scan(&application.ID, &application.OpportunityID, &application.StudentUserID, &application.StudentAvatarURL, &application.ResumeID, &application.CoverLetter, &application.Status, &application.StatusChangedByUserID, &statusChangedAt, &application.CreatedAt, &application.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	if statusChangedAt.Valid {
		application.StatusChangedAt = statusChangedAt.Time
	}
	application.Status = status
	application.StatusChangedByUserID = userID
	application.StatusChangedAt = time.Now()
	application.UpdatedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE applications
SET status = $2, status_changed_by_user_id = $3, status_changed_at = $4, updated_at = $5
WHERE id = $1
`, application.ID, application.Status, application.StatusChangedByUserID, application.StatusChangedAt, application.UpdatedAt); err != nil {
		return nil, fmt.Errorf("update application status: %w", err)
	}
	if _, err := r.db.ExecContext(context.Background(), `
INSERT INTO application_status_history (id, application_id, old_status, new_status, changed_by_user_id, created_at)
VALUES ($1, $2, NULL, $3, $4, $5)
`, uuid.NewString(), application.ID, application.Status, userID, application.StatusChangedAt); err != nil {
		return nil, fmt.Errorf("insert application status history: %w", err)
	}
	r.notify(application.StudentUserID, "application_status_changed", "Статус отклика изменён", fmt.Sprintf("Ваш отклик теперь имеет статус: %s", applicationStatusLabel(status)), "application", application.ID)
	return &application, nil
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

func (r *Repository) recommendationNotificationText(opportunityID string) (string, string) {
	var opportunityTitle sql.NullString
	var companyLegalName sql.NullString
	var contactsInfo sql.NullString
	err := r.db.QueryRowContext(context.Background(), `
SELECT COALESCE(o.title, ''), COALESCE(c.legal_name, ''), COALESCE(o.contacts_info, '')
FROM opportunities o
JOIN companies c ON c.id = o.company_id
WHERE o.id = $1
`, opportunityID).Scan(&opportunityTitle, &companyLegalName, &contactsInfo)
	if err != nil {
		return "Новое приглашение", "Вас пригласили рассмотреть возможность."
	}

	title := "Новое приглашение"
	if companyLegalName.String != "" {
		title = fmt.Sprintf("Приглашение от компании %s", companyLegalName.String)
	}

	bodyParts := []string{}
	if opportunityTitle.String != "" {
		bodyParts = append(bodyParts, fmt.Sprintf("Вас пригласили на возможность \"%s\".", opportunityTitle.String))
	} else {
		bodyParts = append(bodyParts, "Вас пригласили рассмотреть возможность.")
	}
	if companyLegalName.String != "" {
		bodyParts = append(bodyParts, fmt.Sprintf("Компания: %s.", companyLegalName.String))
	}
	if contactsInfo.String != "" {
		bodyParts = append(bodyParts, fmt.Sprintf("Контакты работодателя: %s.", contactsInfo.String))
	}
	return title, strings.Join(bodyParts, " ")
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

func (r *Repository) UpdateEmployerProfile(userID string, profile EmployerProfile, actorID string) (*EmployerProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, err := r.GetEmployerProfile(userID)
	if err != nil {
		return nil, err
	}
	existing.PositionTitle = fallback(profile.PositionTitle, existing.PositionTitle)
	existing.IsCompanyOwner = profile.IsCompanyOwner || existing.IsCompanyOwner
	existing.CanCreateOpportunities = profile.CanCreateOpportunities || existing.CanCreateOpportunities
	existing.CanEditCompanyProfile = profile.CanEditCompanyProfile || existing.CanEditCompanyProfile
	existing.UpdatedAt = time.Now()
	if err := r.upsertEmployerProfileTx(context.Background(), nil, existing); err != nil {
		return nil, err
	}
	r.addAudit("update", "employer_profiles", userID, actorID, "employer profile updated")
	return r.GetEmployerProfile(userID)
}

func (r *Repository) ListModerationQueue() ([]ModerationQueueItem, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, entity_type, entity_id, submitted_by_user_id, COALESCE(assigned_to_user_id::text, ''), status, COALESCE(moderator_comment, ''), created_at, updated_at
FROM moderation_queue
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list moderation queue: %w", err)
	}
	defer rows.Close()
	result := []ModerationQueueItem{}
	for rows.Next() {
		var item ModerationQueueItem
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.SubmittedByUserID, &item.AssignedToUserID, &item.Status, &item.ModeratorComment, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan moderation queue: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ReviewModerationQueueItem(itemID, curatorID, status, comment string) (*ModerationQueueItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	now := time.Now()
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE moderation_queue
SET assigned_to_user_id = $2, status = $3, moderator_comment = NULLIF($4, ''), updated_at = $5
WHERE id = $1
`, itemID, curatorID, status, comment, now); err != nil {
		return nil, fmt.Errorf("review moderation queue item: %w", err)
	}
	var item ModerationQueueItem
	err := r.db.QueryRowContext(context.Background(), `
SELECT id, entity_type, entity_id, submitted_by_user_id, COALESCE(assigned_to_user_id::text, ''), status, COALESCE(moderator_comment, ''), created_at, updated_at
FROM moderation_queue
WHERE id = $1
`, itemID).Scan(&item.ID, &item.EntityType, &item.EntityID, &item.SubmittedByUserID, &item.AssignedToUserID, &item.Status, &item.ModeratorComment, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("moderation item not found")
		}
		return nil, fmt.Errorf("get moderation queue item: %w", err)
	}
	return &item, nil
}

func (r *Repository) ListCompanyVerifications() ([]CompanyVerification, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT cv.id, cv.company_id, COALESCE(c.brand_name, c.legal_name, ''), cv.verification_method, cv.submitted_by_user_id, COALESCE(cv.corporate_email, ''), COALESCE(cv.inn_submitted, ''), COALESCE(cv.documents_comment, ''), cv.status, COALESCE(cv.reviewed_by_user_id::text, ''), COALESCE(cv.review_comment, ''), cv.submitted_at, cv.reviewed_at
FROM company_verifications cv
LEFT JOIN companies c ON c.id = cv.company_id
ORDER BY submitted_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list company verifications: %w", err)
	}
	defer rows.Close()
	result := []CompanyVerification{}
	for rows.Next() {
		var item CompanyVerification
		var reviewedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CompanyID, &item.CompanyName, &item.VerificationMethod, &item.SubmittedByUserID, &item.CorporateEmail, &item.INNSubmitted, &item.DocumentsComment, &item.Status, &item.ReviewedByUserID, &item.ReviewComment, &item.SubmittedAt, &reviewedAt); err != nil {
			return nil, fmt.Errorf("scan company verifications: %w", err)
		}
		if reviewedAt.Valid {
			item.ReviewedAt = reviewedAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ReviewCompanyVerification(verificationID, curatorID, status, comment string) (*CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	var item CompanyVerification
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(context.Background(), `
SELECT cv.id, cv.company_id, COALESCE(c.brand_name, c.legal_name, ''), cv.verification_method, cv.submitted_by_user_id, COALESCE(cv.corporate_email, ''), COALESCE(cv.inn_submitted, ''), COALESCE(cv.documents_comment, ''), cv.status, COALESCE(cv.reviewed_by_user_id::text, ''), COALESCE(cv.review_comment, ''), cv.submitted_at, cv.reviewed_at
FROM company_verifications cv
LEFT JOIN companies c ON c.id = cv.company_id
WHERE cv.id = $1
`, verificationID).Scan(&item.ID, &item.CompanyID, &item.CompanyName, &item.VerificationMethod, &item.SubmittedByUserID, &item.CorporateEmail, &item.INNSubmitted, &item.DocumentsComment, &item.Status, &item.ReviewedByUserID, &item.ReviewComment, &item.SubmittedAt, &reviewedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("verification not found")
		}
		return nil, fmt.Errorf("get verification: %w", err)
	}
	item.Status = status
	item.ReviewedByUserID = curatorID
	item.ReviewComment = comment
	item.ReviewedAt = time.Now()
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE company_verifications
SET status = $2, reviewed_by_user_id = $3, review_comment = NULLIF($4, ''), reviewed_at = $5
WHERE id = $1
`, item.ID, item.Status, item.ReviewedByUserID, item.ReviewComment, item.ReviewedAt); err != nil {
		return nil, fmt.Errorf("review company verification: %w", err)
	}
	if status == "approved" {
		if _, err := r.db.ExecContext(context.Background(), `UPDATE companies SET status = 'verified', updated_at = $2 WHERE id = $1`, item.CompanyID, time.Now()); err != nil {
			return nil, fmt.Errorf("set company verified: %w", err)
		}
		if _, err := r.db.ExecContext(context.Background(), `UPDATE employer_profiles SET can_create_opportunities = TRUE, updated_at = $2 WHERE company_id = $1`, item.CompanyID, time.Now()); err != nil {
			return nil, fmt.Errorf("enable employer opportunities: %w", err)
		}
	} else if status == "rejected" {
		if _, err := r.db.ExecContext(context.Background(), `UPDATE companies SET status = 'rejected', updated_at = $2 WHERE id = $1`, item.CompanyID, time.Now()); err != nil {
			return nil, fmt.Errorf("set company rejected: %w", err)
		}
	}
	return &item, nil
}

func (r *Repository) UpdateOpportunityStatus(curatorID, opportunityID, status string) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	now := time.Now()
	publishedAt := any(nil)
	if status == "published" {
		publishedAt = now
	}
	if _, err := r.db.ExecContext(context.Background(), `
UPDATE opportunities
SET status = $2, published_at = COALESCE(published_at, $3), updated_at = $4
WHERE id = $1
`, opportunityID, status, publishedAt, now); err != nil {
		return nil, fmt.Errorf("update opportunity status: %w", err)
	}
	r.addAudit("status_change", "opportunities", opportunityID, curatorID, status)
	var opp Opportunity
	row := r.db.QueryRowContext(context.Background(), `
SELECT id, company_id, created_by_user_id, title, short_description, full_description, opportunity_type, work_format, COALESCE(location_id::text, ''), published_at, expires_at, status, COALESCE(contacts_info, ''), COALESCE(external_url, ''), views_count, favorites_count, applications_count, created_at, updated_at
FROM opportunities
WHERE id = $1
`, opportunityID)
	err := scanOpportunityBase(row, &opp)
	if err != nil {
		return nil, fmt.Errorf("get opportunity after update: %w", err)
	}
	if err := r.loadOpportunityTypeDetails(context.Background(), &opp); err != nil {
		return nil, err
	}
	return &opp, nil
}

func (r *Repository) ListAuditLogs() ([]AuditLog, error) {
	rows, err := r.db.QueryContext(context.Background(), `
SELECT id, COALESCE(actor_user_id::text, ''), entity_type, entity_id, action, created_at
FROM audit_logs
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()
	result := []AuditLog{}
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.EntityType, &item.EntityID, &item.Action, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit logs: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
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

func (r *Repository) addModerationItem(entityType, entityID, submittedBy string) {
	_, _ = r.db.ExecContext(context.Background(), `
INSERT INTO moderation_queue (id, entity_type, entity_id, submitted_by_user_id, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, 'pending', $5, $6)
`, uuid.NewString(), entityType, entityID, submittedBy, time.Now(), time.Now())
}

func (r *Repository) addAudit(action, entityType, entityID, actorID, details string) {
	_, _ = r.db.ExecContext(context.Background(), `
INSERT INTO audit_logs (id, actor_user_id, entity_type, entity_id, action, new_data_json, created_at)
VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, CASE WHEN NULLIF($6, '') IS NULL THEN NULL ELSE jsonb_build_object('details', $6) END, $7)
`, uuid.NewString(), actorID, entityType, entityID, action, details, time.Now())
}

func (r *Repository) notify(userID, typ, title, body, entityType, entityID string) {
	_, _ = r.db.ExecContext(context.Background(), `
INSERT INTO notifications (id, user_id, type, title, body, related_entity_type, related_entity_id, created_at)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, '')::uuid, $8)
`, uuid.NewString(), userID, typ, title, body, entityType, entityID, time.Now())
}

func (r *Repository) addContact(a, b string) {
	_, _ = r.db.ExecContext(context.Background(), `
INSERT INTO contacts (user_id, contact_user_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, contact_user_id) DO NOTHING
`, a, b, time.Now())
}

func (r *Repository) hasRole(userID, role string) bool {
	var exists bool
	_ = r.db.QueryRowContext(context.Background(), `
SELECT EXISTS (
	SELECT 1
	FROM user_roles ur
	JOIN roles r ON r.id = ur.role_id
	WHERE ur.user_id = $1 AND r.code = $2
)
`, userID, role).Scan(&exists)
	return exists
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

func (r *Repository) loadOpportunityTypeDetails(ctx context.Context, item *Opportunity) error {
	item.VacancyLevel = ""
	item.EmploymentType = ""
	item.SalaryMin = 0
	item.SalaryMax = 0
	item.SalaryCurrency = ""
	item.IsSalaryVisible = false
	item.ApplicationDeadline = time.Time{}
	item.EventStartAt = time.Time{}
	item.EventEndAt = time.Time{}

	switch item.OpportunityType {
	case "internship":
		return r.loadInternshipOpportunityDetails(ctx, item)
	case "vacancy":
		return r.loadVacancyOpportunityDetails(ctx, item)
	case "mentorship":
		return r.loadMentorshipOpportunityDetails(ctx, item)
	case "event":
		return r.loadEventOpportunityDetails(ctx, item)
	default:
		return fmt.Errorf("unsupported opportunity_type: %s", item.OpportunityType)
	}
}

func (r *Repository) loadInternshipOpportunityDetails(ctx context.Context, item *Opportunity) error {
	var applicationDeadline sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(vacancy_level, ''), COALESCE(employment_type, ''), COALESCE(salary_min, 0), COALESCE(salary_max, 0), COALESCE(salary_currency, ''), is_salary_visible, application_deadline
FROM internship_opportunities
WHERE opportunity_id = $1
`, item.ID).Scan(&item.VacancyLevel, &item.EmploymentType, &item.SalaryMin, &item.SalaryMax, &item.SalaryCurrency, &item.IsSalaryVisible, &applicationDeadline)
	if err != nil {
		return fmt.Errorf("load internship opportunity details: %w", err)
	}
	if applicationDeadline.Valid {
		item.ApplicationDeadline = applicationDeadline.Time
	}
	return nil
}

func (r *Repository) loadVacancyOpportunityDetails(ctx context.Context, item *Opportunity) error {
	var applicationDeadline sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT COALESCE(vacancy_level, ''), COALESCE(employment_type, ''), COALESCE(salary_min, 0), COALESCE(salary_max, 0), COALESCE(salary_currency, ''), is_salary_visible, application_deadline
FROM vacancy_opportunities
WHERE opportunity_id = $1
`, item.ID).Scan(&item.VacancyLevel, &item.EmploymentType, &item.SalaryMin, &item.SalaryMax, &item.SalaryCurrency, &item.IsSalaryVisible, &applicationDeadline)
	if err != nil {
		return fmt.Errorf("load vacancy opportunity details: %w", err)
	}
	if applicationDeadline.Valid {
		item.ApplicationDeadline = applicationDeadline.Time
	}
	return nil
}

func (r *Repository) loadMentorshipOpportunityDetails(ctx context.Context, item *Opportunity) error {
	var applicationDeadline sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT application_deadline
FROM mentorship_opportunities
WHERE opportunity_id = $1
`, item.ID).Scan(&applicationDeadline)
	if err != nil {
		return fmt.Errorf("load mentorship opportunity details: %w", err)
	}
	if applicationDeadline.Valid {
		item.ApplicationDeadline = applicationDeadline.Time
	}
	return nil
}

func (r *Repository) loadEventOpportunityDetails(ctx context.Context, item *Opportunity) error {
	var applicationDeadline sql.NullTime
	var eventStartAt sql.NullTime
	var eventEndAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT application_deadline, event_start_at, event_end_at
FROM event_opportunities
WHERE opportunity_id = $1
`, item.ID).Scan(&applicationDeadline, &eventStartAt, &eventEndAt)
	if err != nil {
		return fmt.Errorf("load event opportunity details: %w", err)
	}
	if applicationDeadline.Valid {
		item.ApplicationDeadline = applicationDeadline.Time
	}
	if eventStartAt.Valid {
		item.EventStartAt = eventStartAt.Time
	}
	if eventEndAt.Valid {
		item.EventEndAt = eventEndAt.Time
	}
	return nil
}

func (r *Repository) replaceOpportunityTypeDetails(ctx context.Context, tx *sql.Tx, opportunity *Opportunity) error {
	execTarget := sqlExecTarget(r.db, tx)
	for _, tableName := range []string{"internship_opportunities", "vacancy_opportunities", "mentorship_opportunities", "event_opportunities"} {
		if _, err := execTarget.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE opportunity_id = $1", tableName), opportunity.ID); err != nil {
			return fmt.Errorf("delete old opportunity details from %s: %w", tableName, err)
		}
	}

	switch opportunity.OpportunityType {
	case "internship":
		_, err := execTarget.ExecContext(ctx, `
INSERT INTO internship_opportunities (
	opportunity_id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
)
VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, 0), NULLIF($5, 0), NULLIF($6, ''), $7, $8)
`, opportunity.ID, opportunity.VacancyLevel, opportunity.EmploymentType, zeroableFloat(opportunity.SalaryMin), zeroableFloat(opportunity.SalaryMax), opportunity.SalaryCurrency, opportunity.IsSalaryVisible, nullableTime(opportunity.ApplicationDeadline))
		if err != nil {
			return fmt.Errorf("insert internship opportunity details: %w", err)
		}
	case "vacancy":
		_, err := execTarget.ExecContext(ctx, `
INSERT INTO vacancy_opportunities (
	opportunity_id, vacancy_level, employment_type, salary_min, salary_max, salary_currency, is_salary_visible, application_deadline
)
VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, 0), NULLIF($5, 0), NULLIF($6, ''), $7, $8)
`, opportunity.ID, opportunity.VacancyLevel, opportunity.EmploymentType, zeroableFloat(opportunity.SalaryMin), zeroableFloat(opportunity.SalaryMax), opportunity.SalaryCurrency, opportunity.IsSalaryVisible, nullableTime(opportunity.ApplicationDeadline))
		if err != nil {
			return fmt.Errorf("insert vacancy opportunity details: %w", err)
		}
	case "mentorship":
		_, err := execTarget.ExecContext(ctx, `
INSERT INTO mentorship_opportunities (
	opportunity_id, application_deadline
)
VALUES ($1, $2)
`, opportunity.ID, nullableTime(opportunity.ApplicationDeadline))
		if err != nil {
			return fmt.Errorf("insert mentorship opportunity details: %w", err)
		}
	case "event":
		_, err := execTarget.ExecContext(ctx, `
INSERT INTO event_opportunities (
	opportunity_id, application_deadline, event_start_at, event_end_at
)
VALUES ($1, $2, $3, $4)
`, opportunity.ID, nullableTime(opportunity.ApplicationDeadline), nullableTime(opportunity.EventStartAt), nullableTime(opportunity.EventEndAt))
		if err != nil {
			return fmt.Errorf("insert event opportunity details: %w", err)
		}
	default:
		return fmt.Errorf("opportunity_type must be one of: internship, vacancy, mentorship, event")
	}
	return nil
}

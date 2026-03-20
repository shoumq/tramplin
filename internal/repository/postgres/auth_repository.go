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

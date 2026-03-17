package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	. "tramplin/internal/models"
	"tramplin/internal/repository"
)

type Repository struct {
	mu sync.RWMutex
	db *sql.DB

	users                map[string]*User
	usersByEmail         map[string]string
	userRoles            map[string][]string
	studentProfiles      map[string]*StudentProfile
	employerProfiles     map[string]*EmployerProfile
	curatorProfiles      map[string]*CuratorProfile
	companies            map[string]*Company
	companyLinks         map[string][]CompanyLink
	companyVerifications map[string]*CompanyVerification
	cities               map[int64]*City
	locations            map[string]*Location
	tags                 map[string]*Tag
	opportunities        map[string]*Opportunity
	resumes              map[string]*Resume
	portfolioProjects    map[string]*PortfolioProject
	applications         map[string]*Application
	favoriteOpps         map[string]map[string]bool
	favoriteCompanies    map[string]map[string]bool
	contacts             map[string]map[string]bool
	contactRequests      map[string]*ContactRequest
	recommendations      map[string]*Recommendation
	notifications        map[string][]Notification
	moderationQueue      map[string]*ModerationQueueItem
	auditLogs            []AuditLog
}

type persistentState struct {
	Users                map[string]*User                `json:"users"`
	UsersByEmail         map[string]string               `json:"users_by_email"`
	UserRoles            map[string][]string             `json:"user_roles"`
	StudentProfiles      map[string]*StudentProfile      `json:"student_profiles"`
	EmployerProfiles     map[string]*EmployerProfile     `json:"employer_profiles"`
	CuratorProfiles      map[string]*CuratorProfile      `json:"curator_profiles"`
	Companies            map[string]*Company             `json:"companies"`
	CompanyLinks         map[string][]CompanyLink        `json:"company_links"`
	CompanyVerifications map[string]*CompanyVerification `json:"company_verifications"`
	Cities               map[int64]*City                 `json:"cities"`
	Locations            map[string]*Location            `json:"locations"`
	Tags                 map[string]*Tag                 `json:"tags"`
	Opportunities        map[string]*Opportunity         `json:"opportunities"`
	Resumes              map[string]*Resume              `json:"resumes"`
	PortfolioProjects    map[string]*PortfolioProject    `json:"portfolio_projects"`
	Applications         map[string]*Application         `json:"applications"`
	FavoriteOpps         map[string]map[string]bool      `json:"favorite_opps"`
	FavoriteCompanies    map[string]map[string]bool      `json:"favorite_companies"`
	Contacts             map[string]map[string]bool      `json:"contacts"`
	ContactRequests      map[string]*ContactRequest      `json:"contact_requests"`
	Recommendations      map[string]*Recommendation      `json:"recommendations"`
	Notifications        map[string][]Notification       `json:"notifications"`
	ModerationQueue      map[string]*ModerationQueueItem `json:"moderation_queue"`
	AuditLogs            []AuditLog                      `json:"audit_logs"`
}

func NewRepository(ctx context.Context, dsn string) (*Repository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	r := &Repository{
		db:                   db,
		users:                map[string]*User{},
		usersByEmail:         map[string]string{},
		userRoles:            map[string][]string{},
		studentProfiles:      map[string]*StudentProfile{},
		employerProfiles:     map[string]*EmployerProfile{},
		curatorProfiles:      map[string]*CuratorProfile{},
		companies:            map[string]*Company{},
		companyLinks:         map[string][]CompanyLink{},
		companyVerifications: map[string]*CompanyVerification{},
		cities:               map[int64]*City{},
		locations:            map[string]*Location{},
		tags:                 map[string]*Tag{},
		opportunities:        map[string]*Opportunity{},
		resumes:              map[string]*Resume{},
		portfolioProjects:    map[string]*PortfolioProject{},
		applications:         map[string]*Application{},
		favoriteOpps:         map[string]map[string]bool{},
		favoriteCompanies:    map[string]map[string]bool{},
		contacts:             map[string]map[string]bool{},
		contactRequests:      map[string]*ContactRequest{},
		recommendations:      map[string]*Recommendation{},
		notifications:        map[string][]Notification{},
		moderationQueue:      map[string]*ModerationQueueItem{},
		auditLogs:            []AuditLog{},
	}
	if err := r.ensureStateTable(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := r.loadOrSeed(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *Repository) ensureStateTable(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS repository_state (
	id SMALLINT PRIMARY KEY,
	payload JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("ensure repository_state table: %w", err)
	}
	return nil
}

func (r *Repository) loadOrSeed(ctx context.Context) error {
	var payload []byte
	err := r.db.QueryRowContext(ctx, `SELECT payload FROM repository_state WHERE id = 1`).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		r.seed()
		return r.saveState(ctx)
	}
	if err != nil {
		return fmt.Errorf("load repository state: %w", err)
	}

	var state persistentState
	if err := json.Unmarshal(payload, &state); err != nil {
		return fmt.Errorf("decode repository state: %w", err)
	}
	r.restoreState(state)
	return nil
}

func (r *Repository) saveState(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.saveStateSnapshot(ctx, r.snapshotLocked())
}

func (r *Repository) saveStateLocked() error {
	return r.saveStateSnapshot(context.Background(), r.snapshotLocked())
}

func (r *Repository) saveStateSnapshot(ctx context.Context, state persistentState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal repository state: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
INSERT INTO repository_state (id, payload, updated_at)
VALUES (1, $1, NOW())
ON CONFLICT (id)
DO UPDATE SET payload = EXCLUDED.payload, updated_at = EXCLUDED.updated_at
`, payload)
	if err != nil {
		return fmt.Errorf("persist repository state: %w", err)
	}
	return nil
}

func (r *Repository) snapshotLocked() persistentState {
	return persistentState{
		Users:                r.users,
		UsersByEmail:         r.usersByEmail,
		UserRoles:            r.userRoles,
		StudentProfiles:      r.studentProfiles,
		EmployerProfiles:     r.employerProfiles,
		CuratorProfiles:      r.curatorProfiles,
		Companies:            r.companies,
		CompanyLinks:         r.companyLinks,
		CompanyVerifications: r.companyVerifications,
		Cities:               r.cities,
		Locations:            r.locations,
		Tags:                 r.tags,
		Opportunities:        r.opportunities,
		Resumes:              r.resumes,
		PortfolioProjects:    r.portfolioProjects,
		Applications:         r.applications,
		FavoriteOpps:         r.favoriteOpps,
		FavoriteCompanies:    r.favoriteCompanies,
		Contacts:             r.contacts,
		ContactRequests:      r.contactRequests,
		Recommendations:      r.recommendations,
		Notifications:        r.notifications,
		ModerationQueue:      r.moderationQueue,
		AuditLogs:            r.auditLogs,
	}
}

func (r *Repository) restoreState(state persistentState) {
	r.users = state.Users
	r.usersByEmail = state.UsersByEmail
	r.userRoles = state.UserRoles
	r.studentProfiles = state.StudentProfiles
	r.employerProfiles = state.EmployerProfiles
	r.curatorProfiles = state.CuratorProfiles
	r.companies = state.Companies
	r.companyLinks = state.CompanyLinks
	r.companyVerifications = state.CompanyVerifications
	r.cities = state.Cities
	r.locations = state.Locations
	r.tags = state.Tags
	r.opportunities = state.Opportunities
	r.resumes = state.Resumes
	r.portfolioProjects = state.PortfolioProjects
	r.applications = state.Applications
	r.favoriteOpps = state.FavoriteOpps
	r.favoriteCompanies = state.FavoriteCompanies
	r.contacts = state.Contacts
	r.contactRequests = state.ContactRequests
	r.recommendations = state.Recommendations
	r.notifications = state.Notifications
	r.moderationQueue = state.ModerationQueue
	r.auditLogs = state.AuditLogs
}

func (r *Repository) persistLocked() {
	if err := r.saveStateLocked(); err != nil {
		log.Printf("persist repository state: %v", err)
	}
}

func (r *Repository) seed() {
	now := time.Now()

	r.cities[1] = &City{ID: 1, Country: "Russia", Region: "Moscow", CityName: "Moscow", Latitude: 55.7558, Longitude: 37.6176}
	r.cities[2] = &City{ID: 2, Country: "Russia", Region: "Saint Petersburg", CityName: "Saint Petersburg", Latitude: 59.9343, Longitude: 30.3351}

	tagBackend := Tag{ID: uuid.NewString(), Name: "Go", TagType: "technology", IsSystem: true, IsActive: true, CreatedAt: now}
	tagSQL := Tag{ID: uuid.NewString(), Name: "SQL", TagType: "technology", IsSystem: true, IsActive: true, CreatedAt: now}
	tagJunior := Tag{ID: uuid.NewString(), Name: "Junior", TagType: "level", IsSystem: true, IsActive: true, CreatedAt: now}
	for _, tag := range []Tag{tagBackend, tagSQL, tagJunior} {
		t := tag
		r.tags[t.ID] = &t
	}

	adminID := "00000000-0000-0000-0000-000000000001"
	r.users[adminID] = &User{ID: adminID, Email: "admin@tramplin.local", PasswordHash: hashPassword("admin123"), DisplayName: "System Administrator", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("admin@tramplin.local")] = adminID
	r.userRoles[adminID] = []string{"curator", "admin"}
	r.curatorProfiles[adminID] = &CuratorProfile{UserID: adminID, CuratorType: "administrator", CreatedAt: now, UpdatedAt: now}

	studentID := uuid.NewString()
	r.users[studentID] = &User{ID: studentID, Email: "student@tramplin.local", PasswordHash: hashPassword("student123"), DisplayName: "Ivan Student", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("student@tramplin.local")] = studentID
	r.userRoles[studentID] = []string{"student"}
	r.studentProfiles[studentID] = &StudentProfile{UserID: studentID, FirstName: "Ivan", LastName: "Ivanov", UniversityName: "BMSTU", StudyYear: 3, GraduationYear: now.Year() + 1, ProfileVisibility: "authorized_only", ShowResume: true, ShowApplications: false, ShowCareerInterests: true, CityID: 1, CreatedAt: now, UpdatedAt: now}

	companyID := uuid.NewString()
	r.companies[companyID] = &Company{ID: companyID, LegalName: "Tramplin Tech LLC", BrandName: "Tramplin Tech", Description: "Platform company for internships and jobs.", Industry: "IT", WebsiteURL: "https://tramplin.local", CompanySize: "51-200", FoundedYear: 2020, HQCityID: 1, Status: "verified", CreatedAt: now, UpdatedAt: now}

	employerID := uuid.NewString()
	r.users[employerID] = &User{ID: employerID, Email: "employer@tramplin.local", PasswordHash: hashPassword("employer123"), DisplayName: "Anna Employer", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("employer@tramplin.local")] = employerID
	r.userRoles[employerID] = []string{"employer"}
	r.employerProfiles[employerID] = &EmployerProfile{UserID: employerID, CompanyID: companyID, PositionTitle: "HR Lead", IsCompanyOwner: true, CanCreateOpportunities: true, CanEditCompanyProfile: true, CreatedAt: now, UpdatedAt: now}

	locationID := uuid.NewString()
	r.locations[locationID] = &Location{ID: locationID, CityID: 1, AddressLine: "Red Square 1", Latitude: 55.7539, Longitude: 37.6208, LocationType: "office", DisplayText: "Moscow, Red Square 1", CreatedAt: now}

	oppID := uuid.NewString()
	r.opportunities[oppID] = &Opportunity{ID: oppID, CompanyID: companyID, CreatedByUserID: employerID, Title: "Go Internship", ShortDescription: "Internship for backend students", FullDescription: "Work on Go services, PostgreSQL, and Fiber.", OpportunityType: "internship", VacancyLevel: "intern", EmploymentType: "part_time", WorkFormat: "hybrid", LocationID: locationID, SalaryMin: 40000, SalaryMax: 70000, SalaryCurrency: "RUB", IsSalaryVisible: true, PublishedAt: now, ExpiresAt: now.AddDate(0, 1, 0), Status: "published", ContactsInfo: "hr@tramplin.local", CreatedAt: now, UpdatedAt: now, TagIDs: []string{tagBackend.ID, tagSQL.ID, tagJunior.ID}}

	resumeID := uuid.NewString()
	r.resumes[resumeID] = &Resume{ID: resumeID, StudentUserID: studentID, Title: "Backend Intern Resume", Summary: "Go backend student", ExperienceText: "Pet projects and SQL practice", EducationText: "BMSTU", IsPrimary: true, CreatedAt: now, UpdatedAt: now}

	r.companyLinks[companyID] = []CompanyLink{{ID: uuid.NewString(), CompanyID: companyID, LinkType: "website", URL: "https://tramplin.local", CreatedAt: now}}
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) RegisterUser(params repository.RegisterUserParams) (*User, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()

	emailKey := strings.ToLower(strings.TrimSpace(params.Email))
	if emailKey == "" || strings.TrimSpace(params.Password) == "" || strings.TrimSpace(params.DisplayName) == "" {
		return nil, "", errors.New("email, password and display_name are required")
	}
	if _, exists := r.usersByEmail[emailKey]; exists {
		return nil, "", errors.New("user with this email already exists")
	}
	if params.Role != repository.RoleStudent && params.Role != repository.RoleEmployer {
		return nil, "", fmt.Errorf("role must be one of: %s, %s", repository.RoleStudent, repository.RoleEmployer)
	}

	now := time.Now()
	user := &User{ID: uuid.NewString(), Email: emailKey, PasswordHash: hashPassword(params.Password), DisplayName: params.DisplayName, Status: "active", EmailVerified: false, CreatedAt: now, UpdatedAt: now}
	r.users[user.ID] = user
	r.usersByEmail[emailKey] = user.ID
	r.userRoles[user.ID] = []string{params.Role}

	if params.Role == repository.RoleStudent {
		r.studentProfiles[user.ID] = &StudentProfile{UserID: user.ID, FirstName: params.DisplayName, UniversityName: "", ProfileVisibility: "authorized_only", ShowResume: true, ShowApplications: false, ShowCareerInterests: true, CreatedAt: now, UpdatedAt: now}
	}
	if params.Role == repository.RoleEmployer {
		companyName := strings.TrimSpace(params.CompanyName)
		if companyName == "" {
			companyName = params.DisplayName + " Company"
		}
		company := &Company{ID: uuid.NewString(), LegalName: companyName, BrandName: companyName, Status: "pending_verification", CreatedAt: now, UpdatedAt: now}
		r.companies[company.ID] = company
		r.employerProfiles[user.ID] = &EmployerProfile{UserID: user.ID, CompanyID: company.ID, IsCompanyOwner: true, CanCreateOpportunities: false, CanEditCompanyProfile: true, CreatedAt: now, UpdatedAt: now}
		r.addModerationItem("company", company.ID, user.ID)
	}

	r.addAudit("create", "users", user.ID, user.ID, "user registered")
	return cloneUser(user), params.Role, nil
}

func (r *Repository) Login(email, password string) (*User, []string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()

	id, ok := r.usersByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return nil, nil, errors.New("invalid credentials")
	}
	user := r.users[id]
	if user.PasswordHash != hashPassword(password) {
		return nil, nil, errors.New("invalid credentials")
	}
	user.LastLoginAt = time.Now()
	user.UpdatedAt = time.Now()
	roles := append([]string(nil), r.userRoles[user.ID]...)
	return cloneUser(user), roles, nil
}

func (r *Repository) GetUser(userID string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return cloneUser(user), nil
}

func (r *Repository) GetUserRoles(userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	roles, ok := r.userRoles[userID]
	if !ok {
		return nil, errors.New("roles not found")
	}
	return append([]string(nil), roles...), nil
}

func (r *Repository) CreateCurator(email, password, displayName, curatorType, createdBy string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if !r.hasRole(createdBy, "admin") {
		return nil, errors.New("only administrator can create curator accounts")
	}
	if curatorType == "" {
		curatorType = "moderator"
	}
	key := strings.ToLower(strings.TrimSpace(email))
	if _, exists := r.usersByEmail[key]; exists {
		return nil, errors.New("user with this email already exists")
	}
	now := time.Now()
	user := &User{ID: uuid.NewString(), Email: key, PasswordHash: hashPassword(password), DisplayName: displayName, Status: "active", EmailVerified: true, CreatedAt: now, UpdatedAt: now}
	r.users[user.ID] = user
	r.usersByEmail[key] = user.ID
	r.userRoles[user.ID] = []string{"curator"}
	if curatorType == "administrator" {
		r.userRoles[user.ID] = append(r.userRoles[user.ID], "admin")
	}
	r.curatorProfiles[user.ID] = &CuratorProfile{UserID: user.ID, CuratorType: curatorType, CreatedByUserID: createdBy, CreatedAt: now, UpdatedAt: now}
	r.addAudit("create", "curator_profiles", user.ID, createdBy, "curator created")
	return cloneUser(user), nil
}

func (r *Repository) UpdateUserStatus(userID, status, actorID string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	user.Status = status
	user.UpdatedAt = time.Now()
	r.addAudit("status_change", "users", user.ID, actorID, status)
	return cloneUser(user), nil
}

func (r *Repository) UpdateUserAvatar(userID, avatarObject, avatarURL string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()

	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	user.AvatarObject = avatarObject
	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()
	r.addAudit("update", "users", userID, userID, "avatar updated")
	return cloneUser(user), nil
}

func (r *Repository) GetStudentProfile(userID string) (*StudentProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.studentProfiles[userID]
	if !ok {
		return nil, errors.New("student profile not found")
	}
	cp := *profile
	if user, ok := r.users[userID]; ok {
		cp.AvatarURL = user.AvatarURL
	}
	return &cp, nil
}

func (r *Repository) UpsertStudentProfile(profile StudentProfile, actorID string) (*StudentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if _, ok := r.users[profile.UserID]; !ok {
		return nil, errors.New("user not found")
	}
	now := time.Now()
	existing, ok := r.studentProfiles[profile.UserID]
	if ok {
		profile.CreatedAt = existing.CreatedAt
	} else {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	if profile.ProfileVisibility == "" {
		profile.ProfileVisibility = "authorized_only"
	}
	cp := profile
	if user, ok := r.users[profile.UserID]; ok {
		cp.AvatarURL = user.AvatarURL
	}
	r.studentProfiles[profile.UserID] = &cp
	r.addAudit("update", "student_profiles", profile.UserID, actorID, "student profile upsert")
	return &cp, nil
}

func (r *Repository) ListResumes(studentUserID string) ([]Resume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Resume{}
	for _, resume := range r.resumes {
		if resume.StudentUserID == studentUserID {
			result = append(result, *resume)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) CreateResume(resume Resume) (*Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if _, ok := r.studentProfiles[resume.StudentUserID]; !ok {
		return nil, errors.New("student profile not found")
	}
	now := time.Now()
	resume.ID = uuid.NewString()
	resume.CreatedAt = now
	resume.UpdatedAt = now
	if len(r.resumesByStudent(resume.StudentUserID)) == 0 {
		resume.IsPrimary = true
	}
	cp := resume
	r.resumes[resume.ID] = &cp
	r.addAudit("create", "resumes", resume.ID, resume.StudentUserID, resume.Title)
	return &cp, nil
}

func (r *Repository) SetPrimaryResume(studentUserID, resumeID string) (*Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	resume, ok := r.resumes[resumeID]
	if !ok || resume.StudentUserID != studentUserID {
		return nil, errors.New("resume not found")
	}
	for _, item := range r.resumes {
		if item.StudentUserID == studentUserID {
			item.IsPrimary = item.ID == resumeID
			item.UpdatedAt = time.Now()
		}
	}
	cp := *r.resumes[resumeID]
	return &cp, nil
}

func (r *Repository) ListPortfolioProjects(studentUserID string) ([]PortfolioProject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []PortfolioProject{}
	for _, item := range r.portfolioProjects {
		if item.StudentUserID == studentUserID {
			result = append(result, *item)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) CreatePortfolioProject(project PortfolioProject) (*PortfolioProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	project.ID = uuid.NewString()
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	cp := project
	r.portfolioProjects[project.ID] = &cp
	r.addAudit("create", "portfolio_projects", project.ID, project.StudentUserID, project.Title)
	return &cp, nil
}

func (r *Repository) ListStudentApplications(studentUserID string) ([]Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Application{}
	for _, app := range r.applications {
		if app.StudentUserID == studentUserID {
			cp := *app
			if user, ok := r.users[app.StudentUserID]; ok {
				cp.StudentAvatarURL = user.AvatarURL
			}
			result = append(result, cp)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListFavoriteOpportunities(userID string) ([]PublicOpportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []PublicOpportunity{}
	for opportunityID := range r.favoriteOpps[userID] {
		if opportunity, err := r.publicOpportunityLocked(opportunityID); err == nil {
			result = append(result, *opportunity)
		}
	}
	return result, nil
}

func (r *Repository) AddFavoriteOpportunity(userID, opportunityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	opp, ok := r.opportunities[opportunityID]
	if !ok {
		return errors.New("opportunity not found")
	}
	if r.favoriteOpps[userID] == nil {
		r.favoriteOpps[userID] = map[string]bool{}
	}
	if !r.favoriteOpps[userID][opportunityID] {
		r.favoriteOpps[userID][opportunityID] = true
		opp.FavoritesCount++
	}
	return nil
}

func (r *Repository) RemoveFavoriteOpportunity(userID, opportunityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if r.favoriteOpps[userID] != nil && r.favoriteOpps[userID][opportunityID] {
		delete(r.favoriteOpps[userID], opportunityID)
		if opp, ok := r.opportunities[opportunityID]; ok && opp.FavoritesCount > 0 {
			opp.FavoritesCount--
		}
	}
	return nil
}

func (r *Repository) ListFavoriteCompanies(userID string) ([]Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Company{}
	for companyID := range r.favoriteCompanies[userID] {
		if company, ok := r.companies[companyID]; ok {
			result = append(result, *company)
		}
	}
	return result, nil
}

func (r *Repository) AddFavoriteCompany(userID, companyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if _, ok := r.companies[companyID]; !ok {
		return errors.New("company not found")
	}
	if r.favoriteCompanies[userID] == nil {
		r.favoriteCompanies[userID] = map[string]bool{}
	}
	r.favoriteCompanies[userID][companyID] = true
	return nil
}

func (r *Repository) RemoveFavoriteCompany(userID, companyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	if r.favoriteCompanies[userID] != nil {
		delete(r.favoriteCompanies[userID], companyID)
	}
	return nil
}

func (r *Repository) ListContacts(userID string) ([]User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []User{}
	for contactID := range r.contacts[userID] {
		if user, ok := r.users[contactID]; ok {
			result = append(result, *user)
		}
	}
	return result, nil
}

func (r *Repository) ListContactRequests(userID string) ([]ContactRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []ContactRequest{}
	for _, request := range r.contactRequests {
		if request.SenderUserID == userID || request.ReceiverUserID == userID {
			cp := *request
			if user, ok := r.users[request.SenderUserID]; ok {
				cp.SenderAvatarURL = user.AvatarURL
			}
			if user, ok := r.users[request.ReceiverUserID]; ok {
				cp.ReceiverAvatarURL = user.AvatarURL
			}
			result = append(result, cp)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
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
	r.contactRequests[item.ID] = item
	r.notify(receiverUserID, "contact_request_received", "New contact request", "You have a new professional contact request.", "contact_request", item.ID)
	return r.cloneContactRequest(item), nil
}

func (r *Repository) UpdateContactRequestStatus(requestID, userID, status string) (*ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	item, ok := r.contactRequests[requestID]
	if !ok {
		return nil, errors.New("contact request not found")
	}
	if item.ReceiverUserID != userID && item.SenderUserID != userID {
		return nil, errors.New("forbidden")
	}
	item.Status = status
	item.UpdatedAt = time.Now()
	if status == "accepted" {
		r.addContact(item.SenderUserID, item.ReceiverUserID)
		r.addContact(item.ReceiverUserID, item.SenderUserID)
		r.notify(item.SenderUserID, "contact_request_accepted", "Contact request accepted", "Your contact request has been accepted.", "contact_request", item.ID)
	}
	return r.cloneContactRequest(item), nil
}

func (r *Repository) CreateRecommendation(rec Recommendation) (*Recommendation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	rec.ID = uuid.NewString()
	rec.CreatedAt = time.Now()
	cp := rec
	r.recommendations[rec.ID] = &cp
	r.notify(rec.ToUserID, "recommendation_received", "New recommendation", "Another student recommended an opportunity to you.", "opportunity", rec.OpportunityID)
	return &cp, nil
}

func (r *Repository) ListNotifications(userID string) ([]Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := append([]Notification(nil), r.notifications[userID]...)
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListOpportunities(filter repository.OpportunityFilter) ([]PublicOpportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []PublicOpportunity{}
	for _, opp := range r.opportunities {
		if opp.Status != "published" && opp.Status != "scheduled" {
			continue
		}
		public, err := r.publicOpportunityLocked(opp.ID)
		if err != nil {
			continue
		}
		if matchesFilter(*public, filter) {
			result = append(result, *public)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListOpportunityMarkers(filter repository.OpportunityFilter) ([]OpportunityMarker, error) {
	targets, _ := r.ListOpportunities(filter)
	markers := make([]OpportunityMarker, 0, len(targets))
	for _, item := range targets {
		opp := item.Opportunity
		location := r.locations[opp.LocationID]
		if location == nil {
			continue
		}
		markers = append(markers, OpportunityMarker{ID: opp.ID, Title: opp.Title, CompanyName: item.CompanyName, Latitude: location.Latitude, Longitude: location.Longitude, WorkFormat: opp.WorkFormat, OpportunityType: opp.OpportunityType, SalaryLabel: salaryLabel(opp)})
	}
	return markers, nil
}

func (r *Repository) GetOpportunity(id string) (*PublicOpportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	opp, ok := r.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	opp.ViewsCount++
	return r.publicOpportunityLocked(id)
}

func (r *Repository) CreateApplication(application Application) (*Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	for _, item := range r.applications {
		if item.OpportunityID == application.OpportunityID && item.StudentUserID == application.StudentUserID {
			return nil, errors.New("application already exists")
		}
	}
	opp, ok := r.opportunities[application.OpportunityID]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	application.ID = uuid.NewString()
	application.Status = "submitted"
	application.CreatedAt = time.Now()
	application.UpdatedAt = application.CreatedAt
	cp := application
	if user, ok := r.users[application.StudentUserID]; ok {
		cp.StudentAvatarURL = user.AvatarURL
	}
	r.applications[cp.ID] = &cp
	opp.ApplicationsCount++
	r.notify(opp.CreatedByUserID, "application_submitted", "New application", "A student applied to your opportunity.", "application", cp.ID)
	return &cp, nil
}

func (r *Repository) ListCompanies() ([]Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Company{}
	for _, company := range r.companies {
		result = append(result, *company)
	}
	return result, nil
}

func (r *Repository) GetCompany(id string) (*Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	company, ok := r.companies[id]
	if !ok {
		return nil, errors.New("company not found")
	}
	cp := *company
	return &cp, nil
}

func (r *Repository) ListTags() ([]Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Tag{}
	for _, tag := range r.tags {
		result = append(result, *tag)
	}
	return result, nil
}

func (r *Repository) ListCities() ([]City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []City{}
	for _, city := range r.cities {
		result = append(result, *city)
	}
	return result, nil
}

func (r *Repository) ListLocations() ([]Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []Location{}
	for _, location := range r.locations {
		result = append(result, *location)
	}
	return result, nil
}

func (r *Repository) GetEmployerProfile(userID string) (*EmployerProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	cp := *profile
	if user, ok := r.users[userID]; ok {
		cp.AvatarURL = user.AvatarURL
	}
	return &cp, nil
}

func (r *Repository) GetEmployerCompany(userID string) (*Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	company, ok := r.companies[profile.CompanyID]
	if !ok {
		return nil, errors.New("company not found")
	}
	cp := *company
	return &cp, nil
}

func (r *Repository) UpdateEmployerCompany(userID string, update repository.CompanyUpdate) (*Company, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	company := r.companies[profile.CompanyID]
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
	cp := *company
	r.addAudit("update", "companies", company.ID, userID, "company updated")
	return &cp, nil
}

func (r *Repository) CreateCompanyLink(userID, linkType, url string) (*CompanyLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	link := CompanyLink{ID: uuid.NewString(), CompanyID: profile.CompanyID, LinkType: linkType, URL: url, CreatedAt: time.Now()}
	r.companyLinks[profile.CompanyID] = append(r.companyLinks[profile.CompanyID], link)
	return &link, nil
}

func (r *Repository) SubmitCompanyVerification(userID, method, corporateEmail, inn, comment string) (*CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	item := &CompanyVerification{ID: uuid.NewString(), CompanyID: profile.CompanyID, VerificationMethod: method, SubmittedByUserID: userID, CorporateEmail: corporateEmail, INNSubmitted: inn, DocumentsComment: comment, Status: "pending", SubmittedAt: time.Now()}
	r.companyVerifications[item.ID] = item
	r.addModerationItem("company", profile.CompanyID, userID)
	return cloneVerification(item), nil
}

func (r *Repository) ListEmployerOpportunities(userID string) ([]Opportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	result := []Opportunity{}
	for _, item := range r.opportunities {
		if item.CompanyID == profile.CompanyID {
			result = append(result, r.opportunityWithLocationLocked(item))
		}
	}
	return result, nil
}

func (r *Repository) CreateOpportunity(opportunity Opportunity) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, ok := r.employerProfiles[opportunity.CreatedByUserID]
	if !ok {
		return nil, errors.New("employer profile not found")
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
	cp := opportunity
	r.opportunities[cp.ID] = &cp
	r.addModerationItem("opportunity", cp.ID, opportunity.CreatedByUserID)
	r.addAudit("create", "opportunities", cp.ID, opportunity.CreatedByUserID, cp.Title)
	result := r.opportunityWithLocationLocked(&cp)
	return &result, nil
}

func (r *Repository) GetEmployerOpportunity(userID, opportunityID string) (*Opportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	opp, ok := r.opportunities[opportunityID]
	if !ok || opp.CompanyID != profile.CompanyID {
		return nil, errors.New("opportunity not found")
	}
	cp := r.opportunityWithLocationLocked(opp)
	return &cp, nil
}

func (r *Repository) UpdateEmployerOpportunity(userID string, opportunity Opportunity) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	existing, ok := r.opportunities[opportunity.ID]
	if !ok || existing.CompanyID != profile.CompanyID {
		return nil, errors.New("opportunity not found")
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
	cp := r.opportunityWithLocationLocked(existing)
	return &cp, nil
}

func (r *Repository) ListOpportunityApplications(userID, opportunityID string) ([]Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	opp, ok := r.opportunities[opportunityID]
	if !ok || opp.CompanyID != profile.CompanyID {
		return nil, errors.New("opportunity not found")
	}
	result := []Application{}
	for _, item := range r.applications {
		if item.OpportunityID == opportunityID {
			cp := *item
			if user, ok := r.users[item.StudentUserID]; ok {
				cp.StudentAvatarURL = user.AvatarURL
			}
			result = append(result, cp)
		}
	}
	return result, nil
}

func (r *Repository) UpdateApplicationStatus(userID, applicationID, status string) (*Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	application, ok := r.applications[applicationID]
	if !ok {
		return nil, errors.New("application not found")
	}
	opp, ok := r.opportunities[application.OpportunityID]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	profile, ok := r.employerProfiles[userID]
	if !ok || opp.CompanyID != profile.CompanyID {
		return nil, errors.New("forbidden")
	}
	application.Status = status
	application.StatusChangedByUserID = userID
	application.StatusChangedAt = time.Now()
	application.UpdatedAt = time.Now()
	cp := *application
	if user, ok := r.users[application.StudentUserID]; ok {
		cp.StudentAvatarURL = user.AvatarURL
	}
	r.notify(application.StudentUserID, "application_status_changed", "Application status changed", fmt.Sprintf("Your application status is now: %s", status), "application", application.ID)
	return &cp, nil
}

func (r *Repository) UpdateEmployerProfile(userID string, profile EmployerProfile, actorID string) (*EmployerProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	existing, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	existing.PositionTitle = fallback(profile.PositionTitle, existing.PositionTitle)
	existing.IsCompanyOwner = profile.IsCompanyOwner || existing.IsCompanyOwner
	existing.CanCreateOpportunities = profile.CanCreateOpportunities || existing.CanCreateOpportunities
	existing.CanEditCompanyProfile = profile.CanEditCompanyProfile || existing.CanEditCompanyProfile
	existing.UpdatedAt = time.Now()
	cp := *existing
	if user, ok := r.users[userID]; ok {
		cp.AvatarURL = user.AvatarURL
	}
	r.addAudit("update", "employer_profiles", userID, actorID, "employer profile updated")
	return &cp, nil
}

func (r *Repository) ListModerationQueue() ([]ModerationQueueItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []ModerationQueueItem{}
	for _, item := range r.moderationQueue {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ReviewModerationQueueItem(itemID, curatorID, status, comment string) (*ModerationQueueItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	item, ok := r.moderationQueue[itemID]
	if !ok {
		return nil, errors.New("moderation item not found")
	}
	item.AssignedToUserID = curatorID
	item.Status = status
	item.ModeratorComment = comment
	item.UpdatedAt = time.Now()
	cp := *item
	return &cp, nil
}

func (r *Repository) ListCompanyVerifications() ([]CompanyVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []CompanyVerification{}
	for _, item := range r.companyVerifications {
		result = append(result, *item)
	}
	return result, nil
}

func (r *Repository) ReviewCompanyVerification(verificationID, curatorID, status, comment string) (*CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	item, ok := r.companyVerifications[verificationID]
	if !ok {
		return nil, errors.New("verification not found")
	}
	item.Status = status
	item.ReviewedByUserID = curatorID
	item.ReviewComment = comment
	item.ReviewedAt = time.Now()
	company := r.companies[item.CompanyID]
	if status == "approved" {
		company.Status = "verified"
		for _, profile := range r.employerProfiles {
			if profile.CompanyID == item.CompanyID {
				profile.CanCreateOpportunities = true
			}
		}
	} else if status == "rejected" {
		company.Status = "rejected"
	}
	cp := *item
	return &cp, nil
}

func (r *Repository) UpdateOpportunityStatus(curatorID, opportunityID, status string) (*Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.persistLocked()
	opp, ok := r.opportunities[opportunityID]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	opp.Status = status
	if status == "published" && opp.PublishedAt.IsZero() {
		opp.PublishedAt = time.Now()
	}
	opp.UpdatedAt = time.Now()
	r.addAudit("status_change", "opportunities", opportunityID, curatorID, status)
	cp := r.opportunityWithLocationLocked(opp)
	return &cp, nil
}

func (r *Repository) ListAuditLogs() ([]AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := append([]AuditLog(nil), r.auditLogs...)
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) resumesByStudent(studentUserID string) []Resume {
	result := []Resume{}
	for _, resume := range r.resumes {
		if resume.StudentUserID == studentUserID {
			result = append(result, *resume)
		}
	}
	return result
}

func (r *Repository) publicOpportunityLocked(id string) (*PublicOpportunity, error) {
	opp, ok := r.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	companyName := ""
	if company, ok := r.companies[opp.CompanyID]; ok {
		companyName = firstNonEmpty(company.BrandName, company.LegalName)
	}
	locationLabel := ""
	if location, ok := r.locations[opp.LocationID]; ok {
		locationLabel = location.DisplayText
	}
	tags := []string{}
	for _, tagID := range opp.TagIDs {
		if tag, ok := r.tags[tagID]; ok {
			tags = append(tags, tag.Name)
		}
	}
	enriched := r.opportunityWithLocationLocked(opp)
	return &PublicOpportunity{Opportunity: enriched, CompanyName: companyName, Location: locationLabel, Tags: tags}, nil
}

func (r *Repository) opportunityWithLocationLocked(opp *Opportunity) Opportunity {
	cp := *opp
	if location, ok := r.locations[opp.LocationID]; ok {
		cp.Latitude = location.Latitude
		cp.Longitude = location.Longitude
	}
	return cp
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
	id := uuid.NewString()
	r.moderationQueue[id] = &ModerationQueueItem{ID: id, EntityType: entityType, EntityID: entityID, SubmittedByUserID: submittedBy, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (r *Repository) addAudit(action, entityType, entityID, actorID, details string) {
	r.auditLogs = append(r.auditLogs, AuditLog{ID: uuid.NewString(), ActorUserID: actorID, EntityType: entityType, EntityID: entityID, Action: action, CreatedAt: time.Now(), Details: details})
}

func (r *Repository) notify(userID, typ, title, body, entityType, entityID string) {
	r.notifications[userID] = append(r.notifications[userID], Notification{ID: uuid.NewString(), UserID: userID, Type: typ, Title: title, Body: body, RelatedEntityType: entityType, RelatedEntityID: entityID, CreatedAt: time.Now()})
}

func (r *Repository) addContact(a, b string) {
	if r.contacts[a] == nil {
		r.contacts[a] = map[string]bool{}
	}
	r.contacts[a][b] = true
}

func (r *Repository) hasRole(userID, role string) bool {
	for _, current := range r.userRoles[userID] {
		if current == role {
			return true
		}
	}
	return false
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

func (r *Repository) cloneContactRequest(item *ContactRequest) *ContactRequest {
	cp := *item
	if sender, ok := r.users[item.SenderUserID]; ok {
		cp.SenderAvatarURL = sender.AvatarURL
	}
	if receiver, ok := r.users[item.ReceiverUserID]; ok {
		cp.ReceiverAvatarURL = receiver.AvatarURL
	}
	return &cp
}

func cloneVerification(item *CompanyVerification) *CompanyVerification {
	cp := *item
	return &cp
}

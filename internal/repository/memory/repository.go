package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"tramplin/internal/domain"
	"tramplin/internal/repository"
)

type Repository struct {
	mu sync.RWMutex

	users                map[string]*domain.User
	usersByEmail         map[string]string
	userRoles            map[string][]string
	studentProfiles      map[string]*domain.StudentProfile
	employerProfiles     map[string]*domain.EmployerProfile
	curatorProfiles      map[string]*domain.CuratorProfile
	companies            map[string]*domain.Company
	companyLinks         map[string][]domain.CompanyLink
	companyVerifications map[string]*domain.CompanyVerification
	cities               map[int64]*domain.City
	locations            map[string]*domain.Location
	tags                 map[string]*domain.Tag
	opportunities        map[string]*domain.Opportunity
	resumes              map[string]*domain.Resume
	portfolioProjects    map[string]*domain.PortfolioProject
	applications         map[string]*domain.Application
	favoriteOpps         map[string]map[string]bool
	favoriteCompanies    map[string]map[string]bool
	contacts             map[string]map[string]bool
	contactRequests      map[string]*domain.ContactRequest
	recommendations      map[string]*domain.Recommendation
	notifications        map[string][]domain.Notification
	moderationQueue      map[string]*domain.ModerationQueueItem
	auditLogs            []domain.AuditLog
}

func NewRepository() *Repository {
	r := &Repository{
		users:                map[string]*domain.User{},
		usersByEmail:         map[string]string{},
		userRoles:            map[string][]string{},
		studentProfiles:      map[string]*domain.StudentProfile{},
		employerProfiles:     map[string]*domain.EmployerProfile{},
		curatorProfiles:      map[string]*domain.CuratorProfile{},
		companies:            map[string]*domain.Company{},
		companyLinks:         map[string][]domain.CompanyLink{},
		companyVerifications: map[string]*domain.CompanyVerification{},
		cities:               map[int64]*domain.City{},
		locations:            map[string]*domain.Location{},
		tags:                 map[string]*domain.Tag{},
		opportunities:        map[string]*domain.Opportunity{},
		resumes:              map[string]*domain.Resume{},
		portfolioProjects:    map[string]*domain.PortfolioProject{},
		applications:         map[string]*domain.Application{},
		favoriteOpps:         map[string]map[string]bool{},
		favoriteCompanies:    map[string]map[string]bool{},
		contacts:             map[string]map[string]bool{},
		contactRequests:      map[string]*domain.ContactRequest{},
		recommendations:      map[string]*domain.Recommendation{},
		notifications:        map[string][]domain.Notification{},
		moderationQueue:      map[string]*domain.ModerationQueueItem{},
		auditLogs:            []domain.AuditLog{},
	}
	r.seed()
	return r
}

func (r *Repository) seed() {
	now := time.Now()

	r.cities[1] = &domain.City{ID: 1, Country: "Russia", Region: "Moscow", CityName: "Moscow", Latitude: 55.7558, Longitude: 37.6176}
	r.cities[2] = &domain.City{ID: 2, Country: "Russia", Region: "Saint Petersburg", CityName: "Saint Petersburg", Latitude: 59.9343, Longitude: 30.3351}

	tagBackend := domain.Tag{ID: uuid.NewString(), Name: "Go", TagType: "technology", IsSystem: true, IsActive: true, CreatedAt: now}
	tagSQL := domain.Tag{ID: uuid.NewString(), Name: "SQL", TagType: "technology", IsSystem: true, IsActive: true, CreatedAt: now}
	tagJunior := domain.Tag{ID: uuid.NewString(), Name: "Junior", TagType: "level", IsSystem: true, IsActive: true, CreatedAt: now}
	for _, tag := range []domain.Tag{tagBackend, tagSQL, tagJunior} {
		t := tag
		r.tags[t.ID] = &t
	}

	adminID := "00000000-0000-0000-0000-000000000001"
	r.users[adminID] = &domain.User{ID: adminID, Email: "admin@tramplin.local", PasswordHash: hashPassword("admin123"), DisplayName: "System Administrator", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("admin@tramplin.local")] = adminID
	r.userRoles[adminID] = []string{"curator", "admin"}
	r.curatorProfiles[adminID] = &domain.CuratorProfile{UserID: adminID, CuratorType: "administrator", CreatedAt: now, UpdatedAt: now}

	studentID := uuid.NewString()
	r.users[studentID] = &domain.User{ID: studentID, Email: "student@tramplin.local", PasswordHash: hashPassword("student123"), DisplayName: "Ivan Student", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("student@tramplin.local")] = studentID
	r.userRoles[studentID] = []string{"student"}
	r.studentProfiles[studentID] = &domain.StudentProfile{UserID: studentID, FirstName: "Ivan", LastName: "Ivanov", UniversityName: "BMSTU", StudyYear: 3, GraduationYear: now.Year() + 1, ProfileVisibility: "authorized_only", ShowResume: true, ShowApplications: false, ShowCareerInterests: true, CityID: 1, CreatedAt: now, UpdatedAt: now}

	companyID := uuid.NewString()
	r.companies[companyID] = &domain.Company{ID: companyID, LegalName: "Tramplin Tech LLC", BrandName: "Tramplin Tech", Description: "Platform company for internships and jobs.", Industry: "IT", WebsiteURL: "https://tramplin.local", CompanySize: "51-200", FoundedYear: 2020, HQCityID: 1, Status: "verified", CreatedAt: now, UpdatedAt: now}

	employerID := uuid.NewString()
	r.users[employerID] = &domain.User{ID: employerID, Email: "employer@tramplin.local", PasswordHash: hashPassword("employer123"), DisplayName: "Anna Employer", EmailVerified: true, Status: "active", CreatedAt: now, UpdatedAt: now}
	r.usersByEmail[strings.ToLower("employer@tramplin.local")] = employerID
	r.userRoles[employerID] = []string{"employer"}
	r.employerProfiles[employerID] = &domain.EmployerProfile{UserID: employerID, CompanyID: companyID, PositionTitle: "HR Lead", IsCompanyOwner: true, CanCreateOpportunities: true, CanEditCompanyProfile: true, CreatedAt: now, UpdatedAt: now}

	locationID := uuid.NewString()
	r.locations[locationID] = &domain.Location{ID: locationID, CityID: 1, AddressLine: "Red Square 1", Latitude: 55.7539, Longitude: 37.6208, LocationType: "office", DisplayText: "Moscow, Red Square 1", CreatedAt: now}

	oppID := uuid.NewString()
	r.opportunities[oppID] = &domain.Opportunity{ID: oppID, CompanyID: companyID, CreatedByUserID: employerID, Title: "Go Internship", ShortDescription: "Internship for backend students", FullDescription: "Work on Go services, PostgreSQL, and Fiber.", OpportunityType: "internship", VacancyLevel: "intern", EmploymentType: "part_time", WorkFormat: "hybrid", LocationID: locationID, SalaryMin: 40000, SalaryMax: 70000, SalaryCurrency: "RUB", IsSalaryVisible: true, PublishedAt: now, ExpiresAt: now.AddDate(0, 1, 0), Status: "published", ContactsInfo: "hr@tramplin.local", CreatedAt: now, UpdatedAt: now, TagIDs: []string{tagBackend.ID, tagSQL.ID, tagJunior.ID}}

	resumeID := uuid.NewString()
	r.resumes[resumeID] = &domain.Resume{ID: resumeID, StudentUserID: studentID, Title: "Backend Intern Resume", Summary: "Go backend student", ExperienceText: "Pet projects and SQL practice", EducationText: "BMSTU", IsPrimary: true, CreatedAt: now, UpdatedAt: now}

	r.companyLinks[companyID] = []domain.CompanyLink{{ID: uuid.NewString(), CompanyID: companyID, LinkType: "website", URL: "https://tramplin.local", CreatedAt: now}}
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) RegisterUser(params repository.RegisterUserParams) (*domain.User, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailKey := strings.ToLower(strings.TrimSpace(params.Email))
	if emailKey == "" || strings.TrimSpace(params.Password) == "" || strings.TrimSpace(params.DisplayName) == "" {
		return nil, "", errors.New("email, password and display_name are required")
	}
	if _, exists := r.usersByEmail[emailKey]; exists {
		return nil, "", errors.New("user with this email already exists")
	}
	if params.Role != "student" && params.Role != "employer" {
		return nil, "", errors.New("role must be student or employer")
	}

	now := time.Now()
	user := &domain.User{ID: uuid.NewString(), Email: emailKey, PasswordHash: hashPassword(params.Password), DisplayName: params.DisplayName, Status: "active", EmailVerified: false, CreatedAt: now, UpdatedAt: now}
	r.users[user.ID] = user
	r.usersByEmail[emailKey] = user.ID
	r.userRoles[user.ID] = []string{params.Role}

	if params.Role == "student" {
		r.studentProfiles[user.ID] = &domain.StudentProfile{UserID: user.ID, FirstName: params.DisplayName, UniversityName: "", ProfileVisibility: "authorized_only", ShowResume: true, ShowApplications: false, ShowCareerInterests: true, CreatedAt: now, UpdatedAt: now}
	}
	if params.Role == "employer" {
		companyName := strings.TrimSpace(params.CompanyName)
		if companyName == "" {
			companyName = params.DisplayName + " Company"
		}
		company := &domain.Company{ID: uuid.NewString(), LegalName: companyName, BrandName: companyName, Status: "pending_verification", CreatedAt: now, UpdatedAt: now}
		r.companies[company.ID] = company
		r.employerProfiles[user.ID] = &domain.EmployerProfile{UserID: user.ID, CompanyID: company.ID, IsCompanyOwner: true, CanCreateOpportunities: false, CanEditCompanyProfile: true, CreatedAt: now, UpdatedAt: now}
		r.addModerationItem("company", company.ID, user.ID)
	}

	r.addAudit("create", "users", user.ID, user.ID, "user registered")
	return cloneUser(user), params.Role, nil
}

func (r *Repository) Login(email, password string) (*domain.User, []string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *Repository) GetUser(userID string) (*domain.User, error) {
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

func (r *Repository) CreateCurator(email, password, displayName, curatorType, createdBy string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	user := &domain.User{ID: uuid.NewString(), Email: key, PasswordHash: hashPassword(password), DisplayName: displayName, Status: "active", EmailVerified: true, CreatedAt: now, UpdatedAt: now}
	r.users[user.ID] = user
	r.usersByEmail[key] = user.ID
	r.userRoles[user.ID] = []string{"curator"}
	if curatorType == "administrator" {
		r.userRoles[user.ID] = append(r.userRoles[user.ID], "admin")
	}
	r.curatorProfiles[user.ID] = &domain.CuratorProfile{UserID: user.ID, CuratorType: curatorType, CreatedByUserID: createdBy, CreatedAt: now, UpdatedAt: now}
	r.addAudit("create", "curator_profiles", user.ID, createdBy, "curator created")
	return cloneUser(user), nil
}

func (r *Repository) UpdateUserStatus(userID, status, actorID string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	user.Status = status
	user.UpdatedAt = time.Now()
	r.addAudit("status_change", "users", user.ID, actorID, status)
	return cloneUser(user), nil
}

func (r *Repository) GetStudentProfile(userID string) (*domain.StudentProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.studentProfiles[userID]
	if !ok {
		return nil, errors.New("student profile not found")
	}
	cp := *profile
	return &cp, nil
}

func (r *Repository) UpsertStudentProfile(profile domain.StudentProfile, actorID string) (*domain.StudentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.studentProfiles[profile.UserID] = &cp
	r.addAudit("update", "student_profiles", profile.UserID, actorID, "student profile upsert")
	return &cp, nil
}

func (r *Repository) ListResumes(studentUserID string) ([]domain.Resume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Resume{}
	for _, resume := range r.resumes {
		if resume.StudentUserID == studentUserID {
			result = append(result, *resume)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) CreateResume(resume domain.Resume) (*domain.Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *Repository) SetPrimaryResume(studentUserID, resumeID string) (*domain.Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *Repository) ListPortfolioProjects(studentUserID string) ([]domain.PortfolioProject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.PortfolioProject{}
	for _, item := range r.portfolioProjects {
		if item.StudentUserID == studentUserID {
			result = append(result, *item)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) CreatePortfolioProject(project domain.PortfolioProject) (*domain.PortfolioProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	project.ID = uuid.NewString()
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	cp := project
	r.portfolioProjects[project.ID] = &cp
	r.addAudit("create", "portfolio_projects", project.ID, project.StudentUserID, project.Title)
	return &cp, nil
}

func (r *Repository) ListStudentApplications(studentUserID string) ([]domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Application{}
	for _, app := range r.applications {
		if app.StudentUserID == studentUserID {
			result = append(result, *app)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListFavoriteOpportunities(userID string) ([]domain.PublicOpportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.PublicOpportunity{}
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
	if r.favoriteOpps[userID] != nil && r.favoriteOpps[userID][opportunityID] {
		delete(r.favoriteOpps[userID], opportunityID)
		if opp, ok := r.opportunities[opportunityID]; ok && opp.FavoritesCount > 0 {
			opp.FavoritesCount--
		}
	}
	return nil
}

func (r *Repository) ListFavoriteCompanies(userID string) ([]domain.Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Company{}
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
	if r.favoriteCompanies[userID] != nil {
		delete(r.favoriteCompanies[userID], companyID)
	}
	return nil
}

func (r *Repository) ListContacts(userID string) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.User{}
	for contactID := range r.contacts[userID] {
		if user, ok := r.users[contactID]; ok {
			result = append(result, *user)
		}
	}
	return result, nil
}

func (r *Repository) ListContactRequests(userID string) ([]domain.ContactRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.ContactRequest{}
	for _, request := range r.contactRequests {
		if request.SenderUserID == userID || request.ReceiverUserID == userID {
			result = append(result, *request)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) CreateContactRequest(senderUserID, receiverUserID, message string) (*domain.ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if senderUserID == receiverUserID {
		return nil, errors.New("cannot create contact request to yourself")
	}
	item := &domain.ContactRequest{ID: uuid.NewString(), SenderUserID: senderUserID, ReceiverUserID: receiverUserID, Message: message, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.contactRequests[item.ID] = item
	r.notify(receiverUserID, "contact_request_received", "New contact request", "You have a new professional contact request.", "contact_request", item.ID)
	return cloneContactRequest(item), nil
}

func (r *Repository) UpdateContactRequestStatus(requestID, userID, status string) (*domain.ContactRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	return cloneContactRequest(item), nil
}

func (r *Repository) CreateRecommendation(rec domain.Recommendation) (*domain.Recommendation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec.ID = uuid.NewString()
	rec.CreatedAt = time.Now()
	cp := rec
	r.recommendations[rec.ID] = &cp
	r.notify(rec.ToUserID, "recommendation_received", "New recommendation", "Another student recommended an opportunity to you.", "opportunity", rec.OpportunityID)
	return &cp, nil
}

func (r *Repository) ListNotifications(userID string) ([]domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := append([]domain.Notification(nil), r.notifications[userID]...)
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ListOpportunities(filter repository.OpportunityFilter) ([]domain.PublicOpportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.PublicOpportunity{}
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

func (r *Repository) ListOpportunityMarkers(filter repository.OpportunityFilter) ([]domain.OpportunityMarker, error) {
	targets, _ := r.ListOpportunities(filter)
	markers := make([]domain.OpportunityMarker, 0, len(targets))
	for _, item := range targets {
		opp := item.Opportunity
		location := r.locations[opp.LocationID]
		if location == nil {
			continue
		}
		markers = append(markers, domain.OpportunityMarker{ID: opp.ID, Title: opp.Title, CompanyName: item.CompanyName, Latitude: location.Latitude, Longitude: location.Longitude, WorkFormat: opp.WorkFormat, OpportunityType: opp.OpportunityType, SalaryLabel: salaryLabel(opp)})
	}
	return markers, nil
}

func (r *Repository) GetOpportunity(id string) (*domain.PublicOpportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	opp, ok := r.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	opp.ViewsCount++
	return r.publicOpportunityLocked(id)
}

func (r *Repository) CreateApplication(application domain.Application) (*domain.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.applications[cp.ID] = &cp
	opp.ApplicationsCount++
	r.notify(opp.CreatedByUserID, "application_submitted", "New application", "A student applied to your opportunity.", "application", cp.ID)
	return &cp, nil
}

func (r *Repository) ListCompanies() ([]domain.Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Company{}
	for _, company := range r.companies {
		result = append(result, *company)
	}
	return result, nil
}

func (r *Repository) GetCompany(id string) (*domain.Company, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	company, ok := r.companies[id]
	if !ok {
		return nil, errors.New("company not found")
	}
	cp := *company
	return &cp, nil
}

func (r *Repository) ListTags() ([]domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Tag{}
	for _, tag := range r.tags {
		result = append(result, *tag)
	}
	return result, nil
}

func (r *Repository) ListCities() ([]domain.City, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.City{}
	for _, city := range r.cities {
		result = append(result, *city)
	}
	return result, nil
}

func (r *Repository) ListLocations() ([]domain.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.Location{}
	for _, location := range r.locations {
		result = append(result, *location)
	}
	return result, nil
}

func (r *Repository) GetEmployerProfile(userID string) (*domain.EmployerProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	cp := *profile
	return &cp, nil
}

func (r *Repository) GetEmployerCompany(userID string) (*domain.Company, error) {
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

func (r *Repository) UpdateEmployerCompany(userID string, update repository.CompanyUpdate) (*domain.Company, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *Repository) CreateCompanyLink(userID, linkType, url string) (*domain.CompanyLink, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	link := domain.CompanyLink{ID: uuid.NewString(), CompanyID: profile.CompanyID, LinkType: linkType, URL: url, CreatedAt: time.Now()}
	r.companyLinks[profile.CompanyID] = append(r.companyLinks[profile.CompanyID], link)
	return &link, nil
}

func (r *Repository) SubmitCompanyVerification(userID, method, corporateEmail, inn, comment string) (*domain.CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	item := &domain.CompanyVerification{ID: uuid.NewString(), CompanyID: profile.CompanyID, VerificationMethod: method, SubmittedByUserID: userID, CorporateEmail: corporateEmail, INNSubmitted: inn, DocumentsComment: comment, Status: "pending", SubmittedAt: time.Now()}
	r.companyVerifications[item.ID] = item
	r.addModerationItem("company", profile.CompanyID, userID)
	return cloneVerification(item), nil
}

func (r *Repository) ListEmployerOpportunities(userID string) ([]domain.Opportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	profile, ok := r.employerProfiles[userID]
	if !ok {
		return nil, errors.New("employer profile not found")
	}
	result := []domain.Opportunity{}
	for _, item := range r.opportunities {
		if item.CompanyID == profile.CompanyID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (r *Repository) CreateOpportunity(opportunity domain.Opportunity) (*domain.Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	return &cp, nil
}

func (r *Repository) GetEmployerOpportunity(userID, opportunityID string) (*domain.Opportunity, error) {
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
	cp := *opp
	return &cp, nil
}

func (r *Repository) UpdateEmployerOpportunity(userID string, opportunity domain.Opportunity) (*domain.Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	cp := *existing
	return &cp, nil
}

func (r *Repository) ListOpportunityApplications(userID, opportunityID string) ([]domain.Application, error) {
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
	result := []domain.Application{}
	for _, item := range r.applications {
		if item.OpportunityID == opportunityID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (r *Repository) UpdateApplicationStatus(userID, applicationID, status string) (*domain.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.notify(application.StudentUserID, "application_status_changed", "Application status changed", fmt.Sprintf("Your application status is now: %s", status), "application", application.ID)
	return &cp, nil
}

func (r *Repository) UpdateEmployerProfile(userID string, profile domain.EmployerProfile, actorID string) (*domain.EmployerProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.addAudit("update", "employer_profiles", userID, actorID, "employer profile updated")
	return &cp, nil
}

func (r *Repository) ListModerationQueue() ([]domain.ModerationQueueItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.ModerationQueueItem{}
	for _, item := range r.moderationQueue {
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) ReviewModerationQueueItem(itemID, curatorID, status, comment string) (*domain.ModerationQueueItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *Repository) ListCompanyVerifications() ([]domain.CompanyVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := []domain.CompanyVerification{}
	for _, item := range r.companyVerifications {
		result = append(result, *item)
	}
	return result, nil
}

func (r *Repository) ReviewCompanyVerification(verificationID, curatorID, status, comment string) (*domain.CompanyVerification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func (r *Repository) UpdateOpportunityStatus(curatorID, opportunityID, status string) (*domain.Opportunity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	cp := *opp
	return &cp, nil
}

func (r *Repository) ListAuditLogs() ([]domain.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := append([]domain.AuditLog(nil), r.auditLogs...)
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}

func (r *Repository) resumesByStudent(studentUserID string) []domain.Resume {
	result := []domain.Resume{}
	for _, resume := range r.resumes {
		if resume.StudentUserID == studentUserID {
			result = append(result, *resume)
		}
	}
	return result
}

func (r *Repository) publicOpportunityLocked(id string) (*domain.PublicOpportunity, error) {
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
	return &domain.PublicOpportunity{Opportunity: *opp, CompanyName: companyName, Location: locationLabel, Tags: tags}, nil
}

func matchesFilter(item domain.PublicOpportunity, filter repository.OpportunityFilter) bool {
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

func salaryLabel(opp domain.Opportunity) string {
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
	r.moderationQueue[id] = &domain.ModerationQueueItem{ID: id, EntityType: entityType, EntityID: entityID, SubmittedByUserID: submittedBy, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func (r *Repository) addAudit(action, entityType, entityID, actorID, details string) {
	r.auditLogs = append(r.auditLogs, domain.AuditLog{ID: uuid.NewString(), ActorUserID: actorID, EntityType: entityType, EntityID: entityID, Action: action, CreatedAt: time.Now(), Details: details})
}

func (r *Repository) notify(userID, typ, title, body, entityType, entityID string) {
	r.notifications[userID] = append(r.notifications[userID], domain.Notification{ID: uuid.NewString(), UserID: userID, Type: typ, Title: title, Body: body, RelatedEntityType: entityType, RelatedEntityID: entityID, CreatedAt: time.Now()})
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

func cloneUser(user *domain.User) *domain.User {
	cp := *user
	return &cp
}

func cloneContactRequest(item *domain.ContactRequest) *domain.ContactRequest {
	cp := *item
	return &cp
}

func cloneVerification(item *domain.CompanyVerification) *domain.CompanyVerification {
	cp := *item
	return &cp
}

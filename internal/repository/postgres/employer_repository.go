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

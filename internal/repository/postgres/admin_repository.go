package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	. "tramplin/internal/models"
)

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

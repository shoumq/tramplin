package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	. "tramplin/internal/models"
)

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

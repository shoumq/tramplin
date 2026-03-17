package dto

// RegisterInput is used for self-registration.
// Available roles: student, employer.
// Curator and admin accounts are created separately.
type RegisterInput struct {
	Email       string `json:"email" example:"student@tramplin.local"`
	Password    string `json:"password" example:"secret123"`
	DisplayName string `json:"display_name" example:"Ivan Petrov"`
	Role        string `json:"role" enums:"student,employer" example:"student"`
	CompanyName string `json:"company_name,omitempty" example:"Tramplin Tech"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CuratorCreateInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	CuratorType string `json:"curator_type"`
}

package dto

type StatusPayload struct {
	Status  string `json:"status" example:"approved"`
	Comment string `json:"comment,omitempty" example:"Looks good"`
}

package handlers

type SuccessResponse struct {
	Status string `json:"status" example:"ok"`
	Data   any    `json:"data"`
}

type ErrorResponse struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"bad request"`
}

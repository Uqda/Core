package admin

// ErrorResponse is the JSON representation of an admin request failure.
type ErrorResponse struct {
	Error string `json:"error"`
}

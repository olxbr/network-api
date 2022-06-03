package types

type ErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

func NewSingleErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Errors: map[string]string{
			"_all": message,
		},
	}
}

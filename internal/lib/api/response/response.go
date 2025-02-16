package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ValidationError(errs validator.ValidationErrors) ErrorResponse {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field % s is not valid", err.Field()))
		}
	}

	return ErrorResponse{
		Error: strings.Join(errMsgs, ", "),
	}
}

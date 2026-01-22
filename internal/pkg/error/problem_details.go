package error

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
)

// ProblemDetailsFieldError represents a field validation error
type ProblemDetailsFieldError struct {
	Field   string `json:"field" example:"username"`
	Message string `json:"message" example:"validation failed for 'required' tag"`
}

// NewProblemDetailsFieldError creates a new ProblemDetailsFieldError with the specified field and message.
func NewProblemDetailsFieldError(field, message string) ProblemDetailsFieldError {
	return ProblemDetailsFieldError{
		Field:   field,
		Message: message,
	}
}

// NewProblemDetailsFromStructValidation converts validator.ValidationErrors to ProblemDetailsFieldError slice.
func NewProblemDetailsFromStructValidation(ve validator.ValidationErrors) []ProblemDetailsFieldError {
	var fieldErrors []ProblemDetailsFieldError
	for _, fieldError := range ve {
		fieldErrors = append(fieldErrors, NewProblemDetailsFieldError(fieldError.StructField(), fmt.Sprintf("validation failed for '%s' tag", fieldError.Tag())))
	}
	return fieldErrors
}

// ProblemDetails represents an RFC 7807 problem details response
type ProblemDetails struct {
	Type        string                     `json:"type,omitempty" example:"urn:auth-session-api/healthcheck/check"`
	Title       string                     `json:"title,omitempty" example:"Health check server failed"`
	Status      int                        `json:"status,omitempty" example:"500"`
	Detail      string                     `json:"detail,omitempty" example:"An error occurred while performing the health check"`
	Instance    string                     `json:"instance,omitempty" example:"/health"`
	FieldErrors []ProblemDetailsFieldError `json:"errors,omitempty"`
	Limit       int                        `json:"limit,omitempty" example:"10"`
	Code        int                        `json:"code,omitempty" example:"1001"`
}

// NewProblemDetails creates a new ProblemDetails with default type "about:blank".
func NewProblemDetails() ProblemDetails {
	return ProblemDetails{
		Type: "about:blank",
	}
}

// WithType sets the type field using a formatted URN pattern.
func (p ProblemDetails) WithType(typeContext, t string) ProblemDetails {
	p.Type = fmt.Sprintf("urn:auth-session-api/%s/%s", typeContext, t)
	return p
}

// WithTitle sets the title field of the ProblemDetails.
func (p ProblemDetails) WithTitle(title string) ProblemDetails {
	p.Title = title
	return p
}

// WithStatus sets the HTTP status code of the ProblemDetails.
func (p ProblemDetails) WithStatus(status int) ProblemDetails {
	p.Status = status
	return p
}

// WithDetail sets the detail field of the ProblemDetails.
func (p ProblemDetails) WithDetail(detail string) ProblemDetails {
	p.Detail = detail
	return p
}

// WithInstance sets the instance field of the ProblemDetails.
func (p ProblemDetails) WithInstance(instance string) ProblemDetails {
	p.Instance = instance
	return p
}

// AddFieldErrors appends multiple field errors to the ProblemDetails.
func (p ProblemDetails) AddFieldErrors(errs []ProblemDetailsFieldError) ProblemDetails {
	p.FieldErrors = append(p.FieldErrors, errs...)
	return p
}

// AddFieldError appends a single field error to the ProblemDetails.
func (p ProblemDetails) AddFieldError(err ProblemDetailsFieldError) ProblemDetails {
	p.FieldErrors = append(p.FieldErrors, err)
	return p
}

// WithLimit sets the limit field of the ProblemDetails.
func (p ProblemDetails) WithLimit(limit int) ProblemDetails {
	p.Limit = limit
	return p
}

// WithCode sets the code field of the ProblemDetails.
func (p ProblemDetails) WithCode(code int) ProblemDetails {
	p.Code = code
	return p
}

func (p ProblemDetails) Error() string {
	if p.Title == "" && p.Detail == "" {
		return fmt.Sprintf("status %d", p.Status)
	}
	if p.Detail == "" {
		return fmt.Sprint(p.Title)
	}
	return fmt.Sprintf("%s: %s", p.Title, p.Detail)
}

// ServeJSON writes the ProblemDetails as JSON to the HTTP response writer.
func (p ProblemDetails) ServeJSON(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		return err
	}
	return nil
}

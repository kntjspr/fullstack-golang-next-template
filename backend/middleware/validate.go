package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const maxValidateBodyBytes = 1 << 20 // 1 MiB

type validatedBodyContextKey struct{}

type validationFieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type validationErrorResponse struct {
	Error  string                 `json:"error"`
	Fields []validationFieldError `json:"fields"`
}

var validateBodyContextKey = validatedBodyContextKey{}

// ValidateBody decodes JSON request bodies into T and applies validation constraints.
func ValidateBody[T any](schema T) func(http.Handler) http.Handler {
	validate := validator.New()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxValidateBodyBytes)

			payload := schema
			decoder := json.NewDecoder(r.Body)

			if err := decoder.Decode(&payload); err != nil {
				var maxBytesErr *http.MaxBytesError
				switch {
				case errors.As(err, &maxBytesErr):
					writeValidationErrors(w, http.StatusRequestEntityTooLarge, []validationFieldError{{
						Field:   "body",
						Message: "request body exceeds 1MB limit",
					}})
				case errors.Is(err, io.EOF):
					writeValidationErrors(w, http.StatusUnprocessableEntity, []validationFieldError{{
						Field:   "body",
						Message: "body is required",
					}})
				default:
					writeValidationErrors(w, http.StatusUnprocessableEntity, []validationFieldError{{
						Field:   "body",
						Message: "malformed JSON",
					}})
				}
				return
			}

			if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
				writeValidationErrors(w, http.StatusUnprocessableEntity, []validationFieldError{{
					Field:   "body",
					Message: "body must contain a single JSON value",
				}})
				return
			}

			if err := validate.Struct(payload); err != nil {
				var validationErrors validator.ValidationErrors
				if errors.As(err, &validationErrors) {
					writeValidationErrors(w, http.StatusUnprocessableEntity, formatValidationErrors(payload, validationErrors))
					return
				}

				writeValidationErrors(w, http.StatusUnprocessableEntity, []validationFieldError{{
					Field:   "body",
					Message: "validation failed",
				}})
				return
			}

			ctx := context.WithValue(r.Context(), validateBodyContextKey, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetValidatedBody returns previously validated request payload from context.
func GetValidatedBody[T any](ctx context.Context) T {
	var zero T

	payload, ok := ctx.Value(validateBodyContextKey).(T)
	if !ok {
		return zero
	}

	return payload
}

func formatValidationErrors[T any](schema T, errs validator.ValidationErrors) []validationFieldError {
	result := make([]validationFieldError, 0, len(errs))
	for _, err := range errs {
		result = append(result, validationFieldError{
			Field:   jsonFieldName(schema, err.StructField()),
			Message: validationMessage(err),
		})
	}

	return result
}

func jsonFieldName[T any](schema T, structField string) string {
	t := reflect.TypeOf(schema)
	if t == nil {
		return strings.ToLower(structField)
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}

	field, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(structField)
	}

	parts := strings.Split(jsonTag, ",")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "-" {
		return strings.ToLower(structField)
	}

	return parts[0]
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	default:
		return "is invalid"
	}
}

func writeValidationErrors(w http.ResponseWriter, status int, fields []validationFieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(validationErrorResponse{
		Error:  "validation failed",
		Fields: fields,
	})
}

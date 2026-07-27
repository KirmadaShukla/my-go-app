package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

const validatedBodyKey contextKey = "validated_body"

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	_ = v.RegisterValidation("mobile", validateMobile)
	_ = v.RegisterValidation("gender", validateGender)
	_ = v.RegisterValidation("subject", validateSubject)
	_ = v.RegisterValidation("child_class", validateChildClass)
	return v
}

func validateMobile(fl validator.FieldLevel) bool {
	mobile := strings.TrimSpace(fl.Field().String())
	if mobile == "" {
		return false
	}
	digits := 0
	for _, r := range mobile {
		if unicode.IsDigit(r) {
			digits++
			continue
		}
		if r == '+' || r == '-' || r == ' ' || r == '(' || r == ')' {
			continue
		}
		return false
	}
	return digits >= 10 && digits <= 15
}

func validateGender(fl validator.FieldLevel) bool {
	switch strings.ToLower(strings.TrimSpace(fl.Field().String())) {
	case "male", "female", "other":
		return true
	default:
		return false
	}
}

func validateSubject(fl validator.FieldLevel) bool {
	switch strings.ToLower(strings.TrimSpace(fl.Field().String())) {
	case "maths", "science", "english", "activities":
		return true
	default:
		return false
	}
}

func validateChildClass(fl validator.FieldLevel) bool {
	raw := strings.ToLower(strings.TrimSpace(fl.Field().String()))
	raw = strings.ReplaceAll(raw, "class", "")
	raw = strings.ReplaceAll(raw, "std", "")
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "stndrh")
	raw = strings.TrimSpace(raw)
	switch raw {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10":
		return true
	default:
		return false
	}
}

func ValidateJSON[T any](next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body T
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&body); err != nil {
			writeValidationError(w, "invalid JSON body", nil)
			return
		}

		if err := validate.Struct(body); err != nil {
			details := validationDetails(err)
			writeValidationError(w, "validation failed", details)
			return
		}

		ctx := context.WithValue(r.Context(), validatedBodyKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// BodyFromContext returns the validated request body set by ValidateJSON.
func BodyFromContext[T any](ctx context.Context) (T, bool) {
	body, ok := ctx.Value(validatedBodyKey).(T)
	return body, ok
}

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func writeValidationError(w http.ResponseWriter, message string, details []fieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	payload := map[string]any{"error": message}
	if len(details) > 0 {
		payload["details"] = details
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func validationDetails(err error) []fieldError {
	var details []fieldError
	ves, ok := err.(validator.ValidationErrors)
	if !ok {
		return []fieldError{{Field: "body", Message: err.Error()}}
	}
	for _, fe := range ves {
		details = append(details, fieldError{
			Field:   fe.Field(),
			Message: fieldMessage(fe),
		})
	}
	return details
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "mobile":
		return fmt.Sprintf("%s must be a valid mobile number (10-15 digits)", fe.Field())
	case "gender":
		return fmt.Sprintf("%s must be one of: male, female, other", fe.Field())
	case "subject":
		return fmt.Sprintf("%s must be one of: maths, science, english, activities", fe.Field())
	case "child_class":
		return fmt.Sprintf("%s must be a class from 1 to 10", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

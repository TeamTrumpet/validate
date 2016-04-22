package validate

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gopkg.in/go-playground/validator.v8"
)

//==============================================================================

// validate is used to perform model field validation.
var (
	validate *validator.Validate

	validCoordinateRegex = regexp.MustCompile(`^(\-?\d+)(\.\d+)?,(\-?\d+)(\.\d+)?$`)
	validPhoneRegex      = regexp.MustCompile(`^\(?([0-9]{3})\)?\ [-.●]?([0-9]{3})[-.●]?([0-9]{4})$`)

	// ErrValidation is returned when a field has a validation error.
	ErrValidation = errors.New("validation")
)

func init() {
	config := &validator.Config{
		TagName:      "validate",
		FieldNameTag: "json",
	}

	validate = validator.New(config)

	validate.RegisterValidation("phone", Phone)
	validate.RegisterValidation("timezone", Timezone)
	validate.RegisterValidation("coordinates", Coordinates)

}

// ValidationErrors contains the array of errors
type ValidationErrors struct {
	errors map[string][]string
}

// MarshalJSON implements the Marshaler interface for JSON.
func (e ValidationErrors) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.errors)
}

// HasErrors returns true if there are errors in the struct
func (e ValidationErrors) HasErrors() bool {
	return e.Len() > 0
}

// Len returns the amount of errors that occured
func (e ValidationErrors) Len() int {
	return len(e.errors)
}

// AddError adds an error to the validation error message.
func (e *ValidationErrors) AddError(field, err string) {
	if _, ok := e.errors[field]; ok {
		e.errors[field] = append(e.errors[field], err)
	} else {
		e.errors[field] = []string{err}
	}
}

// Error returns the error string, corresponding to the Error interface
func (e ValidationErrors) Error() string {
	if e.HasErrors() {

		err := "Validation error on fields: "

		keys := make([]string, 0, len(e.errors))
		for key := range e.errors {
			keys = append(keys, key)
		}

		err += strings.Join(keys, ", ")

		return err
	}

	return ""
}

// NewValidationErrors creates a new ValidationErrors object from the foreign
// validator.ValidationErrors structure
func NewValidationErrors(verrs validator.ValidationErrors) *ValidationErrors {
	verr := ValidationErrors{
		errors: make(map[string][]string),
	}

	if verrs != nil {
		for _, err := range verrs {
			// trim off the base struct namespace
			ns := strings.Join(strings.Split(err.NameNamespace, ".")[1:], ".")

			// merge in errors.
			verr.errors[ns] = []string{err.Tag}
		}
	}

	return &verr
}

// Phone validates an phone number, returns true if it is valid, false otherwise
func Phone(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	return validPhoneRegex.MatchString(field.String())
}

// Timezone validates that a given string is recognizable as a timezone by the go standard library
func Timezone(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	_, err := time.LoadLocation(field.String())
	return (err == nil)
}

// Coordinates validates a set of coordinates in the form of 123.00,10.0
func Coordinates(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	if !validCoordinateRegex.MatchString(field.String()) {
		return false
	}

	return true
}

// Struct validates a structure
func Struct(i interface{}) error {
	if errs := validate.Struct(i); errs != nil {
		return NewValidationErrors(errs.(validator.ValidationErrors))
	}

	return nil
}

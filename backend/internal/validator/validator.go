package validator

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// EmailRX is a regular expression for validating email addresses
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// Valid checks if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddNonFieldError adds an error message that is not specific to a single field
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// AddFieldError adds an error message for a specific field, if not already present
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// CheckField checks a condition for a field and adds an error message if the condition is false
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank checks if a string is not empty or just whitespace
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars checks if a string does not exceed a maximum number of characters
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue checks if a value is one of the permitted values
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// MinChars checks if a string has at least a minimum number of characters
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Matches checks if a string matches a regular expression
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func ValidateSearchString(searchString string) error {
	// Regular expression pattern to match alphanumeric characters, spaces, or empty string
	pattern := "^[a-zA-Z0-9 ]*$"

	// Set maximum length for search string
	const maxSearchStringLength = 25
	if len(searchString) > maxSearchStringLength {
		return fmt.Errorf("search string max length exceeded")
	}

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("error compiling regular expression: %w", err)
	}

	if !regex.MatchString(searchString) {
		return fmt.Errorf("search string format is invalid")
	}

	// Check if the search string matches the pattern
	return nil
}

package devify

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
)

// Validation holds validation data and errors for form fields.
type Validation struct {
	Data   url.Values        // Form data from the request
	Errors map[string]string // Validation errors keyed by field name
	Req    *http.Request     // HTTP request for locale and form data
}

// Validator creates a new Validation instance from an HTTP request.
// It parses the request form and initializes the validation state.
func (d *Devify) Validator(r *http.Request) *Validation {
	_ = r.ParseForm() // Ignoring error for simplicity; handle in production if needed
	return &Validation{
		Data:   r.Form,
		Errors: make(map[string]string),
		Req:    r,
	}
}

// Valid reports whether the validation has no errors.
// It returns true if no validation errors exist, false otherwise.
func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message for a field if it doesn’t already exist.
// It associates the message with the specified key in the Errors map.
func (v *Validation) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Has checks if a field exists and is non-empty in the form data.
// It returns true if the field has a non-empty value, false otherwise.
func (v *Validation) Has(field string) bool {
	return strings.TrimSpace(v.Data.Get(field)) != ""
}

// Required ensures the specified fields are present and non-empty.
// It adds a custom or default error for each field that is missing or empty.
func (v *Validation) Required(fields ...string) *Validation {
	message := "This field is required."
	if len(fields) > 1 && fields[len(fields)-1] != "" && !strings.Contains(fields[len(fields)-1], " ") {
		message = fields[len(fields)-1]
		fields = fields[:len(fields)-1]
	}
	for _, field := range fields {
		value := v.Data.Get(field)
		if strings.TrimSpace(value) == "" {
			v.AddError(field, message)
		}
	}
	return v
}

// Check adds an error for a field if the condition is false.
// It uses the provided key and message for the error.
func (v *Validation) Check(ok bool, key, message string) *Validation {
	if !ok {
		v.AddError(key, message)
	}
	return v
}

// IsEmail validates whether a field contains a properly formatted email address.
// It checks against RFC 5322 using mail.ParseAddress and adds a custom or default error if invalid.
func (v *Validation) IsEmail(fieldName string, message ...string) *Validation {
	email := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Invalid email address."
	if email == "" {
		defaultMsg = "Email cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsPhone validates whether a field contains a properly formatted phone number.
// It uses libphonenumber with a region inferred from the request’s Accept-Language header.
// An error (custom or default) is added if the number is invalid or not plausible.
func (v *Validation) IsPhone(fieldName string, message ...string) *Validation {
	phone := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Invalid phone number format."
	if phone == "" {
		defaultMsg = "Phone number cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}

	if phone == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}

	defaultRegion := "US"
	if v.Req != nil {
		acceptLang := v.Req.Header.Get("Accept-Language")
		if acceptLang != "" {
			parts := strings.Split(acceptLang, ",")
			region := strings.SplitN(parts[0], "-", 2)
			if len(region) == 2 {
				defaultRegion = strings.ToUpper(region[1])
			}
		}
	}

	num, err := phonenumbers.Parse(phone, defaultRegion)
	if err != nil {
		v.AddError(fieldName, defaultMsg)
		return v
	}
	if !phonenumbers.IsValidNumber(num) {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// MinLength ensures a field’s value meets a minimum length requirement.
// It adds a custom or default error if the trimmed value is shorter than the minimum.
func (v *Validation) MinLength(fieldName string, min int, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be at least " + strconv.Itoa(min) + " characters long."
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if len(value) < min {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// MaxLength ensures a field’s value does not exceed a maximum length.
// It adds a custom or default error if the trimmed value is longer than the maximum.
func (v *Validation) MaxLength(fieldName string, max int, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must not exceed " + strconv.Itoa(max) + " characters."
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if len(value) > max {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsInt validates whether a field contains a valid integer.
// It adds a custom or default error if the value cannot be parsed as an integer.
func (v *Validation) IsInt(fieldName string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be an integer."
	if value == "" {
		defaultMsg = "Value cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsFloat validates whether a field contains a valid floating-point number.
// It adds a custom or default error if the value cannot be parsed as a float.
func (v *Validation) IsFloat(fieldName string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be a number."
	if value == "" {
		defaultMsg = "Value cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsURL validates whether a field contains a properly formatted URL.
// It checks if the value is a valid URL with a scheme and adds a custom or default error if invalid.
func (v *Validation) IsURL(fieldName string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be a valid URL (e.g., https://example.com)."
	if value == "" {
		defaultMsg = "URL cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}
	u, err := url.Parse(value)
	if err != nil || u.Scheme == "" || u.Host == "" {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsDate validates whether a field contains a valid date.
// It checks the value against common formats (YYYY-MM-DD, MM/DD/YYYY, DD/MM/YYYY, YYYY-MM-DD HH:MM)
// and adds an error if invalid or empty. With one variadic argument, it sets a custom error message.
// With two variadic arguments, it sets a custom message and enforces a single format (e.g., "2006-01-02").
// Default error messages are "Must be a valid date (e.g., YYYY-MM-DD or MM/DD/YYYY)" or "Date cannot be empty."
func (v *Validation) IsDate(fieldName string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be a valid date (e.g., YYYY-MM-DD or MM/DD/YYYY)."
	if value == "" {
		defaultMsg = "Date cannot be empty."
	}
	if len(message) > 0 && message[0] != "" {
		defaultMsg = message[0]
	}
	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}

	formats := []string{
		"2006-01-02",       // YYYY-MM-DD
		"01/02/2006",       // MM/DD/YYYY
		"02/01/2006",       // DD/MM/YYYY
		"2006-01-02 15:04", // YYYY-MM-DD HH:MM
	}
	if len(message) > 1 && message[1] != "" {
		formats = []string{message[1]} // Use custom format if provided
		defaultMsg = "Must match the format: " + message[1]
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return v
		}
	}
	v.AddError(fieldName, defaultMsg)
	return v
}

// Between ensures a field’s value length or numeric value falls within a range.
// It adds a custom or default error if the value is outside the specified range.
func (v *Validation) Between(fieldName string, min, max float64, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := fmt.Sprintf("Must be between %v and %v.", min, max)
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		if f < min || f > max {
			v.AddError(fieldName, defaultMsg)
		}
	} else if float64(len(value)) < min || float64(len(value)) > max {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// In ensures a field’s value is one of the specified options.
// It adds a custom or default error if the value is not in the list.
func (v *Validation) In(fieldName string, options ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be one of: " + strings.Join(options[:len(options)-1], ", ") + "."
	if len(options) > 1 && !strings.Contains(options[len(options)-1], " ") {
		defaultMsg = options[len(options)-1]
		options = options[:len(options)-1]
	}
	for _, opt := range options {
		if value == opt {
			return v
		}
	}
	v.AddError(fieldName, defaultMsg)
	return v
}

// Matches ensures a field’s value matches a regular expression.
// It adds a custom or default error if the value does not match the pattern.
func (v *Validation) Matches(fieldName, pattern string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Does not match the required pattern."
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if matched, _ := regexp.MatchString(pattern, value); !matched {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// HasNoSpaces ensures a field’s value contains no whitespace characters.
// It adds a custom or default error if spaces, tabs, or other whitespace are found.
// Use this when spaces are not allowed (e.g., usernames, codes).
func (v *Validation) HasNoSpaces(fieldName string, message ...string) *Validation {
	value := v.Data.Get(fieldName) // Not trimming to catch all spaces
	defaultMsg := "Must not contain spaces."
	if value == "" {
		defaultMsg = "Value cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}
	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}
	if strings.ContainsAny(value, " \t\n\r") {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// Contains ensures a field's value contains a specific substring.
// It adds a custom or default error if the substring is not found in the value.
func (v *Validation) Contains(fieldName, substring string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := fmt.Sprintf("Must contain '%s'.", substring)
	if value == "" {
		defaultMsg = "Value cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}

	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}

	if !strings.Contains(value, substring) {
		v.AddError(fieldName, defaultMsg)
	}
	return v
}

// IsBoolean validates whether a field's value represents a boolean.
// It checks for common boolean representations ("true", "false", "1", "0")
// and adds a custom or default error if the value is invalid or empty.
func (v *Validation) IsBoolean(fieldName string, message ...string) *Validation {
	value := strings.TrimSpace(v.Data.Get(fieldName))
	defaultMsg := "Must be a valid boolean value (true, false, 1, or 0)."
	if value == "" {
		defaultMsg = "Value cannot be empty."
	}
	if len(message) > 0 {
		defaultMsg = message[0]
	}

	if value == "" {
		v.AddError(fieldName, defaultMsg)
		return v
	}

	// List of valid boolean representations
	validBooleans := []string{"true", "false", "1", "0"}
	for _, valid := range validBooleans {
		if strings.ToLower(value) == valid {
			return v
		}
	}

	v.AddError(fieldName, defaultMsg)
	return v
}

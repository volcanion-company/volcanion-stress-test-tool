package audit

import (
	"regexp"
	"strings"
)

// SensitiveFieldFilter filters sensitive data from audit logs
type SensitiveFieldFilter struct {
	sensitiveFields   []string
	sensitivePatterns []*regexp.Regexp
	maskValue         string
}

// DefaultSensitiveFields lists fields that should be masked in logs
var DefaultSensitiveFields = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"api-key",
	"authorization",
	"auth",
	"bearer",
	"credential",
	"private_key",
	"privatekey",
	"access_token",
	"refresh_token",
	"session_id",
	"sessionid",
	"csrf",
	"ssn",
	"social_security",
	"credit_card",
	"creditcard",
	"card_number",
	"cvv",
	"pin",
}

// DefaultSensitivePatterns lists patterns that should be masked
var DefaultSensitivePatterns = []string{
	`(?i)password["\s:=]+[^,\s\}]+`,
	`(?i)secret["\s:=]+[^,\s\}]+`,
	`(?i)token["\s:=]+[^,\s\}]+`,
	`(?i)api[_-]?key["\s:=]+[^,\s\}]+`,
	`(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`,
	`(?i)basic\s+[A-Za-z0-9+/]+=*`,
	`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, // email
	`\b\d{3}-\d{2}-\d{4}\b`,                               // SSN
	`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,          // Credit card
}

// NewSensitiveFieldFilter creates a new filter with default settings
func NewSensitiveFieldFilter() *SensitiveFieldFilter {
	filter := &SensitiveFieldFilter{
		sensitiveFields: DefaultSensitiveFields,
		maskValue:       "[REDACTED]",
	}

	// Compile patterns
	for _, pattern := range DefaultSensitivePatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			filter.sensitivePatterns = append(filter.sensitivePatterns, re)
		}
	}

	return filter
}

// WithMaskValue sets custom mask value
func (f *SensitiveFieldFilter) WithMaskValue(mask string) *SensitiveFieldFilter {
	f.maskValue = mask
	return f
}

// AddSensitiveField adds a field to the sensitive list
func (f *SensitiveFieldFilter) AddSensitiveField(field string) {
	f.sensitiveFields = append(f.sensitiveFields, strings.ToLower(field))
}

// AddSensitivePattern adds a pattern to the sensitive list
func (f *SensitiveFieldFilter) AddSensitivePattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	f.sensitivePatterns = append(f.sensitivePatterns, re)
	return nil
}

// IsSensitiveField checks if a field name is sensitive
func (f *SensitiveFieldFilter) IsSensitiveField(fieldName string) bool {
	lowered := strings.ToLower(fieldName)
	for _, sensitive := range f.sensitiveFields {
		if strings.Contains(lowered, sensitive) {
			return true
		}
	}
	return false
}

// FilterMap filters sensitive fields from a map
func (f *SensitiveFieldFilter) FilterMap(data map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{}, len(data))

	for key, value := range data {
		if f.IsSensitiveField(key) {
			filtered[key] = f.maskValue
			continue
		}

		// Recursively filter nested maps
		switch v := value.(type) {
		case map[string]interface{}:
			filtered[key] = f.FilterMap(v)
		case map[string]string:
			filtered[key] = f.FilterStringMap(v)
		case string:
			filtered[key] = f.FilterString(v)
		default:
			filtered[key] = value
		}
	}

	return filtered
}

// FilterStringMap filters sensitive fields from a string map
func (f *SensitiveFieldFilter) FilterStringMap(data map[string]string) map[string]string {
	filtered := make(map[string]string, len(data))

	for key, value := range data {
		if f.IsSensitiveField(key) {
			filtered[key] = f.maskValue
		} else {
			filtered[key] = f.FilterString(value)
		}
	}

	return filtered
}

// FilterString filters sensitive patterns from a string
func (f *SensitiveFieldFilter) FilterString(input string) string {
	result := input

	for _, pattern := range f.sensitivePatterns {
		result = pattern.ReplaceAllString(result, f.maskValue)
	}

	return result
}

// FilterHeaders filters sensitive headers
func (f *SensitiveFieldFilter) FilterHeaders(headers map[string][]string) map[string][]string {
	filtered := make(map[string][]string, len(headers))

	sensitiveHeaders := map[string]bool{
		"authorization":   true,
		"x-api-key":       true,
		"x-auth-token":    true,
		"cookie":          true,
		"set-cookie":      true,
		"x-csrf-token":    true,
		"x-forwarded-for": false, // Keep but may want to hash
	}

	for key, values := range headers {
		loweredKey := strings.ToLower(key)
		if sensitiveHeaders[loweredKey] {
			filtered[key] = []string{f.maskValue}
		} else {
			filtered[key] = values
		}
	}

	return filtered
}

// FilterAuditEvent filters sensitive data from an audit event
func (f *SensitiveFieldFilter) FilterAuditEvent(event *AuditEvent) *AuditEvent {
	filtered := *event

	// Filter details if present
	if event.Details != nil {
		filtered.Details = f.FilterMap(event.Details)
	}

	// Filter error messages (may contain sensitive data)
	if event.Error != "" {
		filtered.Error = f.FilterString(event.Error)
	}

	return &filtered
}

// SafeLog creates a log-safe version of any value
func (f *SensitiveFieldFilter) SafeLog(key string, value interface{}) interface{} {
	if f.IsSensitiveField(key) {
		return f.maskValue
	}

	switch v := value.(type) {
	case string:
		return f.FilterString(v)
	case map[string]interface{}:
		return f.FilterMap(v)
	case map[string]string:
		return f.FilterStringMap(v)
	default:
		return value
	}
}

package validation

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidURL          = errors.New("invalid URL format")
	ErrUnsupportedScheme   = errors.New("unsupported URL scheme, only http and https are allowed")
	ErrPrivateIPNotAllowed = errors.New("private/internal IP addresses are not allowed")
	ErrLocalhostNotAllowed = errors.New("localhost targets are not allowed in production mode")
	ErrInvalidHost         = errors.New("invalid or empty host")
	ErrMaliciousInput      = errors.New("potentially malicious input detected")
)

// URLValidator validates and sanitizes URLs for stress testing
type URLValidator struct {
	allowPrivateIPs bool
	allowLocalhost  bool
	allowedSchemes  []string
}

// NewURLValidator creates a new URL validator with default settings
func NewURLValidator() *URLValidator {
	return &URLValidator{
		allowPrivateIPs: false,
		allowLocalhost:  false,
		allowedSchemes:  []string{"http", "https"},
	}
}

// NewURLValidatorDev creates a URL validator for development (allows localhost)
func NewURLValidatorDev() *URLValidator {
	return &URLValidator{
		allowPrivateIPs: true,
		allowLocalhost:  true,
		allowedSchemes:  []string{"http", "https"},
	}
}

// WithAllowPrivateIPs allows or disallows private IP addresses
func (v *URLValidator) WithAllowPrivateIPs(allow bool) *URLValidator {
	v.allowPrivateIPs = allow
	return v
}

// WithAllowLocalhost allows or disallows localhost
func (v *URLValidator) WithAllowLocalhost(allow bool) *URLValidator {
	v.allowLocalhost = allow
	return v
}

// WithAllowedSchemes sets the allowed URL schemes
func (v *URLValidator) WithAllowedSchemes(schemes []string) *URLValidator {
	v.allowedSchemes = schemes
	return v
}

// ValidateURL validates a URL for stress testing
func (v *URLValidator) ValidateURL(rawURL string) (*url.URL, error) {
	// Check for empty URL
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	// Check for potentially malicious patterns
	if containsMaliciousPatterns(rawURL) {
		return nil, ErrMaliciousInput
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// Validate scheme
	if !v.isSchemeAllowed(parsedURL.Scheme) {
		return nil, ErrUnsupportedScheme
	}

	// Validate host
	host := parsedURL.Hostname()
	if host == "" {
		return nil, ErrInvalidHost
	}

	// Check for localhost
	if isLocalhost(host) && !v.allowLocalhost {
		return nil, ErrLocalhostNotAllowed
	}

	// Check for private IP
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) && !v.allowPrivateIPs {
			return nil, ErrPrivateIPNotAllowed
		}
	} else {
		// If it's a hostname, resolve it and check
		addrs, err := net.LookupHost(host)
		if err == nil && !v.allowPrivateIPs {
			for _, addr := range addrs {
				if ip := net.ParseIP(addr); ip != nil && isPrivateIP(ip) {
					return nil, ErrPrivateIPNotAllowed
				}
			}
		}
	}

	return parsedURL, nil
}

// ValidateURLs validates multiple URLs
func (v *URLValidator) ValidateURLs(urls []string) ([]*url.URL, []error) {
	validURLs := make([]*url.URL, 0, len(urls))
	errors := make([]error, 0)

	for _, rawURL := range urls {
		parsedURL, err := v.ValidateURL(rawURL)
		if err != nil {
			errors = append(errors, err)
		} else {
			validURLs = append(validURLs, parsedURL)
		}
	}

	return validURLs, errors
}

// isSchemeAllowed checks if the URL scheme is allowed
func (v *URLValidator) isSchemeAllowed(scheme string) bool {
	scheme = strings.ToLower(scheme)
	for _, allowed := range v.allowedSchemes {
		if scheme == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// isLocalhost checks if the host is localhost
func isLocalhost(host string) bool {
	host = strings.ToLower(host)
	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		host == "[::1]"
}

// isPrivateIP checks if an IP address is private/internal
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check for special ranges
	// 10.0.0.0/8
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 169.254.0.0/16 (link-local)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}

	return false
}

// containsMaliciousPatterns checks for common attack patterns
func containsMaliciousPatterns(input string) bool {
	maliciousPatterns := []string{
		`javascript:`,
		`data:`,
		`vbscript:`,
		`file://`,
		`<script`,
		`</script`,
		`onerror=`,
		`onload=`,
		`onclick=`,
		`eval(`,
		`document.`,
		`window.`,
	}

	lowered := strings.ToLower(input)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}

	// Check for null bytes and other control characters
	if strings.ContainsAny(input, "\x00\x0a\x0d") {
		return true
	}

	return false
}

// SanitizeHeader sanitizes HTTP header values
func SanitizeHeader(value string) string {
	// Remove newlines and carriage returns (HTTP header injection prevention)
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return strings.TrimSpace(value)
}

// SanitizeHeaders sanitizes all headers in a map
func SanitizeHeaders(headers map[string]string) map[string]string {
	sanitized := make(map[string]string, len(headers))
	for key, value := range headers {
		sanitizedKey := SanitizeHeader(key)
		sanitizedValue := SanitizeHeader(value)
		if sanitizedKey != "" {
			sanitized[sanitizedKey] = sanitizedValue
		}
	}
	return sanitized
}

// ValidateTestName validates a test name
func ValidateTestName(name string) error {
	if name == "" {
		return errors.New("test name cannot be empty")
	}
	if len(name) > 200 {
		return errors.New("test name too long (max 200 characters)")
	}
	// Allow alphanumeric, spaces, hyphens, underscores
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.]+$`)
	if !validName.MatchString(name) {
		return errors.New("test name contains invalid characters")
	}
	return nil
}

// ValidateRequestBody validates request body size and content
func ValidateRequestBody(body string, maxSize int) error {
	if len(body) > maxSize {
		return errors.New("request body exceeds maximum size")
	}
	return nil
}

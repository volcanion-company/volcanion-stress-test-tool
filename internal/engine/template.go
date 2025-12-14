package engine

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Pre-compiled regex patterns for better performance
var (
	uuidPattern         = regexp.MustCompile(`\{\{uuid\}\}`)
	timestampPattern    = regexp.MustCompile(`\{\{timestamp\}\}`)
	randomPattern       = regexp.MustCompile(`\{\{random:(\d+)\}\}`)
	randomStringPattern = regexp.MustCompile(`\{\{random_string:(\d+)\}\}`)
	envPattern          = regexp.MustCompile(`\{\{env:(\w+)\}\}`)
	datePattern         = regexp.MustCompile(`\{\{date:([^}]+)\}\}`)
)

// TemplateEngine handles variable substitution in strings
type TemplateEngine struct {
	random *rand.Rand
	mu     sync.Mutex // Protects random
	cache  sync.Map   // Cache for compiled templates
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Process substitutes variables in the input string
// Supported patterns:
// - {{uuid}} - generates a random UUID
// - {{timestamp}} - current Unix timestamp
// - {{random:N}} - random N-digit number
// - {{random_string:N}} - random N-character alphanumeric string
// - {{date:FORMAT}} - current date in specified format
func (t *TemplateEngine) Process(input string) string {
	if input == "" {
		return input
	}

	// Quick check if template markers exist
	if !strings.Contains(input, "{{") {
		return input
	}

	result := input

	// Replace {{uuid}} using pre-compiled pattern
	result = uuidPattern.ReplaceAllStringFunc(result, func(match string) string {
		return t.generateUUID()
	})

	// Replace {{timestamp}} using pre-compiled pattern
	result = timestampPattern.ReplaceAllStringFunc(result, func(match string) string {
		return strconv.FormatInt(time.Now().Unix(), 10)
	})

	// Replace {{random:N}} using pre-compiled pattern
	result = randomPattern.ReplaceAllStringFunc(result, func(match string) string {
		matches := randomPattern.FindStringSubmatch(match)
		if len(matches) == 2 {
			length, _ := strconv.Atoi(matches[1])
			return t.generateRandomNumber(length)
		}
		return match
	})

	// Replace {{random_string:N}} using pre-compiled pattern
	result = randomStringPattern.ReplaceAllStringFunc(result, func(match string) string {
		matches := randomStringPattern.FindStringSubmatch(match)
		if len(matches) == 2 {
			length, _ := strconv.Atoi(matches[1])
			return t.generateRandomString(length)
		}
		return match
	})

	// Replace {{date:FORMAT}} using pre-compiled pattern
	result = datePattern.ReplaceAllStringFunc(result, func(match string) string {
		matches := datePattern.FindStringSubmatch(match)
		if len(matches) == 2 {
			return time.Now().Format(matches[1])
		}
		return match
	})

	return result
}

// ProcessMap applies template substitution to all values in a map
func (t *TemplateEngine) ProcessMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}

	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = t.Process(value)
	}
	return result
}

// generateUUID generates a simple UUID v4 (thread-safe)
func (t *TemplateEngine) generateUUID() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	b := make([]byte, 16)
	t.random.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant
	return strings.ToLower(
		strconv.FormatInt(int64(b[0]), 16) +
			strconv.FormatInt(int64(b[1]), 16) +
			strconv.FormatInt(int64(b[2]), 16) +
			strconv.FormatInt(int64(b[3]), 16) + "-" +
			strconv.FormatInt(int64(b[4]), 16) +
			strconv.FormatInt(int64(b[5]), 16) + "-" +
			strconv.FormatInt(int64(b[6]), 16) +
			strconv.FormatInt(int64(b[7]), 16) + "-" +
			strconv.FormatInt(int64(b[8]), 16) +
			strconv.FormatInt(int64(b[9]), 16) + "-" +
			strconv.FormatInt(int64(b[10]), 16) +
			strconv.FormatInt(int64(b[11]), 16) +
			strconv.FormatInt(int64(b[12]), 16) +
			strconv.FormatInt(int64(b[13]), 16) +
			strconv.FormatInt(int64(b[14]), 16) +
			strconv.FormatInt(int64(b[15]), 16))
}

// generateRandomNumber generates a random N-digit number (thread-safe)
func (t *TemplateEngine) generateRandomNumber(length int) string {
	if length <= 0 {
		return ""
	}
	t.mu.Lock()
	num := t.random.Intn(pow10(length))
	t.mu.Unlock()
	return strconv.Itoa(num)
}

// generateRandomString generates a random alphanumeric string (thread-safe)
func (t *TemplateEngine) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	t.mu.Lock()
	for i := range b {
		b[i] = charset[t.random.Intn(len(charset))]
	}
	t.mu.Unlock()

	return string(b)
}

// pow10 calculates 10^n
func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

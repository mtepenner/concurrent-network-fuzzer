package mutator

import (
	"strings"
)

// FormatStringMutator generates payloads designed to exploit format string vulnerabilities.
type FormatStringMutator struct {
	specifiers int
	attackType string // e.g., "%x" for reading, "%n" for writing/crashing
}

// NewFormatStringMutator creates a new mutator that repeats a format specifier.
// Example: NewFormatStringMutator(100, "%x.") will generate "%x.%x.%x..."
func NewFormatStringMutator(specifiers int, attackType string) *FormatStringMutator {
	return &FormatStringMutator{
		specifiers: specifiers,
		attackType: attackType,
	}
}

// Generate creates the actual byte slice payload.
func (m *FormatStringMutator) Generate() []byte {
	// Create a slice of the specifier repeated 'specifiers' times
	payloadStr := strings.Repeat(m.attackType, m.specifiers)
	return []byte(payloadStr)
}

// Name identifies the mutator in logs or reports.
func (m *FormatStringMutator) Name() string {
	return "FormatString"
}

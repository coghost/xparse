package xparse

import (
	"strings"

	"github.com/spf13/cast"
)

// Splitter provides a fluent interface for splitting strings in sequence
// It allows chaining multiple split operations with specified delimiters and indexes
// When a separator appears at the end of the string, it creates an empty string element
// Example: "a," splits to ["a", ""]
type Splitter struct {
	source string
	rules  []SplitRule

	trimTrailing  bool
	keepLastFound bool // shorter name
}

// NewSplitter creates a StringSplitter with a single rule, similar to SplitAtIndex usage
// Returns *StringSplitter for immediate Split() call
func NewSplitter(raw interface{}, sep string, index int) *Splitter {
	return NewStringSplitter(raw).By(sep, index)
}

// NewStringSplitter creates a new StringSplitter instance
// It accepts any type and attempts to convert it to string
// Returns an empty StringSplitter if conversion fails
//
// Example:
//
//	splitter := NewStringSplitter("hello=world&foo=bar")
//	result := splitter.
//	    WithRule("=", 1).     // gets "world&foo=bar"
//	    WithRule("&", 0).     // gets "world"
//	    Split()
func NewStringSplitter(raw interface{}) *Splitter {
	if raw == nil {
		return &Splitter{trimTrailing: true} // default to true
	}

	str := strings.TrimSpace(cast.ToString(raw))

	return &Splitter{
		source: str,
		rules:  make([]SplitRule, 0),

		trimTrailing:  true,
		keepLastFound: true,
	}
}

// By adds a split rule to the chain
//   - delimiter: the string to split on
//   - indexes: optional slice of indexes, first one is used, defaults to 0
//
// Returns the StringSplitter for method chaining
func (s *Splitter) By(delimiter string, indexes ...int) *Splitter {
	index := FirstOrDefaultArgs(0, indexes...)
	s.rules = append(s.rules, NewSplitRule(delimiter, index))

	return s
}

// TrimTrailing configures whether to trim trailing empty values
func (s *Splitter) TrimTrailing(enabled bool) *Splitter {
	s.trimTrailing = enabled
	return s
}

// KeepLastFound configures whether to keep the last found value as return value of String()
func (s *Splitter) KeepLastFound(enabled bool) *Splitter {
	s.keepLastFound = enabled
	return s
}

// Split applies all rules in sequence and returns the final result.
//   - Returns empty string if input is empty or any split operation fails
//   - Returns interface{} to maintain compatibility with existing code
func (s *Splitter) String() string {
	if s.source == "" {
		return ""
	}

	result := s.source
	for _, rule := range s.rules {
		// If delimiter is empty or not found
		if rule.Delimiter == "" || !strings.Contains(result, rule.Delimiter) {
			if s.keepLastFound {
				result = strings.TrimSpace(result)
				continue
			}
			// Return the last valid result instead of empty string
			return ""
		}

		parts := strings.Split(result, rule.Delimiter)

		// Handle trailing empty values before processing index
		if s.trimTrailing {
			lastNonEmpty := -1

			for i := 0; i < len(parts); i++ {
				if strings.TrimSpace(parts[i]) != "" {
					lastNonEmpty = i
				}
			}

			if lastNonEmpty >= 0 {
				parts = parts[:lastNonEmpty+1]
			} else {
				return ""
			}
		}

		if rule.Index < 0 {
			rule.Index = len(parts) + rule.Index
		}

		// Then check bounds
		if rule.Index < 0 || rule.Index >= len(parts) {
			return ""
		}

		result = strings.TrimSpace(parts[rule.Index])
	}

	return result
}

// Deprecated: Use String() instead.
//
//	Split is an alias for String() for backward compatibility
func (s *Splitter) Split() string {
	return s.String()
}

// SplitRule defines a single split operation configuration
type SplitRule struct {
	Delimiter string // The string to split on
	Index     int    // The index to select after split (negative index counts from end)
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to add a single split rule
func NewSplitRule(delimiter string, index int) SplitRule {
	return SplitRule{
		Delimiter: delimiter,
		Index:     index,
	}
}

// Deprecated: Use StringSplitter.SplitAtIndex instead
func SplitAtIndex(raw interface{}, sep string, index int) string {
	// Handle nil input
	if raw == nil {
		return ""
	}

	// Convert to string and trim spaces
	str := strings.TrimSpace(cast.ToString(raw))
	if str == "" {
		return ""
	}

	// Check separator validity
	if sep == "" || !strings.Contains(str, sep) {
		return str
	}

	// Split string into array
	arr := strings.Split(str, sep)
	arrLen := len(arr)

	// Handle empty array case
	if arrLen == 0 {
		return str
	}

	// Normalize index
	if index < 0 {
		index = arrLen + index
	}

	// Bound check
	switch {
	case index < 0:
		return arr[0]
	case index >= arrLen:
		return arr[arrLen-1]
	default:
		return strings.TrimSpace(arr[index])
	}
}

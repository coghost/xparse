package xparse

import (
	"strings"

	"github.com/spf13/cast"
)

// StringSplitter provides a fluent interface for splitting strings in sequence
// It allows chaining multiple split operations with specified delimiters and indexes
// When a separator appears at the end of the string, it creates an empty string element
// Example: "a," splits to ["a", ""]
type StringSplitter struct {
	raw   string
	rules []SplitRule

	trimTrailingEmpty bool
}

// NewSplitter creates a StringSplitter with a single rule, similar to SplitAtIndex usage
// Returns *StringSplitter for immediate Split() call
func NewSplitter(raw interface{}, sep string, index int) *StringSplitter {
	return NewStringSplitter(raw).WithRule(sep, index)
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
func NewStringSplitter(raw interface{}) *StringSplitter {
	if raw == nil {
		return &StringSplitter{trimTrailingEmpty: true} // default to true
	}

	str := strings.TrimSpace(cast.ToString(raw))

	return &StringSplitter{
		raw:   str,
		rules: make([]SplitRule, 0),

		trimTrailingEmpty: true, // default to true
	}
}

// WithRule adds a split rule to the chain
//   - delimiter: the string to split on
//   - indexes: optional slice of indexes, first one is used, defaults to 0
//
// Returns the StringSplitter for method chaining
func (ss *StringSplitter) WithRule(delimiter string, indexes ...int) *StringSplitter {
	index := FirstOrDefaultArgs(0, indexes...)
	ss.rules = append(ss.rules, NewSplitRule(delimiter, index))

	return ss
}

// SetTrimTrailingEmpty configures whether to trim trailing empty values
func (ss *StringSplitter) SetTrimTrailingEmpty(trim bool) *StringSplitter {
	ss.trimTrailingEmpty = trim
	return ss
}

// Split applies all rules in sequence and returns the final result.
//   - Returns empty string if input is empty or any split operation fails
//   - Returns interface{} to maintain compatibility with existing code
func (ss *StringSplitter) Split() string {
	if ss.raw == "" {
		return ""
	}

	result := ss.raw
	for _, rule := range ss.rules {
		if rule.Delimiter == "" || !strings.Contains(result, rule.Delimiter) {
			return strings.TrimSpace(result)
		}

		parts := strings.Split(result, rule.Delimiter)

		// Handle trailing empty values before processing index
		if ss.trimTrailingEmpty {
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

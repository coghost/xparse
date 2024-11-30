package xparse

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSplitter(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		input    interface{}
		rules    func(*StringSplitter) *StringSplitter
		expected string
	}{
		{
			name:     "nil input",
			input:    nil,
			rules:    func(s *StringSplitter) *StringSplitter { return s },
			expected: "",
		},
		{
			name:  "single split rule",
			input: "hello=world",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule("=", 1)
			},
			expected: "world",
		},
		{
			name:  "multiple split rules",
			input: "hello=world&foo=bar",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule("&", 1).WithRule("=", 1)
			},
			expected: "bar",
		},
		{
			name:  "negative index",
			input: "a,b,c",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", -1)
			},
			expected: "c",
		},
		{
			name:  "index out of bounds",
			input: "a,b,c",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 5)
			},
			expected: "",
		},
		// New test cases for trailing separators
		{
			name:  "trailing separator single rule",
			input: "a,b,c,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", -1)
			},
			expected: "c",
		},
		{
			name:  "trailing separator first element",
			input: "a,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 0)
			},
			expected: "a",
		},
		{
			name:  "trailing separator second element",
			input: "a,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1)
			},
			expected: "",
		},
		{
			name:  "only separator",
			input: ",",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 0)
			},
			expected: "",
		},
		{
			name:  "multiple separators",
			input: ",,,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1)
			},
			expected: "",
		},
		{
			name:  "chained rules with trailing separator",
			input: "key1=value1,key2=value2,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1).WithRule("=", 1)
			},
			expected: "value2",
		},
		{
			name:  "trim spaces input",
			input: "  hello = world  ",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule("=", 1)
			},
			expected: "world",
		},
		{
			name:  "empty delimiter",
			input: "abc",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule("", 0)
			},
			expected: "abc",
		},
		{
			name:  "non-string input",
			input: 123,
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 0)
			},
			expected: "123",
		},
		{
			name:  "delimiter not found",
			input: "hello world",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule("=", 0)
			},
			expected: "hello world",
		},
		{
			name:  "multiple spaces between parts",
			input: "key1  =  value1 , key2  =  value2",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1).WithRule("=", 1)
			},
			expected: "value2",
		},
		// New test cases for trailing empty values
		{
			name:  "trailing empty values trimmed by default",
			input: "a,b,c,,,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", -1)
			},
			expected: "c",
		},
		{
			name:  "keep trailing empty values",
			input: "a,b,c,,,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.SetTrimTrailingEmpty(false).WithRule(",", -1)
			},
			expected: "",
		},
		{
			name:  "trailing empty values with multiple rules",
			input: "key1=value1,key2=value2,,,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1).WithRule("=", 1)
			},
			expected: "value2",
		},
		{
			name:  "all empty values",
			input: ",,,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", 1)
			},
			expected: "",
		},
		{
			name:  "spaces between empty values",
			input: "a, , ,  ,",
			rules: func(s *StringSplitter) *StringSplitter {
				return s.WithRule(",", -1)
			},
			expected: "a",
		},
	}

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := NewStringSplitter(tt.input)
			result := tt.rules(splitter).Split()
			assert.Equal(tt.expected, result, fmt.Sprintf("%v->%d:%s", tt.input, idx, tt.expected))
		})
	}
}

func TestSplitAtIndex(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		input    interface{}
		sep      string
		index    int
		expected string
	}{
		{nil, ",", 0, ""},            // returns ""
		{"", ",", 0, ""},             // returns ""
		{"a,b,c", "", 1, "a,b,c"},    // returns "a,b,c"
		{"a,b,c", ",", 1, "b"},       // returns "b"
		{"a,b,c", ",", -1, "c"},      // returns "c"
		{"a,b,c", ",", 5, ""},        // returns ""
		{"a,b,c", ",", -5, ""},       // returns ""
		{" a , b , c ", ",", 1, "b"}, // returns "b"
		// New test cases for trailing separator
		{"a,b,c,", ",", 0, "a"},  // trailing separator
		{"a,b,c,", ",", -1, "c"}, // last element with trailing separator
		{"a,", ",", 0, "a"},      // single element with trailing separator
		{"a,", ",", 1, ""},       // second element with trailing separator
		{",", ",", 0, ""},        // only separator
		{",,,", ",", 0, ""},      // multiple separators
	}
	for _, tt := range tests {
		result := NewSplitter(tt.input, tt.sep, tt.index).Split()
		assert.Equal(tt.expected, result, fmt.Sprintf("%v->%d:%s", tt.input, tt.index, tt.expected))
	}
}

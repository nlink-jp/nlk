// Package jsonfix extracts and repairs JSON from arbitrary text.
//
// LLM responses often contain JSON wrapped in markdown fences, mixed with
// explanatory text, truncated with missing closing braces, single-quoted,
// with trailing commas, comments, or other syntax errors. This package
// handles all these cases using a recursive descent parser with repair
// heuristics, inspired by Python's json-repair.
//
// Zero external dependencies — standard library only.
//
// Usage:
//
//	raw := "Here is the result:\n```json\n{'key': 'value',}\n```"
//	result, err := jsonfix.Extract(raw)
//	// result == `{"key":"value"}`
package jsonfix

import (
	"encoding/json"
	"errors"
	"strings"
)

var (
	// ErrNoJSON is returned when no JSON structure is found in the input.
	ErrNoJSON = errors.New("jsonfix: no JSON found in input")

	// ErrUnfixable is returned when the repaired output is still not valid JSON.
	ErrUnfixable = errors.New("jsonfix: repaired output is not valid JSON")
)

// Extract finds and repairs JSON in the input text.
// It handles markdown fences, surrounding text, single quotes, trailing commas,
// comments, unquoted keys, missing braces, and many other common LLM output issues.
//
// Returns the repaired JSON string, or an error if no JSON could be extracted.
//
// Note: the input is fully loaded into memory. Callers should limit input size
// before calling Extract if processing untrusted or unbounded data.
//
// Security note: heuristic repairs may produce a JSON structure that differs
// from the LLM's original intent (JSON smuggling). Always validate the
// deserialized output — for example with the validate package — before acting
// on it.
func Extract(input string) (string, error) {
	if input == "" {
		return "", ErrNoJSON
	}

	// Try the input as-is first.
	result, err := tryParse(input)
	if err == nil {
		return result, nil
	}

	// If the input looks like escaped JSON (\"), unescape and retry.
	if strings.Contains(input, `\"`) {
		unescaped := unescapeJSON(input)
		if unescaped != input {
			result, err = tryParse(unescaped)
			if err == nil {
				return result, nil
			}
		}
	}

	return "", err
}

// tryParse runs the repair parser and validates the result.
func tryParse(input string) (string, error) {
	p := newParser(input)
	result := p.repair()

	if result == "" {
		return "", ErrNoJSON
	}

	if !json.Valid([]byte(result)) {
		return "", ErrUnfixable
	}

	return result, nil
}

// unescapeJSON detects and unescapes double-escaped JSON strings.
// Handles: \" → ", \\ → \, \n → newline, \t → tab.
func unescapeJSON(input string) string {
	// Only unescape if the pattern looks like escaped JSON.
	// Check for \" adjacent to structural characters.
	if !strings.Contains(input, `{\"`) && !strings.Contains(input, `[\"`) {
		return input
	}

	var b strings.Builder
	b.Grow(len(input))
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			next := runes[i+1]
			switch next {
			case '"':
				b.WriteRune('"')
				i++
			case '\\':
				b.WriteRune('\\')
				i++
			case 'n':
				b.WriteRune('\n')
				i++
			case 't':
				b.WriteRune('\t')
				i++
			default:
				b.WriteRune(runes[i])
			}
		} else {
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}

// ExtractTo extracts and repairs JSON from input, then unmarshals into target.
func ExtractTo(input string, target any) error {
	s, err := Extract(input)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), target)
}

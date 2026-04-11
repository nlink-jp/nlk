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
func Extract(input string) (string, error) {
	if input == "" {
		return "", ErrNoJSON
	}

	p := newParser(input)
	result := p.repair()

	if result == "" {
		return "", ErrNoJSON
	}

	// Validate the repaired output.
	if !json.Valid([]byte(result)) {
		return "", ErrUnfixable
	}

	return result, nil
}

// ExtractTo extracts and repairs JSON from input, then unmarshals into target.
func ExtractTo(input string, target any) error {
	s, err := Extract(input)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), target)
}

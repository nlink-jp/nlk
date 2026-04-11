// Package jsonfix extracts and repairs JSON from arbitrary text.
//
// LLM responses often contain JSON wrapped in markdown fences, mixed with
// explanatory text, or truncated with missing closing braces. This package
// handles these cases.
//
// Usage:
//
//	raw := "Here is the result:\n```json\n{\"key\": \"value\"}\n```"
//	result, err := jsonfix.Extract(raw)
//	// result == `{"key": "value"}`
package jsonfix

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrNoJSON is returned when no JSON object or array is found in the input.
	ErrNoJSON = errors.New("jsonfix: no JSON found in input")

	// ErrUnfixable is returned when extracted JSON cannot be parsed or repaired.
	ErrUnfixable = errors.New("jsonfix: extracted JSON could not be parsed or repaired")

	// Match a JSON object or array (greedy, dotall).
	reJSON = regexp.MustCompile(`(?s)(\{.*\}|\[.*\])`)

	// Match markdown code fences: ```json ... ``` or ``` ... ```
	reFence = regexp.MustCompile("(?s)```(?:json)?\\s*\n?(.*?)\n?```")

)

// Extract finds the first JSON object or array in the input text,
// attempts to parse it, and repairs common issues (missing closing
// braces/brackets, markdown fences).
//
// Returns the extracted JSON as a string. The JSON is not reformatted.
func Extract(input string) (string, error) {
	// Step 1: Strip markdown fences if present.
	cleaned := stripFences(input)

	// Step 2: Try to extract and parse JSON.
	return extractJSON(cleaned)
}

// ExtractTo extracts JSON from input and unmarshals it into the target value.
func ExtractTo(input string, target any) error {
	s, err := Extract(input)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), target)
}

// stripFences removes markdown code fences and returns the inner content.
func stripFences(input string) string {
	if m := reFence.FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}
	return input
}

// extractJSON finds and parses JSON from text, repairing if needed.
func extractJSON(input string) (string, error) {
	// Strategy 1: Regex match for complete JSON.
	if match := reJSON.FindString(input); match != "" {
		match = strings.TrimSpace(match)
		if json.Valid([]byte(match)) {
			return match, nil
		}
		// Try to fix the regex-matched portion.
		if fixed := fixIncomplete(match); fixed != "" && json.Valid([]byte(fixed)) {
			return fixed, nil
		}
	}

	// Strategy 2: Find the first { or [ and try to fix from there.
	// Handles cases where closing delimiters are missing entirely.
	trimmed := strings.TrimSpace(input)
	startObj := strings.Index(trimmed, "{")
	startArr := strings.Index(trimmed, "[")

	start := -1
	switch {
	case startObj >= 0 && startArr >= 0:
		start = min(startObj, startArr)
	case startObj >= 0:
		start = startObj
	case startArr >= 0:
		start = startArr
	}

	if start < 0 {
		return "", ErrNoJSON
	}

	candidate := trimmed[start:]
	if fixed := fixIncomplete(candidate); fixed != "" && json.Valid([]byte(fixed)) {
		return fixed, nil
	}

	return "", ErrUnfixable
}

// fixIncomplete adds missing closing braces and/or brackets.
func fixIncomplete(s string) string {
	var result strings.Builder
	result.WriteString(s)

	changed := false

	// Fix missing closing braces.
	openBraces := strings.Count(s, "{")
	closeBraces := strings.Count(s, "}")
	if openBraces > closeBraces {
		result.WriteString(strings.Repeat("}", openBraces-closeBraces))
		changed = true
	}

	// Fix missing closing brackets.
	openBrackets := strings.Count(s, "[")
	closeBrackets := strings.Count(s, "]")
	if openBrackets > closeBrackets {
		result.WriteString(strings.Repeat("]", openBrackets-closeBrackets))
		changed = true
	}

	if !changed {
		return ""
	}
	return result.String()
}

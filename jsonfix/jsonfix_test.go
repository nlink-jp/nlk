package jsonfix

import (
	"encoding/json"
	"testing"
)

func TestExtractPlainJSON(t *testing.T) {
	input := `{"key": "value", "num": 42}`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractJSONArray(t *testing.T) {
	input := `[1, 2, 3]`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `[1, 2, 3]` {
		t.Errorf("expected [1, 2, 3], got %s", got)
	}
}

func TestExtractFromMarkdownFence(t *testing.T) {
	input := "Here is the result:\n```json\n{\"key\": \"value\"}\n```\nDone."
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractFromFenceNoLang(t *testing.T) {
	input := "```\n{\"a\": 1}\n```"
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractFromSurroundingText(t *testing.T) {
	input := "The analysis result is:\n{\"category\": \"safe\", \"confidence\": 0.95}\nEnd of response."
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["category"] != "safe" {
		t.Errorf("expected category=safe, got %v", m["category"])
	}
}

func TestExtractFixMissingBraces(t *testing.T) {
	input := `{"key": "value", "nested": {"inner": true}`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractFixMissingBrackets(t *testing.T) {
	// Simple case: array with missing closing bracket.
	input := `[1, 2, 3`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractNoJSON(t *testing.T) {
	input := "This is just plain text with no JSON at all."
	_, err := Extract(input)
	if err != ErrNoJSON {
		t.Errorf("expected ErrNoJSON, got %v", err)
	}
}

func TestExtractEmptyInput(t *testing.T) {
	_, err := Extract("")
	if err != ErrNoJSON {
		t.Errorf("expected ErrNoJSON, got %v", err)
	}
}

func TestExtractUnfixableJSON(t *testing.T) {
	// Malformed JSON that can't be fixed by adding closing braces.
	input := `{"key": value_without_quotes}`
	_, err := Extract(input)
	if err != ErrUnfixable {
		t.Errorf("expected ErrUnfixable, got %v", err)
	}
}

func TestExtractToStruct(t *testing.T) {
	type Result struct {
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}

	input := "```json\n{\"category\": \"phishing\", \"confidence\": 0.87}\n```"
	var r Result
	if err := ExtractTo(input, &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Category != "phishing" {
		t.Errorf("expected phishing, got %s", r.Category)
	}
	if r.Confidence != 0.87 {
		t.Errorf("expected 0.87, got %f", r.Confidence)
	}
}

func TestExtractToInvalidTarget(t *testing.T) {
	input := `{"key": "value"}`
	var s string
	err := ExtractTo(input, &s)
	if err == nil {
		t.Error("expected error unmarshaling object into string")
	}
}

func TestExtractNestedJSON(t *testing.T) {
	input := `{"a": {"b": {"c": [1, 2, {"d": true}]}}}`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractMultilineJSON(t *testing.T) {
	input := `Some text before
{
  "key": "value",
  "list": [
    1,
    2,
    3
  ]
}
Some text after`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(got)) {
		t.Errorf("result is not valid JSON: %s", got)
	}
}

func TestExtractPreservesOriginal(t *testing.T) {
	// Ensure we don't modify the JSON content.
	input := `{"emoji": "🎉", "japanese": "日本語テスト"}`
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["emoji"] != "🎉" {
		t.Errorf("emoji not preserved")
	}
	if m["japanese"] != "日本語テスト" {
		t.Errorf("japanese not preserved")
	}
}

package jsonfix

import (
	"encoding/json"
	"testing"
)

// helper validates that Extract produces valid JSON.
func mustExtract(t *testing.T, name, input string) string {
	t.Helper()
	got, err := Extract(input)
	if err != nil {
		t.Fatalf("[%s] unexpected error: %v\n  input: %s", name, err, input)
	}
	if !json.Valid([]byte(got)) {
		t.Fatalf("[%s] result is not valid JSON: %s", name, got)
	}
	return got
}

// --- Basic extraction ---

func TestPlainObject(t *testing.T) {
	mustExtract(t, "plain object", `{"key": "value", "num": 42}`)
}

func TestPlainArray(t *testing.T) {
	mustExtract(t, "plain array", `[1, 2, 3]`)
}

func TestNestedJSON(t *testing.T) {
	mustExtract(t, "nested", `{"a": {"b": {"c": [1, 2, {"d": true}]}}}`)
}

func TestEmptyObject(t *testing.T) {
	got := mustExtract(t, "empty object", `{}`)
	if got != `{}` {
		t.Errorf("expected {}, got %s", got)
	}
}

func TestEmptyArray(t *testing.T) {
	got := mustExtract(t, "empty array", `[]`)
	if got != `[]` {
		t.Errorf("expected [], got %s", got)
	}
}

// --- Markdown fences ---

func TestMarkdownFenceJSON(t *testing.T) {
	input := "Here is the result:\n```json\n{\"key\": \"value\"}\n```\nDone."
	mustExtract(t, "markdown fence json", input)
}

func TestMarkdownFenceNoLang(t *testing.T) {
	input := "```\n{\"a\": 1}\n```"
	mustExtract(t, "markdown fence no lang", input)
}

func TestMarkdownFenceWithExplanation(t *testing.T) {
	input := "Based on my analysis, here is the JSON output:\n```json\n{\"result\": \"safe\"}\n```\nI hope this helps."
	mustExtract(t, "fence with explanation", input)
}

// --- Surrounding text ---

func TestSurroundingText(t *testing.T) {
	input := "The analysis result is:\n{\"category\": \"safe\", \"confidence\": 0.95}\nEnd of response."
	got := mustExtract(t, "surrounding text", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["category"] != "safe" {
		t.Errorf("expected safe, got %v", m["category"])
	}
}

// --- Single quotes ---

func TestSingleQuotes(t *testing.T) {
	input := `{'key': 'value', 'num': 42}`
	got := mustExtract(t, "single quotes", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestMixedQuotes(t *testing.T) {
	input := `{"key": 'value', 'key2': "value2"}`
	mustExtract(t, "mixed quotes", input)
}

// --- Trailing commas ---

func TestTrailingCommaObject(t *testing.T) {
	input := `{"key": "value", "key2": "value2",}`
	mustExtract(t, "trailing comma object", input)
}

func TestTrailingCommaArray(t *testing.T) {
	input := `[1, 2, 3,]`
	mustExtract(t, "trailing comma array", input)
}

func TestTrailingCommaNestedObject(t *testing.T) {
	input := `{"a": {"b": 1,}, "c": [1, 2,],}`
	mustExtract(t, "trailing comma nested", input)
}

// --- Unquoted keys ---

func TestUnquotedKeys(t *testing.T) {
	input := `{key: "value", key2: 42}`
	got := mustExtract(t, "unquoted keys", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

// --- Comments ---

func TestLineComment(t *testing.T) {
	input := `{
		// This is a comment
		"key": "value"
	}`
	mustExtract(t, "line comment", input)
}

func TestBlockComment(t *testing.T) {
	input := `{
		/* block comment */
		"key": "value"
	}`
	mustExtract(t, "block comment", input)
}

func TestHashComment(t *testing.T) {
	input := `{
		# hash comment
		"key": "value"
	}`
	mustExtract(t, "hash comment", input)
}

// --- Boolean/null case normalization ---

func TestUppercaseTrue(t *testing.T) {
	input := `{"flag": True}`
	got := mustExtract(t, "uppercase True", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["flag"] != true {
		t.Errorf("expected true, got %v", m["flag"])
	}
}

func TestUppercaseFalse(t *testing.T) {
	input := `{"flag": False}`
	got := mustExtract(t, "uppercase False", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["flag"] != false {
		t.Errorf("expected false, got %v", m["flag"])
	}
}

func TestUppercaseNone(t *testing.T) {
	input := `{"value": None}`
	got := mustExtract(t, "Python None", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["value"] != nil {
		t.Errorf("expected null, got %v", m["value"])
	}
}

func TestUppercaseNULL(t *testing.T) {
	input := `{"value": NULL}`
	got := mustExtract(t, "uppercase NULL", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["value"] != nil {
		t.Errorf("expected null, got %v", m["value"])
	}
}

// --- Missing closing braces/brackets ---

func TestMissingClosingBrace(t *testing.T) {
	input := `{"key": "value", "nested": {"inner": true}`
	mustExtract(t, "missing closing brace", input)
}

func TestMissingClosingBracket(t *testing.T) {
	input := `[1, 2, 3`
	mustExtract(t, "missing closing bracket", input)
}

func TestDeepNestingMissing(t *testing.T) {
	input := `{"a": {"b": {"c": [1, 2, 3`
	mustExtract(t, "deep nesting missing", input)
}

// --- Missing commas ---

func TestMissingCommaObject(t *testing.T) {
	input := `{"key1": "val1" "key2": "val2"}`
	mustExtract(t, "missing comma object", input)
}

func TestMissingCommaArray(t *testing.T) {
	input := `["a" "b" "c"]`
	mustExtract(t, "missing comma array", input)
}

// --- Python constructs ---

func TestPythonTuple(t *testing.T) {
	input := `("a", "b", "c")`
	got := mustExtract(t, "python tuple", input)
	var arr []string
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("not valid JSON array: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}

// --- Ellipsis ---

func TestEllipsis(t *testing.T) {
	input := `[1, 2, 3, ...]`
	got := mustExtract(t, "ellipsis", input)
	var arr []any
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}

// --- Number repair ---

func TestLeadingDot(t *testing.T) {
	input := `{"val": .5}`
	got := mustExtract(t, "leading dot", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["val"] != 0.5 {
		t.Errorf("expected 0.5, got %v", m["val"])
	}
}

func TestTrailingDot(t *testing.T) {
	input := `{"val": 1.}`
	got := mustExtract(t, "trailing dot", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["val"] != 1.0 {
		t.Errorf("expected 1.0, got %v", m["val"])
	}
}

func TestNumberWithUnderscore(t *testing.T) {
	input := `{"val": 1_000_000}`
	got := mustExtract(t, "underscore number", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["val"] != 1000000.0 {
		t.Errorf("expected 1000000, got %v", m["val"])
	}
}

// --- Escape handling ---

func TestHexEscape(t *testing.T) {
	input := `{"val": "\x41\x42"}`
	mustExtract(t, "hex escape", input)
}

// --- Error cases ---

func TestEmptyInput(t *testing.T) {
	_, err := Extract("")
	if err != ErrNoJSON {
		t.Errorf("expected ErrNoJSON, got %v", err)
	}
}

func TestNoJSON(t *testing.T) {
	_, err := Extract("This is just plain text with no JSON at all.")
	if err == nil {
		t.Error("expected error for no JSON")
	}
}

// --- ExtractTo ---

func TestExtractToStruct(t *testing.T) {
	type Result struct {
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}

	input := "```json\n{'category': 'phishing', 'confidence': 0.87,}\n```"
	var r Result
	if err := ExtractTo(input, &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Category != "phishing" {
		t.Errorf("expected phishing, got %s", r.Category)
	}
}

// --- Multiline ---

func TestMultilineJSON(t *testing.T) {
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
	mustExtract(t, "multiline", input)
}

// --- Unicode/Japanese ---

func TestJapaneseContent(t *testing.T) {
	input := `{"emoji": "🎉", "japanese": "日本語テスト"}`
	got := mustExtract(t, "japanese", input)
	var m map[string]string
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if m["japanese"] != "日本語テスト" {
		t.Errorf("japanese not preserved")
	}
}

// --- Combined repairs ---

func TestCombinedRepairs(t *testing.T) {
	input := `{name: 'John', age: 30, 'active': True,`
	got := mustExtract(t, "combined", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["name"] != "John" {
		t.Errorf("expected John, got %v", m["name"])
	}
	if m["active"] != true {
		t.Errorf("expected true, got %v", m["active"])
	}
}

// --- String escape edge cases ---

func TestEscapedQuoteInString(t *testing.T) {
	input := `{"msg": "He said \"hello\" to me"}`
	mustExtract(t, "escaped quote", input)
}

func TestEscapedInnerQuote(t *testing.T) {
	// Properly escaped inner quote should be preserved.
	input := `{"msg": "lorem \"ipsum\" dolor"}`
	got := mustExtract(t, "escaped inner quote", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

func TestUnescapedInnerQuote(t *testing.T) {
	input := `{"msg": "lorem "ipsum" dolor"}`
	got := mustExtract(t, "unescaped inner quote", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

func TestUnescapedQuoteWithFollowingKey(t *testing.T) {
	input := `{"a": "he said "hello" to me", "b": 1}`
	got := mustExtract(t, "unescaped quote with following key", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["b"] != 1.0 {
		t.Errorf("expected b=1, got %v", m["b"])
	}
}

func TestUnescapedQuoteLastValue(t *testing.T) {
	input := `{"x": "value with "quotes" inside"}`
	got := mustExtract(t, "unescaped quote last value", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

func TestNewlineInString(t *testing.T) {
	// Newline inside a string should close it.
	input := "{\"key\": \"unterminated\n, \"key2\": \"val2\"}"
	mustExtract(t, "newline in string", input)
}

func TestUnicodeEscape(t *testing.T) {
	input := `{"val": "\u0041\u0042"}`
	got := mustExtract(t, "unicode escape", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["val"] != "AB" {
		t.Errorf("expected AB, got %v", m["val"])
	}
}

func TestSingleQuoteEscape(t *testing.T) {
	input := `{"val": "it\'s fine"}`
	mustExtract(t, "single quote escape", input)
}

func TestBackslashInString(t *testing.T) {
	input := `{"path": "C:\\Users\\test"}`
	mustExtract(t, "backslash in string", input)
}

// --- Unquoted values ---

func TestUnquotedStringValue(t *testing.T) {
	input := `{"city": New York, "state": California}`
	got := mustExtract(t, "unquoted value", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

// --- Number edge cases ---

func TestNegativeNumber(t *testing.T) {
	input := `{"val": -42}`
	got := mustExtract(t, "negative number", input)
	var m map[string]any
	json.Unmarshal([]byte(got), &m)
	if m["val"] != -42.0 {
		t.Errorf("expected -42, got %v", m["val"])
	}
}

func TestExponent(t *testing.T) {
	input := `{"val": 1.5e10}`
	mustExtract(t, "exponent", input)
}

func TestNegativeExponent(t *testing.T) {
	input := `{"val": 3.14e-2}`
	mustExtract(t, "negative exponent", input)
}

func TestZero(t *testing.T) {
	input := `{"val": 0}`
	mustExtract(t, "zero", input)
}

// --- Literal edge cases ---

func TestLowercaseLiterals(t *testing.T) {
	input := `{"a": true, "b": false, "c": null}`
	mustExtract(t, "lowercase literals", input)
}

func TestMixedCaseTRUE(t *testing.T) {
	input := `{"a": TRUE, "b": FALSE}`
	mustExtract(t, "all caps bool", input)
}

// --- Missing value ---

func TestMissingValueInObject(t *testing.T) {
	input := `{"key": , "key2": "val"}`
	mustExtract(t, "missing value", input)
}

// --- Deeply nested ---

func TestDeeplyNested(t *testing.T) {
	input := `{"a":{"b":{"c":{"d":{"e":{"f":"deep"}}}}}}`
	mustExtract(t, "deeply nested", input)
}

// --- ExtractTo error ---

func TestExtractToInvalidTarget(t *testing.T) {
	input := `{"key": "value"}`
	var s string
	err := ExtractTo(input, &s)
	if err == nil {
		t.Error("expected error unmarshaling object into string")
	}
}

func TestExtractToNoJSON(t *testing.T) {
	var m map[string]any
	err := ExtractTo("no json here at all", &m)
	if err == nil {
		t.Error("expected error")
	}
}

// --- Real-world LLM outputs ---

// --- Edge cases: truncated/EOF ---

func TestTruncatedObject(t *testing.T) {
	input := `{"key": "val`
	mustExtract(t, "truncated object", input)
}

func TestTruncatedString(t *testing.T) {
	input := `{"key": "unterminated`
	mustExtract(t, "truncated string", input)
}

func TestTruncatedEscape(t *testing.T) {
	input := `{"key": "value\`
	mustExtract(t, "truncated escape", input)
}

func TestUnquotedValueWithFence(t *testing.T) {
	// Unquoted value ending with a code fence marker.
	input := "{\"key\": somevalue```}"
	mustExtract(t, "unquoted with fence", input)
}

func TestOnlyOpenBrace(t *testing.T) {
	input := `{`
	got := mustExtract(t, "only open brace", input)
	if got != `{}` {
		t.Errorf("expected {}, got %s", got)
	}
}

func TestOnlyOpenBracket(t *testing.T) {
	input := `[`
	got := mustExtract(t, "only open bracket", input)
	if got != `[]` {
		t.Errorf("expected [], got %s", got)
	}
}

func TestUnquotedLiteralFallback(t *testing.T) {
	// Starts with 't' but not "true".
	input := `{"key": test_value}`
	got := mustExtract(t, "literal fallback", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

// --- Escaped JSON ---

func TestEscapedJSON(t *testing.T) {
	input := `{\"key\": \"value\", \"num\": 42}`
	got := mustExtract(t, "escaped JSON", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
	if m["num"] != 42.0 {
		t.Errorf("expected 42, got %v", m["num"])
	}
}

func TestEscapedJSONNested(t *testing.T) {
	input := `{\"outer\": {\"inner\": [1, 2, 3]}}`
	got := mustExtract(t, "escaped nested", input)
	if !json.Valid([]byte(got)) {
		t.Fatalf("result is not valid JSON: %s", got)
	}
}

func TestEscapedJSONWithNewlines(t *testing.T) {
	input := `{\"key\": \"line1\\nline2\"}`
	got := mustExtract(t, "escaped with newlines", input)
	if !json.Valid([]byte(got)) {
		t.Fatalf("result is not valid JSON: %s", got)
	}
}

func TestEscapedJSONInText(t *testing.T) {
	input := "The result is: {\\\"category\\\": \\\"safe\\\", \\\"confidence\\\": 0.9}"
	got := mustExtract(t, "escaped in text", input)
	if !json.Valid([]byte(got)) {
		t.Fatalf("result is not valid JSON: %s", got)
	}
}

func TestNormalJSONNotAffectedByUnescape(t *testing.T) {
	// Normal JSON with escaped quotes inside strings should NOT be double-unescaped.
	input := `{"msg": "He said \"hello\""}`
	got := mustExtract(t, "normal escaped quotes", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

// --- Real-world LLM outputs ---

func TestLLMRealWorldOutput(t *testing.T) {
	input := `Based on my analysis of this email:

` + "```json" + `
{
  'is_suspicious': True,
  'category': 'phishing',
  'confidence': 0.92,
  'summary': 'The email contains suspicious URLs and mismatched sender information.',
  'reasons': [
    'URL points to free hosting service',
    'From/Return-Path mismatch',
  ],
  'tags': ['credential-theft', 'urgent-language'],
}
` + "```" + `

Please let me know if you need more details.`

	got := mustExtract(t, "real world", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["category"] != "phishing" {
		t.Errorf("expected phishing, got %v", m["category"])
	}
	if m["is_suspicious"] != true {
		t.Errorf("expected true, got %v", m["is_suspicious"])
	}
}

// --- Zero-width space handling ---

func TestZeroWidthSpace(t *testing.T) {
	// U+200B (zero-width space) around structural characters.
	input := "{\u200B\"key\"\u200B:\u200B\"value\"\u200B}"
	got := mustExtract(t, "zero-width space", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestBOM(t *testing.T) {
	// U+FEFF (BOM / zero-width no-break space) as leading character.
	input := "\uFEFF{\"key\": \"value\"}"
	got := mustExtract(t, "BOM", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestMongolianVowelSeparator(t *testing.T) {
	// U+180E (Mongolian vowel separator).
	input := "{\u180E\"key\":\u180E\"value\"}"
	got := mustExtract(t, "mongolian vowel separator", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
}

// --- Parenthesized prose should not hijack JSON ---

func TestParenthesizedProseBeforeJSON(t *testing.T) {
	input := "(some clarification):\n{\"key\": \"value\"}"
	got := mustExtract(t, "paren prose before json", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestParenthesizedProseBeforeFencedJSON(t *testing.T) {
	input := "(note: this is important):\n```json\n{\"result\": \"ok\"}\n```"
	got := mustExtract(t, "paren prose before fenced json", input)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("not valid JSON: %v\ngot: %s", err, got)
	}
	if m["result"] != "ok" {
		t.Errorf("expected ok, got %v", m["result"])
	}
}

func TestTupleStillWorks(t *testing.T) {
	// Actual Python tuple should still be parsed.
	input := `("a", "b", "c")`
	got := mustExtract(t, "tuple still works", input)
	var arr []string
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("not valid JSON array: %v\ngot: %s", err, got)
	}
	if len(arr) != 3 || arr[0] != "a" {
		t.Errorf("expected [a,b,c], got %v", arr)
	}
}

func TestTupleWithNumbers(t *testing.T) {
	input := `(1, 2, 3)`
	got := mustExtract(t, "tuple with numbers", input)
	var arr []any
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("not valid JSON array: %v\ngot: %s", err, got)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}

func TestTupleWithBooleans(t *testing.T) {
	input := `(true, false, null)`
	got := mustExtract(t, "tuple with booleans", input)
	var arr []any
	if err := json.Unmarshal([]byte(got), &arr); err != nil {
		t.Fatalf("not valid JSON array: %v\ngot: %s", err, got)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}

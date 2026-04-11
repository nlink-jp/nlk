package strip

import (
	"testing"
)

// --- ThinkTags ---

func TestThinkTagsDeepSeek(t *testing.T) {
	input := `<think>
Let me analyze this step by step...
First, I need to check the sender.
The domain looks suspicious.
</think>
The email appears to be a phishing attempt.`

	got := ThinkTags(input)
	want := "The email appears to be a phishing attempt."
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkTagsEmpty(t *testing.T) {
	input := "<think>\n</think>\nThe answer is 42."
	got := ThinkTags(input)
	want := "The answer is 42."
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkTagsUnclosed(t *testing.T) {
	// Model output was truncated — no closing tag.
	input := `<think>
I'm thinking about this problem...
The key consideration is`

	got := ThinkTags(input)
	if got != "" {
		t.Errorf("expected empty string for all-think content, got: %q", got)
	}
}

func TestThinkTagsUnclosedWithContent(t *testing.T) {
	input := `Some preamble
<think>
Reasoning here without closing tag`

	got := ThinkTags(input)
	want := "Some preamble"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkingTags(t *testing.T) {
	input := "<thinking>Step 1: ...</thinking>\nResult: safe"
	got := ThinkTags(input)
	want := "Result: safe"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestReasoningTags(t *testing.T) {
	input := "<reasoning>Analyzing headers...</reasoning>\n{\"safe\": true}"
	got := ThinkTags(input)
	want := `{"safe": true}`
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestReflectionTags(t *testing.T) {
	input := "<reflection>Let me reconsider...</reflection>\nFinal answer: yes"
	got := ThinkTags(input)
	want := "Final answer: yes"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkTagsCaseInsensitive(t *testing.T) {
	input := "<THINK>Uppercase thinking</THINK>\nResult here"
	got := ThinkTags(input)
	want := "Result here"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkTagsNoTags(t *testing.T) {
	input := "Just normal text without any tags."
	got := ThinkTags(input)
	if got != input {
		t.Errorf("should return input unchanged, got: %q", got)
	}
}

func TestThinkTagsEmptyInput(t *testing.T) {
	got := ThinkTags("")
	if got != "" {
		t.Errorf("expected empty, got: %q", got)
	}
}

func TestThinkTagsWithJSON(t *testing.T) {
	input := `<think>
I need to analyze this email.
The SPF check failed, which is suspicious.
Let me check the URLs...
</think>
{"is_suspicious": true, "category": "phishing", "confidence": 0.92}`

	got := ThinkTags(input)
	want := `{"is_suspicious": true, "category": "phishing", "confidence": 0.92}`
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestThinkTagsMultipleBlocks(t *testing.T) {
	input := `<think>First thought</think>
Part 1
<think>Second thought</think>
Part 2`

	got := ThinkTags(input)
	// Both think blocks should be removed.
	if contains(got, "First thought") || contains(got, "Second thought") {
		t.Errorf("think content should be removed, got: %q", got)
	}
	if !contains(got, "Part 1") || !contains(got, "Part 2") {
		t.Errorf("non-think content should be preserved, got: %q", got)
	}
}

// --- Gemma 4 format ---

func TestGemma4Thought(t *testing.T) {
	input := `<|channel>thought
I need to analyze this carefully.
The sender domain is suspicious.
<channel|>
The email is likely phishing.`

	got := ThinkTags(input)
	want := "The email is likely phishing."
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestGemma4ThoughtUnclosed(t *testing.T) {
	input := `<|channel>thought
Reasoning without closing tag`

	got := ThinkTags(input)
	if got != "" {
		t.Errorf("expected empty, got: %q", got)
	}
}

func TestGemma4ThoughtWithContent(t *testing.T) {
	input := `Preamble text
<|channel>thought
Internal reasoning
<channel|>
Final answer here`

	got := ThinkTags(input)
	if !contains(got, "Preamble text") || !contains(got, "Final answer here") {
		t.Errorf("non-thought content should be preserved, got: %q", got)
	}
	if contains(got, "Internal reasoning") {
		t.Errorf("thought content should be removed, got: %q", got)
	}
}

// --- Tags (custom tag names) ---

func TestCustomTags(t *testing.T) {
	input := "<analysis>Internal notes</analysis>\nPublic result"
	got := Tags(input, "analysis")
	want := "Public result"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestCustomTagsMultiple(t *testing.T) {
	input := "<step1>Plan</step1>\n<step2>Execute</step2>\nDone"
	got := Tags(input, "step1", "step2")
	want := "Done"
	if got != want {
		t.Errorf("got: %q\nwant: %q", got, want)
	}
}

func TestCustomTagsNoMatch(t *testing.T) {
	input := "No matching tags here"
	got := Tags(input, "nonexistent")
	if got != input {
		t.Errorf("should return input unchanged, got: %q", got)
	}
}

// --- Integration: strip then jsonfix ---

func TestStripThenJSON(t *testing.T) {
	// Simulates the recommended usage pattern:
	// raw → strip.ThinkTags() → jsonfix.Extract()
	input := `<think>
Analyzing the email headers...
SPF: fail, DKIM: pass
The URL looks suspicious.
</think>
{"category": "phishing", "confidence": 0.85}`

	stripped := ThinkTags(input)
	if contains(stripped, "<think>") {
		t.Fatal("think tags should be removed")
	}
	if !contains(stripped, `"category"`) {
		t.Fatal("JSON should be preserved")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

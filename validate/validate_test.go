package validate

import (
	"errors"
	"strings"
	"testing"
)

func TestRunAllPass(t *testing.T) {
	err := Run(
		OneOf("category", "safe", "safe", "phishing", "spam"),
		Range("confidence", 0.5, 0, 1),
		MaxLen("tags", 3, 5),
	)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRunSingleFailure(t *testing.T) {
	err := Run(
		OneOf("category", "unknown", "safe", "phishing"),
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "category") {
		t.Errorf("error should mention field name: %v", err)
	}
}

func TestRunMultipleFailures(t *testing.T) {
	err := Run(
		OneOf("category", "bad", "safe", "phishing"),
		Range("confidence", 1.5, 0, 1),
		MaxLen("tags", 10, 5),
	)
	if err == nil {
		t.Fatal("expected error")
	}
	// All three failures should be present.
	s := err.Error()
	if !strings.Contains(s, "category") {
		t.Error("missing category error")
	}
	if !strings.Contains(s, "confidence") {
		t.Error("missing confidence error")
	}
	if !strings.Contains(s, "tags") {
		t.Error("missing tags error")
	}
}

func TestRunEmpty(t *testing.T) {
	err := Run()
	if err != nil {
		t.Errorf("expected nil for no rules, got: %v", err)
	}
}

func TestErrors(t *testing.T) {
	errs := Errors(
		OneOf("a", "x", "y", "z"),
		Range("b", 0.5, 0, 1), // pass
		MaxLen("c", 10, 5),
	)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestErrorsAllPass(t *testing.T) {
	errs := Errors(
		Range("x", 0.5, 0, 1),
	)
	if errs != nil {
		t.Errorf("expected nil, got: %v", errs)
	}
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		value   string
		allowed []string
		wantErr bool
	}{
		{"safe", []string{"safe", "phishing", "spam"}, false},
		{"phishing", []string{"safe", "phishing", "spam"}, false},
		{"unknown", []string{"safe", "phishing", "spam"}, true},
		{"", []string{"safe", "phishing"}, true},
	}
	for _, tt := range tests {
		err := OneOf("field", tt.value, tt.allowed...)()
		if (err != nil) != tt.wantErr {
			t.Errorf("OneOf(%q, %v): err=%v, wantErr=%v", tt.value, tt.allowed, err, tt.wantErr)
		}
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		value   float64
		min     float64
		max     float64
		wantErr bool
	}{
		{0.5, 0, 1, false},
		{0, 0, 1, false},
		{1, 0, 1, false},
		{-0.1, 0, 1, true},
		{1.1, 0, 1, true},
	}
	for _, tt := range tests {
		err := Range("field", tt.value, tt.min, tt.max)()
		if (err != nil) != tt.wantErr {
			t.Errorf("Range(%v, [%v,%v]): err=%v, wantErr=%v", tt.value, tt.min, tt.max, err, tt.wantErr)
		}
	}
}

func TestMaxLen(t *testing.T) {
	tests := []struct {
		length  int
		max     int
		wantErr bool
	}{
		{3, 5, false},
		{5, 5, false},
		{6, 5, true},
		{0, 5, false},
	}
	for _, tt := range tests {
		err := MaxLen("field", tt.length, tt.max)()
		if (err != nil) != tt.wantErr {
			t.Errorf("MaxLen(%d, %d): err=%v, wantErr=%v", tt.length, tt.max, err, tt.wantErr)
		}
	}
}

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"hello", false},
		{"", true},
		{"  ", true},
		{"\t\n", true},
	}
	for _, tt := range tests {
		err := NotEmpty("field", tt.value)()
		if (err != nil) != tt.wantErr {
			t.Errorf("NotEmpty(%q): err=%v, wantErr=%v", tt.value, err, tt.wantErr)
		}
	}
}

func TestCustom(t *testing.T) {
	pass := Custom("field", func() error { return nil })
	if err := pass(); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}

	fail := Custom("field", func() error { return errors.New("bad value") })
	err := fail()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "field") {
		t.Error("error should contain field name")
	}
	if !strings.Contains(err.Error(), "bad value") {
		t.Error("error should contain original message")
	}
}

func TestMailAnalyzerPattern(t *testing.T) {
	// Simulate mail-analyzer's validation pattern.
	type Judgment struct {
		Category   string
		Confidence float64
		Tags       []string
		Reasons    []string
		Summary    string
	}

	j := Judgment{
		Category:   "phishing",
		Confidence: 0.87,
		Tags:       []string{"urgent", "credential-theft"},
		Reasons:    []string{"Suspicious URL", "Mismatched sender"},
		Summary:    "This email appears to be a phishing attempt.",
	}

	err := Run(
		OneOf("category", j.Category,
			"phishing", "spam", "malware-delivery", "bec", "scam", "safe"),
		Range("confidence", j.Confidence, 0, 1),
		MaxLen("tags", len(j.Tags), 5),
		MaxLen("reasons", len(j.Reasons), 5),
		NotEmpty("summary", j.Summary),
	)
	if err != nil {
		t.Errorf("valid judgment should pass: %v", err)
	}

	// Invalid judgment.
	bad := Judgment{
		Category:   "invalid",
		Confidence: 1.5,
		Tags:       []string{"a", "b", "c", "d", "e", "f"},
		Summary:    "",
	}

	err = Run(
		OneOf("category", bad.Category,
			"phishing", "spam", "malware-delivery", "bec", "scam", "safe"),
		Range("confidence", bad.Confidence, 0, 1),
		MaxLen("tags", len(bad.Tags), 5),
		NotEmpty("summary", bad.Summary),
	)
	if err == nil {
		t.Fatal("invalid judgment should fail")
	}

	errs := Errors(
		OneOf("category", bad.Category,
			"phishing", "spam", "malware-delivery", "bec", "scam", "safe"),
		Range("confidence", bad.Confidence, 0, 1),
		MaxLen("tags", len(bad.Tags), 5),
		NotEmpty("summary", bad.Summary),
	)
	if len(errs) != 4 {
		t.Errorf("expected 4 errors, got %d", len(errs))
	}
}

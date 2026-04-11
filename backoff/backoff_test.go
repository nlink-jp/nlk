package backoff

import (
	"testing"
	"time"
)

func TestDurationDefaults(t *testing.T) {
	b := New()

	// attempt 0: base * 2^0 = 5s (± jitter)
	d := b.Duration(0)
	if d < 4*time.Second || d > 6*time.Second {
		t.Errorf("attempt 0: expected ~5s, got %v", d)
	}

	// attempt 1: base * 2^1 = 10s (± jitter)
	d = b.Duration(1)
	if d < 9*time.Second || d > 11*time.Second {
		t.Errorf("attempt 1: expected ~10s, got %v", d)
	}

	// attempt 2: base * 2^2 = 20s (± jitter)
	d = b.Duration(2)
	if d < 19*time.Second || d > 21*time.Second {
		t.Errorf("attempt 2: expected ~20s, got %v", d)
	}
}

func TestDurationMaxCap(t *testing.T) {
	b := New(WithBase(10*time.Second), WithMax(30*time.Second), WithJitter(0))

	// attempt 0: 10s
	if d := b.Duration(0); d != 10*time.Second {
		t.Errorf("attempt 0: expected 10s, got %v", d)
	}

	// attempt 2: 10 * 4 = 40s → capped to 30s
	if d := b.Duration(2); d != 30*time.Second {
		t.Errorf("attempt 2: expected 30s (capped), got %v", d)
	}

	// attempt 10: should still be capped to 30s
	if d := b.Duration(10); d != 30*time.Second {
		t.Errorf("attempt 10: expected 30s (capped), got %v", d)
	}
}

func TestDurationNoJitter(t *testing.T) {
	b := New(WithBase(1*time.Second), WithMax(60*time.Second), WithJitter(0))

	for attempt := range 6 {
		d := b.Duration(attempt)
		expected := time.Duration(1<<uint(attempt)) * time.Second
		if expected > 60*time.Second {
			expected = 60 * time.Second
		}
		if d != expected {
			t.Errorf("attempt %d: expected %v, got %v", attempt, expected, d)
		}
	}
}

func TestDurationJitterRange(t *testing.T) {
	b := New(
		WithBase(10*time.Second),
		WithMax(120*time.Second),
		WithJitter(2*time.Second),
	)

	for range 100 {
		d := b.Duration(0)
		// 10s ± 2s → [8s, 12s]
		if d < 8*time.Second || d > 12*time.Second {
			t.Errorf("expected [8s, 12s], got %v", d)
		}
	}
}

func TestDurationNonNegative(t *testing.T) {
	// Large jitter relative to base could produce negative values without clamping.
	b := New(
		WithBase(100*time.Millisecond),
		WithMax(1*time.Second),
		WithJitter(500*time.Millisecond),
	)

	for range 200 {
		d := b.Duration(0)
		if d < 0 {
			t.Errorf("Duration should never be negative, got %v", d)
		}
	}
}

func TestConvenienceFunction(t *testing.T) {
	d := Duration(0)
	// Default: 5s ± 1s → [4s, 6s]
	if d < 4*time.Second || d > 6*time.Second {
		t.Errorf("Duration(0): expected ~5s, got %v", d)
	}
}

func TestDurationExponentialGrowth(t *testing.T) {
	b := New(WithBase(1*time.Second), WithMax(1000*time.Second), WithJitter(0))

	prev := b.Duration(0)
	for attempt := 1; attempt < 8; attempt++ {
		d := b.Duration(attempt)
		if d < prev {
			t.Errorf("attempt %d (%v) should be >= attempt %d (%v)", attempt, d, attempt-1, prev)
		}
		prev = d
	}
}

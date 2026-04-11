package backoff_test

import (
	"fmt"
	"time"

	"github.com/nlink-jp/nlk/backoff"
)

func ExampleDuration() {
	// Using default settings (base=5s, max=120s, jitter=1s).
	for attempt := 0; attempt < 4; attempt++ {
		d := backoff.Duration(attempt)
		fmt.Printf("attempt %d: ~%v\n", attempt, d.Round(time.Second))
	}
	// Approximate output (jitter varies):
	// attempt 0: ~5s
	// attempt 1: ~10s
	// attempt 2: ~20s
	// attempt 3: ~40s
}

func ExampleNew() {
	bo := backoff.New(
		backoff.WithBase(1*time.Second),
		backoff.WithMax(30*time.Second),
		backoff.WithJitter(0), // no jitter for deterministic output
	)

	for attempt := 0; attempt < 6; attempt++ {
		d := bo.Duration(attempt)
		fmt.Printf("attempt %d: %v\n", attempt, d)
	}
	// Output:
	// attempt 0: 1s
	// attempt 1: 2s
	// attempt 2: 4s
	// attempt 3: 8s
	// attempt 4: 16s
	// attempt 5: 30s
}

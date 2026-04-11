// Package backoff provides exponential backoff duration calculation with jitter.
//
// This package only computes wait durations — it does not perform retries
// or sleep. The calling application controls the retry loop.
//
// Usage:
//
//	for attempt := 0; attempt < 5; attempt++ {
//	    result, err := callAPI()
//	    if err == nil { break }
//	    time.Sleep(backoff.Duration(attempt))
//	}
//
// Custom configuration:
//
//	bo := backoff.New(
//	    backoff.WithBase(2*time.Second),
//	    backoff.WithMax(60*time.Second),
//	    backoff.WithJitter(500*time.Millisecond),
//	)
//	time.Sleep(bo.Duration(attempt))
package backoff

import (
	"math"
	"math/rand" // Intentionally not crypto/rand — jitter doesn't need CSPRNG.
	"time"
)

// Default values.
const (
	DefaultBase   = 5 * time.Second
	DefaultMax    = 120 * time.Second
	DefaultJitter = 1 * time.Second
)

// Backoff computes exponential backoff durations.
type Backoff struct {
	base   time.Duration
	max    time.Duration
	jitter time.Duration
}

// Option configures a Backoff.
type Option func(*Backoff)

// WithBase sets the base delay (default 5s).
func WithBase(d time.Duration) Option {
	return func(b *Backoff) { b.base = d }
}

// WithMax sets the maximum delay cap (default 120s).
func WithMax(d time.Duration) Option {
	return func(b *Backoff) { b.max = d }
}

// WithJitter sets the jitter range (default 1s). The actual jitter is
// uniformly distributed in [-jitter, +jitter].
func WithJitter(d time.Duration) Option {
	return func(b *Backoff) { b.jitter = d }
}

// New creates a Backoff with the given options.
func New(opts ...Option) Backoff {
	b := Backoff{
		base:   DefaultBase,
		max:    DefaultMax,
		jitter: DefaultJitter,
	}
	for _, o := range opts {
		o(&b)
	}
	return b
}

// Duration returns the wait duration for the given attempt (0-indexed).
//
//	delay = min(base * 2^attempt, max) + uniform(-jitter, +jitter)
//
// The result is clamped to a minimum of 0.
func (b Backoff) Duration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	exp := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(b.base) * exp)
	if delay > b.max {
		delay = b.max
	}
	if b.jitter > 0 {
		j := float64(b.jitter)
		delay += time.Duration(rand.Float64()*2*j - j)
	}
	if delay < 0 {
		delay = 0
	}
	return delay
}

// Duration is a convenience function using default settings.
func Duration(attempt int) time.Duration {
	return New().Duration(attempt)
}

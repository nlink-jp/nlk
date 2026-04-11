// Package validate provides a lightweight framework for running validation
// rules against LLM output. Rules are defined by the application; this
// package only provides the execution and error collection mechanism.
//
// Usage:
//
//	errs := validate.Run(result,
//	    validate.OneOf("category", result.Category, "safe", "phishing", "spam"),
//	    validate.Range("confidence", result.Confidence, 0, 1),
//	    validate.MaxLen("tags", len(result.Tags), 5),
//	    validate.Custom("summary", func() error {
//	        if result.Summary == "" { return errors.New("empty summary") }
//	        return nil
//	    }),
//	)
//	if errs != nil {
//	    // handle validation errors
//	}
package validate

import (
	"errors"
	"fmt"
	"strings"
)

// Rule is a validation rule. It returns nil if the value is valid,
// or an error describing the violation.
type Rule func() error

// Run executes all rules and returns a combined error if any fail.
// Returns nil if all rules pass.
func Run(rules ...Rule) error {
	var errs []string
	for _, r := range rules {
		if err := r(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

// Errors executes all rules and returns individual errors as a slice.
// Returns nil if all rules pass.
func Errors(rules ...Rule) []error {
	var errs []error
	for _, r := range rules {
		if err := r(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// --- Built-in rule constructors ---

// OneOf checks that value is one of the allowed values.
func OneOf(field string, value string, allowed ...string) Rule {
	return func() error {
		for _, a := range allowed {
			if value == a {
				return nil
			}
		}
		return fmt.Errorf("%s: %q is not one of [%s]", field, value, strings.Join(allowed, ", "))
	}
}

// Range checks that value is within [min, max].
func Range(field string, value float64, min, max float64) Rule {
	return func() error {
		if value < min || value > max {
			return fmt.Errorf("%s: %v is out of range [%v, %v]", field, value, min, max)
		}
		return nil
	}
}

// MaxLen checks that length does not exceed max.
func MaxLen(field string, length int, max int) Rule {
	return func() error {
		if length > max {
			return fmt.Errorf("%s: length %d exceeds max %d", field, length, max)
		}
		return nil
	}
}

// NotEmpty checks that value is not an empty string.
func NotEmpty(field string, value string) Rule {
	return func() error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s: must not be empty", field)
		}
		return nil
	}
}

// Custom creates a rule from an arbitrary function.
func Custom(field string, fn func() error) Rule {
	return func() error {
		if err := fn(); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		return nil
	}
}

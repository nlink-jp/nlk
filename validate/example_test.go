package validate_test

import (
	"errors"
	"fmt"

	"github.com/nlink-jp/nlk/validate"
)

func ExampleRun() {
	type Judgment struct {
		Category   string
		Confidence float64
		Tags       []string
		Summary    string
	}

	j := Judgment{
		Category:   "phishing",
		Confidence: 0.87,
		Tags:       []string{"credential-theft", "urgent"},
		Summary:    "Suspicious email with mismatched sender.",
	}

	err := validate.Run(
		validate.OneOf("category", j.Category,
			"phishing", "spam", "malware-delivery", "bec", "scam", "safe"),
		validate.Range("confidence", j.Confidence, 0, 1),
		validate.MaxLen("tags", len(j.Tags), 5),
		validate.NotEmpty("summary", j.Summary),
	)

	fmt.Println(err)
	// Output: <nil>
}

func ExampleErrors() {
	err := validate.Errors(
		validate.OneOf("status", "unknown", "active", "inactive"),
		validate.Range("score", 1.5, 0, 1),
	)

	for _, e := range err {
		fmt.Println(e)
	}
	// Output:
	// status: "unknown" is not one of [active, inactive]
	// score: 1.5 is out of range [0, 1]
}

func ExampleCustom() {
	startDate := "2026-04-01"
	endDate := "2026-03-01"

	rule := validate.Custom("dates", func() error {
		if endDate < startDate {
			return errors.New("end date is before start date")
		}
		return nil
	})

	err := validate.Run(rule)
	fmt.Println(err)
	// Output: dates: end date is before start date
}

package jsonfix_test

import (
	"fmt"

	"github.com/nlink-jp/nlk/jsonfix"
)

func ExampleExtract_markdownFence() {
	raw := "Here is the result:\n```json\n{\"status\": \"ok\"}\n```"
	result, _ := jsonfix.Extract(raw)
	fmt.Println(result)
	// Output: {"status":"ok"}
}

func ExampleExtract_singleQuotes() {
	raw := "{'key': 'value', 'active': True}"
	result, _ := jsonfix.Extract(raw)
	fmt.Println(result)
	// Output: {"key":"value","active":true}
}

func ExampleExtract_trailingComma() {
	raw := `{"items": ["a", "b", "c",], "count": 3,}`
	result, _ := jsonfix.Extract(raw)
	fmt.Println(result)
	// Output: {"items":["a","b","c"],"count":3}
}

func ExampleExtract_escapedJSON() {
	raw := `{\"name\": \"Alice\", \"age\": 30}`
	result, _ := jsonfix.Extract(raw)
	fmt.Println(result)
	// Output: {"name":"Alice","age":30}
}

func ExampleExtractTo() {
	type Result struct {
		Category   string  `json:"category"`
		Confidence float64 `json:"confidence"`
	}

	raw := "```json\n{'category': 'safe', 'confidence': 0.95,}\n```"
	var r Result
	_ = jsonfix.ExtractTo(raw, &r)
	fmt.Printf("%s (%.0f%%)\n", r.Category, r.Confidence*100)
	// Output: safe (95%)
}

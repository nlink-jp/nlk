package strip_test

import (
	"fmt"

	"github.com/nlink-jp/nlk/strip"
)

func ExampleThinkTags() {
	raw := `<think>
Let me analyze this step by step...
The URL looks suspicious.
</think>
{"is_suspicious": true, "category": "phishing"}`

	cleaned := strip.ThinkTags(raw)
	fmt.Println(cleaned)
	// Output: {"is_suspicious": true, "category": "phishing"}
}

func ExampleThinkTags_gemma4() {
	raw := `<|channel>thought
Internal reasoning about the problem.
<channel|>
The answer is 42.`

	cleaned := strip.ThinkTags(raw)
	fmt.Println(cleaned)
	// Output: The answer is 42.
}

func ExampleTags() {
	raw := "<internal>Private notes</internal>\nPublic response here."
	cleaned := strip.Tags(raw, "internal")
	fmt.Println(cleaned)
	// Output: Public response here.
}

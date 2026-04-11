// Package strip removes LLM thinking/reasoning tags from model output.
//
// Many LLM models (DeepSeek R1, Qwen QwQ/3, Phi-4, Gemma 4, and most
// open-source reasoning models) emit internal reasoning wrapped in special
// tags. When running these models locally, these tags appear in the text
// output and need to be removed before further processing.
//
// Cloud APIs (Claude, Gemini, OpenAI) separate thinking from the response
// at the API level, so stripping is not needed for those.
//
// Usage:
//
//	cleaned := strip.ThinkTags(llmOutput)
//	// or with custom tag names:
//	cleaned := strip.Tags(llmOutput, "think", "reasoning", "reflection")
package strip

import (
	"strings"
)

// ThinkTags removes all known thinking/reasoning tag patterns from text.
//
// Supported patterns:
//   - <think>...</think> (DeepSeek R1, Qwen, Phi-4, most OSS models)
//   - <thinking>...</thinking>
//   - <reasoning>...</reasoning>
//   - <reflection>...</reflection>
//   - <|channel>thought\n...<channel|> (Gemma 4)
//   - Empty tags: <think>\n</think>
//   - Unclosed tags: <think>... (rest of text consumed)
func ThinkTags(text string) string {
	// Standard XML-style tags.
	result := Tags(text, "think", "thinking", "reasoning", "reflection")

	// Gemma 4 channel format.
	result = stripGemma4Thought(result)

	return result
}

// Tags removes XML-style tag pairs and their content from text.
// For each tag name, removes all occurrences of <name>...</name>.
// Also handles unclosed tags (removes from <name> to end of text).
func Tags(text string, tagNames ...string) string {
	result := text
	for _, name := range tagNames {
		result = stripXMLTag(result, name)
	}
	return result
}

// stripXMLTag removes all occurrences of <name>...</name> from text.
// Handles: normal pairs, empty tags, unclosed tags, case-insensitive matching.
func stripXMLTag(text, name string) string {
	openTag := "<" + name + ">"
	closeTag := "</" + name + ">"

	for {
		// Find opening tag (case-insensitive).
		openIdx := indexCI(text, openTag)
		if openIdx < 0 {
			break
		}

		// Find closing tag after the opening tag.
		searchFrom := openIdx + len(openTag)
		closeIdx := indexCI(text[searchFrom:], closeTag)

		if closeIdx < 0 {
			// Unclosed tag — remove from opening tag to end of text.
			text = strings.TrimSpace(text[:openIdx])
			break
		}

		// Remove the tag pair and its content.
		endIdx := searchFrom + closeIdx + len(closeTag)
		text = text[:openIdx] + text[endIdx:]
	}

	return strings.TrimSpace(text)
}

// stripGemma4Thought removes Gemma 4's channel-based thought format:
// <|channel>thought\n...<channel|>
func stripGemma4Thought(text string) string {
	const openTag = "<|channel>thought"
	const closeTag = "<channel|>"

	for {
		openIdx := strings.Index(text, openTag)
		if openIdx < 0 {
			break
		}

		searchFrom := openIdx + len(openTag)
		closeIdx := strings.Index(text[searchFrom:], closeTag)

		if closeIdx < 0 {
			// Unclosed — remove to end.
			text = strings.TrimSpace(text[:openIdx])
			break
		}

		endIdx := searchFrom + closeIdx + len(closeTag)
		text = text[:openIdx] + text[endIdx:]
	}

	return strings.TrimSpace(text)
}

// indexCI returns the index of the first case-insensitive occurrence of substr in s.
// Returns -1 if not found.
func indexCI(s, substr string) int {
	lower := strings.ToLower(s)
	target := strings.ToLower(substr)
	return strings.Index(lower, target)
}

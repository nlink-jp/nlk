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
// Note: the input is fully loaded into memory. Callers should limit input size
// before calling ThinkTags if processing untrusted or unbounded data.
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
//
// Skips occurrences inside markdown inline-code spans
// (single-backtick on the same line, e.g. `<think>`). LLM responses
// that EXPLAIN the literal tag — common when a user asks "what is
// <think>?" — would otherwise have everything from the literal
// `<think>` to end-of-text stripped under the unclosed-tag rule,
// truncating the explanation mid-sentence. See nlk-rfp note.
//
// Out of scope: triple-backtick fenced code blocks, HTML <code>
// blocks. These are uncommon in LLM output and adding their
// detection requires multi-line state. Add when a real symptom
// motivates it.
func stripXMLTag(text, name string) string {
	openTag := "<" + name + ">"
	closeTag := "</" + name + ">"

	scanFrom := 0
	for {
		// Find opening tag (case-insensitive), skipping past any
		// occurrence inside an inline-code span on its line.
		rel := indexCI(text[scanFrom:], openTag)
		if rel < 0 {
			break
		}
		openIdx := scanFrom + rel

		if isInsideInlineCodeSpan(text, openIdx) {
			// Skip just this occurrence. Advance past the open
			// tag literal so the next scan finds the next match.
			scanFrom = openIdx + len(openTag)
			continue
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
		// Resume scanning from the cut point; new content might
		// have surfaced another open tag.
		scanFrom = openIdx
	}

	return strings.TrimSpace(text)
}

// isInsideInlineCodeSpan reports whether pos in text falls inside a
// markdown single-backtick inline code span on its own line.
// Detection: count backticks on the same line strictly before pos;
// odd count means an unclosed span is currently open. Lines are
// delimited by '\n'.
//
// This deliberately doesn't try to model double-backtick spans,
// fenced code blocks, or anything that crosses line boundaries.
// The narrow rule covers the practical case (LLM writing
// "`<think>` は内部思考マーカー") with no false positives in normal
// prose.
func isInsideInlineCodeSpan(text string, pos int) bool {
	if pos <= 0 || pos > len(text) {
		return false
	}
	lineStart := strings.LastIndexByte(text[:pos], '\n') + 1 // 0 if not found
	count := 0
	for i := lineStart; i < pos; i++ {
		if text[i] == '`' {
			count++
		}
	}
	return count%2 == 1
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

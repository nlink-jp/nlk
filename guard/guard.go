// Package guard provides nonce-tagged XML wrapping for prompt injection defense.
//
// Untrusted data (user input, external content) is wrapped in XML tags
// containing a cryptographic nonce, making it physically distinct from
// system instructions in the prompt.
//
// Usage:
//
//	tag := guard.NewTag()                     // generate random tag
//	wrapped := tag.Wrap(untrustedData)        // <user_data_a1b2c3d4>...</user_data_a1b2c3d4>
//	prompt := tag.Expand(systemPrompt)        // replace {{DATA_TAG}} with tag name
package guard

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// DefaultPlaceholder is the placeholder string replaced by [Tag.Expand].
const DefaultPlaceholder = "{{DATA_TAG}}"

// Tag represents a nonce-based XML tag for isolating untrusted data.
type Tag struct {
	name string
}

// NonceSize is the number of random bytes used for tag nonces.
// 16 bytes = 128 bits of entropy, sufficient to prevent brute-force guessing.
const NonceSize = 16

// NewTag generates a new Tag with a cryptographically random nonce.
// The tag name has the form "user_data_{32 hex chars}" (16 random bytes).
func NewTag() Tag {
	return NewTagWithPrefix("user_data")
}

// NewTagWithPrefix generates a new Tag with the given prefix and a random nonce.
// The tag name has the form "{prefix}_{32 hex chars}".
func NewTagWithPrefix(prefix string) Tag {
	b := make([]byte, NonceSize)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("guard: crypto/rand failed: %v", err))
	}
	return Tag{name: prefix + "_" + hex.EncodeToString(b)}
}

// NewTagWithName creates a Tag with a specific name (for testing).
func NewTagWithName(name string) Tag {
	return Tag{name: name}
}

// Name returns the tag name (e.g. "user_data_a1b2c3d4").
func (t Tag) Name() string {
	return t.name
}

// Wrap encloses data in XML tags: <tagname>data</tagname>.
func (t Tag) Wrap(data string) string {
	return "<" + t.name + ">" + data + "</" + t.name + ">"
}

// Expand replaces DefaultPlaceholder in the template with the tag name.
func (t Tag) Expand(template string) string {
	return strings.ReplaceAll(template, DefaultPlaceholder, t.name)
}

// ExpandPlaceholder replaces a custom placeholder in the template with the tag name.
func (t Tag) ExpandPlaceholder(template, placeholder string) string {
	return strings.ReplaceAll(template, placeholder, t.name)
}

package jsonfix

// Recursive descent JSON parser with repair capabilities.
// Inspired by Python's json-repair, adapted for Go.
//
// The parser walks the input character by character following JSON BNF grammar.
// When it encounters invalid syntax, it applies heuristic repairs:
//   - Single quotes → double quotes
//   - Unquoted keys → quoted keys
//   - Trailing commas → removed
//   - Missing commas → inserted
//   - Comments (// and /* */) → removed
//   - TRUE/FALSE/NULL → true/false/null
//   - Missing closing braces/brackets → appended
//   - Unescaped quotes in strings → escaped
//   - Python tuples () → arrays []
//   - Ellipsis ... → removed

import (
	"strings"
	"unicode"
)

type parser struct {
	input []rune
	pos   int
	out   strings.Builder
}

func newParser(input string) *parser {
	return &parser{input: []rune(input), pos: 0}
}

func (p *parser) repair() string {
	p.skipNonJSON()
	if p.pos >= len(p.input) {
		return ""
	}
	p.parseValue()
	return p.out.String()
}

// --- Navigation helpers ---

func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) peekAt(offset int) rune {
	i := p.pos + offset
	if i >= len(p.input) || i < 0 {
		return 0
	}
	return p.input[i]
}

func (p *parser) advance() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	ch := p.input[p.pos]
	p.pos++
	return ch
}

func (p *parser) emit(r rune) {
	p.out.WriteRune(r)
}

func (p *parser) emitString(s string) {
	p.out.WriteString(s)
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.input)
}

// --- Whitespace and comment handling ---

func (p *parser) skipWhitespace() {
	for !p.atEnd() && unicode.IsSpace(p.peek()) {
		p.advance()
	}
}

func (p *parser) skipWhitespaceAndComments() {
	for {
		p.skipWhitespace()
		if p.atEnd() {
			return
		}
		if p.peek() == '/' && p.peekAt(1) == '/' {
			// Line comment.
			p.advance()
			p.advance()
			for !p.atEnd() && p.peek() != '\n' {
				p.advance()
			}
			continue
		}
		if p.peek() == '/' && p.peekAt(1) == '*' {
			// Block comment.
			p.advance()
			p.advance()
			for !p.atEnd() {
				if p.peek() == '*' && p.peekAt(1) == '/' {
					p.advance()
					p.advance()
					break
				}
				p.advance()
			}
			continue
		}
		if p.peek() == '#' {
			// Hash comment.
			p.advance()
			for !p.atEnd() && p.peek() != '\n' {
				p.advance()
			}
			continue
		}
		break
	}
}

// skipNonJSON skips leading non-JSON content (explanatory text, code fences).
// Only stops at characters that unambiguously start a JSON structure: { [ or (
func (p *parser) skipNonJSON() {
	for !p.atEnd() {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			return
		}

		// Markdown code fence.
		if p.matchPrefix("```") {
			p.pos += 3
			for !p.atEnd() && p.peek() != '\n' {
				p.advance()
			}
			if !p.atEnd() {
				p.advance()
			}
			continue
		}

		ch := p.peek()

		// Unambiguous JSON structure starts.
		if ch == '{' || ch == '[' || ch == '(' {
			return
		}

		// Closing code fence — skip and continue.
		if p.matchPrefix("```") {
			p.pos += 3
			continue
		}

		// Any other character — skip.
		p.advance()
	}
}

func (p *parser) matchPrefix(s string) bool {
	runes := []rune(s)
	for i, r := range runes {
		if p.peekAt(i) != r {
			return false
		}
	}
	return true
}

// --- Value parsing ---

func (p *parser) parseValue() {
	p.skipWhitespaceAndComments()
	if p.atEnd() {
		return
	}

	ch := p.peek()
	switch {
	case ch == '{':
		p.parseObject()
	case ch == '[':
		p.parseArray()
	case ch == '(':
		// Python tuple → array.
		p.parseArray()
	case ch == '"':
		p.parseString('"')
	case ch == '\'':
		// Single-quoted string → double-quoted.
		p.parseString('\'')
	case ch == 't', ch == 'f', ch == 'n':
		p.parseLiteral()
	case ch == 'T', ch == 'F', ch == 'N':
		p.parseLiteralUppercase()
	case ch == '-' || (ch >= '0' && ch <= '9') || ch == '.':
		p.parseNumber()
	default:
		// Unquoted string value — read until a structural character.
		p.parseUnquotedValue()
	}
}

// --- Object parsing ---

func (p *parser) parseObject() {
	p.advance() // consume '{'
	p.emit('{')

	first := true
	for {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			p.emit('}')
			return
		}

		ch := p.peek()

		// End of object.
		if ch == '}' {
			p.advance()
			p.emit('}')
			return
		}

		// Skip trailing comma before '}'.
		if ch == ',' {
			p.advance()
			p.skipWhitespaceAndComments()
			if p.atEnd() || p.peek() == '}' {
				if !p.atEnd() {
					p.advance()
				}
				p.emit('}')
				return
			}
			if !first {
				p.emit(',')
			}
		} else if !first {
			// Missing comma between entries.
			p.emit(',')
		}

		first = false

		// Parse key.
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			p.emit('}')
			return
		}

		p.parseKey()

		// Expect colon.
		p.skipWhitespaceAndComments()
		if !p.atEnd() && p.peek() == ':' {
			p.advance()
		}
		p.emit(':')

		// Parse value.
		p.skipWhitespaceAndComments()
		if p.atEnd() || p.peek() == '}' || p.peek() == ',' {
			// Missing value — emit empty string.
			p.emitString(`""`)
			continue
		}
		p.parseValue()
	}
}

func (p *parser) parseKey() {
	ch := p.peek()
	if ch == '"' {
		p.parseString('"')
	} else if ch == '\'' {
		p.parseString('\'')
	} else {
		// Unquoted key.
		p.parseUnquotedKey()
	}
}

func (p *parser) parseUnquotedKey() {
	p.emit('"')
	for !p.atEnd() {
		ch := p.peek()
		if ch == ':' || ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '}' {
			break
		}
		p.emit(p.advance())
	}
	p.emit('"')
}

// --- Array parsing ---

func (p *parser) parseArray() {
	open := p.advance() // consume '[' or '('
	p.emit('[')

	closing := ']'
	if open == '(' {
		closing = ')'
	}

	first := true
	for {
		p.skipWhitespaceAndComments()
		if p.atEnd() {
			p.emit(']')
			return
		}

		ch := p.peek()

		// End of array.
		if ch == rune(closing) || ch == ']' {
			p.advance()
			p.emit(']')
			return
		}

		// Skip trailing comma.
		if ch == ',' {
			p.advance()
			p.skipWhitespaceAndComments()
			if p.atEnd() || p.peek() == rune(closing) || p.peek() == ']' {
				if !p.atEnd() {
					p.advance()
				}
				p.emit(']')
				return
			}
			if !first {
				p.emit(',')
			}
		} else if !first {
			// Missing comma.
			p.emit(',')
		}

		first = false

		// Handle ellipsis (...) — skip without emitting.
		p.skipWhitespaceAndComments()
		if p.matchPrefix("...") {
			p.pos += 3
			// Undo the comma we just emitted for this element.
			s := p.out.String()
			if len(s) > 0 && s[len(s)-1] == ',' {
				p.out.Reset()
				p.out.WriteString(s[:len(s)-1])
			}
			first = true // reset so next iteration doesn't add comma
			continue
		}

		p.parseValue()
	}
}

// --- String parsing ---

func (p *parser) parseString(quote rune) {
	p.advance() // consume opening quote
	p.emit('"')

	for !p.atEnd() {
		ch := p.peek()

		// Closing quote.
		if ch == quote {
			p.advance()
			p.emit('"')
			return
		}

		// Newline inside string — close it.
		if ch == '\n' || ch == '\r' {
			p.emit('"')
			return
		}

		// Backslash escape.
		if ch == '\\' {
			p.advance()
			if p.atEnd() {
				p.emit('"')
				return
			}
			next := p.advance()
			switch next {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				p.emit('\\')
				p.emit(next)
			case '\'':
				// \' → just '
				p.emit('\'')
			case 'u':
				// Unicode escape — pass through.
				p.emit('\\')
				p.emit('u')
				for i := 0; i < 4 && !p.atEnd(); i++ {
					p.emit(p.advance())
				}
			case 'x':
				// Hex escape \xNN → \u00NN.
				p.emitString("\\u00")
				for i := 0; i < 2 && !p.atEnd(); i++ {
					p.emit(p.advance())
				}
			default:
				// Unknown escape — emit character without backslash.
				p.emit(next)
			}
			continue
		}

		// If we're in a double-quoted string and encounter an unescaped double quote
		// that's not the closing quote, escape it.
		if quote == '"' && ch == '"' {
			// Look ahead to see if this is really the end.
			// If next non-space char is : , } ] or EOF, it's the real end.
			next := p.peekAheadSkipSpace(1)
			if next == ':' || next == ',' || next == '}' || next == ']' || next == 0 {
				p.advance()
				p.emit('"')
				return
			}
			// Otherwise, escape it.
			p.advance()
			p.emitString(`\"`)
			continue
		}

		p.emit(p.advance())
	}

	// Unterminated string — close it.
	p.emit('"')
}

func (p *parser) peekAheadSkipSpace(offset int) rune {
	i := p.pos + offset
	for i < len(p.input) && unicode.IsSpace(p.input[i]) {
		i++
	}
	if i >= len(p.input) {
		return 0
	}
	return p.input[i]
}

// --- Unquoted value parsing ---

func (p *parser) parseUnquotedValue() {
	// Read until a structural character, then emit as a quoted string.
	var buf strings.Builder
	for !p.atEnd() {
		ch := p.peek()
		if ch == ',' || ch == '}' || ch == ']' || ch == ':' || ch == '\n' {
			break
		}
		buf.WriteRune(p.advance())
	}

	val := strings.TrimSpace(buf.String())
	if val == "" {
		p.emitString(`""`)
		return
	}

	// Strip trailing code fence if present.
	if strings.HasSuffix(val, "```") {
		val = strings.TrimSuffix(val, "```")
		val = strings.TrimSpace(val)
	}

	// Emit as quoted string.
	p.emit('"')
	for _, r := range val {
		if r == '"' {
			p.emitString(`\"`)
		} else {
			p.emit(r)
		}
	}
	p.emit('"')
}

// --- Literal parsing ---

func (p *parser) parseLiteral() {
	if p.matchPrefix("true") {
		p.pos += 4
		p.emitString("true")
	} else if p.matchPrefix("false") {
		p.pos += 5
		p.emitString("false")
	} else if p.matchPrefix("null") {
		p.pos += 4
		p.emitString("null")
	} else {
		// Unknown — treat as start of unquoted string.
		p.parseUnquotedValue()
	}
}

func (p *parser) parseLiteralUppercase() {
	if p.matchPrefixCI("true") {
		p.pos += 4
		p.emitString("true")
	} else if p.matchPrefixCI("false") {
		p.pos += 5
		p.emitString("false")
	} else if p.matchPrefixCI("null") {
		p.pos += 4
		p.emitString("null")
	} else if p.matchPrefixCI("none") {
		// Python None → null.
		p.pos += 4
		p.emitString("null")
	} else {
		p.parseUnquotedValue()
	}
}

func (p *parser) matchPrefixCI(s string) bool {
	runes := []rune(s)
	for i, r := range runes {
		ch := p.peekAt(i)
		if unicode.ToLower(ch) != unicode.ToLower(r) {
			return false
		}
	}
	// Ensure it's a complete token (next char is not alphanumeric).
	next := p.peekAt(len(runes))
	return next == 0 || !unicode.IsLetter(next)
}

// --- Number parsing ---

func (p *parser) parseNumber() {
	// Leading dot: .5 → 0.5
	if p.peek() == '.' {
		p.emit('0')
	}

	// Negative sign.
	if p.peek() == '-' {
		p.emit(p.advance())
	}

	hasDigit := false

	// Integer part.
	for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
		p.emit(p.advance())
		hasDigit = true
		// Skip underscores in numbers (e.g., 1_000_000).
		if !p.atEnd() && p.peek() == '_' {
			p.advance()
		}
	}

	// Decimal part.
	if !p.atEnd() && p.peek() == '.' {
		next := p.peekAt(1)
		if next >= '0' && next <= '9' {
			p.emit(p.advance()) // '.'
			for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
				p.emit(p.advance())
				if !p.atEnd() && p.peek() == '_' {
					p.advance()
				}
			}
		} else if hasDigit {
			// Trailing dot: 1. → 1.0
			p.advance()
			p.emitString(".0")
		}
	}

	// Exponent.
	if !p.atEnd() && (p.peek() == 'e' || p.peek() == 'E') {
		p.emit(p.advance())
		if !p.atEnd() && (p.peek() == '+' || p.peek() == '-') {
			p.emit(p.advance())
		}
		for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
			p.emit(p.advance())
		}
	}

	// If no digits were emitted, treat as unquoted value.
	if !hasDigit && p.out.Len() == 0 {
		p.parseUnquotedValue()
	}
}

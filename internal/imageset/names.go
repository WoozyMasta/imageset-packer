package imageset

import "strings"

// NormalizeName converts an arbitrary string to a safe imageset identifier.
// Output contains only 0-9, a-z, A-Z and '_' characters.
// If camel is true, tokens are joined as CamelCase; otherwise snake_case.
func NormalizeName(input string, camel bool) string {
	tokens := splitTokens(input)
	if len(tokens) == 0 {
		return ""
	}

	if camel {
		var b strings.Builder
		for _, t := range tokens {
			b.WriteString(toCamelToken(t))
		}

		return b.String()
	}

	return strings.Join(tokens, "_")
}

// splitTokens splits the input string into tokens.
func splitTokens(input string) []string {
	var tokens []string
	var buf []rune

	flush := func() {
		if len(buf) == 0 {
			return
		}

		tokens = append(tokens, strings.ToLower(string(buf)))
		buf = buf[:0]
	}

	for _, r := range input {
		if isAsciiAlphaNum(r) {
			buf = append(buf, r)
		} else {
			flush()
		}
	}

	flush()

	return tokens
}

// toCamelToken converts a token to CamelCase.
func toCamelToken(token string) string {
	if token == "" {
		return ""
	}

	var b strings.Builder
	for i, r := range token {
		if i == 0 {
			b.WriteRune(toUpperAscii(r))
		} else {
			b.WriteRune(toLowerAscii(r))
		}
	}

	return b.String()
}

// isAsciiAlphaNum checks if a rune is an ASCII alpha-numeric character.
func isAsciiAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// toLowerAscii converts an ASCII uppercase letter to lowercase.
func toLowerAscii(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}

	return r
}

// toUpperAscii converts an ASCII lowercase letter to uppercase.
func toUpperAscii(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}

	return r
}

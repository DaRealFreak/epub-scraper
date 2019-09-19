package unicode

import (
	"strings"
	"unicode"
)

// SanitizeSpaces replaces all space characters with normal spaces
// since f.e. regex doesn't match \s with the non breaking space (0x00A0)
func SanitizeSpaces(s string) string {
	const normalSpace = '\u0020'
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return normalSpace
		}
		return r
	}, s)
}

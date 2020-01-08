package unicode

import (
	"strings"
)

// SanitizeSpaces replaces NBSP with normal spaces (0x20) since f.e. regex \s doesn't match with NBSP (0xA0)
func SanitizeSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		// replace NBSP with normal space due to display problems in some e-book readers
		if uint32(r) == 0xA0 {
			return 0x20
		}

		return r
	}, s)
}

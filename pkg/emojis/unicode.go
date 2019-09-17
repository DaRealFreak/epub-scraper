package emojis

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ToDo:
// - extract to custom project to use in more projects
// - allow loading from local sources
// - allow/disallow by selection version
// - allow/disallow by byte size
// - return error if unicode codes couldn't be loaded instead of failing silently
// - make failing silently optional
// - add configurability before loading/updating the unicode codes

// UnicodeEmojiDataURL is the emoji data we want to parse
const UnicodeEmojiDataURL = "https://unicode.org/Public/emoji/latest/emoji-data.txt"

// nolint: gochecknoglobals
var (
	currentUnicodeEmojiRegexpPattern = getCurrentUnicodeEmojiPattern()
	// AllowedEmojiCodes contains the allowed emojis, which are by default:
	// "#", "*", "[0-9]", "©", "®", "‼", "™"
	AllowedEmojiCodes = []string{"0023", "002A", "0030..0039", "00A9", "00AE", "2122"}
)

// getCurrentUnicodeEmojiPattern fetches the latest emoji data from the official unicode page
// and parses them into a regular expression, on errors it return nil
func getCurrentUnicodeEmojiPattern() *string {
	res, err := http.Get(UnicodeEmojiDataURL)
	if err != nil {
		log.Warningf(
			"could not retrieve emoji data from %s, emojis will not get replaced", UnicodeEmojiDataURL,
		)
		return nil
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warningf(
			"could not read response from %s, emojis will not get replaced", UnicodeEmojiDataURL,
		)
		return nil
	}

	var emojiUnicodeValues []string
	// match [4 bytes] and [4 bytes .. 4 bytes]
	unicodeEmojiLines := regexp.MustCompile(`(?m)^([0-9A-F]{4,5}(\.\.[0-9A-F]{4,5})?)\s+;`)
	for _, line := range strings.Split(string(content), "\n") {
		matches := unicodeEmojiLines.FindStringSubmatch(line)
		if len(matches) > 1 && matches[1] != "" {
			if !isEmojiCodeAllowed(matches[1]) {
				emojiUnicodeValues = append(emojiUnicodeValues, matches[1])
			}
		}
	}

	var emojiUnicodeRegexPattern string
	for _, emojiUnicode := range emojiUnicodeValues {
		if strings.Contains(emojiUnicode, "..") {
			emojiUnicodeRange := strings.Split(emojiUnicode, "..")
			emojiUnicodeRegexPattern += fmt.Sprintf(`\x{%s}-\x{%s}`, emojiUnicodeRange[0], emojiUnicodeRange[1])
		} else {
			emojiUnicodeRegexPattern += fmt.Sprintf(`\x{%s}`, emojiUnicode)
		}
	}
	emojiUnicodeRegexPattern = fmt.Sprintf(`[%s]`, emojiUnicodeRegexPattern)
	if _, err := regexp.Compile(emojiUnicodeRegexPattern); err != nil {
		log.Warningf(
			"could not compile generated emoji regular expression, emojis will not get replaced",
		)
		return nil
	}
	return &emojiUnicodeRegexPattern
}

// isEmojiCodeAllowed checks the whitelist for allowed unicode emojis
func isEmojiCodeAllowed(unicodeCode string) bool {
	for _, allowedUnicodeCode := range AllowedEmojiCodes {
		if allowedUnicodeCode == unicodeCode {
			return true
		}
	}
	return false
}

// ReplaceUnicodeEmojis replaces all unicode emojis from the passed subject with the passed replacement
func ReplaceUnicodeEmojis(subject string, repl string) string {
	if currentUnicodeEmojiRegexpPattern == nil {
		return subject
	}
	pattern := regexp.MustCompile(*currentUnicodeEmojiRegexpPattern)
	return pattern.ReplaceAllString(subject, repl)
}

// StripUnicodeEmojis strips all unicode emojis from the passed subject
func StripUnicodeEmojis(subject string) string {
	return ReplaceUnicodeEmojis(subject, "")
}

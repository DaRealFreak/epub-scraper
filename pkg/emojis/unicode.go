package emojis

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// UnicodeEmojiDataURL is the emoji data we want to parse
const UnicodeEmojiDataURL = "https://unicode.org/Public/emoji/latest/emoji-data.txt"

var (
	currentUnicodeEmojiRegexpPattern = getCurrentUnicodeEmojiPattern()
)

// getCurrentUnicodeEmojiPattern fetches the latest emoji data from the official unicode page
// and parses them into a regular expression, on errors it return nil
func getCurrentUnicodeEmojiPattern() *string {
	res, err := http.Get(UnicodeEmojiDataURL)
	if err != nil {
		logrus.Warningf(
			"could not retrieve emoji data from %s, emojis will not get replaced", UnicodeEmojiDataURL,
		)
		return nil
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warningf(
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
			emojiUnicodeValues = append(emojiUnicodeValues, matches[1])
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
		logrus.Warningf(
			"could not compile generated emoji regular expression, emojis will not get replaced",
		)
		return nil
	}
	return &emojiUnicodeRegexPattern
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

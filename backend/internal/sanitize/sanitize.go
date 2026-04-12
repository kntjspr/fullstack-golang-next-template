package sanitize

import "regexp"

var (
	scriptTagPattern = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	htmlTagPattern   = regexp.MustCompile(`(?s)<[^>]+>`)
)

// StripHTML removes script blocks and HTML tags from input text.
func StripHTML(s string) string {
	withoutScripts := scriptTagPattern.ReplaceAllString(s, "")
	return htmlTagPattern.ReplaceAllString(withoutScripts, "")
}

// TruncateString limits a string to max runes.
func TruncateString(s string, max int) string {
	if max <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	return string(runes[:max])
}

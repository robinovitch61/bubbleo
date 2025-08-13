package linebuffer

import "regexp"

// HighlightData contains information about what to highlight in each item in the viewport.
type HighlightData struct {
	StringToHighlight       string
	RegexPatternToHighlight *regexp.Regexp
	IsRegex                 bool
}

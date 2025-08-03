package viewport

// Configuration consolidates all configuration options for the viewport
type Configuration struct {
	// wrapText is true if the viewport wraps text rather than showing that a line is truncated/horizontally scrollable
	WrapText bool

	// footerEnabled is true if the viewport will show the footer when it overflows
	FooterEnabled bool

	// continuationIndicator is the string to use to indicate that a line has been truncated from the left or right
	ContinuationIndicator string
}

func NewConfiguration() *Configuration {
	return &Configuration{
		WrapText:              false,
		FooterEnabled:         true,
		ContinuationIndicator: "...",
	}
}

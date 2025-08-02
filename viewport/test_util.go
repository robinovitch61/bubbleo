package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
	"strings"
)

// RenderableString is a convenience type
type RenderableString struct {
	LineBuffer linebuffer.LineBufferer
}

func (r RenderableString) Render() linebuffer.LineBufferer {
	return r.LineBuffer
}

// RenderableStringCompareFn is a comparator function for renderableString
func RenderableStringCompareFn(a, b RenderableString) bool {
	if a.LineBuffer == nil || b.LineBuffer == nil {
		return false
	}
	return a.LineBuffer.Content() == b.LineBuffer.Content()
}

// assert RenderableString implements viewport.Renderable
var _ Renderable = RenderableString{}

// pad is a test helper function that pads the given lines to the given width and height.
// for example, pad(5, 4, []string{"a", "b", "c"}) will be padded to:
// "a    "
// "b    "
// "c    "
// "     "
// as a single string
func pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

func setContent(vp *Model[RenderableString], content []string) {
	renderableStrings := make([]RenderableString, len(content))
	for i := range content {
		renderableStrings[i] = RenderableString{LineBuffer: linebuffer.New(content[i])}
	}
	vp.SetContent(renderableStrings)
}

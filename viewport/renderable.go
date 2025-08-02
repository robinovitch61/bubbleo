package viewport

import "github.com/robinovitch61/bubbleo/viewport/linebuffer"

type Renderable interface {
	Render() linebuffer.LineBufferer
}

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

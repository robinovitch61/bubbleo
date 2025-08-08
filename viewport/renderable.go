package viewport

import "github.com/robinovitch61/bubbleo/viewport/linebuffer"

// Renderable represents objects that can be rendered as line buffers.
type Renderable interface {
	Render() linebuffer.LineBufferer
}

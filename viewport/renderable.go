package viewport

import "github.com/robinovitch61/bubbleo/viewport/linebuffer"

type Renderable interface {
	Render() linebuffer.LineBufferer
}

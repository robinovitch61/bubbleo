package viewport

// ContentManager manages the actual content and selection state
type ContentManager[T Renderable] struct {
	// Items is the complete list of items to be rendered in the viewport
	Items []T

	// Header is the fixed header lines at the top of the viewport
	// these lines wrap and are horizontally scrollable similar to other rendered items
	Header []string

	// selectedIdx is the index of Items of the current selection (only relevant when selection is enabled)
	selectedIdx int

	// StringToHighlight is a string to highlight in the viewport wherever it shows up, even wrapped between lines
	// within the same item
	StringToHighlight string

	// CompareFn is an optional function to compare items for maintaining the selection when content changes
	// if set, the viewport will try to maintain the previous selected item when content changes
	CompareFn CompareFn[T]
}

func NewContentManager[T Renderable]() *ContentManager[T] {
	return &ContentManager[T]{
		Items:       []T{},
		Header:      []string{},
		selectedIdx: 0,
	}
}

func (cm *ContentManager[T]) SetSelectedIdx(idx int) {
	cm.selectedIdx = clampValMinMax(idx, 0, len(cm.Items)-1)
}

func (cm *ContentManager[T]) GetSelectedIdx() int {
	return cm.selectedIdx
}

func (cm *ContentManager[T]) GetSelectedItem() *T {
	if cm.selectedIdx >= len(cm.Items) || cm.selectedIdx < 0 {
		return nil
	}
	return &cm.Items[cm.selectedIdx]
}

func (cm *ContentManager[T]) NumItems() int {
	return len(cm.Items)
}

func (cm *ContentManager[T]) IsEmpty() bool {
	return len(cm.Items) == 0
}

func (cm *ContentManager[T]) ValidateSelectedIdx() {
	if len(cm.Items) == 0 {
		cm.selectedIdx = 0
		return
	}
	cm.selectedIdx = clampValMinMax(cm.selectedIdx, 0, len(cm.Items)-1)
}


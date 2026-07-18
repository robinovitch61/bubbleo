package viewport

import "testing"

func TestViewport_ItemMetrics_Top(t *testing.T) {
	w, h := 30, 2
	vp := newViewport(w, h, WithFooterEnabled[object](false))
	setContent(vp, []string{"line 1", "line 2", "line 3"})

	metrics := vp.GetItemMetrics()
	if metrics.TotalItems != 3 {
		t.Errorf("GetItemMetrics returned %d items, expected %d", metrics.TotalItems, 3)
	}
	if metrics.FirstVisibleItemIdx != 0 {
		t.Errorf("GetItemMetrics returned FirstVisibleItemIdx of %d, expected %d", metrics.FirstVisibleItemIdx, 0)
	}
	if metrics.LastVisibleItemIdx != 1 {
		t.Errorf("GetItemMetrics returned LastVisibleItemIdx of %d, expected %d", metrics.LastVisibleItemIdx, 1)
	}
}

func TestViewport_ItemMetrics_Bottom(t *testing.T) {
	w, h := 30, 2
	vp := newViewport(w, h, WithFooterEnabled[object](false))
	setContent(vp, []string{"line 1", "line 2", "line 3"})

	vp.ScrollDown(2)

	metrics := vp.GetItemMetrics()
	if metrics.TotalItems != 3 {
		t.Errorf("GetItemMetrics returned %d items, expected %d", metrics.TotalItems, 3)
	}
	if metrics.FirstVisibleItemIdx != 2 {
		t.Errorf("GetItemMetrics returned FirstVisibleItemIdx of %d, expected %d", metrics.FirstVisibleItemIdx, 2)
	}
	if metrics.LastVisibleItemIdx != 2 {
		t.Errorf("GetItemMetrics returned LastVisibleItemIdx of %d, expected %d", metrics.LastVisibleItemIdx, 2)
	}
}

func TestViewport_ItemMetrics_Middle(t *testing.T) {
	w, h := 30, 2
	vp := newViewport(w, h, WithFooterEnabled[object](false))
	setContent(vp, []string{"line 1", "line 2", "line 3"})

	vp.ScrollDown(1)

	metrics := vp.GetItemMetrics()
	if metrics.TotalItems != 3 {
		t.Errorf("GetItemMetrics returned %d items, expected %d", metrics.TotalItems, 3)
	}
	if metrics.FirstVisibleItemIdx != 1 {
		t.Errorf("GetItemMetrics returned FirstVisibleItemIdx of %d, expected %d", metrics.FirstVisibleItemIdx, 1)
	}
	if metrics.LastVisibleItemIdx != 2 {
		t.Errorf("GetItemMetrics returned LastVisibleItemIdx of %d, expected %d", metrics.LastVisibleItemIdx, 2)
	}
}

func TestViewport_ItemMetrics_NoVisibleItems(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		content    []string
		setHeader  bool
		totalItems int
	}{
		{
			name:       "empty content",
			width:      30,
			height:     2,
			content:    []string{},
			totalItems: 0,
		},
		{
			name:       "zero width",
			width:      0,
			height:     2,
			content:    []string{"line 1"},
			totalItems: 1,
		},
		{
			name:       "header consumes height",
			width:      30,
			height:     1,
			content:    []string{"line 1"},
			setHeader:  true,
			totalItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := newViewport(tt.width, tt.height, WithFooterEnabled[object](false))
			setContent(vp, tt.content)
			if tt.setHeader {
				vp.SetHeader([]string{"header"})
			}

			metrics := vp.GetItemMetrics()
			if metrics.TotalItems != tt.totalItems {
				t.Errorf("GetItemMetrics returned %d items, expected %d", metrics.TotalItems, tt.totalItems)
			}
			if metrics.FirstVisibleItemIdx != -1 {
				t.Errorf("GetItemMetrics returned FirstVisibleItemIdx of %d, expected -1", metrics.FirstVisibleItemIdx)
			}
			if metrics.LastVisibleItemIdx != -1 {
				t.Errorf("GetItemMetrics returned LastVisibleItemIdx of %d, expected -1", metrics.LastVisibleItemIdx)
			}
		})
	}
}

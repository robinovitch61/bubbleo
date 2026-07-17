package viewport

import (
	"testing"
)

func TestItemMetrics0Percent(t *testing.T) {
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

func TestItemMetrics100Percent(t *testing.T) {
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

func TestItemMetrics50Percent(t *testing.T) {
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

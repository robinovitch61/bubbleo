package viewport

import (
	"fmt"
	"github.com/robinovitch61/bubbleo/dev"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	black = lipgloss.Color("#000000")
	blue  = lipgloss.Color("6")
	pink  = lipgloss.Color("#E760FC")
	grey  = lipgloss.Color("#737373")
)

var (
	regular                         = lipgloss.NewStyle()
	bold                            = regular.Copy().Bold(true)
	defaultViewportHeaderStyle      = bold.Copy()
	defaultViewportSelectedRowStyle = regular.Copy().Foreground(black).Background(blue)
	defaultViewportHighlightStyle   = regular.Copy().Foreground(black).Background(pink)
	defaultContentStyle             = regular.Copy()
	defaultViewportFooterStyle      = regular.Copy().Foreground(grey)
)

type Renderable interface {
	Render() string
}

type ContentProvider[T Renderable] interface {
	TakeOneAt(idx int) T
	TakeN(start, n int) []T
	Len() int
}

// Model represents a viewport component
type Model[T Renderable] struct {
	KeyMap                    KeyMap
	LineContinuationIndicator string
	BackgroundStyle           lipgloss.Style
	HeaderStyle               lipgloss.Style
	SelectedContentStyle      lipgloss.Style
	HighlightStyle            lipgloss.Style
	ContentStyle              lipgloss.Style
	FooterStyle               lipgloss.Style

	header            []string
	wrappedHeader     []string
	contentProvider   ContentProvider[T]
	stringToHighlight string
	wrapText          bool
	selectionEnabled  bool

	// TODO
	// when need more lines above or below, refresh and adjust yOffset to allow scrolling line by line even when wrapped
	minContentIdxInLines int
	maxContentIdxInLines int
	lines                []string
	topLineIdx           int // top of screen line index

	// width is the width of the entire viewport in terminal columns
	width int
	// height is the height of the entire viewport in terminal rows
	height int
	// contentHeight is the height of the viewport in terminal rows, excluding the header and footer
	contentHeight int
	// maxLineLength is the maximum line length in terminal characters across header and currently visible visibleContent
	maxLineLength int

	// selectedContentIdx is the current selection's visibleContent index if selectionEnabled is true.
	// m.contentProvider.TakeOneAt(m.selectedContentIdx) gives the currently selected row, regardless of wrapping
	selectedContentIdx int

	//// yOffsetContentIdx is the index of the first item from the contentProvider shown on screen
	//// m.contentProvider.TakeOneAt(m.yOffsetContentIdx) gives the first row shown, regardless of wrapping
	//// it updates every time the screen scrolls up or down due to selection overflow or panning actions
	//yOffsetContentIdx int

	// xOffset is the number of columns scrolled right when visibleContent lines overflow the viewport and wrapText is false
	xOffset int

	//// visibleContent is the non-wrapped rendered array of currently visible lines. It updates every time yOffsetContentIdx changes
	//visibleContent []string
	//
	//// wrappedVisibleContent is the wrapped rendered array of currently visible lines. It updates every time yOffsetContentIdx changes
	//wrappedVisibleContent []string

	// lineIdxToContentIdx maps the item at an index of wrappedVisibleContent to the index of visibleContent it is
	// associated with (many wrappedVisibleContent indexes -> one visibleContent index). It updates every time yOffsetContentIdx changes
	lineIdxToContentIdx map[int]int

	// contentIdxToFirstLineIdx maps the item at an index of visibleContent to the first index of wrappedVisibleContent it
	// is associated with (index of visibleContent -> first index of wrappedVisibleContent). It updates every time yOffsetContentIdx changes
	contentIdxToFirstLineIdx map[int]int

	// contentIdxToNumLines maps the item at an index of visibleContent to its wrapped height in terminal rows
	// It updates every time yOffsetContentIdx changes
	contentIdxToNumLines map[int]int
}

// New creates a new viewport model with reasonable defaults
func New[T Renderable](width, height int, contentProvider ContentProvider[T]) (m Model[T]) {
	m.contentProvider = contentProvider
	m.setWidthAndHeight(width, height)
	m.updateContentHeight()

	m.selectionEnabled = false
	m.wrapText = false

	m.KeyMap = DefaultKeyMap()
	m.LineContinuationIndicator = "..."
	m.BackgroundStyle = regular
	m.HeaderStyle = defaultViewportHeaderStyle
	m.SelectedContentStyle = defaultViewportSelectedRowStyle
	m.HighlightStyle = defaultViewportHighlightStyle
	m.ContentStyle = defaultContentStyle
	m.FooterStyle = defaultViewportFooterStyle
	return m
}

// Update processes messages and updates the model
func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			if m.selectionEnabled {
				m.selectedContentIdxUp(1)
			} else {
				m.viewUp(1)
			}

		case key.Matches(msg, m.KeyMap.Down):
			if m.selectionEnabled {
				m.selectedContentIdxDown(1)
			} else {
				m.viewDown(1)
			}

		case key.Matches(msg, m.KeyMap.Left):
			if !m.wrapText {
				m.viewLeft(m.width / 4)
			}

		case key.Matches(msg, m.KeyMap.Right):
			if !m.wrapText {
				m.viewRight(m.width / 4)
			}

		case key.Matches(msg, m.KeyMap.HalfPageUp):
			offset := max(1, m.getNumVisibleItems()/2)
			m.viewUp(m.contentHeight / 2)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			offset := max(1, m.getNumVisibleItems()/2)
			m.viewDown(m.contentHeight / 2)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.PageUp):
			offset := m.getNumVisibleItems()
			m.viewUp(m.contentHeight)
			if m.selectionEnabled {
				m.selectedContentIdxUp(offset)
			}

		case key.Matches(msg, m.KeyMap.PageDown):
			offset := m.getNumVisibleItems()
			m.viewDown(m.contentHeight)
			if m.selectionEnabled {
				m.selectedContentIdxDown(offset)
			}

		case key.Matches(msg, m.KeyMap.Top):
			if m.selectionEnabled {
				m.SetSelectedContentIdx(0)
			} else {

				//m.setYOffset(0)
			}

		case key.Matches(msg, m.KeyMap.Bottom):
			maxYOffset := m.maxYOffset()
			if m.selectionEnabled {
				m.selectedContentIdxDown(maxYOffset)
			} else {
				m.viewDown(maxYOffset)
			}
		}
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m Model[T]) View() string {
	var viewString string

	footerString, footerHeight := m.getFooter()

	addLineToViewString := func(line string) {
		viewString += line + "\n"
	}

	header := m.getHeader()
	visibleLines := m.getVisibleLines()

	for _, headerLine := range header {
		headerViewLine := m.getVisiblePartOfLine(headerLine)
		addLineToViewString(m.HeaderStyle.Render(headerViewLine))
	}

	hasNoHighlight := stringWidth(m.stringToHighlight) == 0
	for idx, line := range visibleLines {
		contentIdx := m.getContentIdx(m.topLineIdx + idx)
		isSelected := m.selectionEnabled && contentIdx == m.selectedContentIdx

		lineStyle := m.ContentStyle
		if isSelected {
			lineStyle = m.SelectedContentStyle
		}
		contentViewLine := m.getVisiblePartOfLine(line)

		if isSelected && contentViewLine == "" {
			contentViewLine = " "
		}

		if hasNoHighlight {
			addLineToViewString(lineStyle.Render(contentViewLine))
		} else {
			// this splitting and rejoining of styled visibleContent is expensive and causes increased flickering,
			// so only do it if something is actually highlighted
			highlightStyle := m.HighlightStyle
			lineChunks := strings.Split(contentViewLine, m.stringToHighlight)
			var styledChunks []string
			for _, chunk := range lineChunks {
				styledChunks = append(styledChunks, lineStyle.Render(chunk))
			}
			addLineToViewString(strings.Join(styledChunks, highlightStyle.Render(m.stringToHighlight)))
		}
	}

	if footerHeight > 0 {
		// pad so footer shows up at bottom
		padCount := max(0, m.contentHeight-len(visibleLines)-footerHeight)
		viewString += strings.Repeat("\n", padCount)
		viewString += footerString
	}
	renderedViewString := m.BackgroundStyle.Width(m.width).Height(m.height).Render(viewString)

	return renderedViewString
}

// SetContentProvider sets the visibleContent provider, the selectable set of lines in the viewport
func (m *Model[T]) SetContentProvider(contentProvider ContentProvider[T]) {
	m.contentProvider = contentProvider
	m.refreshLines()
	//m.updateForHeaderAndContent()
	m.fixSelection()
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	m.selectionEnabled = selectionEnabled
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m Model[T]) GetSelectionEnabled() bool {
	return m.selectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model[T]) SetWrapText(wrapText bool) {
	m.wrapText = wrapText
	m.updateWrapText()
}

// GetWrapText returns whether the viewport wraps text
func (m Model[T]) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.setWidthAndHeight(width, m.height)
	//m.updateForHeaderAndContent()
}

// GetWidth returns the viewport's width
func (m Model[T]) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthAndHeight(m.width, height)
	//m.updateForHeaderAndContent()
}

// GetHeight returns the viewport's height
func (m Model[T]) GetHeight() int {
	return m.height
}

// SetSelectedContentIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedContentIdx(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelectedIdx := m.maxContentIdx(); n > maxSelectedIdx {
		m.selectedContentIdx = maxSelectedIdx
	} else {
		m.selectedContentIdx = max(0, n)
	}

	//currentLineIdx := m.getCurrentLineIdx()
	//lastVisibleLineIdx := m.lastVisibleLineIdx()
	//offScreenRowCount := currentLineIdx - lastVisibleLineIdx
	//if offScreenRowCount >= 0 || m.lastContentItemSelected() {
	//	heightOffset := m.contentIdxToNumLines[m.selectedContentIdx] - 1
	//	if !m.wrapText {
	//		heightOffset = 0
	//	}
	//	m.viewDown(offScreenRowCount + heightOffset)
	//} else if currentLineIdx < m.yOffset {
	//	m.viewUp(m.yOffset - currentLineIdx)
	//}
	//
	//if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
	//	m.setYOffset(maxYOffset)
	//}

	//m.fixViewForSelection()
}

// GetSelectedContentIdx returns the currently selected visibleContent index
func (m Model[T]) GetSelectedContentIdx() int {
	return m.selectedContentIdx
}

// GetSelectedContent returns the currently selected visibleContent, or nil if there is none
func (m Model[T]) GetSelectedContent() *T {
	if m.selectedContentIdx >= m.contentProvider.Len() || m.selectedContentIdx < 0 {
		return nil
	}
	selectedContent := m.contentProvider.TakeOneAt(m.selectedContentIdx)
	return &selectedContent
}

// SetStringToHighlight sets a string to highlight in the viewport
func (m *Model[T]) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.header = header
	m.updateWrappedHeader()
	//m.updateForHeaderAndContent()
}

// ResetHorizontalOffset resets the horizontal offset to the leftmost position
func (m *Model[T]) ResetHorizontalOffset() {
	m.xOffset = 0
}

// ScrollToTop scrolls the viewport to the top
func (m *Model[T]) ScrollToTop() {
	m.selectedContentIdxUp(m.selectedContentIdx)
	m.viewUp(m.selectedContentIdx)
}

// ScrollToBottom scrolls the viewport to the bottom
func (m *Model[T]) ScrollToBottom() {
	m.selectedContentIdxDown(m.contentProvider.Len())
	m.viewDown(m.contentProvider.Len())
}

func (m *Model[T]) setXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) updateWrappedHeader() {
	var allWrappedHeader []string
	for _, line := range m.header {
		wrappedLinesForLine := m.getWrappedLines(line)
		allWrappedHeader = append(allWrappedHeader, wrappedLinesForLine...)
	}
	m.wrappedHeader = allWrappedHeader
}

func (m *Model[T]) refreshLines() {
	var lines []string
	lineIdxToContentIdx := make(map[int]int)
	contentIdxToFirstLineIdx := make(map[int]int)
	contentIdxToNumLines := make(map[int]int)

	startContentIdx := m.minContentIdxInLines
	endContentIdx := min(startContentIdx+m.contentHeight, m.contentProvider.Len())

	// TODO: if wrap is on, getting more lines than needed here
	lineIdx := 0
	for i, item := range m.contentProvider.TakeN(startContentIdx, endContentIdx-startContentIdx) {
		contentIdx := startContentIdx + i
		rendered := item.Render()
		linesForContent := []string{rendered}
		if m.wrapText {
			linesForContent = m.getWrappedLines(rendered)
		}

		lines = append(lines, linesForContent...)

		for range linesForContent {
			lineIdxToContentIdx[lineIdx] = contentIdx
			if _, exists := contentIdxToFirstLineIdx[contentIdx]; !exists {
				contentIdxToFirstLineIdx[contentIdx] = lineIdx
			}
			lineIdx++
		}

		contentIdxToNumLines[contentIdx] = len(linesForContent)
	}

	m.lines = lines
	dev.Debug(fmt.Sprintf("len lines: %d", len(m.lines)))
	m.lineIdxToContentIdx = lineIdxToContentIdx
	dev.Debug(fmt.Sprintf("lineIdxToContentIdx: %+v", m.lineIdxToContentIdx))
	dev.Debug("")
	m.contentIdxToFirstLineIdx = contentIdxToFirstLineIdx
	dev.Debug(fmt.Sprintf("contentIdxToFirstLineIdx: %+v", m.contentIdxToFirstLineIdx))
	dev.Debug("")
	m.contentIdxToNumLines = contentIdxToNumLines
	dev.Debug(fmt.Sprintf("contentIdxToNumLines: %+v", m.contentIdxToNumLines))
	dev.Debug("")

	m.updateMaxVisibleLineLength()
}

//func (m *Model[T]) updateForHeaderAndContent() {
//	m.updateContentHeight()
//	//m.fixViewForSelection()
//	m.updateMaxVisibleLineLength()
//}

func (m *Model[T]) updateWrapText() {
	// header/footer height could have changed
	m.updateContentHeight()
	m.refreshLines()
	m.ResetHorizontalOffset()
	//m.fixViewForSelection()
}

func (m *Model[T]) updateMaxVisibleLineLength() {
	m.maxLineLength = 0
	header, content := m.getHeader(), m.getVisibleLines()
	for _, line := range append(header, content...) {
		if lineLength := stringWidth(line); lineLength > m.maxLineLength {
			m.maxLineLength = lineLength
		}
	}
}

func (m *Model[T]) setWidthAndHeight(width, height int) {
	m.width, m.height = width, height
	m.updateWrappedHeader()
	m.refreshLines()
}

func (m *Model[T]) fixSelection() {
	if !m.selectionEnabled {
		return
	}
	if m.selectedContentIdx > m.maxContentIdx() {
		m.selectedContentIdx = 0
	}
}

func (m *Model[T]) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

//func (m *Model[T]) setYOffset(n int) {
//if maxYOffset := m.maxYOffset(); n > maxYOffset {
//	m.yOffsetContentIdx = maxYOffset
//} else {
//	m.yOffsetContentIdx = max(0, n)
//}
//m.refreshLines()
//}

func (m *Model[T]) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx + n)
}

func (m *Model[T]) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx - n)
}

func (m *Model[T]) viewDown(n int) {
	m.setYOffset(m.yOffsetContentIdx + n)
}

func (m *Model[T]) viewUp(n int) {
	m.setYOffset(m.yOffsetContentIdx - n)
}

func (m *Model[T]) viewLeft(n int) {
	m.setXOffset(m.xOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.setXOffset(m.xOffset + n)
}

func (m Model[T]) getHeader() []string {
	if m.wrapText {
		return m.wrappedHeader
	}
	return m.header
}

// // getContentStrings returns the rendered lines with wrapping if enabled
//
//	func (m Model[T]) getContentStrings(start, end int) []string {
//		if start >= m.contentProvider.Len() {
//			return []string{}
//		}
//		if end < 0 {
//			return []string{}
//		}
//		start = max(0, start)
//		end = min(end, m.contentProvider.Len())
//
//		var contentStrings []string
//		for _, item := range m.contentProvider.TakeN(start, end-start) {
//			rendered := item.Render()
//			if m.wrapText {
//				contentStrings = append(contentStrings, m.getWrappedLines(rendered)...)
//			} else {
//				contentStrings = append(contentStrings, item.Render())
//			}
//		}
//		return contentStrings[:end-start]
//	}
//
//// maxYOffset returns the maximum yOffsetContentIdx (the yOffsetContentIdx that shows the final screen)
//func (m Model[T]) maxYOffset() int {
//	dev.Debug(fmt.Sprintf("m.contentProvider.Len(): %d", m.contentProvider.Len()))
//	res := max(0, m.contentProvider.Len()-m.contentHeight)
//	dev.Debug(fmt.Sprintf("maxYoffset: %d", res))
//	return res
//}

func (m Model[T]) maxContentIdx() int {
	return m.contentProvider.Len() - 1
}

// getVisibleLines retrieves the visible visibleContent based on the yOffsetContentIdx and contentHeight
func (m Model[T]) getVisibleLines() []string {
	//return m.getContentStrings(m.yOffsetContentIdx, m.yOffsetContentIdx+m.contentHeight)
}

func (m Model[T]) getVisiblePartOfLine(line string) string {
	var lenLineContinuationIndicator = stringWidth(m.LineContinuationIndicator)
	rightTrimmedLineLength := stringWidth(strings.TrimRight(line, " "))
	end := min(stringWidth(line), m.xOffset+m.width)
	start := min(end, m.xOffset)
	line = line[start:end]
	if m.xOffset+m.width < rightTrimmedLineLength {
		truncate := max(0, stringWidth(line)-lenLineContinuationIndicator)
		line = line[:truncate] + m.LineContinuationIndicator
	}
	if m.xOffset > 0 {
		line = m.LineContinuationIndicator + line[min(stringWidth(line), lenLineContinuationIndicator):]
	}
	return line
}

func (m Model[T]) getContentIdx(lineIdx int) int {
	return m.lineIdxToContentIdx[lineIdx]
}

func (m Model[T]) getCurrentLineIdx() int {
	if m.wrapText {
		return m.contentIdxToFirstLineIdx[m.selectedContentIdx]
	}
	return m.selectedContentIdx
}

func (m Model[T]) getWrappedLines(line string) []string {
	if stringWidth(line) < m.width {
		return []string{line}
	}
	line = strings.TrimRight(line, " ")
	return splitLineIntoSizedChunks(line, m.width)
}

func (m Model[T]) getNumVisibleItems() int {
	if !m.wrapText {
		return m.contentHeight
	}

	var itemCount int
	var rowCount int
	contentIdx := m.lineIdxToContentIdx[m.yOffsetContentIdx]
	for rowCount < m.contentHeight {
		if height, exists := m.contentIdxToNumLines[contentIdx]; exists {
			rowCount += height
		} else {
			break
		}
		contentIdx++
		itemCount++
	}
	return itemCount
}

func (m Model[T]) lastContentItemSelected() bool {
	return m.selectedContentIdx == m.contentProvider.Len()-1
}

func (m Model[T]) getFooter() (string, int) {
	numerator := m.selectedContentIdx + 1
	denominator := m.contentProvider.Len()

	// if selection is disabled, percentage should show from the bottom of the visible visibleContent
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffsetContentIdx + m.contentHeight
	}

	if denominator > m.contentHeight {
		footerString := fmt.Sprintf("%d%% (%d/%d)", percent(numerator, denominator), numerator, denominator)
		renderedFooterString := m.FooterStyle.Copy().MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

func splitLineIntoSizedChunks(line string, chunkSize int) []string {
	var wrappedLines []string
	for {
		lineWidth := stringWidth(line)
		if lineWidth == 0 {
			break
		}

		width := chunkSize
		if lineWidth < chunkSize {
			width = lineWidth
		}

		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}
	return wrappedLines
}

// stringWidth is a function in case in the future something like utf8.RuneCountInString or lipgloss.Width is better
func stringWidth(s string) int {
	return len(s)
}

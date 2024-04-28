package viewport

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbles/key"
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

// Model represents a viewport component
type Model struct {
	KeyMap                    KeyMap
	LineContinuationIndicator string
	HeaderStyle               lipgloss.Style
	SelectedContentStyle      lipgloss.Style
	HighlightStyle            lipgloss.Style
	ContentStyle              lipgloss.Style
	FooterStyle               lipgloss.Style

	header         []string
	wrappedHeader  []string
	content        []string
	wrappedContent []string

	// wrappedContentIdxToContentIdx maps the item at an index of wrappedContent to the index of content it is associated with (many wrappedContent indexes -> one content index)
	wrappedContentIdxToContentIdx map[int]int

	// contentIdxToFirstWrappedContentIdx maps the item at an index of content to the first index of wrappedContent it is associated with (index of content -> first index of wrappedContent)
	contentIdxToFirstWrappedContentIdx map[int]int

	// contentIdxToHeight maps the item at an index of content to its wrapped height in terminal rows
	contentIdxToHeight map[int]int

	// selectedContentIdx is the index of content of the currently selected item when selectionEnabled is true
	selectedContentIdx int
	stringToHighlight  string
	selectionEnabled   bool
	wrapText           bool

	// width is the width of the entire viewport in terminal columns
	width int
	// height is the height of the entire viewport in terminal rows
	height int
	// contentHeight is the height of the viewport in terminal rows, excluding the header and footer
	contentHeight int
	// maxLineLength is the maximum line length in terminal characters across header and visible content
	maxLineLength int

	// yOffset is the index of the first row shown on screen - wrappedContent[yOffset] if wrapText, otherwise content[yOffset]
	yOffset int
	// xOffset is the number of columns scrolled right when content lines overflow the viewport and wrapText is false
	xOffset int
}

// New creates a new viewport model with reasonable defaults
func New(width, height int) (m Model) {
	m.setWidthAndHeight(width, height)
	m.updateContentHeight()

	m.selectionEnabled = false
	m.wrapText = false

	m.KeyMap = DefaultKeyMap()
	m.LineContinuationIndicator = "..."
	m.HeaderStyle = defaultViewportHeaderStyle
	m.SelectedContentStyle = defaultViewportSelectedRowStyle
	m.HighlightStyle = defaultViewportHighlightStyle
	m.ContentStyle = defaultContentStyle
	m.FooterStyle = defaultViewportFooterStyle
	return m
}

// Update processes messages and updates the model
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
				m.selectedContentIdxUp(m.yOffset + m.contentHeight)
			} else {
				m.viewUp(m.yOffset + m.contentHeight)
			}

		case key.Matches(msg, m.KeyMap.Bottom):
			if m.selectionEnabled {
				m.selectedContentIdxDown(m.maxVisibleLineIdx())
			} else {
				m.viewDown(m.maxVisibleLineIdx())
			}
		}
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m Model) View() string {
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
		contentIdx := m.getContentIdx(m.yOffset + idx)
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
			// this splitting and rejoining of styled content is expensive and causes increased flickering,
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
	renderedViewString := regular.Width(m.width).Height(m.height).Render(viewString)

	return renderedViewString
}

// SetContent sets the content, the selectable set of lines in the viewport
func (m *Model) SetContent(content []string) {
	m.content = content
	m.updateWrappedContent()
	m.updateForHeaderAndContent()
	m.fixSelection()
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model) SetSelectionEnabled(selectionEnabled bool) {
	m.selectionEnabled = selectionEnabled
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m Model) GetSelectionEnabled() bool {
	return m.selectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model) SetWrapText(wrapText bool) {
	m.wrapText = wrapText
	m.updateWrapText()
}

// GetWrapText returns whether the viewport wraps text
func (m Model) GetWrapText() bool {
	return m.wrapText
}

// SetWidth sets the viewport's width
func (m *Model) SetWidth(width int) {
	m.setWidthAndHeight(width, m.height)
	m.updateForHeaderAndContent()
}

// GetWidth returns the viewport's width
func (m Model) GetWidth() int {
	return m.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model) SetHeight(height int) {
	m.setWidthAndHeight(m.width, height)
	m.updateForHeaderAndContent()
}

// GetHeight returns the viewport's height
func (m Model) GetHeight() int {
	return m.height
}

// SetSelectedContentIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model) SetSelectedContentIdx(n int) {
	if m.contentHeight == 0 {
		return
	}

	if maxSelectedIdx := m.maxContentIdx(); n > maxSelectedIdx {
		m.selectedContentIdx = maxSelectedIdx
	} else {
		m.selectedContentIdx = max(0, n)
	}

	m.fixViewForSelection()
}

// GetSelectedContentIdx returns the currently selected content index
func (m Model) GetSelectedContentIdx() int {
	return m.selectedContentIdx
}

// SetStringToHighlight sets a string to highlight in the viewport
func (m *Model) SetStringToHighlight(h string) {
	m.stringToHighlight = h
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model) SetHeader(header []string) {
	m.header = header
	m.updateWrappedHeader()
	m.updateForHeaderAndContent()
}

// ResetHorizontalOffset resets the horizontal offset to the leftmost position
func (m *Model) ResetHorizontalOffset() {
	m.xOffset = 0
}

// ScrollToTop scrolls the viewport to the top
func (m *Model) ScrollToTop() {
	m.selectedContentIdxUp(m.selectedContentIdx)
	m.viewUp(m.selectedContentIdx)
}

// ScrollToBottom scrolls the viewport to the bottom
func (m *Model) ScrollToBottom() {
	m.selectedContentIdxDown(len(m.content))
	m.viewDown(len(m.content))
}

func (m *Model) setXOffset(n int) {
	maxXOffset := m.maxLineLength - m.width
	m.xOffset = max(0, min(maxXOffset, n))
}

func (m *Model) updateWrappedHeader() {
	var allWrappedHeader []string
	for _, line := range m.header {
		wrappedLinesForLine := m.getWrappedLines(line)
		allWrappedHeader = append(allWrappedHeader, wrappedLinesForLine...)
	}
	m.wrappedHeader = allWrappedHeader
}

func (m *Model) updateWrappedContent() {
	var allWrappedContent []string
	wrappedContentIdxToContentIdx := make(map[int]int)
	contentIdxToFirstWrappedContentIdx := make(map[int]int)
	contentIdxToHeight := make(map[int]int)

	var wrappedContentIdx int
	for contentIdx, line := range m.content {
		wrappedLinesForLine := m.getWrappedLines(line)
		contentIdxToHeight[contentIdx] = len(wrappedLinesForLine)
		for _, wrappedLine := range wrappedLinesForLine {
			allWrappedContent = append(allWrappedContent, wrappedLine)

			wrappedContentIdxToContentIdx[wrappedContentIdx] = contentIdx
			if _, exists := contentIdxToFirstWrappedContentIdx[contentIdx]; !exists {
				contentIdxToFirstWrappedContentIdx[contentIdx] = wrappedContentIdx
			}

			wrappedContentIdx++
		}
	}
	m.wrappedContent = allWrappedContent
	m.wrappedContentIdxToContentIdx = wrappedContentIdxToContentIdx
	m.contentIdxToFirstWrappedContentIdx = contentIdxToFirstWrappedContentIdx
	m.contentIdxToHeight = contentIdxToHeight
}

func (m *Model) updateForHeaderAndContent() {
	m.updateContentHeight()
	m.fixViewForSelection()
	m.updateMaxVisibleLineLength()
}

func (m *Model) updateWrapText() {
	m.updateContentHeight()
	m.updateWrappedContent()
	m.ResetHorizontalOffset()
	m.fixViewForSelection()
	m.updateMaxVisibleLineLength()
}

func (m *Model) updateMaxVisibleLineLength() {
	m.maxLineLength = 0
	header, content := m.getHeader(), m.getVisibleLines()
	for _, line := range append(header, content...) {
		if lineLength := stringWidth(line); lineLength > m.maxLineLength {
			m.maxLineLength = lineLength
		}
	}
}

func (m *Model) setWidthAndHeight(width, height int) {
	m.width, m.height = width, height
	m.updateWrappedHeader()
	m.updateWrappedContent()
}

func (m *Model) fixViewForSelection() {
	currentLineIdx := m.getCurrentLineIdx()
	lastVisibleLineIdx := m.lastVisibleLineIdx()
	offScreenRowCount := currentLineIdx - lastVisibleLineIdx
	if offScreenRowCount >= 0 || m.lastContentItemSelected() {
		heightOffset := m.contentIdxToHeight[m.selectedContentIdx] - 1
		if !m.wrapText {
			heightOffset = 0
		}
		m.viewDown(offScreenRowCount + heightOffset)
	} else if currentLineIdx < m.yOffset {
		m.viewUp(m.yOffset - currentLineIdx)
	}

	if maxYOffset := m.maxYOffset(); m.yOffset > maxYOffset {
		m.setYOffset(maxYOffset)
	}
}

func (m *Model) fixSelection() {
	if !m.selectionEnabled {
		return
	}
	if m.selectedContentIdx > m.maxContentIdx() {
		m.selectedContentIdx = 0
	}
}

func (m *Model) updateContentHeight() {
	_, footerHeight := m.getFooter()
	contentHeight := m.height - len(m.getHeader()) - footerHeight
	m.contentHeight = max(0, contentHeight)
}

func (m *Model) setYOffset(n int) {
	if maxYOffset := m.maxYOffset(); n > maxYOffset {
		m.yOffset = maxYOffset
	} else {
		m.yOffset = max(0, n)
	}
	m.updateMaxVisibleLineLength()
}

func (m *Model) selectedContentIdxDown(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx + n)
}

func (m *Model) selectedContentIdxUp(n int) {
	m.SetSelectedContentIdx(m.selectedContentIdx - n)
}

func (m *Model) viewDown(n int) {
	m.setYOffset(m.yOffset + n)
}

func (m *Model) viewUp(n int) {
	m.setYOffset(m.yOffset - n)
}

func (m *Model) viewLeft(n int) {
	m.setXOffset(m.xOffset - n)
}

func (m *Model) viewRight(n int) {
	m.setXOffset(m.xOffset + n)
}

func (m Model) getHeader() []string {
	if m.wrapText {
		return m.wrappedHeader
	}
	return m.header
}

func (m Model) getContent() []string {
	if m.wrapText {
		return m.wrappedContent
	}
	return m.content
}

// lastVisibleLineIdx returns the maximum visible line index
func (m Model) lastVisibleLineIdx() int {
	return min(m.maxVisibleLineIdx(), m.yOffset+m.contentHeight-1)
}

// maxYOffset returns the maximum yOffset (the yOffset that shows the final screen)
func (m Model) maxYOffset() int {
	if m.maxVisibleLineIdx() < m.contentHeight {
		return 0
	}
	return len(m.getContent()) - m.contentHeight
}

func (m *Model) maxVisibleLineIdx() int {
	return len(m.getContent()) - 1
}

func (m Model) maxContentIdx() int {
	return len(m.content) - 1
}

// getVisibleLines retrieves the visible content based on the yOffset and contentHeight
func (m Model) getVisibleLines() []string {
	maxVisibleLineIdx := m.maxVisibleLineIdx()
	start := max(0, min(maxVisibleLineIdx, m.yOffset))
	end := start + m.contentHeight
	if end > maxVisibleLineIdx {
		return m.getContent()[start:]
	}
	return m.getContent()[start:end]
}

func (m Model) getVisiblePartOfLine(line string) string {
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

func (m Model) getContentIdx(wrappedContentIdx int) int {
	if !m.wrapText {
		return wrappedContentIdx
	}
	return m.wrappedContentIdxToContentIdx[wrappedContentIdx]
}

func (m Model) getCurrentLineIdx() int {
	if m.wrapText {
		return m.contentIdxToFirstWrappedContentIdx[m.selectedContentIdx]
	}
	return m.selectedContentIdx
}

func (m Model) getWrappedLines(line string) []string {
	if stringWidth(line) < m.width {
		return []string{line}
	}
	line = strings.TrimRight(line, " ")
	return splitLineIntoSizedChunks(line, m.width)
}

func (m Model) getNumVisibleItems() int {
	if !m.wrapText {
		return m.contentHeight
	}

	var itemCount int
	var rowCount int
	contentIdx := m.wrappedContentIdxToContentIdx[m.yOffset]
	for rowCount < m.contentHeight {
		if height, exists := m.contentIdxToHeight[contentIdx]; exists {
			rowCount += height
		} else {
			break
		}
		contentIdx++
		itemCount++
	}
	return itemCount
}

func (m Model) lastContentItemSelected() bool {
	return m.selectedContentIdx == len(m.content)-1
}

func (m Model) getFooter() (string, int) {
	numerator := m.selectedContentIdx + 1
	denominator := len(m.content)
	totalNumLines := len(m.getContent())

	// if selection is disabled, percentage should show from the bottom of the visible content
	// such that panning the view to the bottom shows 100%
	if !m.selectionEnabled {
		numerator = m.yOffset + m.contentHeight
		denominator = totalNumLines
	}

	if totalNumLines >= m.height-len(m.getHeader()) {
		percentScrolled := percent(numerator, denominator)
		footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
		renderedFooterString := m.FooterStyle.Copy().MaxWidth(m.width).Render(footerString)
		footerHeight := lipgloss.Height(renderedFooterString)
		return renderedFooterString, footerHeight
	}
	return "", 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// Package termui_addons provides addon components for using with github.com/gizak/termui
package termui_addons

import (
	ui "github.com/gizak/termui"
	rw "github.com/mattn/go-runewidth"
	"strings"
)

// The default position for a ScrollingBlock
type DefaultPosition int

const (
	TOP    DefaultPosition = iota
	BOTTOM DefaultPosition = iota
)

// ScrollingPar is like Par from termui.Par but is scrollable.
type ScrollingPar struct {
	ScrollingBlock
	Text string
}

// NewScrollingPar returns a new *ScrollingPar with current theme.
func NewScrollingPar() *ScrollingPar {
	p := &ScrollingPar{}
	p.Block = *ui.NewBlock()
	p.FgColor = ui.Theme().ParTextFg
	p.BgColor = ui.Theme().ParTextBg
	return p
}

// Buffer implements Bufferer interface.
func (p *ScrollingPar) Buffer() []ui.Point {
	p.Runes = []rune(p.Text)
	return p.ScrollingBlock.Buffer()
}

// ScrollingList is like List from termui.List but is scrollable.
type ScrollingList struct {
	ScrollingBlock
	Items []string
}

// NewScrollingList returns a new *ScrollingList with current theme.
func NewScrollingList() *ScrollingList {
	l := &ScrollingList{}
	l.Block = *ui.NewBlock()
	l.FgColor = ui.Theme().ListItemFg
	l.BgColor = ui.Theme().ListItemBg
	return l
}

// Buffer implements Bufferer interface.
func (l *ScrollingList) Buffer() []ui.Point {
	l.Runes = []rune(strings.Join(l.Items, "\n"))
	return l.ScrollingBlock.Buffer()
}

type ScrollingBlock struct {
	ui.Block
	Runes           []rune
	FgColor         ui.Attribute
	BgColor         ui.Attribute
	DefaultPosition DefaultPosition
	startLine       int
	inScroll        bool
	linesInBuffer   int
}

// Buffer implements Bufferer interface.
func (l *ScrollingBlock) Buffer() []ui.Point {

	innerWidth := l.Width - l.PaddingLeft - l.PaddingRight - 3
	startX := l.X + l.PaddingLeft + 1

	// Make rows
	rows := make([]rowAndSize, 0)
	runeCount := len(l.Runes)
	if runeCount > 0 {
		row := rowAndSize{make([]runeAndSize, 0), 0}
		for i := 0; i < runeCount; i++ {
			r := l.Runes[i]
			w := rw.RuneWidth(r)
			if r == '\n' || row.size+w > innerWidth {
				rows = append(rows, row)
				row = rowAndSize{make([]runeAndSize, 0), 0}
			}
			if r == '\n' {
				continue
			}
			row.rs = append(row.rs, runeAndSize{r, w})
			row.size += w
		}
		rows = append(rows, row)
	}
	l.linesInBuffer = len(rows)

	// Calculate which rows will be on screen
	pageSize := l.PageSize()
	stopLine := 0
	if l.inScroll || l.DefaultPosition == TOP {
		if l.startLine > 0 {
			pageSize -= 1
		}
		stopLine = l.startLine + pageSize
		if l.linesInBuffer > stopLine {
			stopLine -= 1
		} else if l.linesInBuffer < stopLine {
			stopLine = l.linesInBuffer
		}
	} else {
		stopLine = l.linesInBuffer
		l.startLine = l.calcStartFromEnd()
	}

	// Render onscreen rows
	yPos := l.Y + l.PaddingTop + 1
	ps := l.Block.Buffer()
	if l.startLine > 0 {
		ps = l.appendMoreIndicator(startX, yPos, ps)
		yPos++
	}
	for i := l.startLine; i < stopLine; i++ {
		xPos := startX
		for _, rs := range rows[i].rs {
			pi := ui.Point{}
			pi.X = xPos
			pi.Y = yPos
			pi.Ch = rs.r
			pi.Bg = l.BgColor
			pi.Fg = l.FgColor
			ps = append(ps, pi)
			xPos += rs.s
		}
		yPos++
	}
	if stopLine < l.linesInBuffer {
		ps = l.appendMoreIndicator(startX, yPos, ps)
	}
	return ps
}

func (l *ScrollingBlock) SetStartLine(line int) {
	maxStartLine := l.calcStartFromEnd()
	if line > maxStartLine {
		l.startLine = maxStartLine
	} else if line < 0 {
		l.startLine = 0
	} else {
		l.startLine = line
	}
	if (l.startLine == 0 && l.DefaultPosition == TOP) || (l.startLine == maxStartLine && l.DefaultPosition == BOTTOM) {
		l.inScroll = false
	} else {
		l.inScroll = true
	}
}

func (l *ScrollingBlock) ScrollUp() {
	l.SetStartLine(l.startLine - 1)
}

func (l *ScrollingBlock) ScrollDown() {
	l.SetStartLine(l.startLine + 1)
}

func (l *ScrollingBlock) PageUp() {
	l.SetStartLine(l.startLine - l.PageSize())
}

func (l *ScrollingBlock) PageDown() {
	l.SetStartLine(l.startLine + l.PageSize())
}

func (l *ScrollingBlock) Home() {
	l.SetStartLine(0)
}

func (l *ScrollingBlock) End() {
	l.SetStartLine(l.linesInBuffer)
}

func (l *ScrollingBlock) ScrollToDefaultPosition() {
	if l.DefaultPosition == TOP {
		l.Home()
	} else {
		l.End()
	}
}

func (l *ScrollingBlock) LinesInBuffer() int {
	return l.linesInBuffer
}

func (l *ScrollingBlock) PageSize() int {
	return l.Height - l.PaddingTop - l.PaddingBottom - 2
}

func (l *ScrollingBlock) appendMoreIndicator(x int, y int, ps []ui.Point) []ui.Point {
	dot := ([]rune("."))[0]
	for i := 0; i < 3; i++ {
		pi := ui.Point{}
		pi.X = x + i
		pi.Y = y
		pi.Ch = dot
		pi.Bg = l.BgColor
		pi.Fg = l.FgColor
		ps = append(ps, pi)
	}
	return ps
}

func (l *ScrollingBlock) calcStartFromEnd() int {
	start := l.linesInBuffer - l.PageSize()
	if start < 0 {
		start = 0
	} else if start > 0 {
		start += 1
	}
	return start
}

type rowAndSize struct {
	rs   []runeAndSize
	size int
}

type runeAndSize struct {
	r rune
	s int
}

// Package termui_addons provides addon components for using with github.com/gizak/termui
package termui_addons

import (
	ui "github.com/gizak/termui"
	//rw "github.com/mattn/go-runewidth"
	"strings"
)

// The default position for a ScrollingBlock
type DefaultPosition int

const (
	TOP    DefaultPosition = iota
	BOTTOM DefaultPosition = iota
)

type Scroller struct {
	block           *ui.Block
	DefaultPosition DefaultPosition
	startLine       int
	inScroll        bool
	linesInBuffer   int
}

func (s *Scroller) SetStartLine(line int) {
	maxStartLine := s.calcStartFromEnd()
	if line > maxStartLine {
		s.startLine = maxStartLine
	} else if line < 0 {
		s.startLine = 0
	} else {
		s.startLine = line
	}
	if (s.startLine == 0 && s.DefaultPosition == TOP) || (s.startLine == maxStartLine && s.DefaultPosition == BOTTOM) {
		s.inScroll = false
	} else {
		s.inScroll = true
	}
}

func (s *Scroller) ScrollUp() {
	s.SetStartLine(s.startLine - 1)
}

func (s *Scroller) ScrollDown() {
	s.SetStartLine(s.startLine + 1)
}

func (s *Scroller) PageUp() {
	s.SetStartLine(s.startLine - s.block.InnerHeight())
}

func (s *Scroller) PageDown() {
	s.SetStartLine(s.startLine + s.block.InnerHeight())
}

func (s *Scroller) Home() {
	s.SetStartLine(0)
}

func (s *Scroller) End() {
	s.SetStartLine(s.linesInBuffer)
}

func (s *Scroller) ScrollToDefaultPosition() {
	if s.DefaultPosition == TOP {
		s.Home()
	} else {
		s.End()
	}
}

func (s *Scroller) calcStartFromEnd() int {
	start := s.linesInBuffer - s.block.InnerHeight()
	if start < 0 {
		start = 0
	} else if start > 0 {
		start += 1
	}
	return start
}

// ScrollingList is like List from termui.List but is scrollable.
type ScrollingList struct {
	ui.List
	Scroller
}

// NewScrollingList returns a new *ScrollingList with current theme.
func NewScrollingList() *ScrollingList {
	l := &ScrollingList{}
	l.List = *ui.NewList()
	l.Scroller = Scroller{}
	l.Scroller.block = &l.List.Block
	return l
}

type lineAndSize struct {
	cells []ui.Cell
	size  int
}

// Buffer implements Bufferer interface.
func (l *ScrollingList) Buffer() ui.Buffer {

	innerBounds := l.List.Block.InnerBounds()

	// Get lines
	var lines []lineAndSize
	cells := ui.DefaultTxBuilder.Build(strings.Join(l.Items, "\n"), l.ItemFgColor, l.ItemBgColor)
	if len(cells) > 0 {
		line := lineAndSize{make([]ui.Cell, 0), 0}
		for _, cell := range cells {
			w := cell.Width()
			ch := cell.Ch
			if ch == '\n' {
				lines = append(lines, line)
				line = lineAndSize{make([]ui.Cell, 0), 0}
				continue
			}
			if line.size+w > innerBounds.Dx() {
				if l.Overflow == "wrap" {
					lines = append(lines, line)
					line = lineAndSize{make([]ui.Cell, 0), 0}
				} else {
					line.size += w
					continue
				}
			}
			line.cells = append(line.cells, cell)
			line.size += w
		}
	}
	l.linesInBuffer = len(lines)

	// Calculate which rows will be on screen
	pageSize := l.InnerHeight()
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
	buf := l.Block.Buffer()
	yPos := innerBounds.Min.Y
	if l.startLine > 0 {
		l.putMoreIndicator(buf, innerBounds.Min.X, yPos)
		yPos++
	}
	for i := l.startLine; i < stopLine; i++ {
		xPos := innerBounds.Min.X
		for _, cell := range lines[i].cells {
			buf.Set(xPos, yPos, cell)
			xPos += cell.Width()
		}
		yPos++
	}
	if stopLine < l.linesInBuffer {
		l.putMoreIndicator(buf, innerBounds.Min.X, yPos)
	}

	return buf

}

func (l *ScrollingList) putMoreIndicator(b ui.Buffer, x int, y int) {
	dot := ([]rune("."))[0]
	for i := 0; i < 3; i++ {
		cell := ui.Cell{}
		cell.Ch = dot
		cell.Bg = l.ItemBgColor
		cell.Fg = l.ItemFgColor
		b.Set(x+i, y, cell)
	}
}

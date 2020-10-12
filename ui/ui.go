package ui

import (
	"image"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mitchellh/go-wordwrap"
)

const (
	// TopBarHeight is the height of the top bar
	TopBarHeight int = 3
)

// UI is a generic interface for a TUI
type UI interface {
	Init() error                                           // Initialize (non-tcell) resources
	Draw(int, int, tcell.Screen)                           // Draw UI, given width and height
	HandleEvents([]tcell.Event, tcell.Screen) (flush bool) // Return true to flush event buffer, false to accumulate and pass on next event
	Wait() <-chan bool                                     // Used to quit
}

// RunUI is a generic runner for a UI
func RunUI(u UI) error {
	u.Init()
	encoding.Register()
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = s.Init()
	if err != nil {
		return err
	}
	defer s.Fini()
	defer s.Clear()
	evBuf := make([]tcell.Event, 0)

	w, h := s.Size()
	u.Draw(w, h, s)
	s.Show()
	evChan := make(chan tcell.Event)
	go func() {
		for {
			e := s.PollEvent()
			// age := time.Now().Sub(e.When())
			// // exec.Command("notify-send", age.String()).Run()
			// if !(age >= time.Millisecond*10) {
			evChan <- e
			// }
		}
	}()

	for {
		select {
		case <-u.Wait():
			panic("don")
		case ev := <-evChan:

			switch e := ev.(type) {
			case *tcell.EventResize:
				w, h := e.Size()
				u.Draw(w, h, s)
				s.Sync()
			default:
				evBuf = append(evBuf, e)
				flush := u.HandleEvents(evBuf, s)
				if flush {
					evBuf = evBuf[:0]
				}
			}

		}
	}
	return nil
}

// FeefUI is a UI for feef
type FeefUI struct {
	tabs       []string
	currentTab int
	doneChan   chan bool
}

// Init is
func (f *FeefUI) Init() error {
	f.doneChan = make(chan bool)
	f.tabs = []string{"new", "unread", "old", "queue"}
	return nil
}

const (
	topleft     = '┌'
	topright    = '┐'
	bottomleft  = '└'
	bottomright = '┘'

	verticalline   = '│'
	horizontalline = '─'
)

type box struct {
	image.Rectangle
	title, content         string
	borderStyle, textStyle tcell.Style
}

func drawBox(b box, s tcell.Screen) {
	x1 := b.Min.X
	x2 := b.Max.X
	y1 := b.Min.Y
	y2 := b.Max.Y
	s.SetContent(x1, y1, topleft, []rune{}, b.borderStyle)
	s.SetContent(x1, y2, bottomleft, []rune{}, b.borderStyle)
	s.SetContent(x2, y1, topright, []rune{}, b.borderStyle)
	s.SetContent(x2, y2, bottomright, []rune{}, b.borderStyle)
	for x := x1 + 1; x < x2; x++ {
		s.SetContent(x, y1, horizontalline, []rune{}, b.borderStyle)
		s.SetContent(x, y2, horizontalline, []rune{}, b.borderStyle)
	}
	for y := y1 + 1; y < y2; y++ {
		s.SetContent(x1, y, verticalline, []rune{}, b.borderStyle)
		s.SetContent(x2, y, verticalline, []rune{}, b.borderStyle)
	}
	pos := x1 + 1
	maxlength := x2 - x1 - 1

	var trimmed string
	if len(b.title) > maxlength {
		trimmed = b.title[:maxlength]
	} else {
		trimmed = b.title
	}
	cursor := x1 + 1
	for n, line := range strings.Split(wordwrap.WrapString(b.content, uint(x2-x1-2)), "\n") {
		for _, char := range line {
			if n < y2-y1-2 {
				s.SetContent(cursor, y1+1+n, char, []rune{}, b.textStyle)
			}
			cursor++
		}
		cursor = x1 + 1
	}
	for _, c := range trimmed {
		s.SetContent(pos, y1, c, []rune{}, b.borderStyle)
		pos++
	}
}

// Draw is
func (f *FeefUI) Draw(w, h int, s tcell.Screen) {
	drawBox(box{
		Rectangle:   image.Rect(5, 6, 20, 15),
		title:       "arp242.net: This article",
		content:     "Through the technical life, to explore the ultimate value.",
		borderStyle: tcell.StyleDefault},
		s)
	drawBox(box{
		Rectangle:   image.Rect(10, 10, 50, 44),
		title:       "kar.wtf: Hi",
		content:     "this is a post",
		borderStyle: tcell.StyleDefault},
		s)
	// Tabline
	var x int
	for ti, t := range f.tabs {
		for _, r := range t {
			s.SetContent(x, 0, r, []rune{}, tcell.StyleDefault.Bold(ti == f.currentTab))
			x++
		}
		s.SetContent(x, 0, ' ', []rune{}, tcell.StyleDefault)
		x++
	}
}

// HandleEvents is
func (f *FeefUI) HandleEvents(events []tcell.Event, s tcell.Screen) (flush bool) {
	switch e := events[0].(type) {
	case *tcell.EventKey:
		switch e.Rune() {
		case 'q':
			// if f.doneChan == nil {
			// 	panic("aeu")
			// }
			// f.doneChan <- true
			os.Exit(0)
		case 'r':
			s.Clear()
			s.Sync()
		case '1', '2', '3', '4':
			w, h := s.Size()
			in, err := strconv.Atoi(string(e.Rune()))
			if err != nil {
				panic(err)
			}
			if !(f.currentTab == in-1) {
				f.currentTab = in - 1
				f.Draw(w, h, s)
				s.Sync()
			}
		}
	}
	return true
}

// Wait is
func (f *FeefUI) Wait() <-chan bool {
	return f.doneChan
}

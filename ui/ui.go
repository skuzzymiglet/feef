package ui

import (
	"fmt"
	"image"
	"strconv"
	"strings"

	"git.sr.ht/~skuzzymiglet/feef/ui/tilefuncs"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mitchellh/go-wordwrap"
)

// TODO: split this into its own library
const (
	// TopBarHeight is the height of the top bar
	TopBarHeight int = 3 //TODO: actually use this in the code
)

// UI is a generic interface for a TUI
// UI is immediate-mode/stateless - it's drawn on demand. This ensures pretty clean code
type UI interface {
	Init() error                                           // Initialize (non-tcell) resources
	Draw(int, int, tcell.Screen)                           // Draw UI, given width and height (to save on scree.Size() calls)
	HandleEvents([]tcell.Event, tcell.Screen) (flush bool) // Return true to flush event buffer, false to accumulate and pass on next event
}

// RunUI is a generic runner for a UI

func RunUI(u UI, stopChan chan struct{}) error {
	// TODO: add a channel that forces a redraw when an event is sent (e.g. feed updates)
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

	for { // TODO: not have an infinite loop before return
		select {
		case <-stopChan:
			return nil
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
// It will contain everything the UI needs including a Feeds objec
type FeefUI struct {
	tabs       []string
	currentTab int
	DoneChan   chan struct{}
}

// Init is
func (f *FeefUI) Init() error {
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
	// TODO: gofmt -r to rename these, maybe
	x1 := b.Min.X
	x2 := b.Max.X - 1
	y1 := b.Min.Y
	y2 := b.Max.Y - 1
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
	if maxlength < 0 {
		maxlength = 0
	}

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
	s.Clear()
	// Demo boxes
	for _, win := range tilefuncs.Vertical(7, image.Rect(1, 1, w, h)) {
		drawBox(box{
			Rectangle:   win,
			title:       "i love tcell!",
			content:     fmt.Sprintf("%dx%d", win.Dx(), win.Dy()),
			borderStyle: tcell.StyleDefault,
		},
			s)
	}
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

// HandleEvents handles events from an event buffer
func (f *FeefUI) HandleEvents(events []tcell.Event, s tcell.Screen) (flush bool) {
	switch e := events[0].(type) {
	case *tcell.EventKey:
		switch e.Rune() {
		case 'q':
			f.DoneChan <- struct{}{}
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

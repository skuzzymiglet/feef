package ui

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
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
			age := time.Now().Sub(e.When())
			// exec.Command("notify-send", age.String()).Run()
			if !(age >= time.Millisecond*10) {
				evChan <- e
			}
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

// Draw is
func (f *FeefUI) Draw(w, h int, s tcell.Screen) {
	// Colors
	for x := 0; x < w-1; x++ {
		for y := 1; y < h-1; y++ {
			if y == 1 {
				s.SetContent(x, y, '-', []rune{}, tcell.StyleDefault)
			} else {
				s.SetContent(x, y, rune(x+y), []rune{}, tcell.StyleDefault)
			}
		}
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
			panic(nil)
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

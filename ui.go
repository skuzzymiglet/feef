package main

import (
	"fmt"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const (
	// TopBarHeight is the height of the top bar
	TopBarHeight int = 3
)

type Tabs struct {
	names      []string
	widgets    [][]termui.Drawable
	tabpane    *widgets.TabPane
	messageBox *widgets.Paragraph
	gaugeIndex int
}

func InitTabs() Tabs {
	names := []string{"new", "unread", "old", "queue", "jobs"}
	w, h := termui.TerminalDimensions()
	tabpane := widgets.NewTabPane(names...)
	tabpane.SetRect(0, 0, w, TopBarHeight)
	tabpane.Border = true
	tabpane.ActiveTabStyle = termui.Style{
		Fg:       15,
		Bg:       0,
		Modifier: termui.ModifierBold,
	}
	tabpane.InactiveTabStyle = termui.Style{
		Fg: 15,
		Bg: 0,
	}
	emptyParagraph := widgets.NewParagraph()
	emptyParagraph.SetRect(0, TopBarHeight, w, h)

	// message box
	messages := widgets.NewParagraph()
	// TODO: resize this when window resizes
	messages.SetRect(0, 0, w, TopBarHeight)

	return Tabs{
		messageBox: messages,
		tabpane:    tabpane,
		names:      names,
		widgets: [][]termui.Drawable{
			[]termui.Drawable{emptyParagraph},
			[]termui.Drawable{emptyParagraph},
			[]termui.Drawable{emptyParagraph},
			[]termui.Drawable{emptyParagraph},
			[]termui.Drawable{widgets.NewGauge()},
		},
		gaugeIndex: 4,
	}
}

func (t *Tabs) Refresh() {
	w, h := termui.TerminalDimensions()
	t.tabpane.SetRect(0, 0, w, TopBarHeight)
	termui.Render(t.tabpane)
	for _, e := range t.widgets[t.tabpane.ActiveTabIndex] {
		e.SetRect(0, TopBarHeight, w, h)
		termui.Render(e)
	}
}

func (t *Tabs) Go(tab int) {
	t.tabpane.ActiveTabIndex = tab
	t.tabpane.ActiveTabIndex = tab
	if tab != t.gaugeIndex {
		t.widgets[tab][0].(*widgets.Paragraph).Text = fmt.Sprintf("It is: %s", time.Now())
	}
	termui.Render(t.tabpane)
	termui.Render(t.widgets[tab]...)
}

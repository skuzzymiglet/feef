package main

import (
	"io"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/sirupsen/logrus"
)

// BarMessageHook is a Logrus hook for displaying Info/Warn/Error logs in a bar on top
type BarMessageHook struct {
	b *widgets.Paragraph
}

// Levels satisfies logrus.Hook
func (b *BarMessageHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

// Fire satisfies logrus.Hook
func (b *BarMessageHook) Fire(l *logrus.Entry) error {
	var style termui.Style
	switch l.Level {
	case logrus.ErrorLevel:
		style.Fg = 9
		style.Modifier = termui.ModifierBold
	case logrus.WarnLevel:
		style.Fg = 11
		style.Modifier = termui.ModifierBold
	case logrus.InfoLevel:
		style.Fg = 15
		style.Modifier = termui.ModifierReverse
	}
	b.b.TextStyle = style
	b.b.Text = l.Message
	termui.Render(b.b)
	return nil
}

// NewLogger creates a new logger
func NewLogger(messageBox *widgets.Paragraph, output io.Writer) *logrus.Logger {
	// Logging
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	log.AddHook(&BarMessageHook{
		b: messageBox,
	})
	log.Out = output
	return log
}

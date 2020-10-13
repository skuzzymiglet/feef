package tilefuncs

import (
	"image"
	"testing"
)

func TestTileFuncs(t *testing.T) {
	t.Log(Vertical(3, image.Rect(0, 0, 3200, 1800)))
}

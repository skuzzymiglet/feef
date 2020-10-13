package tilefuncs

import "image"

// Tiler is a function which tiles n windows onto the image.Rectangle screen, and returns their rectangles
type Tiler func(n int, screen image.Rectangle) []image.Rectangle

// MasterTiler is a DWM-style master-stack tiling function
// It takes the number of windows, the number of those in master, the width and the height of the master area and image.Rectangle to tile them onto
// The master windows are the first nmaster elements of the returned rectangles
type MasterTiler func(nwindows, nmaster, masterWidth, masterHeight int, screen image.Rectangle) []image.Rectangle

// Vertical tiles windows vertically, each with equal size
func Vertical(n int, screen image.Rectangle) []image.Rectangle {
	tiles := make([]image.Rectangle, n)
	for i := 0; i < n; i++ {
		tiles[i] = image.Rect((screen.Dx()/n)*i, screen.Min.Y, (screen.Dx()/n)*(i+1), screen.Max.Y)
	}
	return tiles
}

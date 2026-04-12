//go:build android

// Package mobile provides the Android entry point for the Tox Messenger application.
// It is used with ebitenmobile bind to produce an .aar library for Android.
//
// Usage:
//
//	ebitenmobile bind -target android -javapkg go -o app.aar ./mobile/
package mobile

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
	mobile.SetGame(&game{})
}

// game is a minimal Ebiten game implementation for the Android entry point.
type game struct{}

// Update runs the game logic once per tick.
func (g *game) Update() error {
	return nil
}

// Draw renders the game frame onto the provided screen image.
func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 25, G: 25, B: 35, A: 255})
}

// Layout returns the logical game screen size for the given outside (device) dimensions.
func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

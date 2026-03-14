//go:build linux

// Package ui provides a platform-independent interface to the UI toolkit.
// On Linux, this package re-exports types from wain for Wayland/X11 support.
package ui

import "github.com/opd-ai/wain"

// Type aliases for zero-cost re-export of wain types
type App = wain.App
type Column = wain.Column
type Row = wain.Row
type Label = wain.Label
type Button = wain.Button
type TextInput = wain.TextInput
type Panel = wain.Panel
type ScrollView = wain.ScrollView
type Size = wain.Size
type Color = wain.Color
type StyleOverride = wain.StyleOverride
type PublicWidget = wain.PublicWidget
type Widget = wain.Widget
type BaseWidget = wain.BaseWidget
type WindowConfig = wain.WindowConfig
type PointerEvent = wain.PointerEvent
type KeyEvent = wain.KeyEvent
type TouchEvent = wain.TouchEvent

var (
	NewApp        = wain.NewApp
	NewRow        = wain.NewRow
	NewColumn     = wain.NewColumn
	NewLabel      = wain.NewLabel
	NewButton     = wain.NewButton
	NewTextInput  = wain.NewTextInput
	NewPanel      = wain.NewPanel
	NewScrollView = wain.NewScrollView
	DefaultDark   = wain.DefaultDark
	RGB           = wain.RGB
	White         = wain.White
)

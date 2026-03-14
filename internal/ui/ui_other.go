//go:build windows || darwin || android || ios

// Package ui provides a platform-independent interface to the UI toolkit.
// On non-Linux platforms, this package provides wrapper types around wayne
// (Ebitengine-based) that add API compatibility methods to match wain's interface.
package ui

import (
	"log"

	"github.com/opd-ai/wayne"
)

// Type aliases for types that don't need extra methods
type Size = wayne.Size
type Color = wayne.Color
type StyleOverride = wayne.StyleOverride
type PublicWidget = wayne.PublicWidget
type Widget = wayne.Widget
type BaseWidget = wayne.BaseWidget
type WindowConfig = wayne.WindowConfig
type PointerEvent = wayne.PointerEvent
type KeyEvent = wayne.KeyEvent
type TouchEvent = wayne.TouchEvent

// App wraps wayne.App to return our wrapped Window type.
type App struct {
	*wayne.App
	onCloseCallbacks []func()
}

// NewApp creates a new application.
func NewApp() *App {
	return &App{App: wayne.NewApp()}
}

// registerOnClose registers a callback to be called when the app quits.
func (a *App) registerOnClose(callback func()) {
	if callback != nil {
		a.onCloseCallbacks = append(a.onCloseCallbacks, callback)
	}
}

// Quit stops the application and invokes all registered OnClose callbacks.
func (a *App) Quit() {
	// Invoke all registered OnClose callbacks before quitting
	for _, cb := range a.onCloseCallbacks {
		cb()
	}
	a.App.Quit()
}

// NewWindow creates a new window with the specified configuration.
func (a *App) NewWindow(cfg WindowConfig) (*Window, error) {
	win, err := a.App.NewWindow(cfg)
	if err != nil {
		return nil, err
	}
	return &Window{Window: win, app: a}, nil
}

// Window wraps wayne.Window to add wain-compatible methods.
type Window struct {
	*wayne.Window
	app             *App
	onCloseCallback func()
}

// SetRootWidget sets the root widget (delegates to SetRoot for wayne compatibility).
// This accepts Widget (which includes PublicWidget) for API compatibility with wain.
// On wayne, we check if the widget implements PublicWidget; if not, we log an error.
func (w *Window) SetRootWidget(widget Widget) {
	// Wayne's SetRoot expects PublicWidget, not Widget
	if pw, ok := widget.(PublicWidget); ok {
		w.Window.SetRoot(pw)
	} else {
		// Log error instead of silently ignoring - helps debugging
		// Using log.Println for consistent logging with the rest of the application
		log.Println("ui: SetRootWidget called with non-PublicWidget; root not set (window will be blank)")
	}
}

// SetRootPublicWidget sets the root widget directly from a PublicWidget.
// This is the preferred method on non-Linux platforms as it avoids type assertion.
func (w *Window) SetRootPublicWidget(pw PublicWidget) {
	w.Window.SetRoot(pw)
}

// OnClose registers a callback to be called when the window is closed.
// Since wayne/Ebitengine doesn't have a native window close event, we store
// the callback and invoke it when the app quits.
func (w *Window) OnClose(callback func()) {
	w.onCloseCallback = callback
	// Register with the app so it can invoke on quit
	if w.app != nil {
		w.app.registerOnClose(callback)
	}
}

// Column wraps wayne.Column to add SetStyle for wain API compatibility.
type Column struct {
	*wayne.Column
	Panel *Panel // Expose Panel as our wrapped type for compatibility
}

// NewColumn creates a new vertical container.
func NewColumn() *Column {
	col := wayne.NewColumn()
	return &Column{
		Column: col,
		Panel:  &Panel{Panel: col.Panel},
	}
}

// SetStyle applies a style override (converts to SetTheme for wayne compatibility).
func (c *Column) SetStyle(override StyleOverride) {
	c.Column.SetStyle(override)
}

// Row wraps wayne.Row to add SetStyle for wain API compatibility.
type Row struct {
	*wayne.Row
	Panel *Panel // Expose Panel as our wrapped type for compatibility
}

// NewRow creates a new horizontal container.
func NewRow() *Row {
	row := wayne.NewRow()
	return &Row{
		Row:   row,
		Panel: &Panel{Panel: row.Panel},
	}
}

// SetStyle applies a style override (converts to SetTheme for wayne compatibility).
func (r *Row) SetStyle(override StyleOverride) {
	r.Row.SetStyle(override)
}

// Panel wraps wayne.Panel.
type Panel struct {
	*wayne.Panel
}

// NewPanel creates a new panel with the specified size.
func NewPanel(size Size) *Panel {
	return &Panel{Panel: wayne.NewPanel(size)}
}

// Label wraps wayne.Label to add SetStyle for wain API compatibility.
type Label struct {
	*wayne.Label
}

// NewLabel creates a new label with the specified text and size.
func NewLabel(text string, size Size) *Label {
	return &Label{Label: wayne.NewLabel(text, size)}
}

// SetStyle applies a style override by converting to theme.
func (l *Label) SetStyle(override StyleOverride) {
	theme := wayne.DefaultDark()
	if override.Background != nil {
		theme.Background = *override.Background
	}
	if override.Foreground != nil {
		theme.Foreground = *override.Foreground
	}
	l.Label.SetTheme(theme)
}

// Button wraps wayne.Button to add SetStyle for wain API compatibility.
type Button struct {
	*wayne.Button
}

// NewButton creates a new button with the specified label and size.
func NewButton(label string, size Size) *Button {
	return &Button{Button: wayne.NewButton(label, size)}
}

// SetStyle applies a style override by converting to theme.
func (b *Button) SetStyle(override StyleOverride) {
	theme := wayne.DefaultDark()
	if override.Background != nil {
		theme.Accent = *override.Background // Buttons use Accent as background in wayne
	}
	if override.Foreground != nil {
		theme.Foreground = *override.Foreground
	}
	b.Button.SetTheme(theme)
}

// TextInput wraps wayne.TextInput to add SetStyle for wain API compatibility.
type TextInput struct {
	*wayne.TextInput
}

// NewTextInput creates a new text input with the specified placeholder and size.
func NewTextInput(placeholder string, size Size) *TextInput {
	return &TextInput{TextInput: wayne.NewTextInput(placeholder, size)}
}

// SetStyle applies a style override by converting to theme.
func (t *TextInput) SetStyle(override StyleOverride) {
	theme := wayne.DefaultDark()
	if override.Background != nil {
		theme.Background = *override.Background
	}
	if override.Foreground != nil {
		theme.Foreground = *override.Foreground
	}
	t.TextInput.SetTheme(theme)
}

// ScrollView wraps wayne.ScrollView.
type ScrollView struct {
	*wayne.ScrollView
}

// NewScrollView creates a new scroll view with the specified size.
func NewScrollView(size Size) *ScrollView {
	return &ScrollView{ScrollView: wayne.NewScrollView(size)}
}

var (
	DefaultDark = wayne.DefaultDark
	RGB         = wayne.RGB
	White       = wayne.White
)

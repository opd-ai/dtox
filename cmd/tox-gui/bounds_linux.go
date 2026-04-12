//go:build linux

package main

import "github.com/opd-ai/dtox/internal/ui"

// publicWidgetSize returns the width and height of a PublicWidget.
// On Linux, wain's PublicWidget provides Bounds().
func publicWidgetSize(pw ui.PublicWidget) (width, height int) {
	return pw.Bounds()
}

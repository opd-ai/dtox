//go:build windows || darwin || android || ios

package main

import "github.com/opd-ai/dtox/internal/ui"

// publicWidgetSize returns the width and height of a PublicWidget.
// On non-Linux platforms, wayne's PublicWidget provides Width() and Height().
func publicWidgetSize(pw ui.PublicWidget) (width, height int) {
	return pw.Width(), pw.Height()
}

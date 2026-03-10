// Package main is the entry point for the Tox Messenger GUI client.
//
// It creates a wain application window, initialises the toxcore backend,
// and starts the Tox event loop in a background goroutine.  All UI updates
// that originate from the Tox callback goroutine are routed through
// app.Notify so they run safely on the UI goroutine.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/wain"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ── wain application ─────────────────────────────────────────────────────
	app := wain.NewApp()
	app.SetTheme(wain.DefaultDark())

	// ── Tox backend (not started yet) ────────────────────────────────────────
	backend := newToxBackend()

	// ── UI state (widgets are built lazily after the window is created) ───────
	ui := newUI(app, backend)
	backend.setUI(ui)

	// ── Schedule window creation for the first event-loop iteration ──────────
	// app.NewWindow requires the app to be running (display server initialised).
	// Placing the call inside app.Notify ensures it runs after app.Run() has
	// completed initialisation but before the first user-visible frame.
	app.Notify(func() {
		win, err := app.NewWindow(wain.WindowConfig{
			Title:  "Tox Messenger",
			Width:  900,
			Height: 650,
		})
		if err != nil {
			log.Printf("failed to create window: %v", err)
			app.Quit()
			return
		}

		// Build the public widget tree and attach it to the window.
		root := ui.buildRoot()
		win.SetRootWidget(newUIRoot(root))
		ui.setWindow(win)
		win.Redraw()

		win.OnClose(func() {
			backend.stop()
			app.Quit()
		})
	})

	// ── Initialise and start the Tox backend ─────────────────────────────────
	if err := backend.init(); err != nil {
		log.Printf("Tox init warning: %v (continuing without network)", err)
	} else {
		go backend.run()
	}

	// ── OS-signal handler for clean shutdown ─────────────────────────────────
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		backend.stop()
		app.Quit()
	}()

	// ── Enter the UI event loop (blocks until app.Quit) ──────────────────────
	if err := app.Run(); err != nil {
		log.Fatalf("wain error: %v", err)
	}
}

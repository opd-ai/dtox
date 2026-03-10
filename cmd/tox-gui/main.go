package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opd-ai/wain"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Starting Tox Messenger GUI...")

	// 1. Create shared state.
	state := NewAppState()

	// 2. Create wain App and set theme.
	app := wain.NewApp()
	app.SetTheme(wain.DefaultDark())

	// 3. Create and start Tox backend (before UI so we have the Tox ID).
	uiRefsPlaceholder := &UIRefs{}
	backend, err := NewToxBackend(state, app, uiRefsPlaceholder)
	if err != nil {
		log.Fatalf("Failed to initialize Tox: %v", err)
	}

	// 4. Build UI widget tree with actions wired in.
	root, uiRefs := BuildUIWithActions(state, backend, app)

	// Point backend at the real UI refs for callback-driven updates.
	backend.uiRefs = uiRefs

	// 5. Start the Tox event loop.
	backend.Start()

	// 6. Create window and set root widget.
	win, err := app.NewWindow(wain.WindowConfig{
		Title:     "Tox Messenger",
		Width:     900,
		Height:    650,
		MinWidth:  640,
		MinHeight: 480,
	})
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	win.SetRootWidget(adaptPublicWidget(root))

	win.OnClose(func() {
		backend.Stop()
		app.Quit()
	})

	// 7. Signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Signal received, shutting down...")
		backend.Stop()
		app.Quit()
	}()

	// 8. Run the application event loop (blocks until Quit).
	log.Println("Tox Messenger running. Tox ID:", state.GetSelfAddress())
	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}

	log.Println("Tox Messenger exited.")
}

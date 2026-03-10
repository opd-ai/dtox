package main

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wain"
)

// ── Widget-bridge: public → internal ─────────────────────────────────────────

// uiRoot bridges the wain public-widget tree (PublicWidget) to the internal
// wain.Widget interface required by Window.SetRootWidget.
//
// wain.BaseWidget provides default implementations for Contains, Children,
// SetFocused and IsFocused.  We override the three event-dispatch methods so
// that pointer, keyboard and touch events are forwarded to the public widget
// tree via PublicWidget.HandleEvent.
type uiRoot struct {
	wain.BaseWidget
	pub wain.PublicWidget
}

// newUIRoot wraps a PublicWidget so it can be passed to Window.SetRootWidget.
func newUIRoot(pub wain.PublicWidget) *uiRoot {
	return &uiRoot{pub: pub}
}

// HandlePointer forwards pointer events to the public widget tree.
func (u *uiRoot) HandlePointer(evt *wain.PointerEvent) {
	u.pub.HandleEvent(evt)
}

// HandleKey forwards keyboard events to the public widget tree.
func (u *uiRoot) HandleKey(evt *wain.KeyEvent) {
	u.pub.HandleEvent(evt)
}

// HandleTouch forwards touch events to the public widget tree.
func (u *uiRoot) HandleTouch(evt *wain.TouchEvent) {
	u.pub.HandleEvent(evt)
}

// ── UI state ─────────────────────────────────────────────────────────────────

// UI holds references to every dynamic widget and encapsulates all layout
// construction.  Methods that update widget text/content are safe to call from
// any goroutine via app.Notify.
type UI struct {
	app     *wain.App
	backend *ToxBackend

	mu     sync.Mutex
	window *wain.Window

	// Header
	titleLabel  *wain.Label
	statusLabel *wain.Label

	// Sidebar
	sidebarCol   *wain.Column
	friendPanels map[uint32]*wain.Panel
	friendLabels map[uint32]*wain.Label

	// Chat area
	msgScroll *wain.ScrollView
	msgCol    *wain.Column
	msgInput  *wain.TextInput

	// Footer
	toxIDInput     *wain.TextInput
	addFriendInput *wain.TextInput

	// Currently selected friend
	activeFriendID uint32
}

// newUI creates a UI state object that references app and backend.
func newUI(app *wain.App, backend *ToxBackend) *UI {
	return &UI{
		app:          app,
		backend:      backend,
		friendPanels: make(map[uint32]*wain.Panel),
		friendLabels: make(map[uint32]*wain.Label),
	}
}

// setWindow stores the window reference after it has been created.
func (u *UI) setWindow(win *wain.Window) {
	u.mu.Lock()
	u.window = win
	u.mu.Unlock()
}

// redraw marks the window dirty so the next frame is re-rendered.
// Must be called on the UI goroutine.
func (u *UI) redraw() {
	u.mu.Lock()
	win := u.window
	u.mu.Unlock()
	if win != nil {
		win.Redraw()
	}
}

// ── Layout construction ───────────────────────────────────────────────────────

// buildRoot assembles the full widget hierarchy and returns the root Column.
//
// Layout:
//
//	Column (root)
//	  Row   (header)  – title label + connection-status label
//	  Row   (main)    – sidebar Column (25 %) + chat Column (75 %)
//	  Row   (footer)  – Tox-ID TextInput + add-friend TextInput + Add button
func (u *UI) buildRoot() *wain.Column {
	root := wain.NewColumn()
	root.SetPadding(0)
	root.SetGap(0)

	root.Add(u.buildHeader())
	root.Add(u.buildMainContent())
	root.Add(u.buildFooter())

	return root
}

// buildHeader creates the application header row.
func (u *UI) buildHeader() *wain.Row {
	header := wain.NewRow()
	header.SetPadding(8)
	header.SetGap(8)

	u.titleLabel = wain.NewLabel("Tox Messenger", wain.Size{Width: 60, Height: 100})
	header.Add(u.titleLabel)

	u.statusLabel = wain.NewLabel("Connecting…", wain.Size{Width: 40, Height: 100})
	header.Add(u.statusLabel)

	return header
}

// buildMainContent creates the main two-column content area.
func (u *UI) buildMainContent() *wain.Row {
	mainRow := wain.NewRow()
	mainRow.SetGap(4)

	mainRow.Add(u.buildSidebar())
	mainRow.Add(u.buildChatArea())

	return mainRow
}

// buildSidebar creates the friend-list sidebar Column (25 % of main width).
func (u *UI) buildSidebar() *wain.Column {
	u.sidebarCol = wain.NewColumn()
	u.sidebarCol.SetPadding(8)
	u.sidebarCol.SetGap(4)

	// Static heading
	heading := wain.NewPanel(wain.Size{Width: 100, Height: 8})
	heading.Add(wain.NewLabel("Friends", wain.Size{Width: 100, Height: 100}))
	u.sidebarCol.Add(heading)

	return u.sidebarCol
}

// buildChatArea creates the chat column (75 % of main width): scroll area + input row.
func (u *UI) buildChatArea() *wain.Column {
	chatCol := wain.NewColumn()
	chatCol.SetPadding(8)
	chatCol.SetGap(4)

	// Scrollable message history (85 % of chat column height)
	u.msgScroll = wain.NewScrollView(wain.Size{Width: 100, Height: 85})
	u.msgScroll.OnScroll(func(_ int) {})

	u.msgCol = wain.NewColumn()
	u.msgCol.SetPadding(4)
	u.msgCol.SetGap(2)
	u.msgScroll.Add(u.msgCol)

	// Wrap the scroll view in a panel so Column.Add accepts it.
	scrollPanel := wain.NewPanel(wain.Size{Width: 100, Height: 85})
	scrollPanel.Add(u.msgScroll)
	chatCol.Add(scrollPanel)

	// Message input row (remaining height)
	chatCol.Add(u.buildInputRow())

	return chatCol
}

// buildInputRow creates the text-input + Send-button row.
func (u *UI) buildInputRow() *wain.Row {
	inputRow := wain.NewRow()
	inputRow.SetGap(4)

	u.msgInput = wain.NewTextInput("Type a message…", wain.Size{Width: 80, Height: 100})
	inputRow.Add(u.msgInput)

	sendBtn := wain.NewButton("Send", wain.Size{Width: 20, Height: 100})
	sendBtn.OnClick(u.onSendMessage)
	inputRow.Add(sendBtn)

	return inputRow
}

// buildFooter creates the footer row: Tox-ID display + add-friend field + Add button.
func (u *UI) buildFooter() *wain.Row {
	footer := wain.NewRow()
	footer.SetPadding(8)
	footer.SetGap(8)

	u.toxIDInput = wain.NewTextInput("Your Tox ID will appear here…", wain.Size{Width: 45, Height: 100})
	footer.Add(u.toxIDInput)

	u.addFriendInput = wain.NewTextInput("Paste friend Tox ID to add…", wain.Size{Width: 40, Height: 100})
	footer.Add(u.addFriendInput)

	addBtn := wain.NewButton("Add Friend", wain.Size{Width: 15, Height: 100})
	addBtn.OnClick(u.onAddFriend)
	footer.Add(addBtn)

	return footer
}

// ── UI event handlers ─────────────────────────────────────────────────────────

// onSendMessage is called when the user clicks the Send button.
// Runs on the UI goroutine (wain callback).
func (u *UI) onSendMessage() {
	text := u.msgInput.Text()
	if text == "" {
		return
	}

	u.mu.Lock()
	friendID := u.activeFriendID
	u.mu.Unlock()

	if friendID == 0 {
		u.appendMessage("[System] Select a friend first")
		return
	}

	if err := u.backend.sendMessage(friendID, text); err != nil {
		u.appendMessage(fmt.Sprintf("[Error] Send failed: %v", err))
		return
	}

	u.appendMessage(fmt.Sprintf("Me: %s", text))
	u.msgInput.SetText("")
	u.redraw()
}

// onAddFriend is called when the user clicks the Add Friend button.
// Runs on the UI goroutine (wain callback).
func (u *UI) onAddFriend() {
	address := u.addFriendInput.Text()
	if address == "" {
		return
	}

	const reqMsg = "Hi! I'd like to add you as a Tox friend."
	if err := u.backend.addFriend(address, reqMsg); err != nil {
		u.appendMessage(fmt.Sprintf("[Error] Add friend failed: %v", err))
		return
	}

	short := address
	if len(short) > 12 {
		short = short[:12] + "…"
	}
	u.appendMessage(fmt.Sprintf("[System] Friend request sent to %s", short))
	u.addFriendInput.SetText("")
	u.redraw()
}

// ── Thread-safe UI update helpers ─────────────────────────────────────────────

// appendMessage adds a line of text to the chat scroll view.
// Must be called on the UI goroutine.
func (u *UI) appendMessage(msg string) {
	label := wain.NewLabel(msg, wain.Size{Width: 100, Height: 5})
	u.msgCol.Add(label)
}

// SetStatus updates the connection-status label from any goroutine.
func (u *UI) SetStatus(status string) {
	u.app.Notify(func() {
		u.statusLabel.SetText(status)
		u.redraw()
	})
}

// SetToxID updates the Tox-ID display field from any goroutine.
func (u *UI) SetToxID(toxID string) {
	u.app.Notify(func() {
		u.toxIDInput.SetText(toxID)
		u.redraw()
	})
}

// AddFriendToSidebar inserts or updates a friend entry in the sidebar.
// Safe to call from any goroutine.
func (u *UI) AddFriendToSidebar(id uint32, name string, online bool) {
	u.app.Notify(func() {
		u.mu.Lock()
		defer u.mu.Unlock()

		statusDot := "○"
		if online {
			statusDot = "●"
		}
		displayName := fmt.Sprintf("%s %s", statusDot, name)

		if lbl, exists := u.friendLabels[id]; exists {
			// Update existing entry.
			lbl.SetText(displayName)
		} else {
			// Create a new sidebar entry panel and label.
			panel := wain.NewPanel(wain.Size{Width: 100, Height: 8})
			lbl = wain.NewLabel(displayName, wain.Size{Width: 100, Height: 100})
			panel.Add(lbl)

			u.friendPanels[id] = panel
			u.friendLabels[id] = lbl
			u.sidebarCol.Add(panel)
		}

		u.redraw()
	})
}

// ReceiveMessage displays an incoming message in the chat area.
// Safe to call from any goroutine.
func (u *UI) ReceiveMessage(friendID uint32, msg string) {
	u.app.Notify(func() {
		friends := u.backend.getFriends()
		name := fmt.Sprintf("Friend %d", friendID)
		if f, ok := friends[friendID]; ok && f.Name != "" {
			name = f.Name
		}
		u.appendMessage(fmt.Sprintf("%s: %s", name, msg))
		u.redraw()
	})
}

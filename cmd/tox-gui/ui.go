package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/wain"
)

// UIRefs holds references to widgets that need dynamic updates.
type UIRefs struct {
	statusLabel    *wain.Label
	friendList     *wain.Column
	chatHeader     *wain.Label
	messageList    *wain.Column
	messageInput   *wain.TextInput
	toxIdDisplay   *wain.TextInput
	addFriendInput *wain.TextInput
}

// BuildUIWithActions constructs the complete widget tree, wires all button
// actions to the backend, and returns the root widget and UI references.
func BuildUIWithActions(state *AppState, backend *ToxBackend, app *wain.App) (*wain.Column, *UIRefs) {
	refs := &UIRefs{}

	// ─── Header row ───
	header := wain.NewRow()
	header.SetPadding(6)
	header.SetGap(8)
	applyHeaderStyle(header.Panel)

	titleLabel := wain.NewLabel("Tox Messenger", wain.Size{Width: 50, Height: 100})
	titleLabel.SetStyle(wain.StyleOverride{
		Foreground: colorPtr(toxGreenLight),
	})
	header.Add(titleLabel)

	refs.statusLabel = wain.NewLabel("Disconnected", wain.Size{Width: 50, Height: 100})
	refs.statusLabel.SetStyle(wain.StyleOverride{
		Foreground: colorPtr(dimWhite),
	})
	header.Add(refs.statusLabel)

	// ─── Sidebar ───
	sidebar := wain.NewColumn()
	sidebar.SetPadding(4)
	sidebar.SetGap(2)
	applySidebarStyle(sidebar.Panel)

	sidebarTitle := wain.NewLabel("Friends", wain.Size{Width: 100, Height: 8})
	sidebarTitle.SetStyle(wain.StyleOverride{
		Foreground: colorPtr(toxGreen),
	})
	sidebar.Add(sidebarTitle)

	friendScroll := wain.NewScrollView(wain.Size{Width: 100, Height: 92})
	refs.friendList = wain.NewColumn()
	refs.friendList.SetGap(2)
	friendScroll.Add(refs.friendList)
	sidebar.Add(friendScroll)

	// ─── Chat area ───
	chatArea := wain.NewColumn()
	chatArea.SetGap(0)
	applyChatAreaStyle(chatArea.Panel)

	refs.chatHeader = wain.NewLabel("Select a friend to chat", wain.Size{Width: 100, Height: 6})
	refs.chatHeader.SetStyle(wain.StyleOverride{
		Foreground: colorPtr(dimWhite),
	})
	chatArea.Add(refs.chatHeader)

	messageScroll := wain.NewScrollView(wain.Size{Width: 100, Height: 82})
	refs.messageList = wain.NewColumn()
	refs.messageList.SetGap(2)
	messageScroll.Add(refs.messageList)
	chatArea.Add(messageScroll)

	inputRow := wain.NewRow()
	inputRow.SetGap(4)
	inputRow.SetPadding(4)

	refs.messageInput = wain.NewTextInput("Type a message...", wain.Size{Width: 80, Height: 100})
	applyInputStyle(refs.messageInput)
	inputRow.Add(refs.messageInput)

	sendBtn := wain.NewButton("Send", wain.Size{Width: 20, Height: 100})
	applySendButtonStyle(sendBtn)
	WireSendButton(sendBtn, refs, backend, state)
	inputRow.Add(sendBtn)

	chatArea.Add(inputRow)

	// ─── Body: sidebar + chat ───
	body := wain.NewRow()
	body.SetGap(0)

	sidebarPanel := wain.NewPanel(wain.Size{Width: 25, Height: 100})
	sidebarPanel.Add(sidebar)
	body.Add(sidebarPanel)

	chatPanel := wain.NewPanel(wain.Size{Width: 75, Height: 100})
	chatPanel.Add(chatArea)
	body.Add(chatPanel)

	// ─── Footer row ───
	footer := wain.NewRow()
	footer.SetPadding(4)
	footer.SetGap(4)
	applyFooterStyle(footer.Panel)

	refs.addFriendInput = wain.NewTextInput("Paste Tox ID...", wain.Size{Width: 55, Height: 100})
	applyInputStyle(refs.addFriendInput)
	footer.Add(refs.addFriendInput)

	addFriendBtn := wain.NewButton("Add Friend", wain.Size{Width: 15, Height: 100})
	addFriendBtn.SetStyle(wain.StyleOverride{
		Background: colorPtr(toxGreenDark),
		Foreground: colorPtr(wain.White),
	})
	WireAddFriendButton(addFriendBtn, refs, backend)
	footer.Add(addFriendBtn)

	refs.toxIdDisplay = wain.NewTextInput("Your Tox ID", wain.Size{Width: 30, Height: 100})
	refs.toxIdDisplay.SetStyle(wain.StyleOverride{
		Background: colorPtr(headerBg),
		Foreground: colorPtr(dimWhite),
	})
	refs.toxIdDisplay.SetText(state.GetSelfAddress())
	footer.Add(refs.toxIdDisplay)

	// ─── Root column ───
	root := wain.NewColumn()

	headerPanel := wain.NewPanel(wain.Size{Width: 100, Height: 8})
	headerPanel.Add(header)
	root.Add(headerPanel)

	bodyPanel := wain.NewPanel(wain.Size{Width: 100, Height: 84})
	bodyPanel.Add(body)
	root.Add(bodyPanel)

	footerPanel := wain.NewPanel(wain.Size{Width: 100, Height: 8})
	footerPanel.Add(footer)
	root.Add(footerPanel)

	// Build the initial friend list view.
	rebuildFriendList(refs, state, backend)

	return root, refs
}

// WireSendButton connects the send button click to sending a message.
// This is called separately since the button reference is captured during BuildUI.
func WireSendButton(btn *wain.Button, refs *UIRefs, backend *ToxBackend, state *AppState) {
	btn.OnClick(func() {
		text := refs.messageInput.Text()
		if text == "" {
			return
		}
		friendID, ok := state.GetSelectedFriendID()
		if !ok {
			log.Println("no friend selected")
			return
		}
		if err := backend.SendMessage(friendID, text); err != nil {
			log.Printf("send message failed: %v", err)
			return
		}
		refs.messageInput.SetText("")
	})
}

// WireAddFriendButton connects the add-friend button click to adding a friend.
func WireAddFriendButton(btn *wain.Button, refs *UIRefs, backend *ToxBackend) {
	btn.OnClick(func() {
		address := refs.addFriendInput.Text()
		if address == "" {
			return
		}
		if err := backend.AddFriend(address); err != nil {
			log.Printf("add friend failed: %v", err)
			if refs.statusLabel != nil {
				refs.statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			}
			return
		}
		refs.addFriendInput.SetText("")
		if refs.statusLabel != nil {
			refs.statusLabel.SetText("Friend request sent")
		}
	})
}

// rebuildFriendList clears and re-populates the friend list column.
// We rebuild by copying the new column's fields into the existing column at
// the same memory address. This preserves pointer identity: the ScrollView's
// internal adapter still references this *Column, so it sees the updated state.
func rebuildFriendList(refs *UIRefs, state *AppState, backend *ToxBackend) {
	if refs.friendList == nil {
		return
	}

	newList := wain.NewColumn()
	newList.SetGap(2)

	selectedIdx := state.GetSelectedFriend()

	// Pending requests first
	pending := state.GetPendingRequests()
	for _, req := range pending {
		pk := req.PublicKey
		label := fmt.Sprintf("[Request] %.16x", pk[:8])
		btn := wain.NewButton(label, wain.Size{Width: 100, Height: 12})
		applyPendingRequestStyle(btn)
		btn.OnClick(func() {
			if err := backend.AcceptFriendRequest(pk); err != nil {
				log.Printf("accept friend request failed: %v", err)
			}
		})
		newList.Add(btn)
	}

	// Friends
	friends := state.GetFriends()
	for i, f := range friends {
		idx := i
		fID := f.ID
		displayName := f.Name
		if f.UnreadCount > 0 {
			displayName = fmt.Sprintf("(%d) %s", f.UnreadCount, displayName)
		}

		btn := wain.NewButton(displayName, wain.Size{Width: 100, Height: 10})
		if idx == selectedIdx {
			applyFriendSelectedStyle(btn)
		} else if f.Online {
			applyFriendOnlineStyle(btn)
		} else {
			applyFriendOfflineStyle(btn)
		}

		btn.OnClick(func() {
			state.SetSelectedFriend(idx)
			state.ClearUnread(fID)
			rebuildFriendList(refs, state, backend)
			rebuildChatView(refs, state)
		})
		newList.Add(btn)
	}

	// Copy the new column's internal state into the existing column.
	// The ScrollView parent still holds a pointer to refs.friendList,
	// so the in-place update ensures the parent sees the new children.
	*refs.friendList = *newList
}

// rebuildChatView clears and re-populates the message list column.
// Uses the same in-place copy approach as rebuildFriendList to preserve
// the parent ScrollView's reference to the *Column pointer.
func rebuildChatView(refs *UIRefs, state *AppState) {
	if refs.messageList == nil {
		return
	}

	friendID, ok := state.GetSelectedFriendID()
	if !ok {
		refs.chatHeader.SetText("Select a friend to chat")
		// Clear stale messages from any previously selected chat.
		emptyList := wain.NewColumn()
		emptyList.SetGap(2)
		*refs.messageList = *emptyList
		return
	}

	friends := state.GetFriends()
	for _, f := range friends {
		if f.ID == friendID {
			status := "offline"
			if f.Online {
				status = f.ConnectionType
			}
			refs.chatHeader.SetText(fmt.Sprintf("%s (%s)", f.Name, status))
			break
		}
	}

	newList := wain.NewColumn()
	newList.SetGap(2)

	messages := state.GetMessages(friendID)
	for _, msg := range messages {
		prefix := msg.Sender
		if msg.Outgoing {
			prefix = "You"
		}
		text := fmt.Sprintf("[%s] %s: %s",
			msg.Timestamp.Format("15:04"),
			prefix,
			msg.Content,
		)
		lbl := wain.NewLabel(text, wain.Size{Width: 100, Height: 6})
		if msg.Outgoing {
			lbl.SetStyle(wain.StyleOverride{
				Background: colorPtr(messageOwnBg),
				Foreground: colorPtr(wain.White),
			})
		} else {
			lbl.SetStyle(wain.StyleOverride{
				Background: colorPtr(messagePeerBg),
				Foreground: colorPtr(wain.White),
			})
		}
		newList.Add(lbl)
	}

	// Copy the new column's internal state into the existing column.
	*refs.messageList = *newList
}

// publicWidgetAdapter bridges a wain.PublicWidget to the internal wain.Widget
// interface, allowing public widgets to be used with Window.SetRootWidget().
// It embeds BaseWidget for default implementations and delegates event handling
// to the wrapped PublicWidget.
type publicWidgetAdapter struct {
	wain.BaseWidget
	public wain.PublicWidget
}

// adaptPublicWidget wraps a PublicWidget so it satisfies the Widget interface.
func adaptPublicWidget(pw wain.PublicWidget) wain.Widget {
	return &publicWidgetAdapter{public: pw}
}

// Contains returns true if the point is within the widget's bounds.
func (a *publicWidgetAdapter) Contains(x, y float64) bool {
	w, h := a.public.Bounds()
	return x >= 0 && x < float64(w) && y >= 0 && y < float64(h)
}

// HandlePointer delegates pointer events to the wrapped PublicWidget.
func (a *publicWidgetAdapter) HandlePointer(evt *wain.PointerEvent) {
	a.public.HandleEvent(evt)
}

// HandleKey delegates keyboard events to the wrapped PublicWidget.
func (a *publicWidgetAdapter) HandleKey(evt *wain.KeyEvent) {
	a.public.HandleEvent(evt)
}

// HandleTouch delegates touch events to the wrapped PublicWidget.
func (a *publicWidgetAdapter) HandleTouch(evt *wain.TouchEvent) {
	a.public.HandleEvent(evt)
}

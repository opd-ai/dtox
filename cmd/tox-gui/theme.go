package main

import "github.com/opd-ai/wain"

// Tox-branded colors for the dark theme.
var (
	toxGreen      = wain.RGB(76, 175, 80)
	toxGreenDark  = wain.RGB(56, 142, 60)
	toxGreenLight = wain.RGB(129, 199, 132)

	headerBg       = wain.RGB(38, 38, 38)
	sidebarBg      = wain.RGB(32, 32, 32)
	chatBg         = wain.RGB(24, 24, 24)
	inputBg        = wain.RGB(45, 45, 45)
	selectedBg     = wain.RGB(55, 55, 55)
	onlineBg       = wain.RGB(35, 50, 35)
	offlineText    = wain.RGB(120, 120, 120)
	messageOwnBg   = wain.RGB(38, 70, 38)
	messagePeerBg  = wain.RGB(50, 50, 50)
	pendingBg      = wain.RGB(70, 50, 30)
	footerBg       = wain.RGB(35, 35, 35)
	accentFg       = wain.RGB(200, 230, 200)
	dimWhite       = wain.RGB(180, 180, 180)
)

// colorPtr returns a pointer to a Color value.
func colorPtr(c wain.Color) *wain.Color {
	return &c
}

// intPtr returns a pointer to an int value.
func intPtr(v int) *int {
	return &v
}

// applyHeaderStyle styles a panel as the header bar.
func applyHeaderStyle(p *wain.Panel) {
	p.SetStyle(wain.StyleOverride{
		Background: colorPtr(headerBg),
	})
}

// applySidebarStyle styles a panel as the sidebar.
func applySidebarStyle(p *wain.Panel) {
	p.SetStyle(wain.StyleOverride{
		Background: colorPtr(sidebarBg),
	})
}

// applyChatAreaStyle styles a panel as the chat area.
func applyChatAreaStyle(p *wain.Panel) {
	p.SetStyle(wain.StyleOverride{
		Background: colorPtr(chatBg),
	})
}

// applyFriendOnlineStyle styles a button for an online friend.
func applyFriendOnlineStyle(b *wain.Button) {
	b.SetStyle(wain.StyleOverride{
		Background: colorPtr(onlineBg),
		Foreground: colorPtr(toxGreenLight),
	})
}

// applyFriendOfflineStyle styles a button for an offline friend.
func applyFriendOfflineStyle(b *wain.Button) {
	b.SetStyle(wain.StyleOverride{
		Foreground: colorPtr(offlineText),
	})
}

// applyFriendSelectedStyle styles a button for the selected friend.
func applyFriendSelectedStyle(b *wain.Button) {
	b.SetStyle(wain.StyleOverride{
		Background: colorPtr(selectedBg),
		Foreground: colorPtr(wain.White),
	})
}

// applyPendingRequestStyle styles a button for a pending friend request.
func applyPendingRequestStyle(b *wain.Button) {
	b.SetStyle(wain.StyleOverride{
		Background: colorPtr(pendingBg),
		Foreground: colorPtr(accentFg),
	})
}

// applySendButtonStyle styles the send button with the Tox green accent.
func applySendButtonStyle(b *wain.Button) {
	b.SetStyle(wain.StyleOverride{
		Background: colorPtr(toxGreen),
		Foreground: colorPtr(wain.White),
	})
}

// applyInputStyle styles a text input field.
func applyInputStyle(ti *wain.TextInput) {
	ti.SetStyle(wain.StyleOverride{
		Background: colorPtr(inputBg),
		Foreground: colorPtr(wain.White),
	})
}

// applyFooterStyle styles a panel as the footer bar.
func applyFooterStyle(p *wain.Panel) {
	p.SetStyle(wain.StyleOverride{
		Background: colorPtr(footerBg),
	})
}

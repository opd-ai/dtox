package main

import "github.com/opd-ai/dtox/internal/ui"

// Tox-branded colors for the dark theme.
var (
	toxGreen      = ui.RGB(76, 175, 80)
	toxGreenDark  = ui.RGB(56, 142, 60)
	toxGreenLight = ui.RGB(129, 199, 132)

	headerBg       = ui.RGB(38, 38, 38)
	sidebarBg      = ui.RGB(32, 32, 32)
	chatBg         = ui.RGB(24, 24, 24)
	inputBg        = ui.RGB(45, 45, 45)
	selectedBg     = ui.RGB(55, 55, 55)
	onlineBg       = ui.RGB(35, 50, 35)
	offlineText    = ui.RGB(120, 120, 120)
	messageOwnBg   = ui.RGB(38, 70, 38)
	messagePeerBg  = ui.RGB(50, 50, 50)
	pendingBg      = ui.RGB(70, 50, 30)
	footerBg       = ui.RGB(35, 35, 35)
	accentFg       = ui.RGB(200, 230, 200)
	dimWhite       = ui.RGB(180, 180, 180)
)

// colorPtr returns a pointer to a Color value.
func colorPtr(c ui.Color) *ui.Color {
	return &c
}

// intPtr returns a pointer to an int value.
func intPtr(v int) *int {
	return &v
}

// applyHeaderStyle styles a panel as the header bar.
func applyHeaderStyle(p *ui.Panel) {
	p.SetStyle(ui.StyleOverride{
		Background: colorPtr(headerBg),
	})
}

// applySidebarStyle styles a panel as the sidebar.
func applySidebarStyle(p *ui.Panel) {
	p.SetStyle(ui.StyleOverride{
		Background: colorPtr(sidebarBg),
	})
}

// applyChatAreaStyle styles a panel as the chat area.
func applyChatAreaStyle(p *ui.Panel) {
	p.SetStyle(ui.StyleOverride{
		Background: colorPtr(chatBg),
	})
}

// applyFriendOnlineStyle styles a button for an online friend.
func applyFriendOnlineStyle(b *ui.Button) {
	b.SetStyle(ui.StyleOverride{
		Background: colorPtr(onlineBg),
		Foreground: colorPtr(toxGreenLight),
	})
}

// applyFriendOfflineStyle styles a button for an offline friend.
func applyFriendOfflineStyle(b *ui.Button) {
	b.SetStyle(ui.StyleOverride{
		Foreground: colorPtr(offlineText),
	})
}

// applyFriendSelectedStyle styles a button for the selected friend.
func applyFriendSelectedStyle(b *ui.Button) {
	b.SetStyle(ui.StyleOverride{
		Background: colorPtr(selectedBg),
		Foreground: colorPtr(ui.White),
	})
}

// applyPendingRequestStyle styles a button for a pending friend request.
func applyPendingRequestStyle(b *ui.Button) {
	b.SetStyle(ui.StyleOverride{
		Background: colorPtr(pendingBg),
		Foreground: colorPtr(accentFg),
	})
}

// applySendButtonStyle styles the send button with the Tox green accent.
func applySendButtonStyle(b *ui.Button) {
	b.SetStyle(ui.StyleOverride{
		Background: colorPtr(toxGreen),
		Foreground: colorPtr(ui.White),
	})
}

// applyInputStyle styles a text input field.
func applyInputStyle(ti *ui.TextInput) {
	ti.SetStyle(ui.StyleOverride{
		Background: colorPtr(inputBg),
		Foreground: colorPtr(ui.White),
	})
}

// applyFooterStyle styles a panel as the footer bar.
func applyFooterStyle(p *ui.Panel) {
	p.SetStyle(ui.StyleOverride{
		Background: colorPtr(footerBg),
	})
}

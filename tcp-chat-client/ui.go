package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type ChatUI struct {
	app     *tview.Application
	chatBox *tview.TextView
	input   *tview.InputField
	layout  *tview.Flex
	OnSend  func(string)
	OnQuit  func()
}

func NewChatUI() *ChatUI {
	app := tview.NewApplication()
	ui := &ChatUI{
		app: app,
	}

	ui.initChatBox()
	ui.initInput()
	ui.createLayout()
	ui.setupHandlers()

	return ui
}

func (ui *ChatUI) initChatBox() {
	ui.chatBox = tview.NewTextView()
	ui.chatBox.
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		}).
		SetBorder(true).
		SetTitle("💬 Chat Log")
}

func (ui *ChatUI) initInput() {
	ui.input = tview.NewInputField()
	ui.input.
		SetLabel("You: ").
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorGreen).
		SetBorder(true).
		SetTitle("Enter messages").
		SetTitleColor(tcell.ColorLightCyan)
}

func (ui *ChatUI) createLayout() {
	ui.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.chatBox, 0, 1, false).
		AddItem(ui.input, 3, 0, true)

	ui.app.SetRoot(ui.layout, true).EnableMouse(true)
}

func (ui *ChatUI) setupHandlers() {
	ui.input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			text := ui.input.GetText()
			if text != "" {
				if ui.OnSend != nil {
					ui.OnSend(text)
				}
				ui.AddMessage("[green]You:[-] " + text)
				ui.input.SetText("")
			}
		case tcell.KeyEsc:
			if ui.OnQuit != nil {
				ui.OnQuit()
			}
		}
	})
}

func (ui *ChatUI) AddMessage(msg string) {
	msg = strings.TrimSpace(msg)
	ui.chatBox.Write([]byte(msg + "\n"))
	ui.chatBox.ScrollToEnd()
}

func (ui *ChatUI) Run() {
	if err := ui.app.Run(); err != nil {
		panic(err)
	}
}

func ShowConnectionDialog() (string, string) {
	app := tview.NewApplication()
	form := tview.NewForm()

	form.AddInputField("Username:", "", 20, nil, nil).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetBorder(true)

	form.AddInputField("Server Address:", "localhost:9000", 20, nil, nil).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetBorder(true)

	form.AddButton("Connect", nil)
	form.SetBorder(true).
		SetTitle("Connect to Chat Server").
		SetTitleColor(tcell.ColorLightCyan)

	var username, address string
	form.GetButton(0).SetSelectedFunc(func() {
		username = form.GetFormItem(0).(*tview.InputField).GetText()
		address = form.GetFormItem(1).(*tview.InputField).GetText()
		app.Stop()
	})

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}

	return username, address
}

func (ui *ChatUI) clearChat() {
	ui.chatBox.Clear()
	ui.chatBox.ScrollToBeginning()
}

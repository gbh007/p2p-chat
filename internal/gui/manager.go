package gui

import (
	"errors"
	"slices"

	"github.com/awesome-gocui/gocui"
	"github.com/gbh007/p2p-chat/internal/entities"
)

const (
	chatConnectNameViewName = "chat-name"
	chatListViewName        = "list"
	chatHistoryViewName     = "chat/"
	chatMessageViewName     = "message"
)

type callbacker interface {
	SendMessage(chat, msg string)
	Connect(name string)
}

type Manager struct {
	g *gocui.Gui

	callbacker callbacker

	currentChatName string
}

func New(callbacker callbacker) *Manager {
	return &Manager{
		callbacker:      callbacker,
		currentChatName: "chat 3",
	}
}

func (gm *Manager) Init() error {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return err
	}

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen

	gm.g = g

	g.SetManager(gm)

	err = gm.setKeybinding()
	if err != nil {
		return err
	}

	return nil
}

func (gm *Manager) MainLoop() error {
	defer gm.g.Close()

	err := gm.g.MainLoop()
	if err != nil && !errors.Is(err, gocui.ErrQuit) {
		return err
	}

	return nil
}

func (gm *Manager) setKeybinding() error {
	if err := gm.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, gm.quit); err != nil {
		return err
	}

	if err := gm.g.SetKeybinding("", 'q', gocui.ModNone, gm.quit); err != nil {
		return err
	}

	if err := gm.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, gm.nextView); err != nil {
		return err
	}

	return nil
}

func (gm *Manager) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (gm *Manager) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	chatSelectorX := maxX/3 - 2

	if chatSelectorX > 20 {
		chatSelectorX = 20
	}

	if v, err := g.SetView(chatConnectNameViewName, 0, 0, chatSelectorX, 2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Connect to"
		v.Editor = gocui.EditorFunc(gm.editChatName)
		v.Editable = true

		_, err = g.SetCurrentView(chatConnectNameViewName)
		if err != nil {
			return err
		}

		g.Cursor = true
	}

	if v, err := g.SetView(chatListViewName, 0, 3, chatSelectorX, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Chat list"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		if err := gm.g.SetKeybinding(chatListViewName, gocui.KeyArrowUp, gocui.ModNone, gm.prevChat); err != nil {
			return err
		}

		if err := gm.g.SetKeybinding(chatListViewName, 'k', gocui.ModNone, gm.prevChat); err != nil {
			return err
		}

		if err := gm.g.SetKeybinding(chatListViewName, gocui.KeyArrowDown, gocui.ModNone, gm.nextChat); err != nil {
			return err
		}

		if err := gm.g.SetKeybinding(chatListViewName, 'j', gocui.ModNone, gm.nextChat); err != nil {
			return err
		}
	}

	if v, err := g.SetView(chatMessageViewName, chatSelectorX+2, maxY-3, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Message"
		v.Editor = gocui.EditorFunc(gm.editMessage)
		v.Editable = true
		v.Visible = false
	}

	return nil
}

func (gm *Manager) nextView(g *gocui.Gui, v *gocui.View) error {
	if gm.currentChatName == "" {
		return nil
	}

	viewNames := map[string]string{
		chatConnectNameViewName:                  chatListViewName,
		chatListViewName:                         chatHistoryViewName + gm.currentChatName,
		chatHistoryViewName + gm.currentChatName: chatMessageViewName,
		chatMessageViewName:                      chatConnectNameViewName,
	}

	next, ok := viewNames[v.Name()]
	if !ok {
		return nil
	}

	nextView, err := g.SetCurrentView(next)
	if err != nil {
		return err
	}

	if nextView.Name() == chatMessageViewName ||
		nextView.Name() == chatConnectNameViewName {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	return nil
}

func (gm *Manager) HandleMessage(msg entities.Message) {
	gm.g.Update(func(g *gocui.Gui) error {
		v, err := g.View(chatHistoryViewName + msg.Chat)
		if err != nil {
			return err
		}

		err = writeMessage(v, msg)
		if err != nil {
			return err
		}

		return nil
	})
}

func (gm *Manager) editMessage(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if key == gocui.KeyEnter {
		msg := v.Buffer()
		gm.callbacker.SendMessage(gm.currentChatName, msg)
		v.Clear()

		return
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

func (gm *Manager) editChatName(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if key == gocui.KeyEnter {
		msg := v.Buffer()
		gm.callbacker.Connect(msg)
		v.Clear()

		_, _ = gm.g.SetCurrentView(chatListViewName)
		gm.g.Cursor = false

		return
	}

	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

func writeMessage(v *gocui.View, msg entities.Message) error {
	v.WriteString(msg.TS.Format("15:04:05"))
	v.WriteString(" (")

	if !msg.IsOwn {
		v.WriteString(msg.User)
	}

	if !msg.IsLocalDomain || msg.IsOwn {
		v.WriteString("@")
	}

	if !msg.IsLocalDomain {
		v.WriteString(msg.Domain)
	}

	v.WriteString("): ")
	v.WriteString(msg.Text)
	v.WriteString("\n")

	err := v.SetHighlight(v.LinesHeight()-2, msg.IsOwn)
	if err != nil {
		return err
	}

	return nil
}

func (gm *Manager) nextChat(g *gocui.Gui, v *gocui.View) error {
	chats := v.BufferLines()

	if len(chats) == 0 {
		return nil
	}

	index := slices.Index(chats, gm.currentChatName)
	nextIndex := (index + 1) % len(chats)

	if index > -1 {
		err := v.SetHighlight(index, false)
		if err != nil {
			return err
		}

		cv, err := g.View(chatHistoryViewName + gm.currentChatName)
		if err != nil {
			return err
		}

		cv.Visible = false
	}

	err := v.SetHighlight(nextIndex, true)
	if err != nil {
		return err
	}

	gm.currentChatName = chats[nextIndex]

	cv, err := g.View(chatHistoryViewName + gm.currentChatName)
	if err != nil {
		return err
	}

	cv.Visible = true

	return nil
}

func (gm *Manager) prevChat(g *gocui.Gui, v *gocui.View) error {
	chats := v.BufferLines()

	if len(chats) == 0 {
		return nil
	}

	index := slices.Index(chats, gm.currentChatName)
	nextIndex := (len(chats) + index - 1) % len(chats)

	if index > -1 {
		err := v.SetHighlight(index, false)
		if err != nil {
			return err
		}

		cv, err := g.View(chatHistoryViewName + gm.currentChatName)
		if err != nil {
			return err
		}

		cv.Visible = false
	}

	if nextIndex > -1 {
		err := v.SetHighlight(nextIndex, true)
		if err != nil {
			return err
		}

		gm.currentChatName = chats[nextIndex]

		cv, err := g.View(chatHistoryViewName + gm.currentChatName)
		if err != nil {
			return err
		}

		cv.Visible = true
	}

	return nil
}

func (gm *Manager) NewChat(name string) {
	gm.g.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		chatSelectorX := maxX/3 - 2

		if chatSelectorX > 20 {
			chatSelectorX = 20
		}

		if v, err := g.SetView(chatHistoryViewName+name, chatSelectorX+2, 0, maxX-1, maxY-4, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}

			v.Title = "Chat " + name
			v.Autoscroll = true
			v.Wrap = true
			v.Visible = false

			lView, err := g.View(chatListViewName)
			if err != nil {
				return err
			}

			if lView.LinesHeight() > 0 {
				lView.WriteString("\n")
			}

			lView.WriteString(name)

			if lView.LinesHeight() == 1 {
				err = lView.SetHighlight(0, true)
				if err != nil {
					return err
				}

				v.Visible = true
				gm.currentChatName = name

				mView, err := g.View(chatMessageViewName)
				if err != nil {
					return err
				}

				mView.Visible = true
			}
		}

		return nil
	})
}

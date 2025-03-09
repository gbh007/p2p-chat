package gui

import (
	"errors"
	"time"

	"github.com/awesome-gocui/gocui"
	"github.com/gbh007/p2p-chat/internal/entities"
)

const (
	chatListViewName    = "list"
	chatHistoryViewName = "chat"
	chatMessageViewName = "message"
)

type callbacker interface {
	SendMessage(chat, msg string)
}

type Manager struct {
	g *gocui.Gui

	callbacker callbacker

	currentChatName string
}

func New(callbacker callbacker) *Manager {
	return &Manager{
		callbacker: callbacker,
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

	if v, err := g.SetView(chatListViewName, 0, 0, maxX/3-2, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Chat list"
		v.SelBgColor = gocui.ColorGreen

		for _, val := range []string{
			"chat 1",
			"chat 2",
			"chat 3",
			"chat 4",
		} {
			v.WriteString(val)
			v.WriteString("\n")
		}

		err := v.SetHighlight(2, true)
		if err != nil {
			return err
		}

		_, err = g.SetCurrentView(chatListViewName)
		if err != nil {
			return err
		}
	}

	if v, err := g.SetView(chatHistoryViewName, maxX/3, 0, maxX-1, maxY-4, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Chat"
		v.Autoscroll = true
		v.Wrap = true

		for _, val := range []entities.Message{
			{
				User:          "",
				IsOwn:         true,
				Domain:        "",
				Text:          "hello",
				TS:            time.Date(0, 0, 0, 1, 14, 15, 0, time.UTC),
				IsLocalDomain: true,
			},
			{
				User:          "user1",
				IsOwn:         false,
				Domain:        "local",
				Text:          "hello",
				TS:            time.Date(0, 0, 0, 1, 14, 16, 0, time.UTC),
				IsLocalDomain: false,
			},
			{
				User:          "user2",
				IsOwn:         false,
				Domain:        "",
				Text:          "hello",
				TS:            time.Date(0, 0, 0, 1, 14, 17, 0, time.UTC),
				IsLocalDomain: true,
			},
			{
				User:          "user3",
				IsOwn:         false,
				Domain:        "example.com",
				Text:          "hello",
				TS:            time.Date(0, 0, 0, 1, 14, 18, 0, time.UTC),
				IsLocalDomain: false,
			},
			{
				User:          "",
				IsOwn:         true,
				Domain:        "",
				Text:          "hello",
				TS:            time.Date(0, 0, 0, 1, 14, 19, 0, time.UTC),
				IsLocalDomain: true,
			},
		} {
			err = writeMessage(v, val)
			if err != nil {
				return err
			}
		}
	}

	if v, err := g.SetView(chatMessageViewName, maxX/3, maxY-3, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Message"
		v.Editor = gocui.EditorFunc(gm.editMessage)
		v.Editable = true
	}

	return nil
}

func (gm *Manager) nextView(g *gocui.Gui, v *gocui.View) error {
	viewNames := map[string]string{
		chatListViewName:    chatHistoryViewName,
		chatHistoryViewName: chatMessageViewName,
		chatMessageViewName: chatListViewName,
	}

	next, ok := viewNames[v.Name()]
	if !ok {
		return nil
	}

	nextView, err := g.SetCurrentView(next)
	if err != nil {
		return err
	}

	if nextView.Name() == chatMessageViewName {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	return nil
}

func (gm *Manager) HandleMessage(msg entities.Message) {
	gm.g.Update(func(g *gocui.Gui) error {
		v, err := g.View(chatHistoryViewName)
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

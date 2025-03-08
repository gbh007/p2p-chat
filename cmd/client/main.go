package main

import (
	"errors"

	"github.com/awesome-gocui/gocui"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		panic(err)
	}

	defer g.Close()

	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SelFrameColor = gocui.ColorGreen

	gm := &GUIManager{}

	g.SetManager(gm)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, gm.nextView); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		panic(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

type GUIManager struct{}

func (gm *GUIManager) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("list", 0, 0, maxX/3-2, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Chat list"

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

		_, err = g.SetCurrentView("list")
		if err != nil {
			return err
		}
	}

	if v, err := g.SetView("chat", maxX/3, 0, maxX-1, maxY-4, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Chat"
		v.Autoscroll = true
		v.Wrap = true

		for i, val := range []string{
			"13:14:00 (@): hello",
			"13:14:15 (user 1@localhost): hello",
			"13:14:16 (user 2@localhost): hello",
			"13:14:17 (user 3): hello",
			"13:14:18 (user 4@localhost): hello",
			"13:14:19 (@): hello",
			"13:14:20 (@): hello",
		} {
			v.WriteString(val)
			v.WriteString("\n")

			if i == 0 || i == 5 || i == 6 {
				err := v.SetHighlight(i, true)
				if err != nil {
					return err
				}
			}
		}
	}

	if v, err := g.SetView("message", maxX/3, maxY-3, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}

		v.Title = "Message"
		v.Editor = gocui.DefaultEditor
		v.Editable = true
	}

	return nil
}

func (gm *GUIManager) nextView(g *gocui.Gui, v *gocui.View) error {
	viewNames := map[string]string{
		"list":    "chat",
		"chat":    "message",
		"message": "list",
	}

	next, ok := viewNames[v.Name()]
	if !ok {
		return nil
	}

	nextView, err := g.SetCurrentView(next)
	if err != nil {
		return err
	}

	if nextView.Name() == "message" {
		g.Cursor = true
	} else {
		g.Cursor = false
	}

	return nil
}

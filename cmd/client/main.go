package main

import (
	"time"

	"github.com/gbh007/p2p-chat/internal/entities"
	"github.com/gbh007/p2p-chat/internal/gui"
)

type controllerMock struct {
	ch chan entities.Message

	gui interface {
		HandleMessage(msg entities.Message)
	}
}

func newControllerMock() *controllerMock {
	return &controllerMock{
		ch: make(chan entities.Message, 10),
	}
}

func (c *controllerMock) SendMessage(chat, msg string) {
	c.ch <- entities.Message{
		Chat:          chat,
		User:          "myuser",
		Domain:        "localhost",
		Text:          msg,
		TS:            time.Now(),
		IsOwn:         true,
		IsLocalDomain: true,
	}
}

func (c *controllerMock) SetGUI(gui interface {
	HandleMessage(msg entities.Message)
}) {
	c.gui = gui
}

func (c *controllerMock) Serve() {
	for v := range c.ch {
		c.gui.HandleMessage(v)
	}
}

func main() {
	cm := newControllerMock()

	gm := gui.New(cm)
	err := gm.Init()
	if err != nil {
		panic(err)
	}

	cm.SetGUI(gm)

	go cm.Serve()

	gm.NewChat("chat 1")
	gm.NewChat("chat 10")
	gm.NewChat("chat 11")

	err = gm.MainLoop()
	if err != nil {
		panic(err)
	}
}

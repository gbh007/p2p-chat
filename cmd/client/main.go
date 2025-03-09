package main

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/gbh007/p2p-chat/internal/entities"
	"github.com/gbh007/p2p-chat/internal/gui"
	"github.com/gbh007/p2p-chat/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ControllerMock struct {
	ch chan entities.Message

	gui interface {
		HandleMessage(msg entities.Message)
		NewChat(name string)
	}
}

func NewControllerMock() *ControllerMock {
	return &ControllerMock{
		ch: make(chan entities.Message, 10),
	}
}

func (c *ControllerMock) SendMessage(chat, msg string) {
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

func (c *ControllerMock) SetGUI(gui interface {
	HandleMessage(msg entities.Message)
	NewChat(name string)
}) {
	c.gui = gui
}

func (c *ControllerMock) Serve() {
	for v := range c.ch {
		c.gui.HandleMessage(v)
	}
}

func (c *ControllerMock) Connect(name string) {
	c.gui.NewChat(name)
}

func main() {
	// cm := NewControllerMock()
	cm, err := NewControllerGRPC()
	if err != nil {
		panic(err)
	}

	gm := gui.New(cm)
	err = gm.Init()
	if err != nil {
		panic(err)
	}

	cm.SetGUI(gm)

	go cm.Serve()

	err = gm.MainLoop()
	if err != nil {
		panic(err)
	}
}

type ControllerGRPC struct {
	client gen.ServerClient
	conn   *grpc.ClientConn
	login  string

	ch chan entities.Message

	gui interface {
		HandleMessage(msg entities.Message)
		NewChat(name string)
	}
}

func NewControllerGRPC() (*ControllerGRPC, error) {
	c := &ControllerGRPC{
		ch:    make(chan entities.Message, 10),
		login: strconv.Itoa(rand.Int()),
	}

	err := c.connect()

	return c, err
}

func (c *ControllerGRPC) SendMessage(chat, msg string) {
	_, err := c.client.SendMessage(context.Background(), &gen.SendMessageRequest{
		Login:   c.login,
		Channel: chat,
		Message: msg,
	})
	if err != nil {
		panic(err)
	}
}

func (c *ControllerGRPC) SetGUI(gui interface {
	HandleMessage(msg entities.Message)
	NewChat(name string)
}) {
	c.gui = gui
}

func (c *ControllerGRPC) connect() error {
	conn, err := grpc.NewClient(
		"localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = gen.NewServerClient(conn)

	return nil
}

func (c *ControllerGRPC) Serve() {}

func (c *ControllerGRPC) Connect(name string) {
	ctx, cancel := context.WithCancel(context.TODO())

	res, err := c.client.ReadMessages(ctx, &gen.ReadMessagesRequest{
		Channel: name,
		Login:   c.login,
	})
	if err == nil {
		c.gui.NewChat(name)
	} else {
		panic(err)
	}

	go func() {
		defer cancel()

		for {
			msg, err := res.Recv()
			if err != nil {
				return
			}

			c.gui.HandleMessage(entities.Message{
				Chat:          name,
				User:          msg.GetLogin(),
				Text:          msg.GetMessage(),
				TS:            msg.GetTs().AsTime(),
				IsOwn:         msg.GetLogin() == c.login,
				IsLocalDomain: true,
			})
		}
	}()
}

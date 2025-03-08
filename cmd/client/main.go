package main

import "github.com/gbh007/p2p-chat/internal/gui"

func main() {
	gm := &gui.Manager{}
	err := gm.Init()
	if err != nil {
		panic(err)
	}

	err = gm.MainLoop()
	if err != nil {
		panic(err)
	}
}

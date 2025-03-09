package entities

import "time"

type Message struct {
	Chat          string
	User          string
	Domain        string
	Text          string
	TS            time.Time
	IsOwn         bool
	IsLocalDomain bool
}

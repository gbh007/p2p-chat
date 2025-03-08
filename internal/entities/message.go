package entities

import "time"

type Message struct {
	User          string
	Domain        string
	Text          string
	TS            time.Time
	IsOwn         bool
	IsLocalDomain bool
}

package network

import "time"

//go:generate mockery --name=Websocket --outpkg=network --inpackage=true
type Websocket interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	SetWriteDeadline(t time.Time) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	Close() error
}

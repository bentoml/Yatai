package controllersv1

import "github.com/gorilla/websocket"

type baseController struct{}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

package controllersv1

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/bentoml/yatai-schemas/schemasv1"
)

func writeWsError(conn *websocket.Conn, err error) {
	if err == nil {
		return
	}

	msg := schemasv1.WsRespSchema{
		Type:    schemasv1.WsRespTypeError,
		Message: err.Error(),
		Payload: nil,
	}
	err_ := conn.WriteJSON(&msg)
	if err_ != nil {
		logrus.Errorf("ws write error: %q", err_.Error())
	}
}

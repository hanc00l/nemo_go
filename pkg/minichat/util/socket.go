package util

import (
	"github.com/gorilla/websocket"
	"log"
)

func SocketSend(conn *websocket.Conn, byteData []byte) error {
	err := conn.WriteMessage(websocket.TextMessage, byteData)
	if err != nil {
		log.Printf("websocket send error, error is %+v", err)
		return err
	}
	return nil
}

func SocketReceive(conn *websocket.Conn) ([]byte, error) {
	_, message, err := conn.ReadMessage()
	//log.Printf("receive message, error is %+v, message type is %d", err, messageType)
	if err != nil {
		return nil, err
		//err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
	}
	return message, nil
}

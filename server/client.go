package server

import (
	"Hermes/common/dto"
	"github.com/gorilla/websocket"
)

type Client struct {
	Info dto.ClientInfo
	Conn *websocket.Conn
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

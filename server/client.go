package server

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	DisplayName string
	Conn        *websocket.Conn
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

package server

import (
	"Hermes/common/dto"
	"Hermes/common/util"
	"fmt"
	"github.com/gorilla/websocket"
)

type Client struct {
	Info *dto.ClientInfo
	Conn *websocket.Conn
}

func (c *Client) SimpleInfo() (s string) {
	info := c.Info
	if info == nil {
		s = "未知设备"
	} else {
		s = fmt.Sprintf("%s(%s/%s)", c.Info.Name, c.Info.Os, c.Info.Arch)
	}
	return
}

func (c *Client) Ip() string {
	ip, _, _ := util.GetIpPortFromAddr(c.Conn.RemoteAddr().String())
	return ip
}

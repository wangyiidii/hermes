package server

import (
	"Hermes/common/dto"
	"Hermes/common/message"
	"container/list"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	clients *list.List
}

var (
	GlobalServer *Server // 全局Server
)

func init() {
	GlobalServer = &Server{
		clients: list.New(),
	}
}

// Start 启动hermes服务
func (s *Server) Start(websocketPort int) {

	var addr = flag.String("addr", ":"+strconv.Itoa(websocketPort), "websocket service address")
	http.HandleFunc("/clipboard", clipboard)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalln("服务启动失败: ", err)
		return
	}

}

func clipboard(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("升级websocket失败:", err)
		return
	}

	GlobalServer.AddClient(conn)

	// 监听客户端上报的
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				GlobalServer.RemoveClient(conn)
				log.Println("server readMessage err: ", err)
				return
			}

			// just for test
			client := GlobalServer.getClient(conn)
			if client != nil && len(client.Info.Name) > 0 {
				log.Printf("%s: %s \n", client.Info.Name, message)
			}

			GlobalServer.processMessage(conn, message)
		}
	}()
}

// 通过websocket.Conn获取Client
func (s *Server) getClient(conn *websocket.Conn) *Client {
	var c *Client
	for i := GlobalServer.clients.Front(); i != nil; i = i.Next() {
		cli := i.Value.(*Client)
		if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
			c = cli
		}
	}
	return c
}

// AddClient 添加一个Client
func (s *Server) AddClient(conn *websocket.Conn) {
	client := NewClient(conn)
	s.clients.PushBack(client)
}

// RemoveClient 移除一个client
func (s *Server) RemoveClient(conn *websocket.Conn) {
	for i := GlobalServer.clients.Front(); i != nil; i = i.Next() {
		cli := i.Value.(*Client)
		if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
			log.Printf("client %s disconnected \n", cli.Conn.RemoteAddr())
			GlobalServer.clients.Remove(i)
			continue
		}
	}
}

// 处理客户端上报的信息
func (s *Server) processMessage(conn *websocket.Conn, messageByte []byte) {
	// 反序列化消息
	var mes message.Message
	err := json.Unmarshal(messageByte, &mes)
	if err != nil {
		return
	}

	switch mes.Typ {
	case message.OnlineType:
		data, err := json.Marshal(mes.Data)
		if err != nil {
			return
		}
		var clientInfo dto.ClientInfo
		json.Unmarshal(data, &clientInfo)

		client := s.getClient(conn)
		client.Info = clientInfo

		log.Printf("%s(%s/%s)上线\n", clientInfo.Name, clientInfo.Os, clientInfo.Arch)
	case message.ClipboardType:
		content := mes.Data.(string)
		for i := s.clients.Front(); i != nil; i = i.Next() {
			cli := i.Value.(*Client)
			if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
				continue
			}
			cli.Conn.WriteMessage(websocket.TextMessage, []byte(content))
		}
	default:
		log.Fatalln("processMessage type err")
		return
	}

}

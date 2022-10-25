package server

import (
	"Hermes/common/dto"
	"Hermes/common/message"
	"Hermes/common/util"
	"container/list"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	Mux     sync.Mutex
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

	// 加入缓存，ip相同的，服务器发送断开纤细，并主动断开
	GlobalServer.Mux.Lock()
	exist, client := GlobalServer.ClientExist(conn)
	if exist {

		conn.WriteJSON(message.NewMessage(message.DisconnectType,
			fmt.Sprintf("相同IP客户端[%s:%s]已连接，本连接服务器主动断开", client.SimpleInfo(), client.Conn.RemoteAddr()),
		))
		conn.Close()
		GlobalServer.Mux.Unlock()
		return
	}
	GlobalServer.AddClient(conn)
	GlobalServer.Mux.Unlock()

	// 监听客户端上报的
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("服务端读取信息错误: ", err)
				GlobalServer.RemoveClient(conn)
				return
			}

			GlobalServer.processMessage(conn, message)
		}
	}()
}

func (s *Server) ClientExist(conn *websocket.Conn) (exist bool, client *Client) {
	ipParam, _, _ := util.GetIpPortFromAddr(conn.RemoteAddr().String())
	if len(ipParam) == 0 {
		return false, nil
	}

	for i := s.clients.Front(); i != nil; i = i.Next() {
		client = i.Value.(*Client)
		ip := client.Ip()
		if len(ip) > 0 && len(ipParam) > 0 && ip == ipParam {
			return true, client
		}
	}

	return false, nil

}

// 通过websocket.Conn获取Client
func (s *Server) getClient(conn *websocket.Conn) *Client {
	var c *Client
	for i := s.clients.Front(); i != nil; i = i.Next() {
		cli := i.Value.(*Client)
		if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
			c = cli
		}
	}
	return c
}

// AddClient 添加一个Client
func (s *Server) AddClient(conn *websocket.Conn) {
	client := &Client{
		Conn: conn,
	}
	s.clients.PushBack(client)
}

// RemoveClient 移除一个client
func (s *Server) RemoveClient(conn *websocket.Conn) {
	for i := s.clients.Front(); i != nil; i = i.Next() {
		cli := i.Value.(*Client)
		if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
			s.Mux.Lock()
			log.Printf("客户端 %s 断开连接 \n", cli.Conn.RemoteAddr())
			s.clients.Remove(i)
			conn.Close()
			s.Mux.Unlock()
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
		log.Println("processMessage json err: ", err)
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
		client.Info = &clientInfo

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

package server

import (
	"Hermes/message"
	"container/list"
	"encoding/json"
	"flag"
	"fmt"
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

	//go func() {
	//	// 1.获取ticker对象
	//	ticker := time.NewTicker(1 * time.Second * 5)
	//	i := 0
	//	// 子协程
	//	go func() {
	//		for {
	//			//<-ticker.C
	//			i++
	//			<-ticker.C
	//
	//			fmt.Println("=================")
	//			for i := GlobalServer.clients.Front(); i != nil; i = i.Next() {
	//				cli := i.Value.(*Client)
	//
	//				fmt.Println("cli cli: ", cli.Conn.RemoteAddr())
	//
	//			}
	//		}
	//	}()
	//	for {
	//	}
	//}()

	var addr = flag.String("addr", ":"+strconv.Itoa(websocketPort), "websocket service address")
	http.HandleFunc("/echo", echo)
	http.ListenAndServe(*addr, nil)

}

func echo(w http.ResponseWriter, r *http.Request) {
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

			processMessage(conn, message)

		}
	}()
}

func (s *Server) AddClient(conn *websocket.Conn) {
	client := NewClient(conn)
	s.clients.PushBack(client)
}

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

func processMessage(conn *websocket.Conn, messageByte []byte) {
	// 反序列化消息
	var mes message.Message
	err := json.Unmarshal(messageByte, &mes)
	if err != nil {
		return
	}

	log.Printf("processMessage mes: %#v", mes)

	switch mes.Typ {
	case message.ClientInfoType:
		data, err := json.Marshal(mes.Data)
		if err != nil {
			return
		}
		var clientInfo message.ClientInfo
		json.Unmarshal(data, &clientInfo)
		log.Println("clientInfo: ", clientInfo.Name)
	case message.ClipboardType:
		content := mes.Data.(string)
		log.Println("ClipboardType: ", content)
		for i := GlobalServer.clients.Front(); i != nil; i = i.Next() {
			cli := i.Value.(*Client)
			if cli.Conn.RemoteAddr() == conn.RemoteAddr() {
				continue
			}
			fmt.Println("cli.Conn send msg: ", cli.Conn.RemoteAddr())
			cli.Conn.WriteMessage(websocket.TextMessage, []byte(content))
		}
	default:
		log.Fatalln("processMessage type err")
		return
	}

}

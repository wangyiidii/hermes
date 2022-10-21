package client

import (
	"Hermes/common/dto"
	"Hermes/common/message"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// GlobalClient 全局client
var GlobalClient *Client

type Client struct {
	Conn *websocket.Conn
	Mux  sync.RWMutex
}

// 初始化全局客户端
func initGlobalClient(conn *websocket.Conn) *Client {
	GlobalClient = &Client{
		Conn: conn,
	}
	return GlobalClient
}

func Start(host string, serverPort int) {
	// 连接websocket
	u := url.URL{Scheme: "ws", Host: host + ":" + strconv.Itoa(serverPort), Path: "/clipboard"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("服务器连接接失败: ", err)
	}
	defer conn.Close()
	log.Println("服务器连接成功")

	// 初始化GlobalClient
	initGlobalClient(conn)

	// 发送上线消息
	GlobalClient.sendOnlineMessage()

	// 测试: 打印服务端推送的消息
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Fatalln("client test read :", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	// 监听粘贴板, 并写给服务端
	GlobalClient.watchClipboard()
}

// 写数据给服务端
func (c *Client) writeMessage(msg *message.Message) {
	// 序列化
	d, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// 写数据
	c.Mux.Lock()
	c.Conn.WriteMessage(websocket.TextMessage, d)
	c.Mux.Unlock()

}

// 发送上线消息
func (c *Client) sendOnlineMessage() {
	// 上线时上报客户端信息
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("获取hostname异常: ", err)
		return
	}

	// Message
	msg := &message.Message{
		Typ: message.OnlineType,
		Data: &dto.ClientInfo{
			Name: hostname,
			Os:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}

	// 上报
	c.writeMessage(msg)
}

// 监听粘贴板
func (c *Client) watchClipboard() {
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		msg := &message.Message{
			Typ:  message.ClipboardType,
			Data: string(data),
		}

		c.writeMessage(msg)
	}
}

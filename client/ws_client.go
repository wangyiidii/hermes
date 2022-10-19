package client

import (
	"Hermes/message"
	"context"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// GlobalClient 全局client
var GlobalClient *Client

type Client struct {
	Conn *websocket.Conn
	Mux  sync.RWMutex
}

func initGlobalClient(conn *websocket.Conn) *Client {
	GlobalClient = &Client{
		Conn: conn,
	}
	return GlobalClient
}

func Start(host string, serverPort int) {
	// 连接websocket
	var addr = flag.String("addr", host+":"+strconv.Itoa(serverPort), "http service address")

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("正在连接websocket %s...", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("websocket连接失败: ", err)
	}
	defer conn.Close()

	// 初始化GlobalClient
	initGlobalClient(conn)

	// 监听粘贴板, 并写给服务端
	go func() {
		ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
		for data := range ch {
			msg := &message.Message{
				Typ:  message.ClipboardType,
				Data: string(data),
			}
			d, err := json.Marshal(msg)
			if err != nil {
				return
			}
			GlobalClient.Mux.Lock()
			GlobalClient.Conn.WriteMessage(websocket.TextMessage, d)
			GlobalClient.Mux.Unlock()
		}
	}()

	// 定时上报客户端信息
	go func() {
		// 1.获取ticker对象
		ticker := time.NewTicker(1 * time.Second * 5)
		// 子协程
		go func() {
			for {
				<-ticker.C

				// 获取hostname
				hostname, err := os.Hostname()
				if err != nil {
					log.Println("获取hostname异常: ", err)
					return
				}

				// ClientInfo
				clientInfo := &message.ClientInfo{
					Name: hostname,
					Os:   runtime.GOOS,
					Arch: runtime.GOARCH,
				}

				// Message
				msg := &message.Message{
					Typ:  message.ClientInfoType,
					Data: clientInfo,
				}
				d, err := json.Marshal(msg)
				if err != nil {
					return
				}

				GlobalClient.Mux.Lock()
				GlobalClient.Conn.WriteMessage(websocket.TextMessage, d)
				GlobalClient.Mux.Unlock()
			}
		}()
		for {
		}
	}()

	// 测试打印服务端推送的消息
	func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

}

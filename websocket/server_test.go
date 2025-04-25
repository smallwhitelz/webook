package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	upgrader := websocket.Upgrader{}
	// 我们假定，websocket请求发到这里
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		// responseHeader 不传
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			writer.Write([]byte("初始化 websocket 失败"))
			return
		}
		// 你要源源不断的从 conn 读取数据
		ws := &Ws{conn: conn}
		go func() {
			ws.ReadCycle()
		}()
		go func() {
			ticker := time.NewTicker(time.Second)
			for now := range ticker.C {
				ws.Write("来自服务端的数据：" + now.String())
			}
		}()
	})
	http.ListenAndServe(":8081", nil)
}

type Ws struct {
	conn *websocket.Conn
}

func (w *Ws) Write(msg string) {
	err := w.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	// 记录下日志，可以考虑关闭连接
	if err != nil {
		log.Println(err)
	}
}

func (w *Ws) ReadCycle() {
	conn := w.conn
	for {
		typ, msg, err := conn.ReadMessage()
		if err != nil {
			// 正常都是代表出了问题，你可以退出循环
			// 记录日志
			return
		}
		switch typ {
		case websocket.CloseMessage:
			conn.Close()
			return
		case websocket.BinaryMessage, websocket.TextMessage:
			log.Println(string(msg))
		default:
			// 不需要管
		}
	}
}

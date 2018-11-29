/**
 * Auth :   liubo
 * Date :   2018/11/29 16:17
 * Comment: 压力测试用的
 */

package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

func init() {
	var moniter func()
	moniter = func() {
		fmt.Println("当前连接数：", len(stressEchos))
		time.AfterFunc(3 * time.Second, func() {
			moniter()
		})
	}
	moniter()
}

func stress(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}


	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	one := OneEcho{conn: conn, msgBuffer: make(chan []byte, 128)}

	addEcho(&one)

	go one.read()
	go one.write()
}

var stressEchos []*OneEcho
var stressEchosMutex sync.RWMutex

func addEcho(one *OneEcho) {
	stressEchosMutex.Lock()
	stressEchos = append(stressEchos, one)
	stressEchosMutex.Unlock()
}
func removeEcho(one *OneEcho) {
	stressEchosMutex.Lock()
	for i, v := range stressEchos {
		if v == one {
			stressEchos = append(stressEchos[0:i], stressEchos[i+1:]...)
		}
	}
	stressEchosMutex.Unlock()
}


type OneEcho struct {
	conn *websocket.Conn
	msgBuffer chan []byte
}

func (self *OneEcho)read() {
	conn := self.conn

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("ws error:", err)
			}
			break
		}

		if t == websocket.CloseMessage {
			break
		}

		// 只转发有效数据。ping包之类的不要转发
		if t == websocket.TextMessage {
			self.msgBuffer <- msg
		}
	}

	removeEcho(self)
}
func (self *OneEcho)write() {

	ping := time.NewTicker(pingPeriod)
	defer func() {
		self.conn.Close()
		ping.Stop()
		removeEcho(self)
	}()


	for {
		select {
		case msg, ok := <-self.msgBuffer:
			self.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				self.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := self.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Println(err)
				return
			}
			w.Write(msg)

			err = w.Close()
			if err != nil {
				fmt.Println(err)
				return
			}

		case <-ping.C:
			self.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := self.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
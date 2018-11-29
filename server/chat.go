/**
 * Auth :   liubo
 * Date :   2018/11/29 15:00
 * Comment: 
 */

package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func wsChat(conn *websocket.Conn) {
	one := OneChatter{conn: conn, msgBuffer: make(chan []byte, 128)}

	allChattersMutex.Lock()
	allChatters = append(allChatters, one)
	allChattersMutex.Unlock()

	go one.read()
	go one.write()
}


type OneChatter struct {
	conn *websocket.Conn
	msgBuffer chan []byte
}

// 为了简单，用mutex锁一下，就不用channel了。
var allChatters []OneChatter
var allChattersMutex sync.RWMutex

func (self *OneChatter)read() {
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

		// 只转发有效数据。ping包之类的不要转发
		if t == websocket.TextMessage {
			broadcast(msg)
		}
	}
}
func (self *OneChatter)write() {

	ping := time.NewTicker(pingPeriod)
	defer func() {
		self.conn.Close()
		ping.Stop()
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

func broadcast(msg []byte) {
	fmt.Println("转发数据", string(msg))

	allChattersMutex.RLock()
	for _, one := range allChatters {
		one.msgBuffer <- msg
	}
	allChattersMutex.RUnlock()
}
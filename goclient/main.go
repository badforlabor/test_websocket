/**
 * Auth :   liubo
 * Date :   2018/11/29 15:58
 * Comment: 
 */

package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"time"
)

func main() {

	thread := 10000
	for i:=0; i<thread; i++ {
		echo(i)
	}

	fmt.Println("创建成功")

	c := make(chan int)
	<-c
	fmt.Println("结束")
}

var allConns []*websocket.Conn

func echo(id int) {

	u := url.URL{Scheme:"ws", Host:"127.0.0.1:8088", Path:"/stress"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("无法打开连接：", u.String(), err)
		return
	}
	allConns = append(allConns, c)

	ticker := time.NewTicker(time.Second)

	go func() {
		for{
			t, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if t == websocket.CloseMessage {
				break
			}
			fmt.Println("echo from server:", id, string(msg))
		}
		c.Close()
		ticker.Stop()
	}()

	go func() {
		for {
			select {
			case t := <- ticker.C:
				err = c.WriteMessage(websocket.TextMessage, []byte(t.String()))
				if err != nil {
					return
				}
			}
		}
	}()
}

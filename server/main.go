/**
 * Auth :   liubo
 * Date :   2018/11/29 14:36
 * Comment: 
 */

package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	http.HandleFunc("/ws", wsPage)
	http.HandleFunc("/stress", stress)
	http.ListenAndServe(":8088", nil)
}

func wsPage(w http.ResponseWriter, r *http.Request) {

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

	wsChat(conn)
}

type errMsg struct {
	Msg string `json:"msg,omitempty"`
}
func ResponseJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

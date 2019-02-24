package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

var connections map[*websocket.Conn]bool = make(map[*websocket.Conn]bool)

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	connections[c] = true
	defer c.Close()
	defer func() {
		delete(connections, c)
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		for conn, _ := range connections {
			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	http.HandleFunc("/", handler)
	fmt.Printf("websocket server run on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

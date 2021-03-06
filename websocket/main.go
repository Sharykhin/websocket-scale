package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-nsq"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	config := nsq.NewConfig()
	p, _ := nsq.NewProducer("nsqd:4150", config)
	defer func() {
		fmt.Println("closing websocket connection")
		err := c.Close()
		if err != nil {
			log.Printf("could not close websocket connection safery: %v", err)
		}
		p.Stop()
		delete(connections, c)
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err, mt)
			break
		}
		log.Printf("recive message from socket: %s", message)
		// TODO: I am concerned that we able to send messages right here but need to send them through the queue.
		//for conn, _ := range connections {
		//	err = conn.WriteMessage(mt, message)
		//	if err != nil {
		//		log.Println("write:", err)
		//		break
		//	}
		//}

		fmt.Println("Publish message to nsq")
		err = p.Publish("write_test", message)
		if err != nil {
			log.Panic("Could not connect")
		}
	}
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	ch := flag.String("ch", "ch", "channel name for nsq consumer")
	flag.Parse()

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT)

	config := nsq.NewConfig()
	q, err := nsq.NewConsumer("write_test", *ch, config)
	if err != nil {
		log.Panicf("could not create a nsq consumer: %v", err)
	}
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("Got a message from nsq: %v, %v", message, string(message.Body))
		var err error
		for conn, _ := range connections {
			err = conn.WriteMessage(websocket.TextMessage, message.Body)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		return err
	}))

	go func() {
		err := q.ConnectToNSQD("nsqd:4150")
		if err != nil {
			log.Panicf("Could not connect to nsq: %v", err)
		}
	}()

	go func() {
		for {
			select {
			case <-q.StopChan:
				log.Fatalln("Consumer has been disconnected")
			case <-shutdown:
				q.Stop()
				log.Fatalln("Graceful shutdown")
			}
		}
	}()

	http.HandleFunc("/", handler)
	fmt.Printf("websocket server run on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

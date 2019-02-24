package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-nsq"
	"github.com/gorilla/websocket"
	"github.com/garyburd/redigo/redis"
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
var pool *redis.Pool

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	connections[c] = true
	config := nsq.NewConfig()
	p, _ := nsq.NewProducer("127.0.0.1:4150", config)
	defer c.Close()
	defer func() {
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
		//for conn, _ := range connections {
		//	err = conn.WriteMessage(mt, message)
		//	if err != nil {
		//		log.Println("write:", err)
		//		break
		//	}
		//}

		conn := pool.Get()
		fmt.Println("Publish message to redis")
		_, err = conn.Do("PUBLISH", "chat", message)
		if err != nil {
			log.Panic("Could not publish message")
		}

		//fmt.Println("Publish message to nsq")
		//err = p.Publish("write_test", message)
		//if err != nil {
		//	log.Panic("Could not connect")
		//}
	}
}

func main() {
	addr := flag.String("addr", ":8080", "http service address")
	flag.Parse()
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT)

	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer("write_test", "ch", config)
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
		err := q.ConnectToNSQD("localhost:4150")
		if err != nil {
			log.Panic("Could not connect")
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

	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		log.Panicf("can't connect to the redis database, got error:\n%v", err)
	}

	go func() {
		rc := pool.Get()
		psc := redis.PubSubConn{Conn: rc}
		fmt.Println("Subscribing on redis chat channel")
		if err := psc.PSubscribe("chat"); err != nil {
			log.Panicf("could noy subscribe: %v",err)
		}

		for {
			switch v := psc.Receive().(type) {
			case redis.PMessage:
				fmt.Println("got message from redis")
				for conn, _ := range connections {
					err = conn.WriteMessage(websocket.TextMessage, v.Data)
					if err != nil {
						log.Println("write:", err)
						break
					}
				}
			}
		}
	}()

	http.HandleFunc("/", handler)
	fmt.Printf("websocket server run on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bitly/go-nsq"
)

func main() {

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT)
	ch := flag.String("ch", "ch", "channel name")
	flag.Parse()
	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer("write_test", *ch, config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("Got a message: %v, %v", message, string(message.Body))
		return nil
	}))

	err := q.ConnectToNSQD("localhost:4150")
	if err != nil {
		log.Panic("Could not connect")
	}

	for {
		select {
		case <-q.StopChan:
			log.Println("Consumer has been disconnected")
			return
		case <-shutdown:
			log.Println("Graceful shutdown")
			q.Stop()
		}
	}
}
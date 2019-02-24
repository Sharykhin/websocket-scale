package main

import (
	"flag"
	"log"
	"github.com/bitly/go-nsq"
)

func main() {
	msg := flag.String("msg", "test", "message to be sent")
	flag.Parse()
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer("127.0.0.1:4150", config)

	err := w.Publish("write_test", []byte(*msg))
	if err != nil {
		log.Panic("Could not connect")
	}

	w.Stop()
}
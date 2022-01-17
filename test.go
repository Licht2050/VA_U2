package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go ping(ch)
	go pong(ch)
	select {}
}

func ping(ch chan string) {
	for {
		time.Sleep(time.Second)
		fmt.Println("Sending message...")
		ch <- "ping"
	}
}

func pong(ch chan string) {
	for {
		select {
		case p := <-ch:
			fmt.Printf("Received message %s\n", p)
			if p == "ping" {
				fmt.Println("pong")
			}
		}
	}
}

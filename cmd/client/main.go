package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	server := "127.0.0.1:8080"
	addr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}

	log.Printf("connect success")

	go receiver(conn)

	sender(conn)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Printf("go-simple-im client get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func sender(conn *net.TCPConn) {
	for {
		var msg string
		fmt.Scanln(&msg)
		bytes := []byte(msg)
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			panic(err)
		}
	}
}

func receiver(conn *net.TCPConn) {
	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadSlice('\n')
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("server: %s", string(data))
	}

}

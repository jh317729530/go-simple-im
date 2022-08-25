package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go-simple-im/internal/model"
)

func main() {
	flag.Parse()

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 8080,
	})
	if err != nil {
		panic(err)
	}
	go acceptTcp(listener)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Printf("go-simple-im-server get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func acceptTcp(listener *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
	)

	for {
		if conn, err = listener.AcceptTCP(); err != nil {
			log.Fatalf("listener.Accept(\"%s\") error(%v)", listener.Addr().String(), err)
			return
		}
		go serveTCP(conn)
	}
}

func serveTCP(conn *net.TCPConn) {
	// 封装连接
	channel := &model.Channel{
		Conn:   conn,
		Signal: make(chan *model.Data, 10),
	}

	// 开启协程处理写事件
	go dispatchTCP(channel)

	// 这个协程处理读事件
	for {
		// 解码数据，封装成Data，进行业务逻辑，然后写入Signal返回resp
		bytes, err := readTCP(channel)
		if err != nil {
			// 断开连接
			break
		}
		log.Printf("server receive info:%s", string(bytes))

		data := &model.Data{
			Content: []byte("server receive!"),
		}

		channel.Signal <- data
	}
}

func readTCP(channel *model.Channel) ([]byte, error) {
	conn := channel.Conn

	reader := bufio.NewReader(conn)

	data, err := reader.ReadSlice('\n')
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		} else {
			return data, err
		}
	}
	return data, nil
}

func dispatchTCP(channel *model.Channel) {
	for {
		data := <-channel.Signal
		data.Content = append(data.Content, '\n')
		channel.Conn.Write(data.Content)
	}
}

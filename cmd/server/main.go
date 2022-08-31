package main

import (
	"bufio"
	"encoding/binary"
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

	addr := conn.RemoteAddr()

	// 封装连接
	channel := &model.Channel{
		Conn:   conn,
		Signal: make(chan *model.Packet, 10),
		Addr:   addr,
	}

	// 开启协程处理写事件
	go dispatchTCP(channel)

	// 这个协程处理读事件
	for {
		// 解码数据，封装成Data，进行业务逻辑，然后写入Signal返回resp
		packet, err := readTCP(channel)
		if err != nil {
			// 断开连接
			break
		}

		switch packet.Operation {
		case model.HeartBeat:
			log.Printf("heart beat from:%s", addr.String())
			// 写入心跳回复
			packet.Length = model.PacketLengthSize + model.VersionSize + model.OperationSize
			packet.Version = model.Version1
			packet.Operation = model.HeartBeatReply

			channel.Signal <- packet
		}

		//channel.Signal <- data
	}
}

func readTCP(channel *model.Channel) (*model.Packet, error) {
	conn := channel.Conn

	reader := bufio.NewReader(conn)

	packetBytes := make([]byte, 4)
	_, err := io.ReadFull(reader, packetBytes)
	if err != nil {
		return nil, err
	}

	packetLength := binary.BigEndian.Uint32(packetBytes)

	bytes := make([]byte, packetLength-4)
	_, err = io.ReadFull(reader, bytes)
	if err != nil {
		return nil, err
	}

	packet := new(model.Packet)
	packet.Length = packetLength
	packet.Version = binary.BigEndian.Uint32(bytes[model.VersionOffset:model.OperationOffset])
	packet.Operation = binary.BigEndian.Uint32(bytes[model.OperationOffset:model.BodyOffset])
	// TODO body解析

	return packet, nil
}

func dispatchTCP(channel *model.Channel) {
	writer := bufio.NewWriter(channel.Conn)
	for {
		packet := <-channel.Signal
		switch packet.Operation {
		case model.HeartBeatReply:
			bytes := make([]byte, packet.Length)
			binary.BigEndian.PutUint32(bytes, packet.Length)
			binary.BigEndian.PutUint32(bytes[model.VersionOffset+model.PacketLengthSize:model.OperationOffset+model.PacketLengthSize], packet.Version)
			binary.BigEndian.PutUint32(bytes[model.OperationOffset+model.PacketLengthSize:model.BodyOffset+model.PacketLengthSize], packet.Operation)
			_, err := writer.Write(bytes)
			if err != nil {
				break
			}
		}

		err := writer.Flush()
		if err != nil {
			break
		}
	}

	channel.Conn.Close()
}

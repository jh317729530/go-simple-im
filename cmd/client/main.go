package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-simple-im/internal/model"
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

	heartBeatTicker := time.Tick(time.Second * 2)
	go heartBeat(conn, heartBeatTicker)

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
	for {
		packet, err := read(conn)
		if err != nil {
			break
		}

		switch packet.Operation {
		case model.HeartBeatReply:
			log.Printf("receive heart beat from server")
		}
	}

	conn.Close()
}

func read(conn *net.TCPConn) (*model.Packet, error) {
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

func heartBeat(conn *net.TCPConn, ticker <-chan time.Time) {
	for {
		select {
		case <-ticker:
			packet := &model.Packet{
				Length:    12,
				Version:   model.Version1,
				Operation: model.HeartBeat,
				Body:      nil,
			}
			log.Println("send heart beat")
			err := send(conn, packet)
			if err != nil {
				panic(err)
			}
		}
	}
}

func send(conn *net.TCPConn, packet *model.Packet) error {
	bytes := make([]byte, packet.Length)
	binary.BigEndian.PutUint32(bytes, packet.Length)
	binary.BigEndian.PutUint32(bytes[model.VersionOffset+model.PacketLengthSize:model.OperationOffset+model.PacketLengthSize], packet.Version)
	binary.BigEndian.PutUint32(bytes[model.OperationOffset+model.PacketLengthSize:model.BodyOffset+model.PacketLengthSize], packet.Operation)
	_, err := conn.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

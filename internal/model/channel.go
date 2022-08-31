package model

import "net"

type Channel struct {
	Conn   *net.TCPConn
	Signal chan *Packet
	Addr   net.Addr
}

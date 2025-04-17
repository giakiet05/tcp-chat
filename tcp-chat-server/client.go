package main

import (
	"net"
	"sync"
)

type Client struct {
	conn      net.Conn
	username  string
	out       chan string
	closeOnce sync.Once
	room      *Room // thêm reference tới room hiện tại
}

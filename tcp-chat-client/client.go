package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Client struct {
	address  string
	username string
	conn     net.Conn
	out      chan string
	ui       *ChatUI
	done     chan struct{}
}

func NewClient(addr string, username string) *Client {
	return &Client{
		address:  addr,
		username: username,
		out:      make(chan string),
		done:     make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	c.conn = conn
	c.ui = NewChatUI()
	fmt.Fprintln(conn, c.username)
	return nil
}

func (c *Client) Run() {
	go c.readLoop()
	go c.writeLoop()

	c.ui.OnSend = func(text string) {
		c.out <- text
	}
	c.ui.OnQuit = func() {
		close(c.done)
		c.conn.Close()
		os.Exit(0)
	}
	c.ui.Run()
}

func (c *Client) readLoop() {
	reader := bufio.NewReader(c.conn)
	for {
		select {
		case <-c.done:
			return
		default:
			msg, err := reader.ReadString('\n')
			if err != nil {
				c.ui.AddMessage("[red]Disconnected from server[-]")
				c.ui.OnQuit()
				return
			}
			c.ui.AddMessage(msg)
		}
	}
}

func (c *Client) writeLoop() {
	for {
		select {
		case msg := <-c.out:
			_, err := fmt.Fprintln(c.conn, msg)
			if err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

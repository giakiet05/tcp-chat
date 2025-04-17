package main

import (
	"fmt"
	"net"
	"sync"
)

type Room struct {
	name    string
	clients map[net.Conn]*Client
	mutex   sync.Mutex
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[net.Conn]*Client),
		mutex:   sync.Mutex{},
	}
}

func (r *Room) AddClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.clients[client.conn] = client
	client.room = r
}

func (r *Room) RemoveClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.clients, client.conn)
	client.room = nil
}

// broadcast ghi tin nhắn từ from client vào channel out của tất cả client khác có trong phòng
func (r *Room) Broadcast(msg string, from *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, client := range r.clients {
		if client != from {
			select {
			case client.out <- fmt.Sprintf("%s: %s", from.username, msg):
			default:
				fmt.Println("Channel is full or closed: ", client.username)
			}
		}

	}
}

func (r *Room) BroadcastRoomMsg(msg string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, client := range r.clients {
		select {
		case client.out <- fmt.Sprintf("ROOM: %s", msg):
		default:
			fmt.Println("Channel is full or closed: ", client.username)
		}
	}
}

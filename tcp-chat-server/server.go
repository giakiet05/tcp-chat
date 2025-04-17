package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

const manual = `Use these commands to interact: 
/room -> see room list
/join {roomName} -> join a room
/user -> see user list
/quit -> quit current room
/create {roomName} -> create a new room`

type Server struct {
	rooms map[string]*Room // thay đổi từ clients sang room
	mutex sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		rooms: make(map[string]*Room),
		mutex: sync.Mutex{},
	}
	s.rooms["general"] = NewRoom("general")
	s.rooms["test"] = NewRoom("test")
	return s
}

func (s *Server) Start(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Server is deaf!!!")
		return
	}

	fmt.Println("Server is running at port 9000!")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error")
			return
		}

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {

	out := make(chan string, 10)

	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}

	username = strings.TrimSpace(username)

	client := &Client{
		conn:     conn,
		username: username,
		out:      out,
	}

	//gửi danh sách phòng cho client biết
	s.sendSystemMsg(fmt.Sprintf("Welcome, %s", username), client)
	s.sendSystemMsg(manual, client)
	s.sendSystemMsg(s.getRoomList(), client)

	go s.readLoop(client)
	go s.writeLoop(client)

}

// đọc tin nhắn của client gửi tới
func (s *Server) readLoop(client *Client) {
	defer func() {
		if client.room != nil {
			client.room.BroadcastRoomMsg(fmt.Sprintf("%s disconnected from server\n", client.username))
			client.room.RemoveClient(client)
		}
		client.closeOnce.Do(func() {
			close(client.out)
		})
		client.conn.Close()
		fmt.Printf("%s disconnected from server\n", client.username)
	}()

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		msg := scanner.Text()
		if strings.HasPrefix(msg, "/") {
			s.handleCommand(msg, client)
		} else {
			if client.room == nil {
				s.sendSystemMsg("You need to join a room first! Use /room to see available rooms.", client)
				continue
			}
			client.room.Broadcast(msg, client)
		}
	}
}

// handle quit chỉ xử lí cho client rời phòng chứ không rời app
func (s *Server) handleQuit(client *Client) {
	if client.room != nil {
		quitInfo := fmt.Sprintf("%s quit %s", client.username, client.room.name)
		fmt.Print(quitInfo)
		client.room.BroadcastRoomMsg(quitInfo)
		client.room.RemoveClient(client)
	} else {
		s.sendSystemMsg("You are not in any room to quit!", client)
	}
}
func (s *Server) handleJoin(client *Client, roomName string) {
	room, ok := s.rooms[roomName]
	if !ok {
		s.sendSystemMsg("Room not found!", client)
		return
	}

	if client.room != nil {
		quitInfo := fmt.Sprintf("%s quit %s", client.username, client.room.name)
		fmt.Print(quitInfo)
		client.room.BroadcastRoomMsg(quitInfo)
		client.room.RemoveClient(client)
	}
	room.AddClient(client)

	joinedInfo := fmt.Sprintf("%s joined %s\n", client.username, roomName)
	fmt.Print(joinedInfo)
	client.room.BroadcastRoomMsg(joinedInfo)
}

func (s *Server) handleCreateRoom(client *Client, roomName string) {
	// Kiểm tra room tồn tại
	s.mutex.Lock()
	if _, exists := s.rooms[roomName]; exists {
		s.mutex.Unlock()
		s.sendSystemMsg("Room already exists!", client)
		return
	}

	// Tạo room mới và thêm vào map
	room := &Room{
		name:    roomName,
		clients: make(map[net.Conn]*Client),
		mutex:   sync.Mutex{},
	}
	s.rooms[room.name] = room
	s.mutex.Unlock()

	// Thông báo room mới được tạo
	s.broadcastSystemMsg(fmt.Sprintf("%s created a new room: %s", client.username, roomName))

	// Join vào room mới tạo
	s.handleJoin(client, roomName)
}

func (s *Server) handleCommand(msg string, client *Client) {
	args := strings.Split(msg, " ")
	cmd := args[0]

	switch cmd {
	case "/join":
		if len(args) < 2 {
			s.sendSystemMsg("Enter room name!", client)
			return
		}
		s.handleJoin(client, args[1])
	case "/quit":
		s.handleQuit(client)
	case "/room":
		s.sendSystemMsg(s.getRoomList(), client)
	case "/create":
		if len(args) < 2 {
			s.sendSystemMsg("Enter room name!", client)
			return
		}
		s.handleCreateRoom(client, args[1])
	case "/user":
		if client.room != nil {
			s.sendSystemMsg(s.getUserList(client.room), client)
		} else {
			s.sendSystemMsg("You are not in any room!", client)
		}
	}

}

// Gửi tin cho client
func (s *Server) writeLoop(client *Client) {
	for msg := range client.out {
		_, err := fmt.Fprintln(client.conn, msg)
		if err != nil {
			return // Just return, let readLoop handle cleanup
		}
	}
}

// broadcast tin nhắn hệ thống tới toàn bộ room
func (s *Server) broadcastSystemMsg(msg string) {
	// Lấy danh sách room an toàn
	s.mutex.Lock()
	rooms := make([]*Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	s.mutex.Unlock()

	// Broadcast cho từng room
	for _, room := range rooms {
		room.BroadcastRoomMsg(msg)
	}
}

// nhắn cho một client duy nhất
func (s *Server) sendSystemMsg(msg string, client *Client) {
	client.out <- fmt.Sprintf("SYSTEM: %s", msg)
}

func (s *Server) getRoomList() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rooms := make([]string, 0, len(s.rooms))
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	return fmt.Sprintf("Available rooms:\n %s", strings.Join(rooms, "\n"))
}

func (s *Server) getUserList(room *Room) string {
	clientNames := make([]string, 0, len(room.clients))
	for _, client := range room.clients {
		clientNames = append(clientNames, client.username)
	}
	return fmt.Sprintf("Users in room %s:\n %s", room.name, strings.Join(clientNames, "\n"))
}

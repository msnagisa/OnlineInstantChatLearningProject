package main

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port string
	// 在线用户列表
	OnlineUserMap map[string]*User
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port string) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineUserMap: make(map[string]*User),
		mapLock: sync.RWMutex{},
		Message: make(chan string),
	}
	return server
}

//广播消息
func (server *Server) broadcast(user *User, message string) {
	sendMessage := "[" + user.Addr + "]" + user.Name + ":" + message
	server.Message <- sendMessage
}

// 监听Message广播消息的channel
func (server *Server) ListenMessage() {
	for {
		message := <-server.Message
		server.mapLock.Lock()
		for _, user := range server.OnlineUserMap {
			user.C <- message
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) Handler(conn net.Conn) {
	fmt.Println("连接建立成功")
	// 添加用户到在线用户列表中
	user := NewUser(conn, server)

	server.mapLock.Lock()
	server.OnlineUserMap[user.Name] = user
	server.mapLock.Unlock()

	// 广播用户上线消息
	user.Online()

	// 监听用户是否活跃
	isLive := make(chan bool)

	// 持续接收客户端消息
	go func () {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				user.Offline()
				runtime.Goexit()
			} else {
				message := string(buf[:n - 1])
				user.HandleMessage(message)
			}
		}
	}()

	// 阻塞当前handler
	for {
		select {
			case <-isLive:
				// 用户活跃
			case <-time.After(time.Second * 300): // 30秒内无消息，则断开连接
				user.SendMessage("timeout exit")
				// user.Offline()
				runtime.Goexit()
		}
	}
}

func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	go server.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}

		// do handler
		go server.Handler(conn)
	}


}
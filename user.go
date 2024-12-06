package main

import (
	"fmt"
	"net"
	"runtime"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
	server *Server
	isLive bool
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
		server: server,
		isLive: true,
	}
	go user.ListenMessage()

	return user
}

// 监听当前user channel的方法
func (user *User) ListenMessage() {
	for {
		if !user.isLive {
			runtime.Goexit()
		}
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
		fmt.Println(msg)
	}
}

func (user *User) Online() {
	user.server.broadcast(user, "login")
}

func (user *User) Offline() {
	user.server.broadcast(user, "logout")
	user.server.mapLock.Lock()
	delete(user.server.OnlineUserMap, user.Name)
	user.server.mapLock.Unlock()
	user.isLive = false
	// 关闭chan
	close(user.C)
	// 关闭connection
	user.conn.Close()
}

func (user *User) SendMessage(message string) {
	user.conn.Write([]byte(message))
}

func (user *User) HandleMessage(message string) {
	if message == "who" { // 查询当前在线用户
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineUserMap {
			onlineMessage := "[" + u.Addr + "]" + u.Name + ":" + " is online\n"
			user.SendMessage(onlineMessage)
		}
		user.server.mapLock.Unlock()
	} else if len(message) > 7 && message[:7] == "rename|" {
		user.server.mapLock.Lock()
		_, ok := user.server.OnlineUserMap[message[7:]];
		if ok {
			user.SendMessage("user name already exist\n")
		} else {
			delete(user.server.OnlineUserMap, user.Name)
			user.Name = message[7:]
			user.server.OnlineUserMap[user.Name] = user
			user.SendMessage("rename success\n")
		}
		user.server.mapLock.Unlock()
	} else if len(message) > 3 && message[:3] == "to|" {
		// 消息格式: to|userName|message
		remoteMsg := strings.Split(message, "|")
		if (len(remoteMsg) != 3) {
			user.SendMessage("message format error\n")
		}
		targetUserName := remoteMsg[1]
		targetMsg := remoteMsg[2]
		user.server.mapLock.Lock()
		targetUser, ok := user.server.OnlineUserMap[targetUserName]
		if !ok {
			user.SendMessage("user not exist\n")
		} else {
			targetUser.SendMessage(user.Name + ":" + targetMsg + "\n")
		}
		user.server.mapLock.Unlock()
	} else {
		user.server.broadcast(user, message)
	}
}
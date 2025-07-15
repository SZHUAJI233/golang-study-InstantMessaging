package main

import (
	"fmt"
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前user channel的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线
func (user *User) Online() {
	server := user.server
	// 用户上线，将用户加入到OnlineMap中
	server.mapLock.Lock()
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	// 广播当前用户上线消息
	server.BroadCast(user, "已上线")
}

// 用户下线
func (user *User) Offline() {
	server := user.server
	// 用户下线，将用户从OnlineMap中删除
	server.mapLock.Lock()
	delete(server.OnlineMap, user.Name)
	server.mapLock.Unlock()

	// 广播当前用户上线消息
	server.BroadCast(user, "用户下线")
}

// 给user对应的客户端发送消息
func (user *User) SendMessage(msg string) {
	user.conn.Write([]byte(msg))

}

// 用户处理消息
func (user *User) DoMessage(msg string) {
	server := user.server

	if msg == "who" {
		// 查询当前在线用户
		server.mapLock.Lock()
		for _, onlineUser := range server.OnlineMap {
			onlineUserMsg := fmt.Sprintf("[%s]%s:在线...\n", onlineUser.Addr, onlineUser.Name)
			user.SendMessage(onlineUserMsg)
		}
		server.mapLock.Unlock()
	} else {
		server.BroadCast(user, msg)
	}

}

// 监听当前User channel的方法，一旦有消息就直接发给对端客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}

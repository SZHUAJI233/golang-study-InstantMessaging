package main

import (
	"fmt"
	"net"
	"strings"
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

	if msg[0] == byte('/') {
		command := strings.Split(msg, " ")
		switch command[0] {
		// 私聊
		case "/tell":
			if len(command) < 3 {
				user.SendMessage("正确用法：/tell <name> <msg>\n")
			} else {
				// 分解指令参数
				toUserName := command[1]
				msg := strings.Join(command[2:], " ")
				// 获取目标用户对象
				toUser, ok := server.OnlineMap[toUserName]
				if !ok {
					user.SendMessage("用户不在线！\n")
				} else {
					toUser.SendMessage(user.Name + "对您说：" + msg + "\n")
				}
			}
		// 查询在线用户
		case "/who":
			// 查询当前在线用户
			server.mapLock.Lock()
			for _, onlineUser := range server.OnlineMap {
				onlineUserMsg := fmt.Sprintf("[%s]%s:在线...\n", onlineUser.Addr, onlineUser.Name)
				user.SendMessage(onlineUserMsg)
			}
			server.mapLock.Unlock()
		// 修改名称
		case "/rename":
			if len(command) != 2 {
				user.SendMessage("正确用法：/rename <name>\n")
			} else {
				newName := command[1]
				_, ok := server.OnlineMap[newName]
				if ok {
					user.SendMessage("用户名被使用！\n")
				} else {
					server.mapLock.Lock()
					delete(server.OnlineMap, user.Name)
					server.OnlineMap[newName] = user
					server.mapLock.Unlock()

					user.Name = newName
					user.SendMessage("用户名修改成功:" + user.Name + "\n")
				}
			}
		default:
			user.SendMessage("指令不存在！请检查！\n")
		}

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

package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message

		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s", user.Addr, user.Name, msg)
	server.Message <- sendMsg
}

// Handler
func (server *Server) Handler(conn net.Conn) {
	//...当前连接的业务
	// fmt.Println("连接建立成功！")

	user := NewUser(conn, server)

	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
				return
			}

			// 提取用户消息，去除\n
			msg := string(buf[:n-1])

			// 将得到的消息进行处理
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的状态
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户活跃，应该重置定时器
			// 当前case执行后当前select结束，开启了新的select，则计时器也重置
		case <-time.After(time.Minute * 10): // 计时10min
			// 10s后超时，发送信息，激活case

			// 强制关闭当前User
			user.SendMessage("你被踢了！\n")
			// 下线用户
			user.Offline()
			// 关闭用户的连接
			conn.Close()

			// 退出当前handler
			return
		}

	}

}

// 启动服务器的接口
func (server *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Nisten err:", err)
		return
	}

	// colose listen socket
	defer listener.Close()
	// 启动监听message的go
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

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接服务器server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn

	// 返回对象
	return client
}

// 处理服务端回应的消息，直接显示到标准输出中
func (client *Client) DealResponse() {
	// 一旦conn中有数据，就直接cooy到stdout标准输出中，且是永久阻塞监听的
	io.Copy(os.Stdout, client.conn)
}

// 菜单
func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	}
	fmt.Println(">>>>请输入合法范围内的数字<<<<")
	return false

}

// 更新用户名
func (client *Client) UpdateName() bool {
	var newName string
	fmt.Println(">>>>请输入用户名：")
	fmt.Scanln(&newName)

	sendMsg := "/rename " + newName + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client.conn.Write error: ", err)
		return false
	}
	client.Name = newName
	return true
}

// 公聊模式
func (client *Client) PublicChat() {
	var chatMsg string
	// 提示用户输入消息
	fmt.Println(">>>>请输入聊天内容(exit 退出)：")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 消息不为空
		if len(chatMsg) != 0 {
			// 发送给服务器
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("client.conn.Write error: ", err)
				break
			}
		}

		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}

}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "/who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client.conn.Write error: ", err)
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var privateUser string
	var chatMsg string

	fmt.Println("在线用户：")
	client.SelectUsers()

	fmt.Println(">>>>请输入私聊用户（exit 退出）：")
	fmt.Scanln(&privateUser)
	if privateUser == "exit" {
		return
	}

	fmt.Println(">>>>请输入聊天内容（exit 退出）：")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := "/tell " + privateUser + " " + chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("client.conn.Write error ", err)
				return
			}
		}

		fmt.Println(">>>>请输入聊天内容（exit 退出）：")
		fmt.Scanln(&chatMsg)

	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			// 公聊模式
			client.PublicChat()
			break
		case 2:
			// 私聊模式
			client.PrivateChat()
			break
		case 3:
			// 更新用户名
			client.UpdateName()
			break
		case 0:
			fmt.Println(">>>>已安全退出<<<<")
			return
		}
	}
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址（默认为127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认为8888）")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>服务器连接失败<<<<")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>服务器连接成功<<<<")

	// 启动客户端业务
	client.Run()
}

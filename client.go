package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Client struct {
	ServerIp   string
	ServerPort string
	Name       string
	Conn      net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort string) *Client {
	client := &Client{
		ServerIp: serverIp, 
		ServerPort: serverPort,
		flag: 999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
	}
	client.Conn = conn
	return client
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式") 
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&flag)
	fmt.Println("current input is: " + strconv.Itoa(flag))
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>> 输入有误，请重新输入")
		return false
	}
}

func (client *Client) UpdateUserName() {
	fmt.Println(">>>> 输入用户名")
	fmt.Scanln(&client.Name)
	client.Conn.Write([]byte("rename|" + client.Name + "\n"))
}

func (client *Client) PublicChat() {
	fmt.Println(">>>> 输入公共聊天内容, exit退出")
	var chatMsg string
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			_, err := client.Conn.Write([]byte(chatMsg + "\n"))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}
		fmt.Println(">>>> 输入公共聊天内容, exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) ShowOnlineUsers() {
	searchStr := "who\n"
	_, err := client.Conn.Write([]byte(searchStr))
	if err != nil {
		fmt.Println("conn.Write err when request online users")
	}
	
}

func (client *Client) PrivateChat() {
	client.ShowOnlineUsers()
	var remoteName string
	fmt.Println(">>>> 请输入私聊对象的用户名, exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		var chatMsg string
		fmt.Println(">>>> 请输入私聊内容, exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				chatMsg = "to|" + remoteName + "|" + chatMsg
				_, err := client.Conn.Write([]byte(chatMsg + "\n"))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}
			fmt.Println(">>>> 请输入私聊内容, exit退出")
			fmt.Scanln(&chatMsg)
		}
		fmt.Println(">>>> 请输入私聊对象的用户名, exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) HandleResponse() {
	// io.Copy(os.Stdout, client.Conn)
	for {
		buf := make([]byte, 4096)
		n, err := client.Conn.Read(buf)
		if err != nil {
			fmt.Println("conn.Read err:", err)
		} else {
			fmt.Println(string(buf[:n]))
		}
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
			continue
		}
		switch client.flag {
			case 1:
				client.PublicChat()
				break
			case 2:
				client.PrivateChat()
				break
			case 3:
				client.UpdateUserName()
				break
			default:
				break;
		}
	}
}

var serverIp string
var serverPort int
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, strconv.Itoa(serverPort))
	if client == nil {
		fmt.Println(">>>> 连接服务器失败")
		return
	}
	fmt.Println(">>>> 连接服务器成功")
	fmt.Println("ip is: " + serverIp + " port is: " + strconv.Itoa(serverPort) + "\n")
	go client.HandleResponse()
	client.Run()
}

func (client *Client) Handler() {
	for {
		time.Sleep(2 * time.Second)
		fmt.Println(">>>> 等待输入")
	}
}


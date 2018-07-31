package main

import (
	"robot"

	//"github.com/golang/snappy"

	//"github.com/golang/protobuf/proto"
)

const (
	addr = "127.0.0.1:7900"
)

func main() {
	robot.CreateRobot()

	for {

	}
}

/*
func main() {
	fmt.Println("Hello World!")

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("连接服务端失败:", err.Error())
		return
	}
	fmt.Println("已连接服务器")
	defer conn.Close()
	Client(conn)
}

func Client(conn net.Conn) {
	sms := make([]byte, 128)
	for {
		fmt.Print("请输入要发送的消息:")
		_, err := fmt.Scan(&sms)
		if err != nil {
			fmt.Println("数据输入异常:", err.Error())
		}
		conn.Write(sms)
		buf := make([]byte, 128)
		c, err := conn.Read(buf)
		if err != nil {
			fmt.Println("读取服务器数据异常:", err.Error())
		}
		fmt.Println(string(buf[0:c]))
	}
}
*/

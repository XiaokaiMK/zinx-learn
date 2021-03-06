package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

//链接模块

type Connection struct {
	//当前链接的socket TCP套接字
	Conn *net.TCPConn

	//链接的ID
	ConnID uint32

	//当前链接的状态
	isClosed bool

	//告知当前链接已经退出的/停止 channel
	ExitChan chan bool

	//无缓冲的管道，用于读写Goroutine之间的通信
	msgChan chan []byte

	//消息的管理MsgID和对应的处理业务的API关系
	MsgHandler ziface.IMsgHandle

}

//初始化链接模块的方法
func NewConnection(conn *net.TCPConn,connID uint32,msgHandler ziface.IMsgHandle ) *Connection {
	c:= &Connection{
		Conn: conn,
		ConnID: connID,
		MsgHandler: msgHandler,
		isClosed:false,
		ExitChan:make(chan bool,1),
		msgChan: make(chan []byte),
	}
	return c
}

//链接的读业务方法
func (c *Connection) StartReader(){
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(),"conn reader exit!")
	defer c.Stop()

	for{
 		//创建一个拆包解包对象
 		dp := NewDataPack()

 		//读取客户端的Msg Head 二进制流 8个字节
		headData := make([]byte, dp.GetHeadLen())
		if _,err := io.ReadFull(c.GetTcpConnection(),headData);err != nil{
			fmt.Println("read msg head error",err)
			break
		}
 		//拆包，得到msgID和msgDatalen放在msg消息中
		msg,err := dp.Unpack(headData)
		if err != nil{
			fmt.Println("unpack error",err)
			break
		}
 		//根据datalen 再次读取Data 放在Msg.Data中
		var data []byte
		if msg.GetDataLen() > 0{
			data = make([] byte, msg.GetDataLen())
			if _,err := io.ReadFull(c.GetTcpConnection(),data);err != nil{
				fmt.Println("read msg data error",err)
				break
			}
		}
		msg.SetData(data)

		//得到当前conn数据的Request请求数据
		req := Request{
			conn:c,
			msg:msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

//写消息的Goroutine，专门发送给客户端消息的模块
func (c *Connection) StartWriter(){
	fmt.Println("[writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(),"conn writer exit!")

	//不断的阻塞等待channel的消息，进行写给客户端
	for  {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _,err := c.Conn.Write(data);err != nil{
				fmt.Println("Send data error",err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，Writer也要退出
			return
		}
	}
}


//启动链接 让当前的链接准备开始工作
func (c *Connection) Start(){
	fmt.Println("Conn Start()... ConnID=",c.ConnID)
	//启动从当前链接的读数据的业务
	go c.StartReader()
	//启动从当前链接写数据的业务
	go c.StartWriter()

	for {
		select {
		case <- c.ExitChan:
			//得到退出消息，不再阻塞
			return
		}
	}
}

//停止链接 结束当前的工作
func (c *Connection) Stop(){
	fmt.Println("Conn Stop().. ConnID=",c.ConnID)

	if c.isClosed ==true{
		return
	}
	c.isClosed = true

	//关闭socket链接
	c.Conn.Close()

	//告知Writer关闭
	c.ExitChan <- true

	//回收资源
	close(c.ExitChan)
	close(c.msgChan)
}

//获取当前链接的绑定socket conn
func (c *Connection) GetTcpConnection() *net.TCPConn{
	return c.Conn
}

//获取当前链接模块的链接ID
func (c *Connection) GetConnId() uint32{
	return c.ConnID
}

//获取远程客户端的TCP状态IP port
func (c *Connection) RemoteAddr() net.Addr{
	return c.Conn.RemoteAddr()
}

//提供一个SendMsg方法 将我们要发送给客户端的数据先进行封包，再发送
func (c *Connection) SendMsg (msgId uint32,data []byte)error{
	if c.isClosed == true{
		return errors.New("Connection closed when send msg")
	}

	//将data进行封包 MsgDataLen/MsgID/Data
	dp := NewDataPack()

	binaryMsg, err := dp.Pack(NewMsgpackage(msgId,data))
	if err != nil{
		fmt.Println("Pack error msg id =", msgId)
		return errors.New("Pack error msg")
	}
	//将数据发送给客户端
	c.msgChan <- binaryMsg

	return nil
}
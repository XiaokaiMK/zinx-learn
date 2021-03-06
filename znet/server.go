package znet

import (

	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Server struct {
	//服务器的名字
	Name string
	//服务器绑定的ip版本
	IPversion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandler ziface.IMsgHandle
}



//启动服务器
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPackageSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPackageSize)

	go func() {
		//0 开启消息队列及worker工作池
		s.MsgHandler.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPversion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Printf("resolve tcp addr error:", err)
			return
		}
		//2 监听服务器地址
		listenner, err := net.ListenTCP(s.IPversion, addr)
		if err != nil {
			fmt.Println("listen", s.IPversion, "err", err)
			return
		}
		fmt.Println("start Zinx server succ,", s.Name, "succ,now Listenning...")
		var cid uint32
		cid = 0

		//3 阻塞的等待客户端链接，处理客户端链接业务（读写）
		for {
			//如果有客户端链接进来，阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			dealConn := NewConnection(conn, cid, s.MsgHandler)
			cid++

			//启动当前的链接业务模块
			go dealConn.Start()
		}
	}()
}

//停止服务器
func (s *Server) Stop() {
	//TODO 将一些服务器的资源、状态或者一些已经开辟的链接信息进行停止或者回收

}

//运行服务器
func (s *Server) Server() {
	//启动server的服务功能
	s.Start()

	//TODO 做一些启动服务器之后的额外业 务

	//阻塞
	select {}
}

//添加路由功能
func (s *Server) AddRouter(msgID uint32,router ziface.IRouter){
	s.MsgHandler.AddRouter(msgID,router)
	fmt.Println("Add Router succ!!")
}

/*
创建一个服务器句柄
*/
func NewServer(name string) ziface.IServer {
	s := &Server{
		utils.GlobalObject.Name,
		"tcp4",
		utils.GlobalObject.Host,
		utils.GlobalObject.TcpPort,
		NewMsgHandle(),
	}
	return s
}

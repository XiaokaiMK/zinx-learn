package utils

import (
	"encoding/json"
	"io/ioutil"
	"zinx/ziface"
)

//存储一切有关Zinx框架的全局参数，供其他模块使用
//一些参数是可以铜鼓片zinx json由用户进行配置

type GlobalObj struct {

	//server
	TcpServer ziface.IServer //当前Zinx全局的server对象
	Host string  			 //当前服务器主机监听的IP
	TcpPort int   			 //当前服务器主机监听的端口号
	Name string 			 //当前服务器的名称

	//zinx
	Version          string //当前Zinx版本号
	MaxPackageSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 //业务工作Worker池的数量
	MaxWorkerTaskLen uint32 //业务工作Worker对应负责的任务队列最大任务存储数量

	//config file path

	ConfFilePath string
}

//定义一个全局的对外Globalobj对象
var GlobalObject *GlobalObj


//从zinx.json去加载用于自定义的参数
func (g *GlobalObj) Reload(){
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil{
		panic(err)
	}

	//将json文件数据解析到struct中
	err = json.Unmarshal(data,&GlobalObject)
	if err != nil {
		panic(err)
	}
}



//提供一个init方法，初始化当前的GlobalObject
func init()  {
	GlobalObject = &GlobalObj{
		Name: "ZinxServerApp",
		Version:"V0.4",
		TcpPort: 8999,
		Host: "0.0.0.0",
		MaxConn: 1000,
		MaxPackageSize: 4096,
		WorkerPoolSize: 10,
		MaxWorkerTaskLen: 1024,
	}

	//应该尝试从conf/zinx.json去加载一些用户自定义的参数
	GlobalObject.Reload()
}
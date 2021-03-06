package ziface

//irequest 接口实际上是把客户端请求的链接信息和请求的数据包装到一个Request中

type IRequest interface {
	//得到当前的链接
	GetConnection() IConnection

	//得到的请求的消息数据
	GetData() []byte

	//得到请求消息的ID
	GetMsgId() uint32
}
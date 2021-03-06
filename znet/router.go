package znet

import "zinx/ziface"

type BaseRouter struct {}

//在处理conn业务之前的钩子方法Hook
func (br *BaseRouter) PreHandle(request ziface.IRequest){}
//在处理conn业务之前的主方法Hook
func (br *BaseRouter) Handle(request ziface.IRequest){}
//在处理conn业务之前的钩子方法Hook
func (br *BaseRouter) PostHandle(request ziface.IRequest){}

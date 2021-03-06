package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandle struct {
	//存放每个MsgId所对应的处理方法
	Apis map[uint32]ziface.IRouter
	//业务工作Worker池的数量
	WorkerPoolSize uint32
	//Worker负责取任务的消息队列
	TaskQueue []chan ziface.IRequest
}

//初始化MsgHandle方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis: make(map[uint32]ziface.IRouter),
		WorkerPoolSize:utils.GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue:make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

//调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	//从Request中找到msgID
	handler,ok := mh.Apis[request.GetMsgId()]
	if !ok{
		fmt.Println("api msgID=",request.GetMsgId(),"is NOT FOUND Need Register!")
	}
	//根据MsgID调度对应的router业务即可
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

//为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//判断当前msg绑定的处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//ID已经注册
		panic("repeat api ,msgID =" + strconv.Itoa(int(msgID)))
	}
	//判断msg与API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID=", msgID, "succ!")
}

//启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

//启动worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	//遍历需要启动worker的数量，依此启动 根据workPoolSize分别开启worker
	for i:= 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

//将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle)SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnId() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnId()," request msgID=", request.GetMsgId(), "to workerID=", workerID)
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}
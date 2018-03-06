package main

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/zx9229/zxgo"
)

type CbDisConnected func(error)
type CbReceivedData func([]byte)

type Connection struct {
	cbRecv  CbReceivedData //收到消息时的回调
	cbDis   CbDisConnected //断开时的回调函数
	disFlag int32          //断开的标志.
	conn    net.Conn       //用于connect/accept的socket.
	queue   *zxgo.Queue    //缓冲队列
}

func NewConnection(c net.Conn, onRecv CbReceivedData, onDis CbDisConnected) *Connection {
	connection := &Connection{cbRecv: onRecv, cbDis: onDis, disFlag: 0, conn: c}
	connection.queue = zxgo.NewQueue(connection.queueCallbackFun)
	go connection.recvData()
	return connection
}

func (self *Connection) handleDisConnected(err error) {
	if atomic.AddInt32(&self.disFlag, 1) == 1 {
		self.conn.Close()
		self.queue.ExitGoroutine()
		self.cbDis(err)
	}
}

func (self *Connection) recvData() {
	buf := make([]byte, 2048)
	var err error = nil
	var lenght int = 0

	for err == nil {
		lenght, err = self.conn.Read(buf)
		if err != nil {
			break
		}
		if lenght > 0 {
			buf[lenght] = 0
		}
		self.cbRecv(buf[:lenght])
	}

	self.handleDisConnected(err)
}

func (self *Connection) queueCallbackFun(iData interface{}) {
	data := iData.([]byte)

	var totalLen int = len(data)
	var sendLen int = 0
	for i := 0; i < totalLen; i++ { //本来想写一个死循环呢,后来想了一下,如果一次发送一字节都发不出去的话,那还不如断了呢.
		if n, err := self.conn.Write(data); err != nil {
			self.handleDisConnected(err)
		} else {
			sendLen += n
			if sendLen < totalLen {
				data = data[n:]
			} else if sendLen > totalLen {
				panic(fmt.Sprintf("逻辑错误:totalLen=%v,sendLen=%v", totalLen, sendLen))
			}
		}
	}
	if sendLen != totalLen {
		err := errors.New(fmt.Sprintf("totalLen=%v,sendLen=%v,没有发送完毕", totalLen, sendLen))
		self.handleDisConnected(err)
	}
}

// 缓存了多少条消息.
func (self *Connection) CacheCnt() int {
	return self.queue.Size()
}

// 返回值表示:是否成功受理,不表示发送成功.
func (self *Connection) Send(data []byte) bool {
	if 0 < self.disFlag {
		return false
	} else {
		self.queue.Push(data)
		return true
	}
}

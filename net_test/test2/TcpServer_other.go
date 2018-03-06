package net_bak1

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type disconnectedChanType struct {
	tcpConn *net.TCPConn
	err     error
}

type TcpClient struct {
	address        string
	alive          bool
	conn           net.Conn
	tcpConn        *net.TCPConn
	tcpConnChan    chan disconnectedChanType
	OnConnected    func(*TcpClient)
	OnDisconnected func(*TcpClient, error)
	OnReceivedData func(*TcpClient, []byte)
}

func NewTcpClient(connection *net.TCPConn, chanObj chan disconnectedChanType) *TcpClient {
	if connection == nil && chanObj == nil {
		return &TcpClient{}
	}
	//服务器accept的socket,使用它进行构造
	if connection != nil && chanObj != nil {
		return &TcpClient{"", true, nil, connection, chanObj, nil, nil, nil}
	}
	return nil
}

func (this *TcpClient) Start(address string) error {
	var err error
	if this.conn != nil || this.tcpConn != nil {
		err = errors.New(fmt.Sprintf("连接已存在,conn=%p,tcpConn=%p", this.conn, this.tcpConn))
		fmt.Println(err)
		return err
	}

	this.conn, err = net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return err
	}

	this.doConnect()
	return nil
}

func (this *TcpClient) doConnect() {
	go func() {
		var err error
		for this.alive {
			if this.conn, err = net.Dial("tcp", this.address); err != nil {
				time.Sleep(time.Second * 5)
			} else {
				this.doReadConn()
			}
		}
	}()
}

func (this *TcpClient) doReadConn() {
	this.OnConnected(this)

	go func() {
		buf := make([]byte, 2048)
		for {
			if num, err := this.conn.Read(buf); err != nil {
				this.OnDisconnected(this, err)
				this.doConnect()
				break
			} else {
				if num == 0 {
					panic("逻辑错误")
				}
				data := buf[:num]
				this.OnReceivedData(this, data)
			}
		}
	}()
}

func (this *TcpClient) doReadTcpConn() {
	this.OnConnected(this)

	go func() {
		buf := make([]byte, 1536)
		for {
			if num, err := this.tcpConn.Read(buf); err != nil {
				if true {
					this.tcpConnChan <- disconnectedChanType{this.tcpConn, err}
					this.OnDisconnected(this, err)
				}
				break
			} else {
				if num == 0 {
					panic("逻辑错误")
				}
				data := buf[:num]
				this.OnReceivedData(this, data)
			}
		}
	}()
}

type TcpServer struct {
	svrAddr        *net.TCPAddr
	svrListener    *net.TCPListener
	tcpConnAddChan chan *net.TCPConn
	tcpConnDelChan chan disconnectedChanType
	exitChan       chan bool
	tcpConnMap     map[*net.TCPConn]*TcpClient
	OnConnected    func(*TcpClient)
	OnDisconnected func(*TcpClient, error)
	OnReceivedData func(*TcpClient, []byte)
}

func NewTcpServer() *TcpServer {
	chanAdd := make(chan *net.TCPConn, 65535)
	chanDel := make(chan disconnectedChanType, 65535)
	chanExt := make(chan bool, 1)
	connMap := map[*net.TCPConn]*TcpClient{}
	server := &TcpServer{nil, nil, chanAdd, chanDel, chanExt, connMap, nil, nil, nil}
	return server
}

func (this *TcpServer) Start(network, address string) error {
	var err error = nil
	this.svrAddr, err = net.ResolveTCPAddr(network, address)
	if err != nil {
		return err
	}
	this.svrListener, err = net.ListenTCP("tcp", this.svrAddr)
	if err != nil {
		return err
	}
	go this.doAccept()
	go this.doSth()
	return err
}

func (this *TcpServer) doAccept() {
	for {
		if tcpConn, err := this.svrListener.AcceptTCP(); err != nil {
			fmt.Println(err)
			break
		} else {
			this.tcpConnAddChan <- tcpConn
		}
	}
	fmt.Println("acceptTCPConn finish")
}

func (this *TcpServer) doSth() {
	for {
		select {
		case connection, ok := <-this.tcpConnAddChan:
			if ok {
				if _, ok := this.tcpConnMap[connection]; ok {
					fmt.Println("ERROR,", connection)
				} else {
					client := NewTcpClient(connection, this.tcpConnDelChan)
					client.OnConnected = this.OnConnected
					client.OnDisconnected = this.OnDisconnected
					client.OnReceivedData = this.OnReceivedData
					this.tcpConnMap[connection] = client
					client.doReadTcpConn()
				}
			}
		case disData, ok := <-this.tcpConnDelChan:
			if ok {
				if _, ok := this.tcpConnMap[disData.tcpConn]; ok {
					delete(this.tcpConnMap, disData.tcpConn)
				} else {
					fmt.Println("ERROR,", disData.tcpConn)
				}
			}
			if this.svrListener == nil && len(this.tcpConnMap) == 0 {
				break
			}
		case <-this.exitChan:
			for k, _ := range this.tcpConnMap {
				k.Close()
			}
			if this.svrListener == nil && len(this.tcpConnMap) == 0 {
				break
			}
		}
	}
}

func (this *TcpServer) Stop() {
	err := this.svrListener.Close()
	if err != nil {
		fmt.Println(err)
	}
	this.svrListener = nil
	this.exitChan <- true
}

////////////////////////////////////
func funOnConnected(client *TcpClient) {
	fmt.Println("funOnConnected")
}
func funOnDisconnected(client *TcpClient, err error) {
	fmt.Println("funOnDisconnected", err)
}

func funOnReceivedData(client *TcpClient, data []byte) {
	fmt.Println("funOnReceivedData,", data)
}
func main_1() {
	server := NewTcpServer()
	server.OnConnected = funOnConnected
	server.OnDisconnected = funOnDisconnected
	server.OnReceivedData = funOnReceivedData

	if err := server.Start("tcp", "0.0.0.0:12345"); err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

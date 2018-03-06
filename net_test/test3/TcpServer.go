package net_bak2

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
	address        string                    //用于connect的地址(例如127.0.0.1:12345).
	isAlive        bool                      //在封装层面上,是不是活着的.
	isConnected    bool                      //在封装层面上,是不是连接着的.
	conn           net.Conn                  //用于connect的socket.
	tcpConn        *net.TCPConn              //用于accept的socket.
	tcpConnChan    chan disconnectedChanType //用于accept的socket.
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
		return &TcpClient{"", false, false, nil, connection, chanObj, nil, nil, nil}
	}
	return nil
}

func (this *TcpClient) LocalAddr() net.Addr {
	if this.conn != nil {
		return this.conn.LocalAddr()
	} else if this.tcpConn != nil {
		return this.tcpConn.LocalAddr()
	} else {
		return nil
	}
}

func (this *TcpClient) RemoteAddr() net.Addr {
	if this.conn != nil {
		return this.conn.RemoteAddr()
	} else if this.tcpConn != nil {
		return this.tcpConn.RemoteAddr()
	} else {
		return nil
	}
}

func (this *TcpClient) Start(address string) error {
	var err error
	if this.conn != nil || this.tcpConn != nil {
		err = errors.New(fmt.Sprintf("内部对象已初始化,conn=%p,tcpConn=%p", this.conn, this.tcpConn))
		fmt.Println(err)
		return err
	}

	if _, err = net.ResolveTCPAddr("tcp", address); err != nil {
		fmt.Println(err)
		return err
	}

	this.address = address
	this.isAlive = true
	this.isConnected = false
	this.doConnect(nil)

	return nil
}

func (this *TcpClient) doConnect(err error) {
	if this.isConnected {
		this.isConnected = false
		this.OnDisconnected(this, err)
	}

	go func() {
		var err error
		for this.isAlive && !this.isConnected {
			if this.conn, err = net.Dial("tcp", this.address); err != nil {
				time.Sleep(time.Second * 5)
			} else {
				this.isConnected = true
				this.OnConnected(this)

				this.doReadConn()
				break
			}
		}
	}()
}

func (this *TcpClient) doReadConn() {
	go func() {
		buf := make([]byte, 2048)
		for {
			if num, err := this.conn.Read(buf); err != nil {
				this.doConnect(err)
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

func (this *TcpClient) Write(data []byte) (n int, err error) {
	if this.isAlive && this.isConnected {
		if this.conn != nil {
			return this.conn.Write(data)
		} else if this.tcpConn != nil {
			return this.Write(data)
		} else {
			panic("逻辑错误")
		}
	} else {
		err := errors.New(fmt.Sprintf("无法工作,isAlive=%v,isConnected=%v", this.isAlive, this.isConnected))
		return 0, err
	}
}

func (this *TcpClient) doReadTcpConn() {
	go func() {
		if true {
			this.isAlive = true
			this.isConnected = true
			this.OnConnected(this)
		}

		buf := make([]byte, 1536)
		for {
			if num, err := this.tcpConn.Read(buf); err != nil {
				if true {
					this.isAlive = false
					this.isConnected = false
					this.OnDisconnected(this, err)
					this.tcpConnChan <- disconnectedChanType{this.tcpConn, err}
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

	this.doAccept()
	this.doSth()

	return err
}

func (this *TcpServer) doAccept() {
	go func() {
		for {
			if tcpConn, err := this.svrListener.AcceptTCP(); err != nil {
				fmt.Println(err)
				break
			} else {
				this.tcpConnAddChan <- tcpConn
			}
		}
		fmt.Println("acceptTCPConn finish")
	}()
}

func (this *TcpServer) doSth() {
	go func() {
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
	}()
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
	fmt.Println("funOnConnected", client.LocalAddr(), client.RemoteAddr())
}
func funOnDisconnected(client *TcpClient, err error) {
	fmt.Println("funOnDisconnected", err, client.LocalAddr(), client.RemoteAddr())
}

func funOnReceivedData(client *TcpClient, data []byte) {
	fmt.Println("funOnReceivedData,", data)
}
func main_bak2() {
	e := easylog.Elog.InitFromFile("config.json")
	defer easylog.Elog.Terminate()
	if logger, ok := easylog.Elog.GetDefault(); ok {
		logger.Println("1", "2", "3")
		easylog.Elog.Printf(easylog.DN, "a=%v,b=%v,c=%v", ok, e, "123456")
	}
	fmt.Println(e)
	return
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

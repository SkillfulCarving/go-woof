package client

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go-woof/proxy/tcp/public"
	"go-woof/utils"
)

type node struct {
	name       string
	proxyAddr  string
	remotePort uint16
}

func (n *node) register(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	//   服务端写入注册信息
	msg := public.Message{
		Method: public.TypeRegister,
		Name:   n.name,
		Port:   n.remotePort,
		Status: 1,
	}
	if err = msg.Write(conn); err != nil {
		return err
	}

	// 获取服务端返回的信息，判断是否注册成功
	var respMsg public.Message
	_ = respMsg.Read(conn)

	if respMsg.Status != 0 && respMsg.Method == public.TypeResponse {
		log.Println(respMsg.Msg) //
	} else {
		return errors.New(respMsg.Msg)
	}

	for {
		var connInfo public.ConnInfo
		if err = connInfo.Read(conn); err != nil {
			return err
		}
		if connInfo.Method == public.TypeRequest { // 进行注册
			go n.clientHandle(connInfo.Name, connInfo.ConnId, addr)
		}
		runtime.GC()
	}
}

func (n *node) clientHandle(name [16]byte, connId uint32, addr string) {
	if connId == 0 {
		return
	}
	proxyConn, err := net.Dial("tcp", n.proxyAddr)
	if err != nil {
		return
	}

	// 连接服务端
	sConn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	// 数据请求建立连接通道
	msg := public.Message{
		Method: public.TypeRequest,
		ConnId: connId,
		Name:   fmt.Sprintf("%x", name),
		Status: 1,
	}
	_ = msg.Write(sConn)
	// 是否请求建立成功
	var respMsg public.Message
	err = respMsg.Read(sConn)
	if err != nil || respMsg.Status == 0 || respMsg.Method != public.TypeResponse {
		return
	}
	public.DataExchange(sConn, proxyConn)
	runtime.GC()
}

type Client struct {
	conf  utils.TCPClientConf
	nodes []node
}

func (c *Client) handleNodes() {
	for _, v := range strings.Split(c.conf.ProxyAddr, "\n") {
		nodeStr := strings.TrimSpace(v)
		if nodeStr == "" {
			continue
		}
		nodeStr = regexp.MustCompile("\\s+").ReplaceAllString(nodeStr, "-")
		nodeSplit := strings.Split(nodeStr, "-")
		if len(nodeSplit) != 3 {
			continue
		}
		port, err := strconv.Atoi(nodeSplit[2])
		if err != nil {
			continue
		}
		c.nodes = append(c.nodes, node{
			name:       nodeSplit[0],
			proxyAddr:  nodeSplit[1],
			remotePort: uint16(port),
		})
	}
}

func NewClient(conf utils.TCPClientConf) *Client {
	c := &Client{
		conf: conf,
	}
	c.handleNodes()
	return c
}

func (c *Client) registerNode(n node) {
	for {
		err := n.register(c.conf.ServerAddr)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 10)
	}
}

func (c *Client) Run() {
	for _, n := range c.nodes {
		log.Printf("register name: %v, addr: %v, remote_port %v", n.name, n.proxyAddr, n.remotePort)
		go c.registerNode(n)
	}
	select {}
}

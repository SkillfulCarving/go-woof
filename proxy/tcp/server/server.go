package server

import (
	"crypto/md5"
	"fmt"
	"log"
	"net"
	"sync"

	"go-woof/proxy/tcp/public"
	"go-woof/utils"
)

type Server struct {
	conf utils.TCPServerConf
	pool *sync.Map
}

func NewServer(conf utils.TCPServerConf) *Server {
	s := &Server{
		conf: conf,
		pool: &sync.Map{},
	}
	return s
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", s.conf.ServerAddr)
	// 属于注册端口
	if err != nil {
		log.Fatalf("server run error：%s", err.Error())
		return
	}

	log.Printf("proxy server run on %v，Waiting for client access...", s.conf.ServerAddr)
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Sprintf("addr%v, conn error:%v", clientConn.RemoteAddr(), err.Error()))
		}
		// 处理每一个端口l
		go s.handleConn(clientConn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	// 获取客户端返回的信息
	var msg public.Message
	err := msg.Read(conn)
	if err != nil || msg.Status == 0 {
		conn.Close()
		return
	}
	if msg.Method == public.TypeRegister && msg.Status != 0 {
		s.registerClient(conn, msg.Name, msg.Port)
	} else if msg.Method == public.TypeRequest && msg.Status != 0 {
		s.handleClient(conn, msg.Name, msg.ConnId)
	} else {
		conn.Close()
	}
}

// 客户端注册
func (s *Server) registerClient(conn net.Conn, name string, port uint16) {
	defer conn.Close()
	md5Name, bMd5Name := NewName(name)
	var respData public.Message
	// 进行服务注册
	if _, ok := s.pool.Load(md5Name); ok {
		respData.Msg = fmt.Sprintf("name-%v registered，cannot register repeatedly", name)
		respData.Method = public.TypeResponse
		_ = respData.Write(conn)
		log.Println(fmt.Sprintf("[-]access failed, duplicate service name! addr: %v name: %v proxy_port: %v", conn.RemoteAddr(), name, port))
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", port))
	if err != nil {
		respData.Msg = fmt.Sprintf("register failed! %v", err)
		respData.Method = public.TypeResponse
		_ = respData.Write(conn)
		log.Println(fmt.Sprintf("addr %v register failed! %v", conn.RemoteAddr(), err))
		return
	}

	defer func() {
		log.Println(fmt.Sprintf("[-]break conn! addr: %v name: %v proxy_port: %v", conn.RemoteAddr(), name, port))
		if val, ok := s.pool.LoadAndDelete(md5Name); ok {
			pool := val.(*ProxyPool)
			pool.ProxyPoolClose()
		}
		_ = listener.Close()
	}()

	s.pool.Store(md5Name, NewProxyPool())
	respData = public.Message{
		Method: public.TypeResponse,
		Status: 1,
		Msg:    fmt.Sprintf("name-%v register success", name),
	}
	_ = respData.Write(conn)
	log.Println(fmt.Sprintf("[+]register success! addr: %v name: %v proxy_port: %v", conn.RemoteAddr(), name, port))
	go func() {
		for {
			reqConn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				if val, ok := s.pool.Load(md5Name); ok {
					pool := val.(*ProxyPool)
					proxyConn := pool.CreateProxyConn(reqConn)
					connInfo := public.ConnInfo{
						Method: public.TypeRequest,
						Name:   bMd5Name,
						ConnId: proxyConn.ID,
					}
					if err = connInfo.Write(conn); err != nil {
						return
					}
				}
			}()
		}
	}()
	// 检测是否断开连接
	buffer := make([]byte, 1024)
	for {
		_, err = conn.Read(buffer)
		if err != nil {
			return
		}
	}
}

// 建立数据通信
func (s *Server) handleClient(conn net.Conn, name string, id uint32) {
	if val, ok := s.pool.Load(name); !ok {
		return
	} else {
		pool := val.(*ProxyPool)
		proxyConn := pool.GetProxyConn(id)
		if proxyConn.ID == 0 {
			return
		}
		defer pool.CloseProxyConn(id)

		// 返回建立成功
		msg := public.Message{
			Method: public.TypeResponse,
			Status: 1,
		}
		_ = msg.Write(conn)
		public.DataExchange(proxyConn.Conn, conn)
	}
}

func NewName(s string) (string, [16]byte) {
	data := []byte(s)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has), has
}

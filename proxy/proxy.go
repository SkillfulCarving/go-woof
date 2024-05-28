package proxy

import (
	"go-woof/proxy/http"
	"log"

	"go-woof/proxy/tcp/client"
	"go-woof/proxy/tcp/server"
	"go-woof/utils"
)

type empty struct {
}

func (e *empty) Run() {
	log.Println("please select a mode")
}

type Proxy interface {
	Run()
}

func NewProxy(service string, conf utils.Config) Proxy {

	switch service {
	case "http":
		return http.NewHttp(conf.Http)
	case "server":
		return server.NewServer(conf.TCP.Server)
	case "client":
		return client.NewClient(conf.TCP.Client)
	default:
	}
	return new(empty)
}

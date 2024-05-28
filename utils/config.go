package utils

import (
	"github.com/BurntSushi/toml"
	"log"
)

type HttpConf struct {
	ServerAddr string `toml:"listen"` // 服务器地址
	Whitelist  string `toml:"whitelist"`
	Username   string `toml:"username"`
	Password   string `toml:"password"`
}

type TCPServerConf struct {
	ServerAddr string `toml:"listen"` // 服务器地址
}

type TCPClientConf struct {
	ServerAddr string `toml:"remote"`     // 服务器地址
	ProxyAddr  string `toml:"proxy_addr"` // 代理地址
	ProxyFile  string `toml:"proxy_file"` // 代理文件地址
}

type TCPConf struct {
	Client TCPClientConf `toml:"client"`
	Server TCPServerConf `toml:"server"`
}

type Config struct {
	Http HttpConf `toml:"http"`
	TCP  TCPConf  `toml:"tcp"`
}

func LoadConfig(file string) Config {
	conf := Config{}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		log.Fatal("load config err", err)
	}
	return conf
}

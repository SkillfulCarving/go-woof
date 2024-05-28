package main

import (
	"github.com/urfave/cli/v2"
	"go-woof/proxy"
	"go-woof/utils"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Usage = "GoWoof"

	var conf utils.Config

	app.Commands = []*cli.Command{
		{
			Name:        "server",
			Usage:       "Run Woof server",
			Description: "run tcp-proxy server.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Usage:   "select config file",
					Value:   "",
				},
				&cli.StringFlag{
					Name:        "listen",
					Aliases:     []string{"l"},
					Usage:       "select config file",
					Value:       "0.0.0.0:5000",
					Destination: &conf.TCP.Server.ServerAddr,
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("config") != "" {
					conf = utils.LoadConfig(c.String("config"))
				}
				p := proxy.NewProxy("server", conf)
				p.Run()
				return nil
			},
		},
		{
			Name:        "client",
			Usage:       "Run Woof client",
			Description: "run tcp-proxy client.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Usage:   "select config file",
					Value:   "",
				},
				&cli.StringFlag{
					Name:        "remote",
					Aliases:     []string{"r"},
					Usage:       "remote addr",
					Value:       "",
					Destination: &conf.TCP.Client.ServerAddr,
				},
				&cli.StringFlag{
					Name:        "proxy",
					Aliases:     []string{"p"},
					Usage:       "node1-127.0.0.1:3389-3389",
					Value:       "",
					Destination: &conf.TCP.Client.ProxyAddr,
				},
				&cli.StringFlag{
					Name:        "file",
					Aliases:     []string{"f"},
					Usage:       "need to proxy addr file",
					Value:       "",
					Destination: &conf.TCP.Client.ProxyFile,
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("config") != "" {
					conf = utils.LoadConfig(c.String("config"))
				}
				p := proxy.NewProxy("client", conf)
				p.Run()
				return nil
			},
		}, {
			Name:        "http",
			Usage:       "Run Http server",
			Description: "run http-proxy server.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Usage:   "select config file",
					Value:   "",
				},
				&cli.StringFlag{
					Name:        "listen",
					Aliases:     []string{"l"},
					Usage:       "listen addr",
					Value:       "0.0.0.0:6000",
					Destination: &conf.Http.ServerAddr,
				},
			},
			Action: func(c *cli.Context) error {
				if c.String("config") != "" {
					conf = utils.LoadConfig(c.String("config"))
				}
				p := proxy.NewProxy("http", conf)
				p.Run()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

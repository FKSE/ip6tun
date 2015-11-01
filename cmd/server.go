package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fkse/ip6tun"
	"github.com/kelseyhightower/envconfig"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// get config from environment
	var conf ip6tun.ServerConfig
	envconfig.MustProcess("ip6tun", &conf)
	// start server & broker
	srv, err := ip6tun.NewServer(&conf)
	if err != nil {
		panic(err)
	}
	go srv.Run()
	// SIGINT and SIGTERM handling
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info(<-ch)
	// stop server
	srv.Stop()
}

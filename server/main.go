package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Debug bool   `default:"false"`
	Port  int    `default:"10026"`
	Key   string `required:"true"`
}

func main() {
	// get config from enviroment
	var c config
	envconfig.MustProcess("ip6tun", &c)

	fmt.Println(c)
}

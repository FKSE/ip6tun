package main

import (
	"fmt"
	"github.com/fkse/ip6update/client"
	"github.com/fkse/ip6update/server"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func main() {
	// validate count
	if len(os.Args) != 3 {
		help()
	}
	if _, err := os.Stat(os.Args[2]); os.IsNotExist(err) {
		fmt.Printf("File %s not found\n", os.Args[2])
		help()
	}
	switch os.Args[1] {
	case "server":
		c := new(server.Config)
		unmarshal(os.Args[2], c)
		server.Run(c)
	case "client":
		c := new(client.Config)
		unmarshal(os.Args[2], c)
		client.Run(c)
	}
}

func help() {
	fmt.Printf("Usage: %s (server|client) /path/to/config.yaml\n", os.Args[0])
	os.Exit(0)
}

func unmarshal(file string, v interface{}) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = yaml.Unmarshal(b, v)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

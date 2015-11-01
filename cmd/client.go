package main

import (
	"fmt"
	"github.com/fkse/ip6tun"
	"github.com/kelseyhightower/envconfig"
	"os"
	"strconv"
	"strings"
)

// replaced at compile time using -ldflags
const Version = "1.0.0-Alpha"

func main() {
	// process config
	var conf struct {
		Host string `default:"localhost"`
		Port int    `default:"8080"`
		Key  string `required:"true"`
	}
	envconfig.MustProcess("ip6tun", &conf)
	// validate args
	if len(os.Args) != 4 {
		showHelp()
	}
	// validate args
	lPort, err := strconv.Atoi(os.Args[2])
	showError(err)
	sPort, err := strconv.Atoi(os.Args[3])
	showError(err)
	tName := os.Args[1]
	// create client instance
	client := ip6tun.NewClient(conf.Host, conf.Port, conf.Key)
	// try to update if tunnel exists
	tunnels, err := client.List()
	showError(err)
	updated := false
	for _, tun := range tunnels {
		if tun.Name == tName {
			// update tunnel
			_, err := client.Update(tun.Id, tun.Name, uint16(lPort), uint16(sPort))
			showError(err)
			// success
			fmt.Println("Tunnel updated")
			updated = true
		}
	}
	if !updated {
		// create new tunnel
		_, err = client.Create(tName, uint16(lPort), uint16(sPort))
		showError(err)
		// Show
		fmt.Println("Tunnel created")

	}
	// list all tunnels
	fmt.Println("Listing all Tunnels")
	list, err := client.List()
	showError(err)
	for _, t := range list {
		fmt.Printf("%d, Name: %s, LocalPort: %d, RemoteHost: %s, RemotePort: %d\n",
			t.Id,
			t.Name,
			t.LocalPort,
			t.RemoteHost,
			t.RemotePort)
		fmt.Printf("Message-Log:\n\t%s\n", strings.Join(t.MessageLog, "\n\t"))
		fmt.Println("---------")
	}
}

// show help an exit
func showHelp() {
	fmt.Printf("Version: %s\nUsage: %s <name> <local-port> <server-port>\n", Version, os.Args[0])
	os.Exit(0)
}

// show error an exit
func showError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

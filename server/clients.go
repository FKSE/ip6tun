package server

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

type Client struct {
	ClientId   [32]byte
	Address    net.Addr
	LocalPort  uint16
	RemotePort uint16
	RemoveAt   time.Time
	Cmd        *exec.Cmd
}

func (c *Client) String() string {
	return fmt.Sprintf("Id: %s, Address: %s, Local-Port: %d, Remote-Port: %d, Remove-At: %s", c.ClientId, c.Address, c.LocalPort, c.RemotePort, c.RemoveAt)
}

// List all clients
func listClients() {
	for _, c := range clients {
		fmt.Println(c.String())
	}
}

// Remove old entries
func cleanup() {
	for _ = range tickerCleanup.C {
		fmt.Println("Starting cleanup")
		for k, c := range clients {
			if c.RemoveAt.Before(time.Now()) {
				mutex.Lock()
				fmt.Printf("Removing Client %s\n", c.String())
				delete(clients, k)
				mutex.Unlock()
			}
		}
	}
}

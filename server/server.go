package server

import (
	"fmt"
	"github.com/fkse/ip6update/protocol"
	"net"
	"os"
	"sync"
	"time"
	"crypto/cipher"
	"crypto/aes"
)

// Server configuration
type Config struct {
	Port int
	SecretKey string `yaml:"secret_key"`
}

type Client struct {
	ClientId   [32]byte
	Address    net.Addr
	LocalPort  uint16
	RemotePort uint16
	RemoveAt time.Time
}

func (c *Client) String() string {
	return fmt.Sprintf("Id: %s, Address: %s, Local-Port: %d, Remote-Port: %d, Remove-At: %s", c.ClientId, c.Address, c.LocalPort, c.RemotePort, c.RemoveAt)
}

// List of all clients
var clients map[string]Client
// Mutex for accessing clients list
var mutex *sync.Mutex
// AES cipher block
var ciph cipher.Block
// Cleanup ticker
var tickerCleanup *time.Ticker

func Run(c *Config) {
	key := []byte(c.SecretKey)
	//validate key length
	if (len(key) < 32) {
		fmt.Printf("The given key %s is to short. Expected length 32 given %d.\n", c.SecretKey, len(key))
		return
	}
	//init cipher
	var err error
	ciph, err = aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// init mutex
	mutex = &sync.Mutex{}
	// Map containing  all clients
	clients = make(map[string]Client)

	// Cleanup of old clients
	tickerCleanup = time.NewTicker(time.Minute * 15)
	go cleanup()

	// Start server
	ln, err := net.Listen("tcp6", fmt.Sprintf(":%d", c.Port))
	fmt.Printf("Server running at port %d\n", c.Port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}
		go handle(conn)
	}
}

func handle(c net.Conn) {
	// get message
	buf := make([]byte, 1024)
	// read input
	l, err := c.Read(buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	// Decrypt message
	ciph.Decrypt(buf[:l], buf[:l])
	// decode
	m, err := protocol.Unmarshal(buf[:l])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	// get index as string
	key := string(m.ClientId[:])
	// Update/Create
	if m.Type == protocol.TypeUpdate {
		mutex.Lock()
		clients[key] = Client{
			ClientId:m.ClientId,
			Address:c.RemoteAddr(),
			LocalPort:m.LocalPort,
			RemotePort:m.RemotePort,
			RemoveAt:time.Now().Add(12 * time.Hour),
		}
		mutex.Unlock()
	} else if m.Type == protocol.TypeDelete {
		mutex.Lock()
		delete(clients, key)
		mutex.Unlock()
	} else {
		fmt.Fprintln(os.Stderr, "Invalid message")
	}

	listClients()
}

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
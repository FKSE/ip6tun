package client

import (
	"crypto/aes"
	"fmt"
	"github.com/fkse/ip6update/protocol"
	"net"
)

type Config struct {
	Address   string
	Uid       string
	SecretKey string `yaml:"secret_key"`
}

func Run(c *Config) {

	id := []byte(c.Uid)
	key := []byte(c.SecretKey)

	if len(key) < 32 {
		fmt.Printf("The given key %s is to short. Expected length 32 given %d.\n", c.SecretKey, len(key))
		return
	}
	// Connect to server
	conn, err := net.Dial("tcp6", c.Address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// Create message
	m := &protocol.Message{
		Type:       protocol.TypeUpdate,
		LocalPort:  5000,
		RemotePort: 1237,
	}
	copy(m.ClientId[0:32], id[:])
	// Marshal message
	b, err := m.Marshal()
	if err != nil {
		panic(err)
	}
	// Init cipher
	ciph, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciph.Encrypt(b, b)
	// send data
	conn.Write(b)
}

package server

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fkse/ip6update/protocol"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// Server configuration
type Config struct {
	Port       int    `yaml:"port"`
	HttpPort   int    `yaml:"http_port"`
	SecretKey  string `yaml:"secret_key"`
	Bin6Tunnel string `yaml:"bin_6tunnel"`
}

// App config
var conf *Config

// List of all clients
var clients map[string]Client

// Mutex for accessing clients list
var mutex *sync.Mutex

// AES cipher block
var ciph cipher.Block

// Cleanup ticker
var tickerCleanup *time.Ticker

func Run(c *Config) {
	conf = c
	key := []byte(c.SecretKey)
	//validate key length
	if len(key) < 32 {
		log.Errorf("The given key %s is to short. Expected length 32 given %d.", c.SecretKey, len(key))
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

	// Start http server if set
	if c.HttpPort != 0 {
		http.HandleFunc("/", handleHttp)
		go http.ListenAndServe(fmt.Sprintf(":%d", c.HttpPort), nil)
		log.Infof("HTTP-Server running at port %d", c.HttpPort)
	}

	// Start server
	ln, err := net.Listen("tcp6", fmt.Sprintf(":%d", c.Port))
	log.Infof("Server running at port %d", c.Port)
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
	// Close connection later
	defer c.Close()
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
		cl := Client{
			ClientId:   m.ClientId,
			Address:    c.RemoteAddr().(*net.TCPAddr).IP,
			LocalPort:  m.LocalPort,
			RemotePort: m.RemotePort,
			RemoveAt:   time.Now().Add(12 * time.Hour),
		}
		// set cmd
		cl.Cmd = exec.Command(conf.Bin6Tunnel, "-d", strconv.Itoa(int(m.RemotePort)), cl.Address.String(), strconv.Itoa(int(m.LocalPort)))
		log.Infof("%s -d %d %s %d", conf.Bin6Tunnel, m.RemotePort, cl.Address.String(), m.LocalPort)
		err := cl.Cmd.Run()
		if err != nil {
			log.Errorf("Unable to start tunnel6: %s", err.Error())
			// Send error to client; Message is basically the same besides the type
			m.Type = protocol.TypeErrorNoTunnel
			// encode and encrypt
			b, err := m.Marshal()
			if err != nil {
				log.Error(err)
				return
			}
			ciph.Encrypt(b, b)
			// send message to client
			c.Write(b)
			return
		}
		mutex.Lock()
		clients[key] = cl
		mutex.Unlock()
	} else if m.Type == protocol.TypeDelete {
		mutex.Lock()
		delete(clients, key)
		mutex.Unlock()
	} else {
		log.Warn("Received invalid message")
	}

	listClients()
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	// Check for key
	if r.URL.Query().Get("key") != conf.SecretKey {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	fmt.Printf("Hi there, I love %s!\n", r.URL)
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL)
}

package ip6tun

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrTunnelNotFound = errors.New("Tunnel not found")
)

// Tunnel is a 4in6 tunnel
type Tunnel struct {
	Id         uint32     `json:"id"`
	Name       string     `json:"name"`
	LocalPort  uint16     `json:"local_port"`
	RemoteHost string     `json:"remote_host"`
	RemotePort uint16     `json:"remote_port"`
	MessageLog []string   `json:"message_log"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	ln         net.Listener
	quit       chan bool
	mutex      sync.Mutex
	wg         *sync.WaitGroup
}

// Update allows updating the remote host and port
// TODO: allow updating name an local port
func (t *Tunnel) Update(rHost string, rPort uint16) {
	n := time.Now()
	t.UpdatedAt = &n
	t.RemoteHost = rHost
	t.RemotePort = rPort
}

// Close the tunnel graceful
func (t *Tunnel) Close() {
	//t.quit <- true
	close(t.quit)
	// wait for all connections to close
	t.wg.Wait()
}

// Start the tunnel
func (t *Tunnel) Start() {
	t.log("Tunnel started")
	t.wg.Add(1)
	defer t.wg.Done()
	// graceful stopping
	go func() {
		<-t.quit
		t.ln.Close()
	}()
	for {
		conn, err := t.ln.Accept()
		if err != nil {
			// check if channel is closed
			if _, ok := <-t.quit; !ok {
				return
			}
			log.Error(err)
			continue
		}
		t.log(fmt.Sprintf("Accept connection from %s", conn.RemoteAddr()))
		t.wg.Add(1)
		go t.handleConnection(conn)
	}
}

func (t *Tunnel) log(msg string) {
	msg = fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), msg)
	t.mutex.Lock()
	t.MessageLog = append(t.MessageLog, msg)
	t.mutex.Unlock()
	log.Debugf("[Tunnel:%d]%s", t.Id, msg)
}

func (t *Tunnel) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer t.wg.Done()
	// connect to remote host
	rc, err := net.Dial("tcp6", fmt.Sprintf("[%s]:%d", t.RemoteHost, t.RemotePort))
	if err != nil {
		t.log(err.Error())
		return
	}
	// graceful stopping
	go func() {
		<-t.quit
		conn.Close()
		rc.Close()
	}()
	defer rc.Close()
	var wg sync.WaitGroup
	// bidirectional copying of the data
	wg.Add(1)
	go t.copy(rc, conn, &wg)
	wg.Add(1)
	go t.copy(conn, rc, &wg)
	// wait for both connections to EOF
	wg.Wait()
	t.log(fmt.Sprintf("Closed connection from %s", conn.RemoteAddr()))
}

// simple wrapper for io.Copy
func (t *Tunnel) copy(dst net.Conn, src net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	l, err := io.Copy(dst, src)
	if err != nil {
		t.log(err.Error())
		return
	}
	t.log(fmt.Sprintf("Copied %d bytes from %s to %s", l, src.RemoteAddr(), dst.RemoteAddr()))
}

// Broker
type Broker struct {
	currentId     uint32
	tunnels       map[uint32]*Tunnel
	idleTimeout   time.Duration
	cleanupTicker *time.Ticker
}

func NewBroker(idleTimeout time.Duration) *Broker {
	b := &Broker{
		currentId:     0,
		tunnels:       make(map[uint32]*Tunnel),
		idleTimeout:   idleTimeout,
		cleanupTicker: time.NewTicker(30 * time.Second),
	}
	// cleanup
	go func() {
		for range b.cleanupTicker.C {
			b.Cleanup()
		}
	}()

	return b
}

// Close closes all open tunnels
func (b *Broker) Close() {
	// stop cleanup
	b.cleanupTicker.Stop()
	// close all tunnels
	for id, tun := range b.tunnels {
		tun.Close()
		log.Debugf("Closed tunnel %s with id %d", tun.Name, id)
	}
}

// Tunnels returns a list of all active tunnels
func (b *Broker) Tunnels() []*Tunnel {
	list := make([]*Tunnel, 0, len(b.tunnels))
	for _, client := range b.tunnels {
		list = append(list, client)
	}
	return list
}

// Add creates a new new tunnel from rHost:rPort to lPort
func (b *Broker) Add(name, rHost string, rPort, lPort uint16) (*Tunnel, error) {
	// create ipv4 listener
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", lPort))
	if err != nil {
		return nil, err
	}
	t := time.Now()
	// create client
	client := &Tunnel{
		Id:         b.currentId,
		Name:       name,
		LocalPort:  lPort,
		RemoteHost: rHost,
		RemotePort: rPort,
		CreatedAt:  &t,
		ln:         ln,
		quit:       make(chan bool),
		wg:         &sync.WaitGroup{},
	}
	// add client to repository
	b.tunnels[b.currentId] = client
	b.currentId++
	// start tunnel
	go client.Start()
	// return the created client
	return client, nil
}

// Update allows the modification of the remote host and port.
func (b *Broker) Update(id uint32, rHost string, rPort uint16) (*Tunnel, error) {
	tun, err := b.Get(id)
	if err != nil {
		return nil, err
	}
	tun.Update(rHost, rPort)
	return tun, err
}

// Get returns a tunnel by its id
func (b *Broker) Get(id uint32) (*Tunnel, error) {
	if tun, ok := b.tunnels[id]; ok {
		return tun, nil
	}
	return nil, ErrTunnelNotFound
}

// Delete closes the tunnel graceful and removes it from the list
func (b *Broker) Delete(id uint32) error {
	if tun, ok := b.tunnels[id]; ok {
		// close tunnel
		tun.Close()
		// remove from struct
		delete(b.tunnels, id)
	}
	return ErrTunnelNotFound
}

// Cleanup deletes clients which haven't been updated for some time
func (b *Broker) Cleanup() {

}

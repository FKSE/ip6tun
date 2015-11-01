package protocol

import (
	"crypto/cipher"
	"fmt"
	"net"
)

type IP6TunListener struct {
	listener *net.TCPListener
	gcm      cipher.AEAD
}

func ListenIP6Tun(port int, key string) (*IP6TunListener, error) {
	laddr, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[::1]:%d", port))
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp6", laddr)
	if err != nil {
		return nil, err
	}
	// init gcm cipher
	gcm, err := initCipher(key)
	if err != nil {
		return nil, err
	}
	return &IP6TunListener{ln, gcm}, nil
}

func (l *IP6TunListener) Accept() (c *IP6TunConn, err error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &IP6TunConn{conn.(*net.TCPConn), l.gcm}, nil
}

func (l *IP6TunListener) Close() error {
	return l.listener.Close()
}

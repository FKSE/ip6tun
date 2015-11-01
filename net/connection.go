package protocol

import (
	"crypto/cipher"
	"fmt"
	"net"
)

type IP6TunConn struct {
	conn *net.TCPConn
	gcm  cipher.AEAD
}

func Dial(host string, port int, key string) (*IP6TunConn, error) {
	// resolver remote address
	raddr, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[%s]:%d", host, port))
	if err != nil {
		return nil, err
	}
	// solve local address
	laddr, err := net.ResolveTCPAddr("tcp6", "[::1]:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp6", laddr, raddr)
	if err != nil {
		return nil, err
	}
	// gcm cipher
	gcm, err := initCipher(key)
	if err != nil {
		return nil, err
	}
	return &IP6TunConn{conn, gcm}, nil
}

func (c *IP6TunConn) Close() error {
	return c.conn.Close()
}

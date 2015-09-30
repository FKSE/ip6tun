package protocol

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
)

const (
	TypeUpdate byte = iota
	TypeDelete
	TypeAcknowledge
	TypeErrorNoTunnel
)

type Message struct {
	Type       byte
	ClientId   [32]byte
	PortMapping map[uint16]uint16 // local => remote
}

func (m *Message) String() string {
	return fmt.Sprintf("Type: %x, Sender-Id: % x, Local-Port: %d, Remote-Port: %d", m.Type, m.ClientId, m.LocalPort, m.RemotePort)
}

func (m *Message) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(b []byte) (m *Message, err error) {
	m = new(Message)
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func initCipher(key string) (cipher.AEAD, error) {
	// Init aes cipher
	ciph, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	// gcm cipher
	gcm, err := cipher.NewGCM(ciph)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}

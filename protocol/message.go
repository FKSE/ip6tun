package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	TypeUpdate      = 0x01
	TypeDelete      = 0x02
	TypeAcknowledge = 0x03
	TypeErrorNoTunnel = 0x04 // Unable to create tunnel
)

type Message struct {
	Type       byte
	ClientId   [32]byte
	LocalPort  uint16
	RemotePort uint16
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

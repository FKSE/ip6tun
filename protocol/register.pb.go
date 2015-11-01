// Code generated by protoc-gen-go.
// source: register.proto
// DO NOT EDIT!

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:
	register.proto

It has these top-level messages:
	PortMapping
	Register
*/
package protocol

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type PortMapping struct {
	LocalPort        *uint32 `protobuf:"varint,1,req,name=localPort" json:"localPort,omitempty"`
	RemotePort       *uint32 `protobuf:"varint,2,req,name=remotePort" json:"remotePort,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PortMapping) Reset()         { *m = PortMapping{} }
func (m *PortMapping) String() string { return proto.CompactTextString(m) }
func (*PortMapping) ProtoMessage()    {}

func (m *PortMapping) GetLocalPort() uint32 {
	if m != nil && m.LocalPort != nil {
		return *m.LocalPort
	}
	return 0
}

func (m *PortMapping) GetRemotePort() uint32 {
	if m != nil && m.RemotePort != nil {
		return *m.RemotePort
	}
	return 0
}

type Register struct {
	ClientId         *string        `protobuf:"bytes,1,req,name=clientId" json:"clientId,omitempty"`
	Mapping          []*PortMapping `protobuf:"bytes,2,rep,name=mapping" json:"mapping,omitempty"`
	XXX_unrecognized []byte         `json:"-"`
}

func (m *Register) Reset()         { *m = Register{} }
func (m *Register) String() string { return proto.CompactTextString(m) }
func (*Register) ProtoMessage()    {}

func (m *Register) GetClientId() string {
	if m != nil && m.ClientId != nil {
		return *m.ClientId
	}
	return ""
}

func (m *Register) GetMapping() []*PortMapping {
	if m != nil {
		return m.Mapping
	}
	return nil
}

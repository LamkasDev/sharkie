package structs

import "sync"

var GlobalIpmiManager = NewIpmiManager()

const IPMI_MAGIC = 0xDEADBADECAFEBEAF

const (
	IMPI_CREATE_SERVER = 0x0
	IMPI_CREATE_CLIENT = 0x2
	IMPI_DESTROY       = 0x3
	IMPI_DISCONNECT    = 0x310
	IMPI_INVOKE_SYNC   = 0x320
	IMPI_CONNECT       = 0x400
)

type IpmiManager struct {
	Clients    map[uint32]*IpmiClient
	Servers    map[uint32]*IpmiServer
	Lock       sync.RWMutex
	NextHandle uint32
}

func NewIpmiManager() *IpmiManager {
	return &IpmiManager{
		Clients:    map[uint32]*IpmiClient{},
		Servers:    map[uint32]*IpmiServer{},
		NextHandle: 0x40000001,
	}
}

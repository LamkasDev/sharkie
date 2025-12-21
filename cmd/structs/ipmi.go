package structs

import "sync"

var GlobalIpmiManager = NewIpmiManager()

type IpmiClient struct {
	Handle  uint32
	Name    string
	UserPtr uintptr
}

const IPMI_MAGIC = 0xDEADBADECAFEBEAF

const (
	IMPI_CREATE      = 0x2
	IMPI_DESTROY     = 0x3
	IMPI_INVOKE_SYNC = 0x320
	IMPI_CONNECT     = 0x400
)

type IpmiManager struct {
	Clients          map[uint32]*IpmiClient
	ClientsLock      sync.RWMutex
	NextClientHandle uint32
}

func NewIpmiManager() *IpmiManager {
	return &IpmiManager{
		Clients:          map[uint32]*IpmiClient{},
		NextClientHandle: 0x40000001,
	}
}

func CreateImpiClient(name string, userPtr uintptr) *IpmiClient {
	GlobalIpmiManager.ClientsLock.Lock()
	defer GlobalIpmiManager.ClientsLock.Unlock()

	client := &IpmiClient{
		Handle:  GlobalIpmiManager.NextClientHandle,
		Name:    name,
		UserPtr: userPtr,
	}
	GlobalIpmiManager.Clients[client.Handle] = client
	GlobalIpmiManager.NextClientHandle++
	return client
}

func GetImpiClient(handle uint32) *IpmiClient {
	GlobalIpmiManager.ClientsLock.RLock()
	defer GlobalIpmiManager.ClientsLock.RUnlock()
	return GlobalIpmiManager.Clients[handle]
}

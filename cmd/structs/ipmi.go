package structs

import "sync"

var GlobalIpmiManager = NewIpmiManager()

type IpmiClient struct {
	Handle uint32
	Name   string
	ObjPtr uintptr
}

type IpmiServer struct {
	Handle uint32
	Name   string
	ObjPtr uintptr
}

const IPMI_MAGIC = 0xDEADBADECAFEBEAF

const (
	IMPI_CREATE_SERVER = 0x0
	IMPI_CREATE_CLIENT = 0x2
	IMPI_DESTROY       = 0x3
	IMPI_DISCONNECT    = 0x310
	IMPI_INVOKE_SYNC   = 0x320
	IMPI_CONNECT       = 0x400
)

const ImpiBufferDefault = 4096

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

func CreateImpiClient(name string, userPtr uintptr) *IpmiClient {
	GlobalIpmiManager.Lock.Lock()
	defer GlobalIpmiManager.Lock.Unlock()

	client := &IpmiClient{
		Handle: GlobalIpmiManager.NextHandle,
		Name:   name,
		ObjPtr: userPtr,
	}
	GlobalIpmiManager.Clients[client.Handle] = client
	GlobalIpmiManager.NextHandle++
	return client
}

func GetImpiClient(handle uint32) *IpmiClient {
	GlobalIpmiManager.Lock.RLock()
	defer GlobalIpmiManager.Lock.RUnlock()
	return GlobalIpmiManager.Clients[handle]
}

func CreateImpiServer(name string, userPtr uintptr) *IpmiServer {
	GlobalIpmiManager.Lock.Lock()
	defer GlobalIpmiManager.Lock.Unlock()

	server := &IpmiServer{
		Handle: GlobalIpmiManager.NextHandle,
		Name:   name,
		ObjPtr: userPtr,
	}
	GlobalIpmiManager.Servers[server.Handle] = server
	GlobalIpmiManager.NextHandle++
	return server
}

func GetImpiServer(handle uint32) *IpmiServer {
	GlobalIpmiManager.Lock.RLock()
	defer GlobalIpmiManager.Lock.RUnlock()
	return GlobalIpmiManager.Servers[handle]
}

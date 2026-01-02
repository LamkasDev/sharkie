package structs

import "fmt"

type IpmiClient struct {
	Handle    uint32
	Name      string
	ObjPtr    uintptr
	EventFlag *EventFlag
}

type IpmiPollEventFlag struct {
	WaitPattern   uint32
	_             [12]byte
	WaitMode      uint32
	_             [4]byte
	OutPatternPtr uintptr
}

func CreateImpiClient(name string, userPtr uintptr) *IpmiClient {
	GlobalIpmiManager.Lock.Lock()
	defer GlobalIpmiManager.Lock.Unlock()

	client := &IpmiClient{
		Handle:    GlobalIpmiManager.NextHandle,
		Name:      name,
		ObjPtr:    userPtr,
		EventFlag: NewEventFlag(fmt.Sprintf("%sEvf", name), EVF_ATTR_TH_FIFO|EVF_ATTR_MULTI, 0, 0),
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

func GetImpiClientByName(name string) *IpmiClient {
	GlobalIpmiManager.Lock.RLock()
	defer GlobalIpmiManager.Lock.RUnlock()

	for _, client := range GlobalIpmiManager.Clients {
		if client.Name == name {
			return client
		}
	}

	return nil
}

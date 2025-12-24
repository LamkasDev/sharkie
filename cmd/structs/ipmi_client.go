package structs

type IpmiClient struct {
	Handle uint32
	Name   string
	ObjPtr uintptr
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

package structs

type IpmiServer struct {
	Handle    uint32
	Name      string
	ObjPtr    uintptr
	EventFlag *EventFlag
}

func (server *IpmiServer) CreateEventFlag(name string) {
	server.EventFlag = CreateEventFlag(name, EVF_ATTR_TH_FIFO|EVF_ATTR_MULTI, 0, 0)
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

func GetImpiServerByName(name string) *IpmiServer {
	GlobalIpmiManager.Lock.RLock()
	defer GlobalIpmiManager.Lock.RUnlock()

	for _, server := range GlobalIpmiManager.Servers {
		if server.Name == name {
			return server
		}
	}

	return nil
}

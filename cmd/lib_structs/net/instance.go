package net

import "net"

var GlobalNetConnectionInstance *NetConnectionInstance

type NetConnectionInstance struct {
	MacAddress [6]byte
}

func NewNetConnectionInstance() *NetConnectionInstance {
	inst := &NetConnectionInstance{}

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			// skip loopback
			if i.Flags&net.FlagLoopback != 0 {
				continue
			}
			if len(i.HardwareAddr) >= 6 {
				copy(inst.MacAddress[:], i.HardwareAddr)
				break
			}
		}
	}

	return inst
}

func SetupNetConnectionInstance() {
	GlobalNetConnectionInstance = NewNetConnectionInstance()
}

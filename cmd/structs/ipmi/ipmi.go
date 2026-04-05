package ipmi

import (
	"fmt"
	"sync"

	"github.com/LamkasDev/sharkie/cmd/structs"
)

var GlobalIpmiManager *IpmiManager

const IPMI_MAGIC = 0xDEADBADECAFEBEAF

const (
	IMPI_CREATE_SERVER   = 0x0
	IMPI_DESTROY_SERVER  = 0x1
	IMPI_CREATE_CLIENT   = 0x2
	IMPI_DESTROY_CLIENT  = 0x3
	IMPI_CREATE_SESSION  = 0x4
	IMPI_DESTROY_SESSION = 0x5
	_                    = 0x10

	IMPI_RECEIVE_PACKET           = 0x201
	IMPI_SHUTDOWN_DISPATCHER      = 0x202
	IMPI_SEND_CONNECT_RESPONSE    = 0x212
	IMPI_SEND_DISCONNECT_RESPONSE = 0x222
	IMPI_RESPOND_TO_SYNC_REQUEST  = 0x232
	IMPI_INVOKE_ASYNC_CLIENT      = 0x241
	IMPI_RESPOND_TO_ASYNC_REQUEST = 0x242
	IMPI_TRY_GET_RESULT_CLIENT    = 0x243
	IMPI_GET_MESSAGE_CLIENT       = 0x251
	IMPI_TRY_GET_MESSAGE_CLIENT   = 0x252
	IMPI_SEND_MESSAGE             = 0x253
	IMPI_TRY_SEND_MESSAGE         = 0x254
	IMPI_EMPTY_MESSAGE_QUEUE      = 0x255

	IMPI_GET_CLIENT_PROCESS_ID = 0x302
	_                          = 0x303
	IMPI_DISCONNECT_CLIENT     = 0x310
	IMPI_INVOKE_SYNC_CLIENT    = 0x320

	IMPI_CONNECT                       = 0x400
	IMPI_COMPLETE_ASYNC_DISPATCH       = 0x444
	IMPI_GET_MESSAGE                   = 0x456
	IMPI_TRY_GET_MESSAGE               = 0x457
	_                                  = 0x463
	IMPI_IS_PEER_PRIVILEGED            = 0x464
	IMPI_GET_SERVER                    = 0x465
	IMPI_GET_USER_DATA_SERVER          = 0x466
	IMPI_GET_USER_DATA_CLIENT          = 0x467
	IMPI_GET_USER_DATA_SESSION         = 0x468
	IMPI_GET_SESSION_PID_BY_CLIENT_KID = 0x469
	IMPI_TRY_DISPATCH                  = 0x46A
	IPMI_REPORT_LONG_CALL              = 0x46B

	IMPI_WAIT_EVENT_FLAG = 0x490
	IMPI_POLL_EVENT_FLAG = 0x491
	_                    = 0x492
	IMPI_SET_EVENT_FLAG  = 0x493

	IMPI_TERMINATE_CONNECTION_CLIENT = 0x520
)

type IpmiManager struct {
	Clients    map[uint32]*IpmiClient
	Servers    map[uint32]*IpmiServer
	Lock       sync.RWMutex
	NextHandle uint32
}

// NewIpmiManager creates a new instance of IpmiManager.
func NewIpmiManager() *IpmiManager {
	return &IpmiManager{
		Clients:    map[uint32]*IpmiClient{},
		Servers:    map[uint32]*IpmiServer{},
		Lock:       sync.RWMutex{},
		NextHandle: 0x40000001,
	}
}

func SetupImpiManager() {
	GlobalIpmiManager = NewIpmiManager()
	CreateImpiServer("SceLncService", 0)
	CreateImpiServer("SceShellCoreUtil", 0)
	structs.CreateDefaultEventFlags([]string{
		fmt.Sprintf("SceShellCoreUtil%x", structs.GlobalAppInfo.AppId),
		"SceShellCoreUtilAppFocus",
		"SceShellCoreUtilCtrlFocus",
		"SceShellCoreUtilPowerControl",
	})
	CreateImpiServer("SceAppMessaging", 0)
	structs.CreateDefaultEventFlags([]string{
		fmt.Sprintf("SceAppMessaging%x", structs.GlobalAppInfo.AppId),
	})
	structs.CreateSemaphore(fmt.Sprintf("SceAppMessaging%x", structs.GlobalAppInfo.AppId), 0, 0, 255)
	npMgrIpc := CreateImpiServer("SceNpMgrIpc", 0)
	npMgrIpc.CreateEventFlag("SceNpMgrEvf")
	npService := CreateImpiServer("SceNpService", 0)
	npService.CreateEventFlag("SceNpServiceEvf")
	netCtl := CreateImpiServer("SceNetCtl", 0)
	netCtl.CreateEventFlag("SceNetCtlEvf")
	npTrophyIpc := CreateImpiServer("SceNpTrophyIpc", 0)
	npTrophyIpc.CreateEventFlag("SceNpTrophyEvf")
	CreateImpiServer("SceAppContent", 0)
	CreateImpiServer("SceMbusIpc", 0)
	CreateImpiServer("SceSysAudioSystemIpc", 0)
	structs.CreateDefaultEventFlags([]string{
		fmt.Sprintf("sceAudioOutMix%x", 1001),
	})
	avSetting := CreateImpiServer("SceAvSettingIpc", 0)
	avSetting.CreateEventFlag("SceAvSettingEvf")
	CreateImpiServer("SceSaveData", 0)
	CreateImpiServer("SceUserService", 0)
	CreateImpiServer("SceRemoteplayIpc", 0)
}

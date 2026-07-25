package pad

type PadDevice interface {
	Read(data *PadData)
	GetControllerInformation(info *PadControllerInformation)
}

type PadHandle struct {
	Id     uint32
	Device PadDevice
}

const (
	PadMaxTouchNum             = 2
	PadMaxDeviceUniqueDataSize = 12
)

const (
	PadButtonL3          = 0x2
	PadButtonR3          = 0x4
	PadButtonOptions     = 0x8
	PadButtonUp          = 0x10
	PadButtonRight       = 0x20
	PadButtonDown        = 0x40
	PadButtonLeft        = 0x80
	PadButtonL2          = 0x100
	PadButtonR2          = 0x200
	PadButtonL1          = 0x400
	PadButtonR1          = 0x800
	PadButtonTriangle    = 0x1000
	PadButtonCircle      = 0x2000
	PadButtonCross       = 0x4000
	PadButtonSquare      = 0x8000
	PadButtonTouchPad    = 0x100000
	PadButtonIntercepted = 0x80000000
)

type PadData struct {
	Buttons             uint32
	LeftStick           PadAnalogStick
	RightStick          PadAnalogStick
	AnalogButtons       PadAnalogButtons
	Orientation         FQuaternion
	Acceleration        FVector3
	AngularVelocity     FVector3
	TouchData           PadTouchData
	Connected           bool
	_                   [7]uint8
	Timestamp           uint64
	ExtensionUnitData   PadExtensionUnitData
	ConnectedCount      uint8
	Reserved            [2]uint8
	DeviceUniqueDataLen uint8
	DeviceUniqueData    [PadMaxDeviceUniqueDataSize]uint8
}

type PadAnalogStick struct {
	X uint8
	Y uint8
}

type PadAnalogButtons struct {
	L2 uint8
	R2 uint8
	_  [2]uint8
}

type FQuaternion struct {
	X, Y, Z, W float32
}

type FVector3 struct {
	X, Y, Z float32
}

type PadTouchData struct {
	TouchNum               uint8
	Reserved               [3]uint8
	TimeSinceTouchHeldDown uint32
	Touch                  [PadMaxTouchNum]PadTouch
}

type PadTouch struct {
	X        uint16
	Y        uint16
	Id       uint8
	Reserved [3]uint8
}

type PadExtensionUnitData struct {
	ExtensionUnitId uint32
	Reserved        [1]uint8
	DataLength      uint8
	Data            [10]uint8
}

type PadControllerInformation struct {
	TouchPadInfo   PadTouchPadInformation
	StickInfo      PadStickInformation
	ConnectionType uint8
	ConnectedCount uint8
	Connected      bool
	_              [1]uint8
	DeviceClass    uint32
	Reserved       [8]uint8
}

type PadTouchPadInformation struct {
	PixelDensity float32
	ResolutionX  uint16
	ResolutionY  uint16
}

type PadStickInformation struct {
	DeadZoneLeft  uint8
	DeadZoneRight uint8
}

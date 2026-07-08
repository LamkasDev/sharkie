package pad

type PadHandle struct {
	Id uint32
}

const (
	ORBIS_PAD_MAX_TOUCH_NUM               = 2
	ORBIS_PAD_MAX_DEVICE_UNIQUE_DATA_SIZE = 12
)

const (
	OrbisPadButtonL3          = 0x2
	OrbisPadButtonR3          = 0x4
	OrbisPadButtonOptions     = 0x8
	OrbisPadButtonUp          = 0x10
	OrbisPadButtonRight       = 0x20
	OrbisPadButtonDown        = 0x40
	OrbisPadButtonLeft        = 0x80
	OrbisPadButtonL2          = 0x100
	OrbisPadButtonR2          = 0x200
	OrbisPadButtonL1          = 0x400
	OrbisPadButtonR1          = 0x800
	OrbisPadButtonTriangle    = 0x1000
	OrbisPadButtonCircle      = 0x2000
	OrbisPadButtonCross       = 0x4000
	OrbisPadButtonSquare      = 0x8000
	OrbisPadButtonTouchPad    = 0x100000
	OrbisPadButtonIntercepted = 0x80000000
)

type OrbisPadAnalogStick struct {
	X uint8
	Y uint8
}

type OrbisPadAnalogButtons struct {
	L2 uint8
	R2 uint8
	_  [2]uint8
}

type OrbisFQuaternion struct {
	X, Y, Z, W float32
}

type OrbisFVector3 struct {
	X, Y, Z float32
}

type OrbisPadTouch struct {
	X        uint16
	Y        uint16
	Id       uint8
	Reserved [3]uint8
}

type OrbisPadTouchData struct {
	TouchNum               uint8
	Reserved               [3]uint8
	TimeSinceTouchHeldDown uint32
	Touch                  [ORBIS_PAD_MAX_TOUCH_NUM]OrbisPadTouch
}

type OrbisPadExtensionUnitData struct {
	ExtensionUnitId uint32
	Reserved        [1]uint8
	DataLength      uint8
	Data            [10]uint8
}

type OrbisPadData struct {
	Buttons             uint32
	LeftStick           OrbisPadAnalogStick
	RightStick          OrbisPadAnalogStick
	AnalogButtons       OrbisPadAnalogButtons
	Orientation         OrbisFQuaternion
	Acceleration        OrbisFVector3
	AngularVelocity     OrbisFVector3
	TouchData           OrbisPadTouchData
	Connected           bool
	_                   [7]uint8
	Timestamp           uint64
	ExtensionUnitData   OrbisPadExtensionUnitData
	ConnectedCount      uint8
	Reserved            [2]uint8
	DeviceUniqueDataLen uint8
	DeviceUniqueData    [ORBIS_PAD_MAX_DEVICE_UNIQUE_DATA_SIZE]uint8
}

type OrbisPadTouchPadInformation struct {
	PixelDensity float32
	ResolutionX  uint16
	ResolutionY  uint16
}

type OrbisPadStickInformation struct {
	DeadZoneLeft  uint8
	DeadZoneRight uint8
}

type OrbisPadControllerInformation struct {
	TouchPadInfo   OrbisPadTouchPadInformation
	StickInfo      OrbisPadStickInformation
	ConnectionType uint8
	ConnectedCount uint8
	Connected      bool
	_              [1]uint8
	DeviceClass    uint32
	Reserved       [8]uint8
}

package app_content

const (
	NpUnifiedEntitlementLabelSize    = 17
	AppContentEntitlementLabelOffset = 20
)

type AppContentAddcontDownloadStatus uint32

const (
	AppContentAddcontDownloadStatusNoExtraData       = AppContentAddcontDownloadStatus(0)
	AppContentAddcontDownloadStatusNoInQueue         = AppContentAddcontDownloadStatus(1)
	AppContentAddcontDownloadStatusDownloading       = AppContentAddcontDownloadStatus(2)
	AppContentAddcontDownloadStatusDownloadSuspended = AppContentAddcontDownloadStatus(3)
	AppContentAddcontDownloadStatusInstalled         = AppContentAddcontDownloadStatus(4)
)

type AppContentAppParamId int32

const (
	AppContentAppParamIdSkuFlag           = AppContentAppParamId(0)
	AppContentAppParamIdUserDefinedParam1 = AppContentAppParamId(1)
	AppContentAppParamIdUserDefinedParam2 = AppContentAppParamId(2)
	AppContentAppParamIdUserDefinedParam3 = AppContentAppParamId(3)
	AppContentAppParamIdUserDefinedParam4 = AppContentAppParamId(4)
	AppContentAppParamSkuFlagFull         = 3
)

type AppContentInitParam struct {
	Reserved [32]byte
}

type AppContentBootParam struct {
	Reserved1 [4]byte
	Attr      uint32
	Reserved2 [32]byte
}

type AppContentAddcontInfo struct {
	EntitlementLabel NpUnifiedEntitlementLabel
	Status           AppContentAddcontDownloadStatus
}

type NpUnifiedEntitlementLabel struct {
	Data    [NpUnifiedEntitlementLabelSize]byte
	Padding [3]byte
}

package app_content

type OrbisAppContentInitParam struct {
	Reserved [32]byte
}

type OrbisAppContentBootParam struct {
	Reserved1 [4]byte
	Attr      uint32
	Reserved2 [32]byte
}

const (
	OrbisNpUnifiedEntitlementLabelSize    = 17
	OrbisAppContentEntitlementLabelOffset = 20
)

type OrbisNpUnifiedEntitlementLabel struct {
	Data    [OrbisNpUnifiedEntitlementLabelSize]byte
	Padding [3]byte
}

type OrbisAppContentAddcontDownloadStatus uint32

const (
	OrbisAppContentAddcontDownloadStatusNoExtraData       OrbisAppContentAddcontDownloadStatus = 0
	OrbisAppContentAddcontDownloadStatusNoInQueue         OrbisAppContentAddcontDownloadStatus = 1
	OrbisAppContentAddcontDownloadStatusDownloading       OrbisAppContentAddcontDownloadStatus = 2
	OrbisAppContentAddcontDownloadStatusDownloadSuspended OrbisAppContentAddcontDownloadStatus = 3
	OrbisAppContentAddcontDownloadStatusInstalled         OrbisAppContentAddcontDownloadStatus = 4
)

type OrbisAppContentAddcontInfo struct {
	EntitlementLabel OrbisNpUnifiedEntitlementLabel
	Status           OrbisAppContentAddcontDownloadStatus
}

type OrbisAppContentAppParamId int32

const (
	OrbisAppContentAppParamIdSkuFlag           OrbisAppContentAppParamId = 0
	OrbisAppContentAppParamIdUserDefinedParam1 OrbisAppContentAppParamId = 1
	OrbisAppContentAppParamIdUserDefinedParam2 OrbisAppContentAppParamId = 2
	OrbisAppContentAppParamIdUserDefinedParam3 OrbisAppContentAppParamId = 3
	OrbisAppContentAppParamIdUserDefinedParam4 OrbisAppContentAppParamId = 4
)

const OrbisAppContentAppParamSkuFlagFull = 3

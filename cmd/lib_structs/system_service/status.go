package system_service

type SystemServiceStatus struct {
	EventNumber              int32
	IsSystemUiOverlaid       bool
	IsInBackgroundExecution  bool
	IsCpuMode7CpuNormal      bool
	IsGameLiveStreamingOnAir bool
	IsOutOfVrPlayArea        bool
	// Reserved.
}

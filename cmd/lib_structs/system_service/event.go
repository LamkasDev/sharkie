package system_service

type SystemServiceEventType int32

const (
	SystemServiceEventTypeInvalid = SystemServiceEventType(-1)
)

type SystemServiceEvent struct {
	Type SystemServiceEventType
	Data [8192]byte
}

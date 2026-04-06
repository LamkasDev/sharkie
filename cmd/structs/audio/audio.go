package audio

const AudioInBufferDefault = 4096
const AudioInBufferName = "/vmicDdShmAin"
const AudioInEventFlagName = "/vmicDdEvfAin"

const AudioVideoSettingsDefault = 4096
const AudioVideoSettingsName = "/SceAvSetting"

type AudioOutHandle struct {
	Id uint32
}

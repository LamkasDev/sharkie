package gpu

import (
	"hash/adler32"
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type ConstRamSnapshots map[uint32]LiverpoolConstRam

var GlobalConstRamSnapshots = ConstRamSnapshots{}

func (l *Liverpool) SnapshotConstRam() uint32 {
	constRam := l.DrawState.ConstRam
	constRamBytes := unsafe.Slice((*byte)(unsafe.Pointer(&constRam[0])), LiverpoolConstRamSize*4)
	hash := adler32.Checksum(constRamBytes)
	if _, ok := GlobalConstRamSnapshots[hash]; !ok {
		GlobalConstRamSnapshots[hash] = constRam
	}

	return hash
}

type UserDataSnapshots map[uint32]UserData

var GlobalUserDataSnapshots = UserDataSnapshots{}

func (l *Liverpool) SnapshotUserData() uint32 {
	var userData UserData
	copy(userData[UserDataOffsetVertex:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_VS_0:GREG_MM_SPI_SHADER_USER_DATA_VS_15+1])
	copy(userData[UserDataOffsetHull:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_HS_0:GREG_MM_SPI_SHADER_USER_DATA_HS_15+1])
	copy(userData[UserDataOffsetEvaluation:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_ES_0:GREG_MM_SPI_SHADER_USER_DATA_ES_15+1])
	copy(userData[UserDataOffsetGeometry:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_GS_0:GREG_MM_SPI_SHADER_USER_DATA_GS_15+1])
	copy(userData[UserDataOffsetFragment:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_PS_0:GREG_MM_SPI_SHADER_USER_DATA_PS_15+1])
	copy(userData[UserDataOffsetCompute:], l.Registers.Shader[GREG_MM_COMPUTE_USER_DATA_0:GREG_MM_COMPUTE_USER_DATA_15+1])
	userDataBytes := unsafe.Slice((*byte)(unsafe.Pointer(&userData[0])), len(userData)*4)
	hash := adler32.Checksum(userDataBytes)
	if _, ok := GlobalUserDataSnapshots[hash]; !ok {
		GlobalUserDataSnapshots[hash] = userData
	}

	return hash
}

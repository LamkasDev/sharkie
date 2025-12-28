package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000017F0
// __int64 __fastcall _sys_regmgr_call()
func libKernel___sys_regmgr_call(op, id, resultPtr, valuePtr, size uintptr) uintptr {
	if valuePtr == 0 {
		logger.Printf("%-120s %s failed due to invalid value pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
		)
		return EFAULT
	}

	switch op {
	case REGMGR_GET_INT, REGMGR_NONSYS_GET_INT:
		valueSlice := unsafe.Slice((*byte)(unsafe.Pointer(valuePtr)), 4)
		keyName, ok := RegistryNames[id]
		if !ok {
			keyName = fmt.Sprintf("UNKNOWN KEY 0x%X", id)
		}
		switch id {
		case REG_DEVENV_TOOL_sce_module_dbg, REG_DEVENV_TOOL_preload_chk_off,
			REG_DEVENV_TOOL_fake_xxx_mode:
			binary.LittleEndian.PutUint32(valueSlice, 1)
			break
		default:
			binary.LittleEndian.PutUint32(valueSlice, 0)
			break
		}

		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		logger.Printf("%-120s %s requested int for %s (resultPtr=%s, valuePtr=%s, size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
			color.Blue.Sprint(keyName),
			color.Yellow.Sprintf("0x%X", resultPtr),
			color.Yellow.Sprintf("0x%X", valuePtr),
			color.Green.Sprintf("%d", size),
		)
		return 0
	case REGMGR_GET_BIN, REGMGR_NONSYS_GET_BIN:
		valueSlice := unsafe.Slice((*byte)(unsafe.Pointer(valuePtr)), size)
		keyName, ok := RegistryNames[id]
		if !ok {
			keyName = fmt.Sprintf("UNKNOWN KEY 0x%X", id)
		}
		switch id {
		default:
			binary.LittleEndian.PutUint32(valueSlice, 0)
			break
		}

		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		logger.Printf("%-120s %s requested binary data for %s (resultPtr=%s, valuePtr=%s, size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
			color.Blue.Sprint(keyName),
			color.Yellow.Sprintf("0x%X", resultPtr),
			color.Yellow.Sprintf("0x%X", valuePtr),
			color.Green.Sprintf("%d", size),
		)
		return 0
	case REGMGR_GET_STRING, REGMGR_NONSYS_GET_STRING:
		if size < 1 {
			logger.Printf("%-120s %s failed due to invalid size %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__sys_regmgr_call"),
				color.Green.Sprintf("%d", size),
			)
			return EFAULT
		}

		keyName, ok := RegistryNames[id]
		if !ok {
			keyName = fmt.Sprintf("UNKNOWN KEY 0x%X", id)
		}
		switch id {
		default:
			WriteCString(valuePtr, "")
			break
		}

		if resultPtr != 0 {
			resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
			binary.LittleEndian.PutUint32(resultSlice, 0)
		}

		logger.Printf("%-120s %s requested string for %s (resultPtr=%s, valuePtr=%s, size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
			color.Blue.Sprint(keyName),
			color.Yellow.Sprintf("0x%X", resultPtr),
			color.Yellow.Sprintf("0x%X", valuePtr),
			color.Green.Sprintf("%d", size),
		)
		return 0
	}

	logger.Printf("%-120s %s failed due to unknown operation %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_regmgr_call"),
		color.Green.Sprintf("%d", op),
	)
	return ENOENT
}

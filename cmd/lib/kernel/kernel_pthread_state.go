package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// 0x0000000000014560
// __int64 __fastcall scePthreadGetaffinity(signed __int32 *, _QWORD *)
func libKernel_scePthreadGetaffinity(threadPtr uintptr, mask *ThreadAffinityMask) uintptr {
	cpuSet := ThreadCpuSet{}
	err := posix.Pthread_getaffinity_np(threadPtr, ThreadCpuSetSize, &cpuSet)
	if err != 0 {
		return err - SonyErrorOffset
	}
	*mask = ThreadAffinityMask(cpuSet.Low)

	return 0
}

// 0x0000000000798B20
// __int64 scePthreadSetaffinity()
func libKernel_scePthreadSetaffinity(threadPtr uintptr, mask uint64) uintptr {
	cpuSet := ThreadCpuSet{
		Low: mask,
	}
	err := posix.Pthread_setaffinity_np(threadPtr, ThreadCpuSetSize, &cpuSet)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014210
// __int64 scePthreadGetschedparam()
func libKernel_scePthreadGetschedparam(attrHandlePtr *uintptr, schedulingParameterPtr *int32) uintptr {
	err := posix.Pthread_attr_getschedparam(attrHandlePtr, schedulingParameterPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000014420
// __int64 scePthreadSetschedparam()
func libKernel_scePthreadSetschedparam(attrHandlePtr *uintptr, schedulingParameterPtr *int32) uintptr {
	err := posix.Pthread_attr_setschedparam(attrHandlePtr, schedulingParameterPtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

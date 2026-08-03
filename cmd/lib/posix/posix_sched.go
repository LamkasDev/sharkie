package posix

import (
	"runtime"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
)

func libScePosix_sched_yield() uintptr {
	runtime.Gosched()
	return 0
}

func libScePosix_sched_get_priority_min(policy PthreadSchedulingPolicy) uintptr {
	if policy != PthreadSchedulingPolicyFifo && policy != PthreadSchedulingPolicyRoundRobin {
		return EINVAL
	}

	return 256
}

func libScePosix_sched_get_priority_max(policy PthreadSchedulingPolicy) uintptr {
	if policy != PthreadSchedulingPolicyFifo && policy != PthreadSchedulingPolicyRoundRobin {
		return EINVAL
	}

	return 767
}

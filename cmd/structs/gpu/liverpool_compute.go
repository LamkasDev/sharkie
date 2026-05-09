package gpu

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// LiverpoolComputeDispatch is a snapshot of GPU state needed to issue a compute dispatch.
type LiverpoolComputeDispatch struct {
	// Compute parameters.
	DimX, DimY, DimZ uint32

	// Compute shader program
	ComputeShPgmLo, ComputeShPgmHi uint32
	ComputeShRsrc1, ComputeShRsrc2 uint32
	ComputeShader                  *GcnShader

	// Snapshots of register states.
	UserDataHash uint32
}

// CsGpuAddress returns the full compute shader GPU address.
func (d *LiverpoolComputeDispatch) CsGpuAddress() uintptr {
	return (uintptr(d.ComputeShPgmLo) | uintptr(d.ComputeShPgmHi)<<32) << 8
}

// NewComputeDispatch captures the current register state into a LiverpoolComputeDispatch.
func (l *Liverpool) NewComputeDispatch(dimX, dimY, dimZ uint32) LiverpoolComputeDispatch {
	l.StateMutex.Lock()
	computeDispatch := LiverpoolComputeDispatch{
		DimX: dimX, DimY: dimY, DimZ: dimZ,

		ComputeShPgmLo: l.Registers.Shader[GREG_MM_COMPUTE_PGM_LO],
		ComputeShPgmHi: l.Registers.Shader[GREG_MM_COMPUTE_PGM_HI],
		ComputeShRsrc1: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC1],
		ComputeShRsrc2: l.Registers.Shader[GREG_MM_COMPUTE_PGM_RSRC2],

		UserDataHash: l.SnapshotUserData(),
	}
	l.StateMutex.Unlock()

	computeDispatch.ComputeShader = l.GetShader(GcnShaderStageCompute, computeDispatch.CsGpuAddress())

	return computeDispatch
}

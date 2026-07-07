package gpu

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/cespare/xxhash"
)

type LiverpoolCommandType int

const (
	LiverpoolCommandTypeDraw LiverpoolCommandType = iota
	LiverpoolCommandTypeDispatch
	LiverpoolCommandTypeDmaCopy
	LiverpoolCommandTypeBindPipeline
	LiverpoolCommandTypeSetDynamicState
)

type LiverpoolCommand struct {
	Type  LiverpoolCommandType
	Index uint32
}

type LiverpoolCommandStream struct {
	Commands      []LiverpoolCommand
	Pipelines     []LiverpoolBindPipeline
	DynamicStates []LiverpoolSetDynamicState
	Draws         []LiverpoolDraw
	Dispatches    []LiverpoolDispatch
	DmaCopies     []LiverpoolDmaCopy

	PipelinesMap     map[uint64]uint32
	DynamicStatesMap map[uint64]uint32
}

func (s *LiverpoolCommandStream) Reset() {
	s.Commands = s.Commands[:0]
	s.Pipelines = s.Pipelines[:0]
	s.DynamicStates = s.DynamicStates[:0]
	s.Draws = s.Draws[:0]
	s.Dispatches = s.Dispatches[:0]
	s.DmaCopies = s.DmaCopies[:0]

	clear(s.PipelinesMap)
	clear(s.DynamicStatesMap)
}

type LiverpoolBindPipeline struct {
	// Pointers to parsed shader programs.
	VertexShader   *GcnShader
	HullShader     *GcnShader
	EvalShader     *GcnShader
	GeometryShader *GcnShader
	PixelShader    *GcnShader

	LiverpoolBindPipelineInternal
}

func (z *LiverpoolBindPipeline) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolSetDynamicState struct {
	LiverpoolSetDynamicStateInternal
}

func (z *LiverpoolSetDynamicState) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolDraw struct {
	LiverpoolDrawInternal
}

func (z *LiverpoolDraw) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolDispatch struct {
	// Pointers to parsed shader programs.
	ComputeShader *GcnShader

	LiverpoolDispatchInternal
}

func (z *LiverpoolDispatch) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

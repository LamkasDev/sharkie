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
	LiverpoolCommandTypeBindComputePipeline
	LiverpoolCommandTypeBindResources
	LiverpoolCommandTypeSetDynamicState
	LiverpoolCommandTypeWriteData
	LiverpoolCommandTypeWaitRegMemory
	LiverpoolCommandTypeFlip
)

type LiverpoolCommand struct {
	Type  LiverpoolCommandType
	Index uint32
}

type LiverpoolCommandStream struct {
	Name         string
	CommandIndex int

	Commands         []LiverpoolCommand
	Pipelines        []LiverpoolBindPipeline
	ComputePipelines []LiverpoolBindComputePipeline
	BindResources    []LiverpoolBindResources
	DynamicStates    []LiverpoolSetDynamicState
	Draws            []LiverpoolDraw
	Dispatches       []LiverpoolDispatch
	DmaCopies        []LiverpoolDmaCopy
	WriteDatas       []LiverpoolWriteData
	WaitRegMems      []LiverpoolWaitRegMemory

	PipelinesMap        map[uint64]uint32
	ComputePipelinesMap map[uint64]uint32
	BindResourcesMap    map[uint64]uint32
	DynamicStatesMap    map[uint64]uint32
	DrawsMap            map[uint64]uint32
	DispatchesMap       map[uint64]uint32
	DmaCopiesMap        map[uint64]uint32
	WriteDatasMap       map[uint64]uint32
	WaitRegMemsMap      map[uint64]uint32
}

func NewLiverpoolCommandStream(name string) *LiverpoolCommandStream {
	return &LiverpoolCommandStream{
		Name: name,

		Commands:         []LiverpoolCommand{},
		Pipelines:        []LiverpoolBindPipeline{},
		ComputePipelines: []LiverpoolBindComputePipeline{},
		BindResources:    []LiverpoolBindResources{},
		DynamicStates:    []LiverpoolSetDynamicState{},
		Draws:            []LiverpoolDraw{},
		Dispatches:       []LiverpoolDispatch{},
		DmaCopies:        []LiverpoolDmaCopy{},
		WriteDatas:       []LiverpoolWriteData{},
		WaitRegMems:      []LiverpoolWaitRegMemory{},

		PipelinesMap:        map[uint64]uint32{},
		ComputePipelinesMap: map[uint64]uint32{},
		BindResourcesMap:    map[uint64]uint32{},
		DynamicStatesMap:    map[uint64]uint32{},
		DrawsMap:            map[uint64]uint32{},
		DispatchesMap:       map[uint64]uint32{},
		DmaCopiesMap:        map[uint64]uint32{},
		WriteDatasMap:       map[uint64]uint32{},
		WaitRegMemsMap:      map[uint64]uint32{},
	}
}

func (s *LiverpoolCommandStream) Reset() {
	s.Commands = s.Commands[:0]
	s.Pipelines = s.Pipelines[:0]
	s.ComputePipelines = s.ComputePipelines[:0]
	s.BindResources = s.BindResources[:0]
	s.DynamicStates = s.DynamicStates[:0]
	s.Draws = s.Draws[:0]
	s.Dispatches = s.Dispatches[:0]
	s.DmaCopies = s.DmaCopies[:0]
	s.WriteDatas = s.WriteDatas[:0]
	s.WaitRegMems = s.WaitRegMems[:0]

	clear(s.PipelinesMap)
	clear(s.ComputePipelinesMap)
	clear(s.BindResourcesMap)
	clear(s.DynamicStatesMap)
}

type LiverpoolBindResources struct {
	// Pointers to parsed shader programs.
	VertexShader   *GcnShader
	FragmentShader *GcnShader
	HullShader     *GcnShader
	EvalShader     *GcnShader
	GeometryShader *GcnShader
	ComputeShader  *GcnShader

	LiverpoolBindResourcesInternal
}

func (z *LiverpoolBindResources) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolBindComputePipeline struct {
	LiverpoolBindComputePipelineInternal
}

func (z *LiverpoolBindComputePipeline) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolBindPipeline struct {
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
	LiverpoolDispatchInternal
}

func (z *LiverpoolDispatch) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolDmaCopy struct {
	LiverpoolDmaCopyInternal
}

func (z *LiverpoolDmaCopy) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolWaitRegMemory struct {
	LiverpoolWaitRegMemoryInternal
}

func (z *LiverpoolWaitRegMemory) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

type LiverpoolWriteData struct {
	LiverpoolWriteDataInternal
}

func (z *LiverpoolWriteData) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

// Package gpu contains structs to emulate the AMD Liverpool GPU.
package gpu

import (
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

var GlobalLiverpool *Liverpool

// Liverpool keeps state of the AMD Liverpool GPU.
type Liverpool struct {
	FrameNumber uint64

	StateMutex sync.Mutex
	Registers  LiverpoolRegisters
	DrawState  LiverpoolDrawState

	ShadersMutex  sync.Mutex
	LoadedShaders map[uintptr]*GcnShader

	DisplaySurfaces map[uintptr]*LiverpoolDisplaySurface
	PM4Handlers     map[uint8]PM4Handler

	OnFlip                   func(flip *VideoOutFlip)
	OnRingWork               func(ringWork *RingWork)
	OnRegisterDisplaySurface func(address uintptr, attribute *VideoOutBufferAttribute)
	WaitOnFence              func()
}

type RingWork struct {
	GraphicsRing *LiverpoolCommandRing
	ComputeRing  *LiverpoolCommandRing
}

func NewLiverpool() *Liverpool {
	l := &Liverpool{
		StateMutex: sync.Mutex{},

		ShadersMutex:  sync.Mutex{},
		LoadedShaders: map[uintptr]*GcnShader{},

		DisplaySurfaces: map[uintptr]*LiverpoolDisplaySurface{},
		PM4Handlers:     map[uint8]PM4Handler{},
	}
	l.SetupPM4Handlers()

	return l
}

func (l *Liverpool) RegisterDisplaySurface(address uintptr, attribute *VideoOutBufferAttribute, attributeIndex uint32) {
	l.DisplaySurfaces[address] = &LiverpoolDisplaySurface{
		GpuAddress:     address,
		PixelFormat:    attribute.PixelFormat,
		TilingMode:     attribute.TilingMode,
		Width:          attribute.Width,
		Height:         attribute.Height,
		PitchPixels:    attribute.PitchInPixel,
		AttributeIndex: attributeIndex,
	}
	if l.OnRegisterDisplaySurface != nil {
		l.OnRegisterDisplaySurface(address, attribute)
	}
}

func (l *Liverpool) SubmitCommandBuffers(indirectBuffers []PM4IndirectBuffer) {
	ringWork := &RingWork{
		GraphicsRing: &LiverpoolCommandRing{},
		ComputeRing:  &LiverpoolCommandRing{},
	}
	for _, indirectBuffer := range indirectBuffers {
		opcode := (indirectBuffer.Header >> 8) & 0xFF
		switch opcode {
		case PM4_IT_INDIRECT_BUFFER:
			ringWork.GraphicsRing.Pending = append(ringWork.GraphicsRing.Pending, indirectBuffer)
		case PM4_IT_INDIRECT_BUFFER_CNST:
			ringWork.ComputeRing.Pending = append(ringWork.ComputeRing.Pending, indirectBuffer)
		default:
			continue
		}
	}
	l.OnRingWork(ringWork)
}

func (l *Liverpool) GetShader(stage GcnShaderStage, address uintptr) *GcnShader {
	// Get already loaded shader.
	l.ShadersMutex.Lock()
	shader, ok := l.LoadedShaders[address]
	l.ShadersMutex.Unlock()
	if ok {
		return shader
	}

	// Load the shader.
	l.ShadersMutex.Lock()
	shader, err := NewGcnShader(stage, address)
	if err != nil {
		panic(err)
	}
	logger.Printf("[%s] Loaded %s shader %s of %s bytes...\n",
		color.Blue.Sprint("SHADER"),
		color.Blue.Sprint(stage),
		color.Yellow.Sprintf("0x%X", address),
		color.Green.Sprint(shader.DwordLength*4),
	)
	if err = l.DumpShaderOnce(shader); err != nil {
		panic(err)
	}
	l.LoadedShaders[address] = shader
	l.ShadersMutex.Unlock()

	return shader
}

func SetupLiverpool() {
	GlobalLiverpool = NewLiverpool()
}

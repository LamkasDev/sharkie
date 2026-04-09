// Package gpu contains structs to emulate the AMD Liverpool GPU.
package gpu

import (
	"sync"

	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
	"github.com/gookit/color"
)

var GlobalLiverpool *Liverpool

// Liverpool keeps state of the AMD Liverpool GPU.
type Liverpool struct {
	RingMutex    sync.Mutex
	GraphicsRing *LiverpoolCommandRing
	ComputeRing  *LiverpoolCommandRing

	StateMutex       sync.Mutex
	Registers        LiverpoolRegisters
	DrawState        LiverpoolDrawState
	PendingDrawCalls []LiverpoolDrawCall
	ConstRam         [LiverpoolConstRamSize]uint32

	ShadersMutex  sync.Mutex
	LoadedShaders map[uintptr]*GcnShader

	DisplaySurfaces map[uintptr]*LiverpoolDisplaySurface
	PM4Handlers     map[uint8]PM4Handler

	OnFlip                   func(gpuAddress uintptr, flipArg uint64)
	OnRegisterDisplaySurface func(address uintptr, attribute *VideoOutBufferAttribute)
}

func NewLiverpool() *Liverpool {
	l := &Liverpool{
		RingMutex:    sync.Mutex{},
		GraphicsRing: &LiverpoolCommandRing{},
		ComputeRing:  &LiverpoolCommandRing{},

		StateMutex: sync.Mutex{},

		LoadedShaders: map[uintptr]*GcnShader{},
		ShadersMutex:  sync.Mutex{},

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
	l.RingMutex.Lock()
	defer l.RingMutex.Unlock()
	for _, indirectBuffer := range indirectBuffers {
		opcode := (indirectBuffer.Header >> 8) & 0xFF
		switch opcode {
		case PM4_IT_INDIRECT_BUFFER:
			l.GraphicsRing.Pending = append(l.GraphicsRing.Pending, indirectBuffer)
		case PM4_IT_INDIRECT_BUFFER_CNST:
			l.ComputeRing.Pending = append(l.ComputeRing.Pending, indirectBuffer)
		}
	}
}

func (l *Liverpool) FlushDrawCalls() []LiverpoolDrawCall {
	l.StateMutex.Lock()
	drawCalls := l.PendingDrawCalls
	l.PendingDrawCalls = l.PendingDrawCalls[:0]
	l.StateMutex.Unlock()

	return drawCalls
}

func (l *Liverpool) Flip(gpuAddress uintptr, flipArg uint64) {
	if l.OnFlip != nil {
		l.OnFlip(gpuAddress, flipArg)
	}
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

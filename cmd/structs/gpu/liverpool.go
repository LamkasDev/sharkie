// Package gpu contains structs to emulate the AMD Liverpool GPU.
package gpu

import . "github.com/LamkasDev/sharkie/cmd/structs/video"

var GlobalLiverpool *Liverpool

// Liverpool keeps state of the AMD Liverpool GPU.
type Liverpool struct {
	GraphicsRing    *LiverpoolCommandRing
	ComputeRing     *LiverpoolCommandRing
	DisplaySurfaces map[uintptr]*LiverpoolDisplaySurface

	OnFlip                   func(gpuAddress uintptr, flipArg uint64)
	OnRegisterDisplaySurface func(address uintptr, attribute *VideoOutBufferAttribute)
}

func NewLiverpool() *Liverpool {
	return &Liverpool{
		GraphicsRing:    &LiverpoolCommandRing{},
		ComputeRing:     &LiverpoolCommandRing{},
		DisplaySurfaces: map[uintptr]*LiverpoolDisplaySurface{},
	}
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

func (l *Liverpool) Flip(gpuAddress uintptr, flipArg uint64) {
	if l.OnFlip != nil {
		l.OnFlip(gpuAddress, flipArg)
	}
}

func SetupLiverpool() {
	GlobalLiverpool = NewLiverpool()
}

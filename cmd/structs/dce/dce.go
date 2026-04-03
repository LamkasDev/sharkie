// Package dce contains structs to emulate the Display Core Engine (/dev/dce device).
package dce

import (
	"errors"
	"io/fs"

	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
)

var GlobalDisplayCoreEngine *DisplayCoreEngine

// DisplayCoreEngine keeps state of the /dev/dce device.
type DisplayCoreEngine struct {
	Handles [VideoOutMaxHandles]VideoOutHandle
}

func NewDisplayCoreEngine() *DisplayCoreEngine {
	dce := &DisplayCoreEngine{}
	for i := range dce.Handles {
		dce.Handles[i].Id = i + 1
		dce.Handles[i].LabelBufferAddress = GlobalGoAllocator.Malloc(uintptr(VideoOutMaxBuffers) * 8)
	}

	return dce
}

func (dce *DisplayCoreEngine) Read(b []byte) (int, error) {
	return 0, errors.New("dce read not implemented")
}

func (dce *DisplayCoreEngine) Write(b []byte) (int, error) {
	return 0, errors.New("dce write not implemented")
}

func (dce *DisplayCoreEngine) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("dce seek not implemented")
}

func (dce *DisplayCoreEngine) Close() error {
	return nil
}

func (dce *DisplayCoreEngine) Stat() (fs.FileInfo, error) {
	return nil, errors.New("dce stat not implemented")
}

func (dce *DisplayCoreEngine) Truncate(size int64) error {
	return errors.New("dce truncate not implemented")
}

func (dce *DisplayCoreEngine) Ioctl(request uint32, argPtr uintptr) error {
	// TODO: this shouldn't get called, we handle video ourselves now (maybe remove it completely?).
	return errors.New("unknown dce ioctl")
}

func (dce *DisplayCoreEngine) GetHandleById(id int) *VideoOutHandle {
	if id < 1 || id > VideoOutMaxHandles {
		return nil
	}

	return &dce.Handles[id-1]
}

func SetupDisplayCoreEngine() {
	GlobalDisplayCoreEngine = NewDisplayCoreEngine()
}

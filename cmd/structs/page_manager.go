package structs

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

// PageProtState mirrors the C SIGSEGV handler's prot_state values.
const (
	PageProtNone = 0 // PROT_NONE - read faults (GPU owns)
	PageProtRead = 1 // PROT_READ - write faults (CPU may read)
	PageProtRW   = 2 // PROT_READ|PROT_WRITE - no traps
)

type pageState struct {
	readWatchers  uint8
	writeWatchers uint8
	protState     int
}

// PageManager batches mprotect calls based on refcounted read/write watchers.
type PageManager struct {
	pages map[uintptr]*pageState
	mu    sync.Mutex
}

func newPageManager() *PageManager {
	return &PageManager{pages: make(map[uintptr]*pageState)}
}

func pagePermissions(readWatchers, writeWatchers uint8) int {
	var prot int
	if readWatchers == 0 {
		prot |= unix.PROT_READ
	}
	if writeWatchers == 0 {
		prot |= unix.PROT_WRITE
	}
	return prot
}

func protStateFromPermissions(prot int) int {
	switch {
	case prot&(unix.PROT_READ|unix.PROT_WRITE) == unix.PROT_READ|unix.PROT_WRITE:
		return PageProtRW
	case prot&unix.PROT_READ != 0:
		return PageProtRead
	default:
		return PageProtNone
	}
}

func (pm *PageManager) stateForPage(page uintptr) *pageState {
	st, ok := pm.pages[page]
	if !ok {
		st = &pageState{}
		pm.pages[page] = st
	}
	return st
}

func (pm *PageManager) applyProtection(page uintptr, st *pageState) {
	prot := pagePermissions(st.readWatchers, st.writeWatchers)
	newState := protStateFromPermissions(prot)
	if st.protState == newState {
		return
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(page)), SystemPageSize)
	_ = unix.Mprotect(b, prot)
	st.protState = newState
	if st.readWatchers == 0 && st.writeWatchers == 0 {
		cUntrackPage(page)
		delete(pm.pages, page)
		return
	}
	cTrackPage(page, st.protState)
}

// UpdatePageWatchers adds or removes read/write watchers across [addr, addr+size).
func (pm *PageManager) UpdatePageWatchers(addr, size uintptr, track, isRead bool) {
	if addr == 0 || size == 0 {
		return
	}

	start := addr & ^(SystemPageSize - 1)
	end := addr + size
	pageEnd := (end + SystemPageSize - 1) & ^(SystemPageSize - 1)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for page := start; page < pageEnd; page += SystemPageSize {
		st := pm.stateForPage(page)
		if isRead {
			if track {
				st.readWatchers = 1
			} else {
				st.readWatchers = 0
			}
		} else if track {
			if st.writeWatchers < 127 {
				st.writeWatchers++
			}
		} else if st.writeWatchers > 0 {
			st.writeWatchers--
		}
		pm.applyProtection(page, st)
	}
}

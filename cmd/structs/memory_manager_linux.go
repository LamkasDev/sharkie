//go:build linux

package structs

/*
#include <signal.h>
#include <string.h>
#include <stdint.h>
#include <sys/mman.h>
#include <stdio.h>
#include <pthread.h>

#define MAX_TRACKED_PAGES 262144

typedef struct {
    uintptr_t addr;
    int is_dirty;
    int prot_state; // 0 = PROT_NONE, 1 = PROT_READ, 2 = PROT_READ|PROT_WRITE
} TrackedPage;

static TrackedPage tracked_pages[MAX_TRACKED_PAGES];
static int num_tracked_pages = 0;

static pthread_mutex_t sync_mutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t sync_cond = PTHREAD_COND_INITIALIZER;
static pthread_cond_t worker_cond = PTHREAD_COND_INITIALIZER;
static uintptr_t pending_sync_addr = 0;
static int pending_sync_is_write = 0;
static int sync_done = 0;

static struct sigaction old_segv_sa;

static void segv_handler(int sig, siginfo_t* info, void* ctx) {
    uintptr_t fault_addr = (uintptr_t)info->si_addr;
    uintptr_t aligned_addr = fault_addr & ~(4096 - 1);

    int handled = 0;
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == aligned_addr) {
            if (tracked_pages[i].prot_state == 1) {
                // If it was PROT_READ, this must be a CPU write!
                // We do NOT need to sync GPU -> CPU, we just mark it dirty and unprotect it!
                mprotect((void*)aligned_addr, 4096, PROT_READ | PROT_WRITE);
                tracked_pages[i].is_dirty = 1;
                tracked_pages[i].prot_state = 2; // PROT_READ | PROT_WRITE
                handled = 1;
                break;
            }

            // Otherwise, it was PROT_NONE, so we must fetch from GPU.
            // Signal Go worker to perform Vulkan sync
            pthread_mutex_lock(&sync_mutex);
            pending_sync_addr = aligned_addr;
            pending_sync_is_write = 0;
            sync_done = 0;
            pthread_cond_signal(&worker_cond);

            while (!sync_done) {
                pthread_cond_wait(&sync_cond, &sync_mutex);
            }
            pthread_mutex_unlock(&sync_mutex);

            // Unprotect the page to PROT_READ so we can catch future writes!
            mprotect((void*)aligned_addr, 4096, PROT_READ);
            tracked_pages[i].prot_state = 1;
            handled = 1;
            break;
        }
    }

    if (handled) {
        return; // Handled!
    }

    // Not handled, forward to Go's original handler
    if (old_segv_sa.sa_sigaction != NULL) {
        old_segv_sa.sa_sigaction(sig, info, ctx);
    } else if (old_segv_sa.sa_handler != NULL) {
        old_segv_sa.sa_handler(sig);
    }
}

static void install_segv_handler() {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_sigaction = segv_handler;
    sa.sa_flags = SA_SIGINFO | SA_NODEFER | SA_ONSTACK;
    sigemptyset(&sa.sa_mask);
    sigaction(SIGSEGV, &sa, &old_segv_sa);
}

static void c_track_page(uintptr_t addr) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == addr) {
            return;
        }
    }
    if (num_tracked_pages < MAX_TRACKED_PAGES) {
        tracked_pages[num_tracked_pages].addr = addr;
        tracked_pages[num_tracked_pages].is_dirty = 0;
        tracked_pages[num_tracked_pages].prot_state = 0; // PROT_NONE
        num_tracked_pages++;
    }
}

static void c_untrack_page(uintptr_t addr) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == addr) {
            tracked_pages[i] = tracked_pages[num_tracked_pages - 1];
            num_tracked_pages--;
            return;
        }
    }
}

static uintptr_t c_wait_for_sync_request() {
    pthread_mutex_lock(&sync_mutex);
    while (pending_sync_addr == 0) {
        pthread_cond_wait(&worker_cond, &sync_mutex);
    }
    uintptr_t addr = pending_sync_addr;
    pthread_mutex_unlock(&sync_mutex);
    return addr;
}

static void c_complete_sync_request() {
    pthread_mutex_lock(&sync_mutex);
    pending_sync_addr = 0;
    sync_done = 1;
    pthread_cond_signal(&sync_cond);
    pthread_mutex_unlock(&sync_mutex);
}

static int c_is_page_dirty(uintptr_t addr) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == addr) {
            return tracked_pages[i].is_dirty;
        }
    }
    return 0;
}

static int c_is_region_dirty(uintptr_t addr, uintptr_t size) {
    uintptr_t aligned_addr = addr & ~(4096ULL - 1);
    uintptr_t end = addr + size;
    uintptr_t aligned_end = (end + 4095) & ~(4096ULL - 1);
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr >= aligned_addr && tracked_pages[i].addr < aligned_end) {
            if (tracked_pages[i].is_dirty) return 1;
        }
    }
    return 0;
}

static void c_clear_region_dirty(uintptr_t addr, uintptr_t size) {
    uintptr_t aligned_addr = addr & ~(4096ULL - 1);
    uintptr_t end = addr + size;
    uintptr_t aligned_end = (end + 4095) & ~(4096ULL - 1);
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr >= aligned_addr && tracked_pages[i].addr < aligned_end) {
            tracked_pages[i].is_dirty = 0;
            tracked_pages[i].prot_state = 1; // back to PROT_READ
            mprotect((void*)tracked_pages[i].addr, 4096, PROT_READ);
        }
    }
}

static void c_set_region_prot_state(uintptr_t addr, uintptr_t size, int state) {
    uintptr_t aligned_addr = addr & ~(4096ULL - 1);
    uintptr_t end = addr + size;
    uintptr_t aligned_end = (end + 4095) & ~(4096ULL - 1);
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr >= aligned_addr && tracked_pages[i].addr < aligned_end) {
            tracked_pages[i].prot_state = state;
        }
    }
}
*/
import "C"

func SetupMemoryManagerSignalHandler() {
	C.install_segv_handler()
}

func cTrackPage(addr uintptr) {
	C.c_track_page(C.uintptr_t(addr))
}

func cUntrackPage(addr uintptr) {
	C.c_untrack_page(C.uintptr_t(addr))
}

func WaitForSyncRequest() uintptr {
	return uintptr(C.c_wait_for_sync_request())
}

func CompleteSyncRequest() {
	C.c_complete_sync_request()
}

func IsPageDirty(addr uintptr) bool {
	return C.c_is_page_dirty(C.uintptr_t(addr)) != 0
}

func IsRegionDirty(addr uintptr, size uintptr) bool {
	return C.c_is_region_dirty(C.uintptr_t(addr), C.uintptr_t(size)) != 0
}

func ClearRegionDirty(addr uintptr, size uintptr) {
	C.c_clear_region_dirty(C.uintptr_t(addr), C.uintptr_t(size))
}

func SetRegionProtState(addr uintptr, size uintptr, state int) {
	C.c_set_region_prot_state(C.uintptr_t(addr), C.uintptr_t(size), C.int(state))
}

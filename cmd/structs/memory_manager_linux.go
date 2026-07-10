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

static int find_tracked_page(uintptr_t aligned_addr) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == aligned_addr) {
            return i;
        }
    }
    return -1;
}

static void segv_handler(int sig, siginfo_t* info, void* ctx) {
    uintptr_t fault_addr = (uintptr_t)info->si_addr;
    uintptr_t aligned_addr = fault_addr & ~(4096 - 1);

    int page_idx = find_tracked_page(aligned_addr);
    if (page_idx < 0) {
        if (old_segv_sa.sa_sigaction != NULL) {
            old_segv_sa.sa_sigaction(sig, info, ctx);
        } else if (old_segv_sa.sa_handler != NULL) {
            old_segv_sa.sa_handler(sig);
        }
        return;
    }

    int is_write = 0;
    if (tracked_pages[page_idx].prot_state == 1) {
        // PROT_READ: CPU write fault.
        is_write = 1;
    } else if (tracked_pages[page_idx].prot_state != 0) {
        // PROT_RW should not fault; forward.
        if (old_segv_sa.sa_sigaction != NULL) {
            old_segv_sa.sa_sigaction(sig, info, ctx);
        } else if (old_segv_sa.sa_handler != NULL) {
            old_segv_sa.sa_handler(sig);
        }
        return;
    }

    pthread_mutex_lock(&sync_mutex);
    pending_sync_addr = aligned_addr;
    pending_sync_is_write = is_write;
    sync_done = 0;
    pthread_cond_signal(&worker_cond);

    while (!sync_done) {
        pthread_cond_wait(&sync_cond, &sync_mutex);
    }
    pthread_mutex_unlock(&sync_mutex);

    // Re-lookup: the Go worker may have untracked this page (CPU invalidate).
    page_idx = find_tracked_page(aligned_addr);
    if (is_write || page_idx < 0) {
        mprotect((void*)aligned_addr, 4096, PROT_READ | PROT_WRITE);
        if (page_idx >= 0) {
            tracked_pages[page_idx].prot_state = 2;
        }
    } else {
        mprotect((void*)aligned_addr, 4096, PROT_READ);
        tracked_pages[page_idx].prot_state = 1;
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

static void c_track_page(uintptr_t addr, int prot_state) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == addr) {
            tracked_pages[i].prot_state = prot_state;
            return;
        }
    }
    if (num_tracked_pages < MAX_TRACKED_PAGES) {
        tracked_pages[num_tracked_pages].addr = addr;
        tracked_pages[num_tracked_pages].prot_state = prot_state;
        num_tracked_pages++;
    }
}

static void c_set_prot_state(uintptr_t addr, int prot_state) {
    for (int i = 0; i < num_tracked_pages; i++) {
        if (tracked_pages[i].addr == addr) {
            tracked_pages[i].prot_state = prot_state;
            return;
        }
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

static uintptr_t c_wait_for_sync_request(int* is_write_out) {
    pthread_mutex_lock(&sync_mutex);
    while (pending_sync_addr == 0) {
        pthread_cond_wait(&worker_cond, &sync_mutex);
    }
    uintptr_t addr = pending_sync_addr;
    if (is_write_out) {
        *is_write_out = pending_sync_is_write;
    }
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
*/
import "C"

func SetupMemoryManagerSignalHandler() {
	C.install_segv_handler()
}

func cTrackPage(addr uintptr, protState int) {
	C.c_track_page(C.uintptr_t(addr), C.int(protState))
}

func cSetProtState(addr uintptr, protState int) {
	C.c_set_prot_state(C.uintptr_t(addr), C.int(protState))
}

func cUntrackPage(addr uintptr) {
	C.c_untrack_page(C.uintptr_t(addr))
}

// SyncRequest is delivered by the SIGSEGV handler to the memory sync worker.
type SyncRequest struct {
	Addr    uintptr
	IsWrite bool
}

func WaitForSyncRequest() SyncRequest {
	var isWrite C.int
	addr := uintptr(C.c_wait_for_sync_request(&isWrite))
	return SyncRequest{Addr: addr, IsWrite: isWrite != 0}
}

func CompleteSyncRequest() {
	C.c_complete_sync_request()
}

package gpu

import (
	"testing"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
)

func TestSnapshotUserDataUniqueIDs(t *testing.T) {
	oldSnapshots := GlobalUserDataSnapshots
	oldDedup := userDataDedup
	oldNext := nextUserDataID
	t.Cleanup(func() {
		GlobalUserDataSnapshots = oldSnapshots
		userDataDedup = oldDedup
		nextUserDataID = oldNext
	})

	GlobalUserDataSnapshots = map[uint32]UserData{}
	userDataDedup = map[UserData]uint32{}
	nextUserDataID = 1

	l := &Liverpool{}
	id1 := l.SnapshotUserData()

	l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_VS_0] = 42
	id2 := l.SnapshotUserData()
	if id1 == id2 {
		t.Fatalf("expected different user data ids, both were %d", id1)
	}

	id3 := l.SnapshotUserData()
	if id2 != id3 {
		t.Fatalf("expected identical user data to reuse id %d, got %d", id2, id3)
	}
}

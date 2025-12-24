package structs

type FileStat struct {
	Device                uint32    // st_dev
	Inodes                uint32    // st_ino
	Mode                  uint16    // st_mode
	HardLinkCount         uint16    // st_nlink
	OwnerUser             uint32    // st_uid
	OwnerGroup            uint32    // st_gid
	SpecialDevice         uint32    // st_rdev
	AccessTime            Timestamp // st_atim
	ModifyTime            Timestamp // st_mtim
	ChangeStatusTime      Timestamp // st_ctim
	Size                  int64     // st_size
	Blocks                int64     // st_blocks
	BlockSize             uint32    // st_blksize
	Flags                 uint32    // st_flags
	GenerationNumber      uint32    // st_gen
	ImplementationDetails int32     // st_lspare
	CreateTime            Timestamp // st_birthtim
}

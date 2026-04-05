package structs

type FileFlags int32

const (
	O_RDONLY  = FileFlags(0x0000) /* open for reading only */
	O_WRONLY  = FileFlags(0x0001) /* open for writing only */
	O_RDWR    = FileFlags(0x0002) /* open for reading and writing */
	O_ACCMODE = FileFlags(0x0003) /* mask for above modes */

	FREAD      = FileFlags(0x0001)
	FWRITE     = FileFlags(0x0002)
	O_NONBLOCK = FileFlags(0x0004) /* no delay */
	O_APPEND   = FileFlags(0x0008) /* set append mode */
	O_SHLOCK   = FileFlags(0x0010) /* open with shared file lock */
	O_EXLOCK   = FileFlags(0x0020) /* open with exclusive file lock */
	O_ASYNC    = FileFlags(0x0040) /* signal pgrp when data ready */
	O_FSYNC    = FileFlags(0x0080) /* synchronous writes */
	O_SYNC     = O_FSYNC           /* POSIX synonym for O_FSYNC */
	O_NOFOLLOW = FileFlags(0x0100) /* don't follow symlinks */
	O_CREAT    = FileFlags(0x0200) /* create if nonexistent */
	O_TRUNC    = FileFlags(0x0400) /* truncate to zero length */
	O_EXCL     = FileFlags(0x0800) /* error if already exists */
	FHASLOCK   = FileFlags(0x4000) /* descriptor holds advisory lock */
	O_NOCTTY   = FileFlags(0x8000) /* don't assign controlling terminal */

	O_DIRECT          = FileFlags(0x00010000) /* attempt to bypass buffer cache */
	O_DIRECTORY       = FileFlags(0x00020000) /* fail if not directory */
	O_EXEC            = FileFlags(0x00040000) /* open for execute only */
	O_SEARCH          = O_EXEC
	FEXEC             = O_EXEC
	FSEARCH           = O_SEARCH
	O_TTY_INIT        = FileFlags(0x00080000) /* restore default termios attributes */
	O_CLOEXEC         = FileFlags(0x00100000)
	O_VERIFY          = FileFlags(0x00200000) /* open only after verification */
	O_PATH            = FileFlags(0x00400000) /* fd is only a path */
	O_RESOLVE_BENEATH = FileFlags(0x00800000) /* do not allow name resolution to walk out of cwd */
	O_DSYNC           = FileFlags(0x01000000) /* POSIX data sync */
	O_EMPTY_PATH      = FileFlags(0x02000000)
	O_NAMEDATTR       = FileFlags(0x04000000) /* NFSv4 named attributes */
	O_XATTR           = O_NAMEDATTR           /* Solaris compatibility */
	O_CLOFORK         = FileFlags(0x08000000)
)

type FileMode uint16

const (
	S_IRWXU = FileMode(0000700) /* RWX mask for owner */
	S_IRUSR = FileMode(0000400) /* R for owner */
	S_IWUSR = FileMode(0000200) /* W for owner */
	S_IXUSR = FileMode(0000100) /* X for owner */

	S_IREAD  = S_IRUSR
	S_IWRITE = S_IWUSR
	S_IEXEC  = S_IXUSR

	S_IRWXG = FileMode(0000070) /* RWX mask for group */
	S_IRGRP = FileMode(0000040) /* R for group */
	S_IWGRP = FileMode(0000020) /* W for group */
	S_IXGRP = FileMode(0000010) /* X for group */

	S_IRWXO = FileMode(0000007) /* RWX mask for other */
	S_IROTH = FileMode(0000004) /* R for other */
	S_IWOTH = FileMode(0000002) /* W for other */
	S_IXOTH = FileMode(0000001) /* X for other */
)

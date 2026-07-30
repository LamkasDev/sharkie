package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/net"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/net/file"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func libScePosix_socketpair(domain, sockType, protocol, svPtr uintptr) uintptr {
	if svPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid sv pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("socketpair"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	// Create socket pair.
	sock1 := &SocketPair{
		Name:     "socketpair-0",
		Domain:   int32(domain),
		Type:     int32(sockType),
		Protocol: int32(protocol),
	}
	sock2 := &SocketPair{
		Name:     "socketpair-1",
		Domain:   int32(domain),
		Type:     int32(sockType),
		Protocol: int32(protocol),
	}
	sock1.Peer = sock2
	sock2.Peer = sock1
	fd1 := GlobalFilesystem.AllocateFd(sock1.Name, sock1)
	fd2 := GlobalFilesystem.AllocateFd(sock2.Name, sock2)
	outArray := (*[2]int32)(unsafe.Pointer(svPtr))
	outArray[0] = int32(fd1)
	outArray[1] = int32(fd2)

	logger.Printf("%-132s %s created socket pair %s <-> %s (domain=%s, sockType=%s, protocol=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("socketpair"),
		color.Yellow.Sprintf("0x%X", fd1),
		color.Yellow.Sprintf("0x%X", fd2),
		color.Yellow.Sprintf("0x%X", domain),
		color.Yellow.Sprintf("0x%X", sockType),
		color.Yellow.Sprintf("0x%X", protocol),
	)
	return 0
}

func libScePosix_recvmsg(fd FileDescriptor, msg *Msghdr, flags int32) uintptr {
	if msg == nil {
		logger.Printf("%-132s %s failed due to invalid message pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("recvmsg"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	totalRead := int64(0)
	iovecs := unsafe.Slice((*Iovec)(unsafe.Pointer(msg.Iov)), msg.IovLen)
	for i := range msg.IovLen {
		iov := iovecs[i]
		if iov.Len == 0 {
			continue
		}
		buf := unsafe.Slice((*byte)(unsafe.Pointer(iov.Base)), iov.Len)
		n, err := GlobalFilesystem.ReadFd(fd, buf)
		if n > 0 {
			totalRead += int64(n)
		}
		if err != nil {
			if totalRead > 0 {
				break
			}
			emu.SetErrno(EINVAL)
			return ERR_PTR
		}
		if uint64(n) < iov.Len {
			break
		}
	}

	logger.Printf("%-132s %s read %s bytes from %s (flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("recvmsg"),
		color.Green.Sprint(totalRead),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return uintptr(totalRead)
}

func libScePosix_sendmsg(fd FileDescriptor, msg *Msghdr, flags int32) uintptr {
	if msg == nil {
		logger.Printf("%-132s %s failed due to invalid message pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sendmsg"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	totalWritten := int64(0)
	iovecs := unsafe.Slice((*Iovec)(unsafe.Pointer(msg.Iov)), msg.IovLen)
	for i := range msg.IovLen {
		iov := iovecs[i]
		if iov.Len == 0 {
			continue
		}
		buf := unsafe.Slice((*byte)(unsafe.Pointer(iov.Base)), iov.Len)
		n, err := GlobalFilesystem.WriteFd(fd, buf)
		if n > 0 {
			totalWritten += int64(n)
		}
		if err != nil {
			if totalWritten > 0 {
				break
			}
			emu.SetErrno(EINVAL)
			return ERR_PTR
		}
	}

	logger.Printf("%-132s %s wrote %s bytes to %s (flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_sendmsg"),
		color.Green.Sprint(totalWritten),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return uintptr(totalWritten)
}

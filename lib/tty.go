package lib

import (
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

const (
	WATTSON_BAUD = unix.B19200
)

// PrepareFd prepares the given file descriptor for talking to a Wattson monitoring device
// including acquiring an exclusive lock.
func PrepareFd(fd uintptr) error {
	err := syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return err // could not get exclusive lock
	}

	var adtio unix.Termios
	adtio.Cflag = 0
	adtio.Cflag |= uint32(unix.CLOCAL)  // Ignore modem control lines
	adtio.Cflag |= uint32(unix.CREAD)   // Enable Receiver
	adtio.Cflag |= uint32(unix.CS8)     // Character size 8 bits
	adtio.Cflag |= uint32(WATTSON_BAUD) // Baud rate
	adtio.Lflag = 0
	adtio.Lflag |= unix.NOFLSH
	adtio.Iflag = 0
	adtio.Iflag |= uint32(unix.IGNPAR)
	adtio.Iflag |= uint32(unix.IGNCR)
	adtio.Cc[unix.VTIME] = 10 // timer 1s

	p := uintptr(unsafe.Pointer(&adtio))
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(syscall.TCSETS), p)
	if errno != 0 {
		return error(errno)
	}
	return nil
}

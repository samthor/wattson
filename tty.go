package main

import (
	"golang.org/x/sys/unix"
	"log"
	"os"
	"syscall" // needed as TCSETS is not in unix
	"unsafe"
)

const (
	WATTSON_BAUD = unix.B19200
)

// openPath accepts a path to a USB serial device, returning an open *os.File
// corresponding to a connection.
func openPath(path string) (file *os.File, err error) {
	file, err = os.OpenFile(path, unix.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return
	}

	fd := file.Fd()
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

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&adtio)))
	if errno != 0 {
		file.Close()
		return nil, error(errno)
	}
	log.Printf("set ioctl (%+v) on %v", adtio, path)
	return file, nil
}

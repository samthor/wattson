package main

import (
	"github.com/samthor/wattson/lib"
	"golang.org/x/sys/unix"
	"os"
)

// openPath accepts a path to a USB serial device, returning an open *os.File
// corresponding to a connection.
func openPath(path string) (file *os.File, err error) {
	file, err = os.OpenFile(path, unix.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return
	}
	err = lib.PrepareFd(file.Fd())
	if err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

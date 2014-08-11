package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

type Serial struct {
	rw   *bufio.ReadWriter
	file *os.File
}

// NewSerial opens a new serial connection to the given device. Assumes \r\n
// as the line ending.
func NewSerial(device string) (serial *Serial, err error) {
	file, err := os.OpenFile(device, os.O_RDWR|syscall.O_NOCTTY, 0666)
	if err != nil {
		return nil, err
	}
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
	serial = &Serial{rw, file}
	serial.prime()

	return serial, nil
}

func (s *Serial) prime() {
	done := make(chan bool)
	tick := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			s.rw.WriteString("nowv\r\n")
			s.rw.Flush()
			select {
			case <-done:
				return
			case <-tick.C:
			}
		}
	}()

	_, _, err := s.rw.ReadLine()
	if err != nil {
		log.Println("error reading primer", err)
	}
	done <- true
}

// Do sends a line, reads a line.
func (s *Serial) Do(command string) (output string, err error) {
	/*	for {
		log.Println("peeking")
		ret, err := s.rw.Peek(2)
		log.Println("peeking done", err, ret)
		if err != nil || len(ret) == 0 {
			break // nodata
		}
		s.rw.Read(make([]byte, 32))
	}*/

	out := fmt.Sprintf("%s\r\n", command)
	_, err = s.rw.WriteString(out)
	if err != nil {
		return
	}
	s.rw.Flush()
	raw, _, err := s.rw.ReadLine()
	return string(raw), err
}

// Close shuts down the serial connection. Do not use this instance after
// this point.
func (s *Serial) Close() {
	s.file.Close()
}

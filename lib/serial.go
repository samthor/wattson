package lib

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"
)

const (
	sep = "\r\n"
)

type readLine struct {
	line string
	err  error
}

type Serial struct {
	rw    *bufio.ReadWriter
	lines chan readLine
}

// NewSerial creates a new line-buffered connection to the given ReadWriter.
// Assumes \r\n as the line ending.
func NewSerial(x io.ReadWriter) (serial *Serial, err error) {
	rw := bufio.NewReadWriter(bufio.NewReader(x), bufio.NewWriter(x))
	serial = &Serial{rw, make(chan readLine, 100)}

	go func() {
		for {
			raw, _, err := serial.rw.ReadLine()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("got err reading: %v, %v", raw, err)
			}
			line := string(raw)
			serial.lines <- readLine{line, err}
		}
	}()

	//	serial.prime()
	return serial, nil
}

func (s *Serial) prime() {
	done := make(chan bool)
	tick := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			log.Printf("sending nowv primer")
			s.rw.WriteString("nowv" + sep)
			s.rw.Flush()
			select {
			case <-done:
				return
			case <-tick.C:
			}
		}
	}()

	line := <-s.lines
	log.Printf("got resp from primer: %v", line)
	done <- true
}

// Do sends a line, reads a line.
func (s *Serial) Do(command string) (output string, err error) {
	out := fmt.Sprintf("%s%s", command, sep)
	_, err = s.rw.WriteString(out)
	if err != nil {
		return
	}
	s.rw.Flush()

	readLine := <-s.lines
	return readLine.line, readLine.err
}

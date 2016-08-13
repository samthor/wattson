package lib

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
func NewSerial(x io.ReadWriter) *Serial {
	rw := bufio.NewReadWriter(bufio.NewReader(x), bufio.NewWriter(x))
	serial := &Serial{rw, make(chan readLine, 100)}

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

	return serial
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

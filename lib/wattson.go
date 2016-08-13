package lib

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type WattsonBridge struct {
	Serial *Serial
	Silent bool
}

// BaseValue performs the given cmd, parsing its result as a value in the given
// base.
func (w *WattsonBridge) BaseValue(cmd rune, base int) int {
	raw := w.Do(cmd)

	if len(raw) == 0 || rune(raw[0]) != cmd {
		return -1
	}

	v, err := strconv.ParseInt(raw[1:], base, 64)
	if err != nil {
		return -1
	}

	return int(v)
}

// HexValue perfomrs the given cmd, assuming the result is a hex value.
func (w *WattsonBridge) HexValue(cmd rune) int {
	return w.BaseValue(cmd, 16)
}

// DecValue performs the given cmd, assuming the result is a dec value.
func (w *WattsonBridge) DecValue(cmd rune) int {
	return w.BaseValue(cmd, 10)
}

// Series performs a command, treating the result as a comma-seperated list
// of hex values (e.g., 'l' => "l123,AF31,...").
func (w *WattsonBridge) Series(cmd rune, arg string) []int {
	raw := w.DoArg(cmd, arg)

	if len(raw) == 0 || rune(raw[0]) != cmd {
		return nil
	}

	out := make([]int, 0)
	parts := strings.Split(raw[1:], ",")
	for _, str := range parts {
		v, err := strconv.ParseInt(str, 16, 64)
		if err != nil {
			v = -1
		}
		out = append(out, int(v))
	}
	return out
}

// Do performs the specific cmd passed, with no arguments.
func (w *WattsonBridge) Do(cmd rune) string {
	return w.DoArg(cmd, "")
}

// DoArg performs the specific cmd passed, with the specified argument..
func (w *WattsonBridge) DoArg(cmd rune, arg string) string {
	out, err := w.Serial.Do(fmt.Sprintf("now%c%s", cmd, arg))

	if !w.Silent {
		var note string
		if arg != "" {
			note = fmt.Sprintf("%c.%s", cmd, arg)
		} else {
			note = fmt.Sprintf("%c", cmd)
		}
		if err != nil {
			log.Printf("err(%s): %s", note, err)
		}
		if out != "" {
			log.Printf("out(%s): %s", note, out)
		} else {
			log.Printf("out(%s)- NONE", note)
		}
	}

	return out
}

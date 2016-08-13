package main

import (
	"flag"
	"fmt"
	"github.com/samthor/wattson/lib"
	"log"
	"time"
)

const (
	iterations = 5
	reads      = 3
)

var (
	devicePath      = flag.String("device", "/dev/ttyUSB0", "serial device path")
	generationLimit = flag.Int("generation_limit", 3000, "if generation is above this, discard values")
	usageLimit      = flag.Int("usage_limit", 10000, "if usage is above this, discard values")
	latitude        = flag.Int("latitude", -30, "rough latitude")
)

func main() {
	flag.Parse()
	now := time.Now()

	file, err := openPath(*devicePath)
	if err != nil {
		log.Fatalln("could not open path", err)
	}
	defer file.Close()

	// Connect and send an initial dummy command (flush null bytes etc).
	bridge := lib.New(file)
	bridge.Do('v')

	for i := 0; i < iterations; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		use, gen := safeReadValues(bridge)
		if *generationLimit > 0 && gen > *generationLimit {
			continue
		}
		if *usageLimit > 0 && use > *usageLimit {
			log.Printf("got extreme usage value: (%v) %v", now, use)
			continue
		}
		fmt.Printf("%d,%d\n", use, gen)
		break
	}
}

func safeReadValues(bridge *lib.WattsonBridge) (use, gen int) {
	for count := 0; count < reads; count++ {
		tu, tg := readValues(bridge)
		if tu != use || tg != gen {
			use, gen = tu, tg
			count = 0
		}
	}
	return use, gen
}

func readValues(bridge *lib.WattsonBridge) (use, gen int) {
	factor := bridge.HexValue('n')
	if factor < 0 {
		log.Fatalf("invalid factor: %v", factor)
	}
	use = bridge.HexValue('p') * (factor + 1)
	gen = bridge.HexValue('w')
	return
}

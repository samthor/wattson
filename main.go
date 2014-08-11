package main

import (
	"flag"
	"fmt"
	"github.com/samthor/wattson/lib"
	"log"
	"time"
)

var (
	devicePath = flag.String("device", "/dev/ttyUSB0", "serial device path")
	blockLimit = flag.Int("block_limit", 100, "number of 5-minute blocks to get")

	wattsonVolts = flag.Int("wattson_volts", 230, "assumed voltage wattson is built for")
	volts        = flag.Int("volts", 0, "local voltage override")

	// TODO: this may also be available in the result to "nown" (+1).
	powerUseFactor = flag.Float64("use_factor", 1.0, "power use scale factor")
)

// DataBlock is a 5-minute block of data. It contains use and generation values.
type DataBlock struct {
	When time.Time
	Instant bool

	Use int // watts
	Gen int // watts

	AUse int // adjusted
	AGen int // adjusted
}

func main() {
	flag.Parse()

	serial, err := lib.NewSerial(*devicePath)
	if err != nil {
		log.Fatalln("could not connect", err)
	}
	log.Println("connected to", *devicePath)

	bridge := lib.WattsonBridge{Serial: serial}
	//wattDoArg(serial, 'T', now.Format("06/01/02 15:04:05"))
	//wattDo(serial, 'p')

	//for _, cmd := range "abcdfghijklmnpqrstuvwxyz" {
	//	bridge.Do(cmd)
	//}

	gen, use := FindLatestDay(&bridge)
	log.Printf("FOUND gen=%d, use=%d", gen, use)

	fmt.Printf("when,gen,agen,use,ause\n")

	// display instant use for funsies
	instant := DataBlock{
		When: time.Now(),
		Instant: true,
		Use: bridge.HexValue('p'),
		Gen: bridge.HexValue('w'),
	}
	instant.Adjust()

	// try to work out csv-ish data
	output := ReadData(&bridge)

	// output
	output = append([]DataBlock{instant}, output...)
	for _, block := range output {
		fmt.Printf("%v,%v,%v,%v,%v\n", block.When.Format(time.RFC3339), block.Gen, block.AGen, block.Use, block.AUse)
	}
}

func FindLatestDay(bridge *lib.WattsonBridge) (gen, use int) {
	now := time.Now()
	segment := (now.Hour() + 2) / 2
	for days := 1; days <= 50; days++ {
		format := "%02d%02d"
		useValues := bridge.Series('l', fmt.Sprintf(format, days, segment))
		genValues := bridge.Series('h', fmt.Sprintf(format, days, segment))

		for _, v := range useValues {
			if v == 65534 {
				log.Printf("found use: day=%d", days)
				use = days
			}
		}
		for _, v := range genValues {
			if v == 65534 {
				log.Printf("found gen: days=%d", days)
				gen = days
			}
		}

		if gen != 0 && use != 0 {
			return gen, use
		}
	}
	return gen, use
}

func ReadData(bridge *lib.WattsonBridge) (output []DataBlock) {
	days := bridge.HexValue('d') + 1
	if days <= 0 {
		panic("wrong number of days")
	}
	now := time.Now()

	// Generation day starts at 1, then goes to (q-response + 3) before counting down. This will
	// probably be never < than the number of use days.
	genday := bridge.HexValue('x') + 1

	// To the Wattson, we're currently at about days/segment. Grab these power readings and then walk
	// backwards until there is enough records.
	// NOTE: If the wattson is +day from device, this might get confused/wrong.
	segment := (now.Hour() + 2) / 2
	for days > 0 {
		for segment > 0 {
			if len(output) >= *blockLimit {
				return output
			}
			format := "%02d%02d"
			useValues := bridge.Series('l', fmt.Sprintf(format, days, segment))
			genValues := bridge.Series('h', fmt.Sprintf(format, genday, segment))

			if len(useValues) != 25 || len(genValues) != 25 {
				panic("did not get 25 values")
			}
			for j := 23; j >= 0; j-- {
				use := useValues[j]
				if use >= 65534 {
					if genValues[j] != use {
						panic("future values should match gen/use")
					}
					continue // this block isn't ready yet, probably in future
				}
				gen := genValues[j]

				hour, minute := convertSegmentRowHourMinute(segment, j)
				when := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
				block := DataBlock{
					When: when,
					Use:  use,
					Gen:  gen,
				}
				block.Adjust()
				output = append(output, block)
				if len(output) >= *blockLimit {
					return output
				}
			}

			segment--
		}

		segment = 12
		days--
		genday--

		if genday == 0 || days == 0 {
			log.Printf("no more data, genday=%d days=%d", genday, days)
			return output
		}
		now = now.AddDate(0, 0, -1)
	}
	return output
}

// Adjust updates this DataBlock to have adjusted values based on user flags.
func (db *DataBlock) Adjust() {
	use := *powerUseFactor * float64(db.Use)
	gen := float64(db.Gen)

	if *volts != 0 {
		vratio := float64(*volts) / float64(*wattsonVolts)
		use *= vratio
		gen *= vratio
	}

	db.AUse, db.AGen = int(use), int(gen)
}

func convertSegmentRowHourMinute(segment, row int) (hour, minute int) {
	hour = (segment - 1) * 2
	if row >= 12 {
		hour++
	}
	minute = 5 * (row % 12)
	return hour, minute
}

package main

import (
	"time"
	"math"
)

// solarHigh performs a quick-and-dirty calculation to guess the solar duration.
func solarHigh(now time.Time, lat int) (hours int) {
	alat := lat
	if alat < 0 {
		alat = -lat
	}

	yd := (now.YearDay() + 10) % 365
	ratio := math.Sin((float64(yd) / 365) * math.Pi) * 2 - 1

	hours = 14
	hours -= (alat / 20)
	hours -= int(ratio * 4)
	return hours
}



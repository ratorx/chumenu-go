package main

import "fmt"

type hourMinute struct {
	Hour   uint8
	Minute uint8
}

func (hm hourMinute) Before(minutes uint8) hourMinute {
	// Setup for overflow
	newHM := hourMinute{hm.Hour + 24, hm.Minute + 60}

	hours := minutes / 60
	minutes = minutes % 60

	// Apply difference
	if minutes > hm.Minute {
		newHM.Hour -= hours + 1
	} else {
		newHM.Hour -= hours
	}
	newHM.Minute -= minutes

	// Translate back to original interval
	newHM.Hour %= 24
	newHM.Minute %= 60

	return newHM
}

func (hm hourMinute) IsAfter(other hourMinute) bool {
	return hm.Hour > other.Hour || (hm.Hour == other.Hour && hm.Minute > other.Minute)
}

func (hm hourMinute) String() string {
	return fmt.Sprintf("%02d:%02d", hm.Hour, hm.Minute)
}

type mealTime struct {
	Start hourMinute
	End   hourMinute
}

func (mt mealTime) String() string {
	return fmt.Sprintf("%s - %s", mt.Start, mt.End)
}

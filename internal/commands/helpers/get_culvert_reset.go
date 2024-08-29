package helpers

import "time"

var culvertResetChangedFromSundayToWednesday, _ = time.Parse("2006-01-02", "2024-08-28") // hardcoded as of the maple GMS patch on this day

func GetCulvertResetDay(asOf time.Time) time.Weekday {
	if asOf.Before(culvertResetChangedFromSundayToWednesday) {
		return time.Sunday
	}

	return time.Wednesday
}

func GetCulvertResetDate(thisWeek time.Time) time.Time {
	for thisWeek.Weekday() != GetCulvertResetDay(thisWeek) {
		thisWeek = thisWeek.Add(time.Hour * -24)
	}
	return thisWeek
}

func GetCulvertPreviousDate(thisWeek time.Time) time.Time {
	thisWeek = thisWeek.Add(-24 * time.Hour)
	for thisWeek.Weekday() != GetCulvertResetDay(thisWeek) {
		thisWeek = thisWeek.Add(time.Hour * -24)
	}
	return thisWeek
}

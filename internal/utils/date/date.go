package date

import (
	"log"
	"time"
)

type GetRelativeDateParams struct {
	Day      *int
	Hour     *int
	Minute   *int
	Location *string
}

func GetRelativeDate(params *GetRelativeDateParams) int64 {
	var day int = 1
	var hour int = 21
	var minute int = 30
	var location *time.Location

	if params.Location != nil {
		var err error
		location, err = time.LoadLocation(*params.Location)
		if err != nil {
			log.Printf("Error loading location: %v", err)
			location = time.FixedZone("UTC+3", 3*60*60)
		}
	}

	if params.Hour != nil {
		hour = *params.Hour
	}
	if params.Minute != nil {
		minute = *params.Minute
	}
	if params.Day != nil {
		day = *params.Day
	}

	now := time.Now()

	current := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, location)

	weekday := int(current.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	var daysToAdd int

	if weekday == day {
		// If today is day
		daysToAdd = 7
	} else if weekday < day {
		// If before day
		daysToAdd = day - weekday
	} else {
		// If after day
		daysToAdd = 7 - (weekday - day)
	}

	targetDate := current.AddDate(0, 0, daysToAdd)
	targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, minute, 0, 0, targetDate.Location())

	return targetDate.Unix()
}

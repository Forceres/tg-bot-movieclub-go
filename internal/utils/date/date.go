package date

import (
	"log"
	"time"

	"github.com/Forceres/tg-bot-movieclub-go/internal/utils/telegram/datepicker"
	"github.com/go-telegram/bot"
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

var DatePicker *datepicker.DatePicker

func InitDatePicker(b *bot.Bot, f datepicker.OnCancelHandler, h datepicker.OnSelectHandler) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	date := time.Date(year, month, 0, 0, 0, 0, 0, time.UTC)
	curr := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	curr = curr.AddDate(0, 0, -1)
	daysToAdd := curr.Day() - 1
	opts := []datepicker.Option{
		datepicker.From(date),
		datepicker.OnCancel(f),
		datepicker.To(date.AddDate(0, 0, daysToAdd)),
		datepicker.Language("ru"),
		datepicker.WithPrefix("datepicker"),
	}
	DatePicker = datepicker.New(
		b,
		h,
		opts...,
	)
}

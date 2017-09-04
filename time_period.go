package main

import (
	"errors"
	"log"
	"strings"
	"time"
)

type TimePeriod struct {
	Start time.Time
	End   time.Time
}

func (p TimePeriod) String() string {
	return p.Start.Format(DATE_FMT_SPACES) + " - " + p.End.Format(DATE_FMT_SPACES)
}

func (p TimePeriod) StringDay() string {
	return p.Start.Format(DAY_FMT) + " " + p.End.Format(DAY_FMT)
}

// Valid formats: '2006-01-02', '2006-01-02 2006-01-02', '2006-01', '2006-01 2006-01', '2006' and '2006 2006'
// Note that the end date will be at the end of the smallest time unit specified.
// Ex: '2017' -> '2017-01-01 00:00:00' to '2017-12-31 23:59:59'
// Ex: '2017 2018' -> '2017-01-01 00:00:00' to '2018-12-31 23:59:59'
func ParseTimePeriod(input string) (TimePeriod, error) {
	var err, err1, err2 error
	var start, end, date time.Time

	// Try '2006-01-02'
	date, err = time.Parse(DAY_FMT, input)
	if err == nil {
		end = EndOfDay(start)
		return TimePeriod{Start: date, End: date}, nil
	}
	// Try '2006-01'
	start, err = time.Parse(MONTH_FMT, input)
	if err == nil {
		end = EndOfMonth(start)
		return TimePeriod{Start: start, End: date}, nil
	}
	// Try '2006'
	start, err = time.Parse(YEAR_FMT, input)
	if err == nil {
		end = EndOfYear(start)
		return TimePeriod{Start: start, End: date}, nil
	}

	// Try multipart
	parts := strings.Split(input, " ")
	if len(parts) != 2 {
		log.Println("Failed to parse time period: " + input)
		log.Println("Too many or too few parts:", parts)
		return TimePeriod{}, errors.New("failed to parse: " + input)
	}
	// Try '2006-01-02 2006-01-02'
	start, err1 = time.Parse(DAY_FMT, parts[0])
	end, err2 = time.Parse(DAY_FMT, parts[1])
	if err1 == nil && err2 == nil {
		end = EndOfDay(start)
		return TimePeriod{Start: start, End: end}, nil
	}
	// Try '2006-01 2006-01'
	start, err1 = time.Parse(MONTH_FMT, parts[0])
	end, err2 = time.Parse(MONTH_FMT, parts[1])
	if err1 == nil && err2 == nil {
		end = EndOfMonth(start)
		return TimePeriod{Start: start, End: end}, nil
	}
	// Try '2006 2006'
	start, err1 = time.Parse(YEAR_FMT, parts[0])
	end, err2 = time.Parse(YEAR_FMT, parts[1])
	if err1 == nil && err2 == nil {
		end = EndOfYear(start)
		return TimePeriod{Start: start, End: end}, nil
	}

	// Everything failed
	return TimePeriod{}, errors.New("failed to parse: " + input)
}

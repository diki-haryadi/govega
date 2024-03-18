package util

import (
	"fmt"
	"log"
	"time"
)

func ToClock(timeData time.Time) string {
	hour, min, _ := timeData.Clock()
	if min < 10 {
		return fmt.Sprintf("%v:0%v", hour, min)
	}
	return fmt.Sprintf("%v:%v", hour, min)
}

func TimeToDateTimezone(setTime time.Time, timeZone string) time.Time {
	//Set Timezone
	loc, _ := time.LoadLocation(timeZone)
	//Convert to Timezone
	now := setTime.In(loc)
	//Change hh:mm:ss to 00:00:00
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

func EpochToTime(epoch int64) time.Time {
	t := time.Unix(epoch, 0)
	return t
}

func StringToDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Println("invalid duration value, fallback to default timeout")
		return 10 * time.Second
	}

	return d
}

package service

import (
	"os"
	"time"

	"github.com/Adapptor/service/v2/log"
)

// Integer min/max
func Min(x, y int32) int32 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

func MapFromStringSlice(strs []string) map[string]bool {
	result := make(map[string]bool)
	for _, s := range strs {
		result[s] = true
	}
	return result
}

func PerthLocation() *time.Location {
	awst, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		log.Log(log.Error, "PerthLocation() unable to find Perth timezone in local timezone db", err, nil)
		os.Exit(1)
	}

	return awst
}

func PerthNow() time.Time {
	return time.Now().In(PerthLocation())
}

func ParseProtoTime(timeStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return date, err
	}

	date = time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		date.Hour(),
		date.Minute(),
		date.Second(),
		date.Nanosecond(),
		PerthLocation(),
	)

	return date, nil
}

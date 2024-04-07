package utils

import (
	"os"
	"strconv"
	"time"
)

func EpochMilliToTime(epochMilli uint64) time.Time {
	sec := int64(epochMilli / 1000)
	nsec := int64((epochMilli % 1000) * 1000000)
	return time.Unix(sec, nsec)
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetEnvAsInt(key string, fallback int) int {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return fallback
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}
	return value
}

func Map[T any, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

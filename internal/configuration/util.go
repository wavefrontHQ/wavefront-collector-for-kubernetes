package configuration

import "time"

func GetStringValue(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func GetDurationValue(value, defaultValue time.Duration) time.Duration {
	if value != 0 {
		return value
	}
	return defaultValue
}

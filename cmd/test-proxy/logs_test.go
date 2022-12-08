package main

import (
	"testing"
)

func TestLogFormat(t *testing.T) {
	t.Run("logs are in json_array format", func(t *testing.T) {
		jsonArray := `[
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"},
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					]`
		VerifyJsonArray(jsonArray)
	})

	t.Run("logs are in json lines format", func(t *testing.T) {
		jsonLines := `{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}
					{"cluster":"testk8scluster","message":"testlogmessage","service":"none"}`
		VerifyJsonLines(jsonLines)
	})
}



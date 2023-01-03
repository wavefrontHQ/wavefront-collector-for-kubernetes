package logs

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	expectedTags = [...]string{"service", "application", "source", "cluster", "timestamp", "pod_name", "container_name", "namespace_name", "pod_id", "container_id"}
)

func VerifyJsonArray(line string) bool {
	var logLines []interface{}
	err := json.Unmarshal([]byte(line), &logLines)
	if err != nil {
		fmt.Println("Data is not in json array format:", err)
		return false
	}

	for _, logLine := range logLines {
		if validateLogLine(logLine) != true {
			return false
		}
	}
	return true
}

func VerifyJsonLines(line string) bool {
	decoder := json.NewDecoder(strings.NewReader(line))
	for decoder.More() {
		var jsonLine interface{}
		if err := decoder.Decode(&jsonLine); err != nil {
			fmt.Println("Data is not in json line format:", err)
			return false
		}
	}

	return true
}

func validateLogLine(logLine interface{}) bool {
	//return validateTags(logLine)
	return true
}

func validateTags(logLine interface{}) bool {
	myMap := logLine.(map[string]interface{})

	for _, expectedTag := range expectedTags {
		if val, ok := myMap[expectedTag]; ok {
			if len(val.(string)) == 0 {
				fmt.Printf("Empty expected tag: %s\n", expectedTag)
				return false
			}
		} else {
			fmt.Printf("Missing expected tag: %s\n", expectedTag)
			return false
		}

	}
	return true
}

package logs

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	expectedTags = [...]string{"user_defined_tag", "service", "application", "source", "cluster", "timestamp", "pod_name", "container_name", "namespace_name", "pod_id", "container_id"}
)

func VerifyJsonArrayFormat(line string) (bool, []interface{}) {
	var logLines []interface{}
	err := json.Unmarshal([]byte(line), &logLines)
	if err != nil {
		fmt.Println("Data is not in json array format:", err)
		return false, logLines
	}

	return true, logLines
}

func VerifyJsonLinesFormat(line string) (bool, []interface{}) {
	var logLines []interface{}
	decoder := json.NewDecoder(strings.NewReader(line))
	for decoder.More() {
		var jsonLine interface{}
		if err := decoder.Decode(&jsonLine); err != nil {
			fmt.Println("Data is not in json line format:", err)
			return false, nil
		}

		logLines = append(logLines, jsonLine)
	}

	return true, logLines
}

func ValidateTags(logLines []interface{}) (bool, []string) {
	valid := true
	missing := make(map[string]interface{})

	for _, logLine := range logLines {
		myMap := logLine.(map[string]interface{})

		for _, expectedTag := range expectedTags {
			if val, ok := myMap[expectedTag]; ok {
				if val == nil {
					fmt.Printf("Empty expected tag: %s\n", expectedTag)
					valid = false
					missing[expectedTag] = nil
				}
			} else {
				fmt.Printf("Missing expected tag: %s\n", expectedTag)
				valid = false
				missing[expectedTag] = nil
			}
		}
	}

	var missingTags []string
	for k := range missing {
		missingTags = append(missingTags, k)
	}

	return valid, missingTags
}

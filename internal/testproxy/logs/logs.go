package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type LogVerifier struct {
	expectedTags          []string
	allowListFilteredTags map[string]string
	denyListFilteredTags  map[string]string
}

func NewLogVerifier(expectedTags []string, allowListFilteredTags map[string]string, denyListFilteredTags map[string]string) *LogVerifier {
	return &LogVerifier{
		expectedTags:          expectedTags,
		allowListFilteredTags: allowListFilteredTags,
		denyListFilteredTags:  denyListFilteredTags,
	}
}

func (l *LogVerifier) VerifyJsonArrayFormat(line []byte) (bool, []interface{}) {
	if len(line) == 0 {
		fmt.Println("Data for verifying json array format is empty")
		return false, nil
	}

	if line[0] != '[' {
		fmt.Printf("Data is not in json array format, first character was '%s'\n", string(line[0]))
		return false, nil
	}

	var logLines []interface{}
	err := json.Unmarshal(line, &logLines)
	if err != nil {
		fmt.Println("Data is not in json array format:", err)
		return false, nil
	}

	if len(logLines) == 0 {
		fmt.Println("Json array was empty")
		return false, nil
	}

	return true, logLines
}

func (l *LogVerifier) VerifyJsonLinesFormat(line []byte) (bool, []interface{}) {
	if len(line) == 0 {
		fmt.Println("Data for verifying json lines format is empty")
		return false, nil
	}

	if line[0] != '{' {
		fmt.Printf("Data is not in json line format, first character was '%s'\n", string(line[0]))
		return false, nil
	}

	var logLines []interface{}
	decoder := json.NewDecoder(bytes.NewReader(line))
	for decoder.More() {
		var jsonLine interface{}
		if err := decoder.Decode(&jsonLine); err != nil {
			fmt.Println("Data is not in json line format:", err)
			return false, nil
		}

		if len(jsonLine.(map[string]interface{})) == 0 {
			fmt.Println("Json line was empty")
			return false, nil
		}

		logLines = append(logLines, jsonLine)
	}

	return true, logLines
}

func (l *LogVerifier) ValidateExpectedTags(logLines []interface{}) (bool, []string, []string) {
	valid := true
	missingTags := make(map[string]interface{})
	emptyExpectedTags := make(map[string]interface{})

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		for _, expectedTag := range l.expectedTags {
			if tagVal, ok := logTags[expectedTag]; ok {
				if tagVal == nil {
					fmt.Printf("Empty expected tag: %s\n", expectedTag)
					valid = false
					emptyExpectedTags[expectedTag] = nil
				}
			} else {
				fmt.Printf("Missing expected tag: %s\n", expectedTag)
				valid = false
				missingTags[expectedTag] = nil
			}
		}
	}

	var missingTagsList []string
	for k := range missingTags {
		missingTagsList = append(missingTagsList, k)
	}

	var emptyTagsList []string
	for k := range emptyExpectedTags {
		emptyTagsList = append(emptyTagsList, k)
	}

	return valid, missingTagsList, emptyTagsList
}

// TODO: add logic for missingAllowListTags, extraDenyListTags

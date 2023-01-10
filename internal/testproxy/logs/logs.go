package logs

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"reflect"
)

type LogVerifier struct {
	expectedTags          []string
	allowListFilteredTags map[string][]string
	denyListFilteredTags  map[string][]string
}

func NewLogVerifier(expectedTags []string, allowListFilteredTags map[string][]string, denyListFilteredTags map[string][]string) *LogVerifier {
	return &LogVerifier{
		expectedTags:          expectedTags,
		allowListFilteredTags: allowListFilteredTags,
		denyListFilteredTags:  denyListFilteredTags,
	}
}

func (l *LogVerifier) VerifyJsonArrayFormat(line []byte) (bool, []interface{}) {
	if len(line) == 0 {
		log.Println("Data for verifying json array format is empty")
		return false, nil
	}

	if line[0] != '[' {
		log.Printf("Data is not in json array format, first character was '%s'\n", string(line[0]))
		return false, nil
	}

	var logLines []interface{}
	err := json.Unmarshal(line, &logLines)
	if err != nil {
		log.Println("Data is not in json array format:", err)
		return false, nil
	}

	if len(logLines) == 0 {
		log.Println("Json array was empty")
		return false, nil
	}

	return true, logLines
}

func (l *LogVerifier) VerifyJsonLinesFormat(line []byte) (bool, []interface{}) {
	if len(line) == 0 {
		log.Println("Data for verifying json lines format is empty")
		return false, nil
	}

	if line[0] != '{' {
		log.Printf("Data is not in json line format, first character was '%s'\n", string(line[0]))
		return false, nil
	}

	var logLines []interface{}
	decoder := json.NewDecoder(bytes.NewReader(line))
	for decoder.More() {
		var jsonLine interface{}
		if err := decoder.Decode(&jsonLine); err != nil {
			log.Println("Data is not in json line format:", err)
			return false, nil
		}

		if len(jsonLine.(map[string]interface{})) == 0 {
			log.Println("Json line was empty")
			return false, nil
		}

		logLines = append(logLines, jsonLine)
	}

	return true, logLines
}

func (l *LogVerifier) ValidateExpectedTags(logLines []interface{}) (bool, map[string]interface{}, map[string]interface{}) {
	valid := true
	missingTags := make(map[string]interface{})
	emptyExpectedTags := make(map[string]interface{})

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		for _, expectedTag := range l.expectedTags {
			if tagVal, ok := logTags[expectedTag]; ok {
				if tagVal == nil || (reflect.TypeOf(tagVal).String() == "string" && len(tagVal.(string)) == 0) {
					log.Printf("Empty expected tag: %s\n", expectedTag)
					valid = false
					emptyExpectedTags[expectedTag] = nil
				}
			} else {
				log.Printf("Missing expected tag: %s\n", expectedTag)
				valid = false
				missingTags[expectedTag] = nil
			}
		}
	}

	return valid, missingTags, emptyExpectedTags
}

func (l *LogVerifier) ValidateAllowedTags(logLines []interface{}) (bool, []interface{}) {
	valid := true
	var unexpectedLogs []interface{}

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		foundAllowedTag := false
		for tagKey, tagVal := range logTags {
			if foundAllowedTag {
				break
			}

			if allowedValsForKey, ok := l.allowListFilteredTags[tagKey]; ok && allowedValsForKey != nil {
				for _, allowedVal := range allowedValsForKey {
					if tagVal == allowedVal {
						foundAllowedTag = true
						break
					}
				}
			}
		}

		if !foundAllowedTag {
			log.Println("Expected to find a tag from the allowed list")
			unexpectedLogs = append(unexpectedLogs, logLine)
		}
		valid = valid && foundAllowedTag
	}

	return valid, unexpectedLogs
}

func (l *LogVerifier) ValidateDeniedTags(logLines []interface{}) (bool, map[string]interface{}) {
	valid := true
	unexpectedTags := make(map[string]interface{})

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		for tagDenyKey, tagDenyValList := range l.denyListFilteredTags {
			if tagVal, ok := logTags[tagDenyKey]; ok {
				for _, tagDenyVal := range tagDenyValList {
					if tagVal == tagDenyVal {
						log.Printf("Unexpected deny list tag: key=\"%s\", value=\"%s\"\n", tagDenyKey, tagVal)
						valid = false
						unexpectedTags[tagDenyKey] = tagVal
					}
				}
			}
		}
	}

	return valid, unexpectedTags
}

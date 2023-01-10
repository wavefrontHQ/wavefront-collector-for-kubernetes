package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

type LogVerifier struct {
	results               *Results
	expectedTags          []string
	allowListFilteredTags map[string][]string
	denyListFilteredTags  map[string][]string
}

func NewLogVerifier(results *Results, expectedTags []string, allowListFilteredTags map[string][]string, denyListFilteredTags map[string][]string) *LogVerifier {
	return &LogVerifier{
		results:               results,
		expectedTags:          expectedTags,
		allowListFilteredTags: allowListFilteredTags,
		denyListFilteredTags:  denyListFilteredTags,
	}
}

func (l *LogVerifier) VerifyJsonArrayFormat(data []byte) []interface{} {
	var isValid bool
	defer func() {
		l.results.SetHasValidFormat(isValid)
	}()

	if len(data) == 0 {
		log.Println("Data for verifying json array format is empty")
		return nil
	}

	if data[0] != '[' {
		log.Printf("Data is not in json array format, first character was '%s'\n", string(data[0]))
		return nil
	}

	var logLines []interface{}
	err := json.Unmarshal(data, &logLines)
	if err != nil {
		log.Println("Data is not in json array format:", err)
		return nil
	}

	logCount := len(logLines)
	if logCount == 0 {
		log.Println("Json array was empty")
		return nil
	}

	isValid = true
	l.results.AddReceivedCount(logCount)

	return logLines
}

func (l *LogVerifier) VerifyJsonLinesFormat(line []byte) []interface{} {
	if len(line) == 0 {
		log.Println("Data for verifying json lines format is empty")
		l.results.SetHasValidFormat(false)
		return nil
	}

	if line[0] != '{' {
		log.Printf("Data is not in json line format, first character was '%s'\n", string(line[0]))
		l.results.SetHasValidFormat(false)
		return nil
	}

	var logLines []interface{}
	decoder := json.NewDecoder(bytes.NewReader(line))
	for decoder.More() {
		var jsonLine interface{}
		if err := decoder.Decode(&jsonLine); err != nil {
			log.Println("Data is not in json line format:", err)
			l.results.SetHasValidFormat(false)
			return nil
		}

		if len(jsonLine.(map[string]interface{})) == 0 {
			log.Println("Json line was empty")
			l.results.SetHasValidFormat(false)
			return nil
		}

		logLines = append(logLines, jsonLine)
	}

	l.results.SetHasValidFormat(true)
	l.results.AddReceivedCount(len(logLines))

	return logLines
}

func (l *LogVerifier) ValidateExpectedTags(logLines []interface{}) {
	valid := true

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		for _, expectedTag := range l.expectedTags {
			if tagVal, ok := logTags[expectedTag]; ok {
				if tagVal == nil || (reflect.TypeOf(tagVal).String() == "string" && len(tagVal.(string)) == 0) {
					log.Printf("Empty expected tag: %s\n", expectedTag)
					valid = false

					l.results.AddEmptyExpectedTag(expectedTag)
					l.results.IncrementEmptyExpectedTagsCount()
				}
			} else {
				log.Printf("Missing expected tag: %s\n", expectedTag)
				valid = false
				l.results.AddMissingExpectedTag(expectedTag)
				l.results.IncrementMissingExpectedTagsCount()
			}
		}
	}

	l.results.SetHasValidTags(valid)

	return
}

func (l *LogVerifier) ValidateAllowedTags(logLines []interface{}) {
	valid := true
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
			l.results.AddUnexpectedAllowedLog(logLine)
			l.results.IncrementUnexpectedAllowedLogsCount()
		}
		valid = valid && foundAllowedTag
	}

	l.results.SetHasValidTags(valid)

	return
}

func (l *LogVerifier) ValidateDeniedTags(logLines []interface{}) {
	valid := true

	for _, logLine := range logLines {
		logTags := logLine.(map[string]interface{})

		for tagDenyKey, tagDenyValList := range l.denyListFilteredTags {
			if tagVal, ok := logTags[tagDenyKey]; ok {
				for _, tagDenyVal := range tagDenyValList {
					if tagVal == tagDenyVal {
						log.Printf("Unexpected deny list tag: key=\"%s\", value=\"%s\"\n", tagDenyKey, tagVal)
						valid = false
						l.results.AddUnexpectedDeniedTag(fmt.Sprintf("%s:%s", tagDenyKey, tagVal))
						l.results.IncrementUnexpectedDeniedTagsCount()
					}
				}
			}
		}
	}

	l.results.SetHasValidTags(valid)

	return
}

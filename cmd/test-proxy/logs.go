package main

import (
	"encoding/json"
	"fmt"
)

func VerifyJsonArray(line string) bool{
	var jsonarray []interface{}
	err := json.Unmarshal([]byte(line), &jsonarray)
	if err != nil {
		fmt.Println("Data is not in json array format:", err)
		return false
	}

	return true
}

func VerifyJsonLines(line string) bool{
	var jsonlines []interface{}
	err := json.Unmarshal([]byte(line), &jsonlines)
	if err == nil {
		fmt.Println("Data is not in json lines format :", err)
		return false
	}
	fmt.Println(err)
	return true
}
package main

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"time"
)

func main() {
	encoded := make([]byte, base64.StdEncoding.EncodedLen(1024))
	for {
		data := make([]byte, 1024)
		rand.Read(data)
		base64.StdEncoding.Encode(encoded, data)
		os.Stdout.Write(encoded)
		time.Sleep(1 * time.Second)
	}
}

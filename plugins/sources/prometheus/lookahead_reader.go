package prometheus

import (
	"bufio"
	"io"
)

type LookaheadReader struct {
	scanner   *bufio.Scanner
	done      bool
	lineCount int
}

func NewLookaheadReader(reader io.Reader) *LookaheadReader {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	laReader := &LookaheadReader{
		scanner: scanner,
	}

	laReader.advanceLine()
	return laReader
}

func (laReader *LookaheadReader) Done() bool {
	return laReader.done
}

func (laReader *LookaheadReader) Read() []byte {
	if laReader.done {
		return []byte{}
	}
	// the scanner slice changes unexpectedly in some cases so a defensive copy
	// protects us from that
	retVal:=makeDefensiveCopy(laReader.scanner.Bytes())
	laReader.advanceLine()
	return retVal
}

func makeDefensiveCopy(buf []byte) []byte{
	retVal := make([]byte, len(buf))
	copy(retVal, buf)
	return retVal
}

func (laReader *LookaheadReader) advanceLine() {
	laReader.lineCount++
	laReader.done = !laReader.scanner.Scan()
}

func (laReader *LookaheadReader) Peek() []byte {
	return laReader.scanner.Bytes()
}

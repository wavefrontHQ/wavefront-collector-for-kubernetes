package prometheus_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"testing"
)

func TestLAEmptyFile(t *testing.T) {
	byteReader := bytes.NewReader([]byte{})
	reader := prometheus.NewLookaheadReader(byteReader)
	assertEmptyLAReader(t, reader)
}

func TestLAOneLineRead(t *testing.T) {
	line := []byte(`ONE LINE`)
	byteReader := bytes.NewReader(line)
	reader := prometheus.NewLookaheadReader(byteReader)

	assert.False(t, reader.Done())
	assert.Equal(t, line, reader.Peek())
	assert.Equal(t, line, reader.Read())

	assertEmptyLAReader(t, reader)
}

func TestLAMultiLineRead(t *testing.T) {
	lineOne := []byte("ONE LINE")
	lineTwo := []byte("TWO LINE")

	file := append(lineOne,'\n')
	file = append(file, lineTwo...)
	file = append(file, '\n')

	byteReader := bytes.NewReader(file)
	reader := prometheus.NewLookaheadReader(byteReader)

	assert.False(t, reader.Done())
	assert.Equal(t, lineOne, reader.Peek())
	assert.Equal(t, lineOne, reader.Read())

	assert.False(t, reader.Done())
	assert.Equal(t, lineTwo, reader.Peek())
	assert.Equal(t, lineTwo, reader.Read())

	assertEmptyLAReader(t, reader)
}

func assertEmptyLAReader(t *testing.T, reader *prometheus.LookaheadReader) {
	assert.True(t, reader.Done())
	assert.Empty(t, reader.Peek())
	assert.Empty(t, reader.Read())
}

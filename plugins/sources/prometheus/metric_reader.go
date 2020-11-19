package prometheus

import (
	"bytes"
	"io"
)

type MetricReader struct {
	lar *LookaheadReader
}

func NewMetricReader(reader io.Reader) *MetricReader {
	return &MetricReader{
		lar: NewLookaheadReader(reader),
	}
}

// Done tells us if there is anything left to read
func (mReader *MetricReader) Done() bool {
	return mReader.lar.Done()
}

func (mReader *MetricReader) Read() []byte {
	buffer := bytes.NewBuffer(nil)

	mReader.readComments(buffer)
	mReader.readMetrics(buffer)

	return buffer.Bytes()
}

func (mReader *MetricReader) readComments(buffer *bytes.Buffer) {
	for !mReader.lar.Done() {
		if !mReader.commentNext() {
			break
		}
		buffer.Write(mReader.lar.Read())
		buffer.Write([]byte("\n"))
	}
}

func (mReader *MetricReader) commentNext() bool {
	trimmed := bytes.TrimSpace(mReader.lar.Peek())
	if 0 == bytes.Compare([]byte{}, trimmed) {
		return true
	}
	return bytes.HasPrefix(mReader.lar.Peek(), []byte("#"))
}

func (mReader *MetricReader) readMetrics(buffer *bytes.Buffer) {
	for !mReader.lar.Done() {
		if bytes.HasPrefix(mReader.lar.Peek(), []byte("#")) {
			break
		}
		buffer.Write(mReader.lar.Read())
		buffer.Write([]byte("\n"))
	}
}

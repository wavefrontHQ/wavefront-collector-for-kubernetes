package util

import (
	"bytes"
	"net/url"
)

type TagsEncoder interface {
	Encode(map[string]string) string
}

func NewTagsEncoder() TagsEncoder {
	return &tagsEncoder{
		buf: bytes.NewBufferString(""),
	}
}

type tagsEncoder struct {
	buf *bytes.Buffer
}

func (te *tagsEncoder) Encode(tags map[string]string) string {
	te.buf.Reset()
	for k, v := range tags {
		if len(v) > 0 {
			te.buf.WriteString(" ")
			te.buf.WriteString(url.QueryEscape(k))
			te.buf.WriteString("=")
			te.buf.WriteString(url.QueryEscape(v))
		}
	}
	return te.buf.String()
}

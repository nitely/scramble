package scrambler

import (
	"bytes"
	"testing"
)

func encodeText(s string) (string, error) {
	reader := bytes.NewBuffer([]byte(s))
	writer := bytes.NewBuffer([]byte{})
	err := PipeEncoder(reader, writer)
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}

func TestEncoderEncodeFoo(t *testing.T) {
	want := "[124807030]"
	in := "foo"
	got, err := encodeText(in)
	if err != nil {
		t.Errorf("encodeText(%s) error %s", in, err)
	}
	if got != want {
		t.Errorf("encodeText(%s) == %s, want %s", in, got, want)
	}
}

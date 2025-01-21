package scrambler

import (
	"bytes"
	"testing"
)

func decodeText(s string) (string, error) {
	reader := bytes.NewBuffer([]byte(s))
	writer := bytes.NewBuffer([]byte{})
	err := PipeDecoder(reader, writer)
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}

func TestDecoderFoo(t *testing.T) {
	want := "foo"
	in := "[124807030]"
	got, err := decodeText(in)
	if err != nil {
		t.Errorf("decodeText(%s) error %s", in, err)
	}
	if got != want {
		t.Errorf("decodeText(%s) == %s, want %s", in, got, want)
	}
}

func TestDecoderExamples(t *testing.T) {
	cases := []struct {
		want string
		in   string
	}{
		{"foo", "[124807030]"},
		{" foo", "[250662636]"},
		{"foot", "[267939702]"},
		{"BIRD", "[251930706]"},
		{"....", "[15794160]"},
		{"^^^^", "[252706800]"},
		{"Woot", "[266956663]"},
		{"no", "[53490482]"},
		{"tacocat", "[267487694, 125043731]"},
		{"never odd or even", "[267657050, 233917524, 234374596, 250875466, 17830160]"},
		{"lager, sir, is regal", "[267394382, 167322264, 66212897, 200937635, 267422503]"},
		{
			"go hang a salami, I'm a lasagna hog",
			"[200319795, 133178981, 234094669, 267441422, 78666124, 99619077, 267653454, 133178165, 124794470]",
		},
		{
			"egad, a base tone denotes a bad age",
			"[267389735, 82841860, 267651166, 250793668, 233835785, 267665210, 99680277, 133170194, 124782119]",
		},
	}
	for _, c := range cases {
		got, err := decodeText(c.in)
		if err != nil {
			t.Errorf("decodeText(%s) error %s", c.in, err)
		}
		if got != c.want {
			t.Errorf("decodeText(%s) == %s, want %s", c.in, got, c.want)
		}
	}
}

func TestDecoderMissingEnd(t *testing.T) {
	_, err := decodeText("[123")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "missing ]"; err.Error() != expected {
		t.Errorf("error == %s, want %s", err.Error(), expected)
	}
}

func TestDecoderMissingStart(t *testing.T) {
	_, err := decodeText("123]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "expected [ at pos 0"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderMissingSeparator(t *testing.T) {
	_, err := decodeText("[123 123]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "unexpected space at pos 4"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderMissingSpace(t *testing.T) {
	_, err := decodeText("[123,123]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "expected space at pos 5"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderBigNumber(t *testing.T) {
	_, err := decodeText("[123123123123123123123]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "token too long error at pos 1"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderBadNumber(t *testing.T) {
	_, err := decodeText("[abc]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "failed to parse number at pos 1"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderExtraSeparator(t *testing.T) {
	_, err := decodeText("[123,,123]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "unexpected comma at pos 5"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderComma(t *testing.T) {
	_, err := decodeText("[123,]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "unexpected char at pos 5"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderSpace(t *testing.T) {
	_, err := decodeText("[ ]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "unexpected char at pos 1"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderEmptyBrackets(t *testing.T) {
	_, err := decodeText("[]")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "empty list"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderEmptyInput(t *testing.T) {
	_, err := decodeText("")
	if err == nil {
		t.Errorf("decoder error expected")
		return
	}
	if expected := "empty input"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

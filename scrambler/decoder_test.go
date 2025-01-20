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

func TestDecoderExamples(t *testing.T) {
	testCases := []struct {
		decoded string
		encoded string
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
	for _, tc := range testCases {
		got, err := decodeText(tc.encoded)
		if err != nil {
			t.Errorf("decoder error %s for %s", err, tc.encoded)
		}
		if got != tc.decoded {
			t.Errorf("expected %s, got %s", tc.decoded, got)
		}
	}
}

func TestDecoderMissingEnd(t *testing.T) {
	_, err := decodeText("[123")
	if err == nil {
		t.Errorf("decoder error expected")
	}
	if expected := "missing ]"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderMissingStart(t *testing.T) {
	_, err := decodeText("123]")
	if err == nil {
		t.Errorf("decoder error expected")
	}
	if expected := "missing ["; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

func TestDecoderMissingSeparator(t *testing.T) {
	_, err := decodeText("[123 123]")
	if err == nil {
		t.Errorf("decoder error expected")
	}
	if expected := "missing comma separator"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

//func TestDecoderMissingSpace(t *testing.T) {
//	_, err := decodeText("[123,123]")
//	if err == nil {
//		t.Errorf("decoder error expected")
//	}
//	if expected := ""; err.Error() != expected {
//		t.Errorf("expected message %s got %s", expected, err.Error())
//	}
//}

func TestDecoderBigNumer(t *testing.T) {
	_, err := decodeText("[123123123123123123123123123]")
	if err == nil {
		t.Errorf("decoder error expected")
	}
	if expected := "missing space separator"; err.Error() != expected {
		t.Errorf("expected message %s got %s", expected, err.Error())
	}
}

//func TestDecoderExtraSeparator(t *testing.T) {
//	_, err := decodeText("[123,,123]")
//	if err == nil {
//		t.Errorf("decoder error expected")
//	}
//	if expected := ""; err.Error() != expected {
//		t.Errorf("expected message %s got %s", expected, err.Error())
//	}
//}

package scrambler

import (
	"bytes"
	"fmt"
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

func TestEncoderExamples(t *testing.T) {
	cases := []struct {
		in   string
		want string
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
		got, err := encodeText(c.in)
		if err != nil {
			t.Errorf("encodeText(%s) error %s", c.in, err)
		}
		if got != c.want {
			t.Errorf("encodeText(%s) == %s, want %s", c.in, got, c.want)
		}
	}
}

func TestEncoderEmptyInput(t *testing.T) {
	_, err := encodeText("")
	if err == nil {
		t.Errorf("encoder error expected")
		return
	}
	if expected := "empty input"; err.Error() != expected {
		t.Errorf("error == %s, want %s", err.Error(), expected)
	}
}

func ExamplePipeEncoder() {
	reader := bytes.NewBuffer([]byte("foo"))
	writer := bytes.NewBuffer([]byte{})
	err := PipeEncoder(reader, writer)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(writer.String())
	// Output: [124807030]
}

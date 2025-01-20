package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// Convert string of up to 4 chars into
// a decoded/raw integer.
func rawInteger(s string) (uint32, error) {
	if len(s) > 4 {
		return 0, errors.New("string must be <= 4 chars long")
	}
	// Zero padding.
	for len(s) < 4 {
		s += "\x00"
	}
	result := uint32(0)
	for i := range 4 {
		result |= uint32(s[i]) << (8 * i)
	}
	return result, nil
}

// Encode string of up to 4 chars into
// a Weird Text Format-8 integer.
func encode(s string) (uint32, error) {
	decoded, err := rawInteger(s)
	if err != nil {
		return 0, err
	}
	nimble := 0
	result := uint32(0)
	for i := range 32 {
		bit := (decoded & (1 << i)) >> i      // Get bit at position i.
		result |= bit << (nimble*4 + (i / 8)) // Shift the bit to the right nimble start position + bit offset.
		nimble = (nimble + 1) % 8             // Move to the next nimble wrapping around.
	}
	return result, nil
}

// Decode a Weird Text Format-8 encoded integer.
func decode(enc uint32) string {
	decoded := uint32(0)
	nimble := 0
	for i := range 32 {
		sourcePos := nimble*4 + (i / 8)              // Position where the bit was placed during encoding.
		bit := (enc & (1 << sourcePos)) >> sourcePos // Extract the bit from its encoded position.
		decoded |= bit << i                          // Place the bit back in its original position.
		nimble = (nimble + 1) % 8                    // Move to the next nimble wrapping around.
	}
	// Uint32 to 4 chars string.
	result := make([]byte, 4)
	result[3] = byte((decoded >> 24) & 0xFF)
	result[2] = byte((decoded >> 16) & 0xFF)
	result[1] = byte((decoded >> 8) & 0xFF)
	result[0] = byte(decoded & 0xFF)
	return strings.TrimRight(string(result), "\000")
}

// Weird Text Format-8 Encoder.
type Encoder struct {
	writer    io.Writer
	buffer    []byte // Write buffer.
	isAtStart bool   // Start of the stream.
}

// Format a encoded integer in Weird Text Format-8.
func piece(n uint32, isStart, isEnd bool) string {
	if isStart && isEnd {
		return fmt.Sprintf("[%d]", n)
	}
	if isStart {
		return fmt.Sprintf("[%d", n)
	}
	if isEnd {
		return fmt.Sprintf(", %d]", n)
	}
	return fmt.Sprintf(", %d", n)
}

// Encode a data stream into Weird Text Format-8.
func (e *Encoder) Write(p []byte) (n int, err error) {
	e.buffer = append(e.buffer, p...)
	written := 0
	for len(e.buffer) >= 4 {
		chunk := e.buffer[:4]
		encoded, err := encode(string(chunk))
		if err != nil {
			return written, err
		}
		out := piece(encoded, e.isAtStart, false)
		n, err := e.writer.Write([]byte(out))
		if err != nil {
			return written, err
		}
		written += n
		e.isAtStart = false
		e.buffer = e.buffer[4:]
	}
	return len(p), nil
}

// Flush remaining buffer data into the writter.
// That's raw data lesser than 4 chars long.
// Write closing list symbol.
func (e *Encoder) Flush() error {
	out := "]"
	if e.isAtStart {
		out = "[]"
	}
	if len(e.buffer) > 0 {
		chunk := e.buffer[:]
		encoded, err := encode(string(chunk))
		if err != nil {
			return err
		}
		out = piece(encoded, e.isAtStart, true)
		e.buffer = nil
	}
	_, err := e.writer.Write([]byte(out))
	if err != nil {
		return err
	}
	return nil
}

// Weird Text Format-8 Decoder.
type Decoder struct {
	reader    io.Reader
	buffer    []byte // Read buffer.
	decoded   []byte // Decoded data buffer.
	isAtStart bool   // Start of the stream.
	isAtEnd   bool   // End of the stream.
}

// Decode a data stream encoded in Weird Text Format-8.
func (d *Decoder) Read(p []byte) (n int, err error) {
	// If there is decoded data left from a previous call,
	// copy as much as the caller allows into their buffer.
	if len(d.decoded) > 0 {
		n = copy(p, d.decoded)
		d.decoded = d.decoded[n:]
		return n, nil
	}
	if d.isAtEnd {
		return 0, io.EOF
	}
	buf := make([]byte, len(p))
	n, err = d.reader.Read(buf)
	if err != nil && err != io.EOF {
		return n, err
	}
	d.buffer = append(d.buffer, buf[:n]...)
	if err == io.EOF {
		if len(d.buffer) == 0 || d.buffer[len(d.buffer)-1] != ']' {
			return 0, errors.New("missing ]")
		}
		d.isAtEnd = true
	}
	if d.isAtStart && len(d.buffer) > 0 {
		if d.buffer[0] != '[' {
			return n, errors.New("missing [")
		}
		d.isAtStart = false
		d.buffer = d.buffer[1:]
	}
	const minChunkSize = 12 // 10 digits (max uint32) + comma + space.
	// Extract full numbers from the read buffer and
	// decode them into the decoded buffer.
	for len(d.buffer) > 0 {
		idx := bytes.IndexByte(d.buffer, byte(' '))
		if idx == -1 && d.isAtEnd {
			idx = len(d.buffer) - 1
		}
		if idx == -1 && len(d.buffer) > minChunkSize {
			return 0, errors.New("missing space separator")
		}
		if idx == -1 {
			break
		}
		if d.buffer[idx] == ' ' && d.buffer[max(0, idx-1)] != ',' {
			return 0, errors.New("missing comma separator")
		}
		end := idx
		if d.buffer[idx] == ' ' {
			end = idx - 1
		}
		chunk := d.buffer[:end]
		num, err := strconv.ParseUint(string(chunk), 10, 32) // base 10; 32 bits.
		if err != nil {
			return 0, err
		}
		decoded := decode(uint32(num))
		d.decoded = append(d.decoded, []byte(decoded)...)
		d.buffer = d.buffer[idx+1:]
	}
	n = copy(p, d.decoded)
	d.decoded = d.decoded[n:]
	return n, nil
}

// Encode reader data stream into writer Weird Text Format-8 data stream.
func PipeEncoder(reader io.Reader, writer io.Writer) error {
	encoder := &Encoder{writer: writer, isAtStart: true}
	_, err := io.Copy(encoder, reader)
	if err != nil {
		return err
	}
	err = encoder.Flush()
	if err != nil {
		return err
	}
	return nil
}

// Decode reader Weird Text Format-8 data stream into writer data stream.
func PipeDecoder(reader io.Reader, writer io.Writer) error {
	decoder := &Decoder{reader: reader, isAtStart: true}
	_, err := io.Copy(writer, decoder)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	reader := bytes.NewBuffer([]byte("ecastro"))
	writer := bytes.NewBuffer([]byte{})
	err := PipeEncoder(reader, writer)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(writer.String())
	reader2 := bytes.NewBuffer([]byte("[267911599, 124994916]"))
	writer2 := bytes.NewBuffer([]byte{})
	err = PipeDecoder(reader2, writer2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(writer2.String())
	//fmt.Println(decode(124994916))
	fmt.Println("ok")
}

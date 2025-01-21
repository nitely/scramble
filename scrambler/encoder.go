package scrambler

import (
	"fmt"
	"io"
)

// Convert string of up to 4 chars into
// a decoded/raw integer.
func rawInteger(s string) (uint32, error) {
	if len(s) > 4 {
		return 0, fmt.Errorf("string must be <= 4 chars long")
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

// Weird Text Format-8 Encoder.
type encoder struct {
	writer    io.Writer
	buffer    []byte // Write buffer.
	isAtStart bool   // Start of the stream.
}

// Format a encoded integer to Weird Text Format-8.
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

// Encode a data stream to Weird Text Format-8.
func (e *encoder) Write(p []byte) (n int, err error) {
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

// Flush remaining buffer data if any and closing list symbol into the writter.
func (e *encoder) Flush() error {
	if e.isAtStart && len(e.buffer) == 0 {
		return fmt.Errorf("empty input")
	}
	if e.buffer == nil { // Already flushed.
		return nil
	}
	out := "]"
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

// Encode reader data stream into writer data stream.
func PipeEncoder(reader io.Reader, writer io.Writer) error {
	encoder := &encoder{writer: writer, isAtStart: true}
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

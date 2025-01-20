package scrambler

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Decode an integer encoded in Weird Text Format-8.
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

// Decode reader data stream into writer data stream.
func PipeDecoder(reader io.Reader, writer io.Writer) error {
	decoder := &Decoder{reader: reader, isAtStart: true}
	_, err := io.Copy(writer, decoder)
	if err != nil {
		return err
	}
	return nil
}

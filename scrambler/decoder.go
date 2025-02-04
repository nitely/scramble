package scrambler

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Decode an integer encoded in Scramble.
func decode(enc uint32) string {
	decoded := uint32(0)
	nibble := 0
	for i := range 32 {
		sourcePos := nibble*4 + (i / 8)              // Position where the bit was placed during encoding.
		bit := (enc & (1 << sourcePos)) >> sourcePos // Extract the bit from its encoded position.
		decoded |= bit << i                          // Place the bit back in its original position.
		nibble = (nibble + 1) % 8                    // Move to the next nibble wrapping around.
	}
	// Uint32 to 4 chars string.
	result := make([]byte, 4)
	result[3] = byte((decoded >> 24) & 0xFF)
	result[2] = byte((decoded >> 16) & 0xFF)
	result[1] = byte((decoded >> 8) & 0xFF)
	result[0] = byte(decoded & 0xFF)
	return strings.TrimRight(string(result), "\000")
}

// The max token size allowed
const maxTokenSize = 20

// Split data into tokens; includes the separators
func scanToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b == ' ' || b == ',' || b == '[' || b == ']' {
			if i == 0 {
				return 1, data[:1], nil
			}
			return i, data[:i], nil
		}
		if i > maxTokenSize {
			return 0, nil, fmt.Errorf("token too long")
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return
}

// Decode reader data stream into writer data stream.
func PipeDecoder(reader io.Reader, writer io.Writer) error {
	// This is a fairly strict parser, crafted to handle untrusted text.
	// It won't consume more than the default buffer size (64KB) for bad inputs.
	// Input data format: ``[number(, number)*]``.
	scanner := bufio.NewScanner(reader)
	scanner.Split(scanToken)
	atEnd := false
	spaces, commas, pos := 0, 0, 0
	for scanner.Scan() {
		tok := scanner.Text()
		switch {
		case pos == 0:
			if tok != "[" {
				return fmt.Errorf("expected [ at pos %d", pos)
			}
		case tok == "[":
			return fmt.Errorf("unexpected [ at pos %d", pos)
		case tok == "]":
			if pos == 1 {
				return fmt.Errorf("empty list")
			}
			if commas+spaces != 0 {
				return fmt.Errorf("unexpected char at pos %d", pos)
			}
			atEnd = true
			if scanner.Scan() {
				return fmt.Errorf("unexpected char after ] at pos %d", pos)
			}
		case tok == ",":
			commas++
			if commas > 1 {
				return fmt.Errorf("unexpected comma at pos %d", pos)
			}
		case tok == " ":
			spaces++
			if spaces > 1 || commas != 1 {
				return fmt.Errorf("unexpected space at pos %d", pos)
			}
		default:
			if commas != 1 && pos > 1 {
				return fmt.Errorf("expected comma at pos %d", pos)
			}
			if spaces != 1 && pos > 1 {
				return fmt.Errorf("expected space at pos %d", pos)
			}
			commas, spaces = 0, 0
			num, err := strconv.ParseUint(tok, 10, 32) // base 10; 32 bits.
			if err != nil {
				return fmt.Errorf("failed to parse number at pos %d", pos)
			}
			decoded := decode(uint32(num))
			_, err = writer.Write([]byte(decoded))
			if err != nil {
				return err
			}
		}
		pos += len(tok)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s error at pos %d", err, pos)
	}
	if pos == 0 {
		return fmt.Errorf("empty input")
	}
	if !atEnd {
		return fmt.Errorf("missing ]")
	}
	return nil
}

/*
Simple CLI to encode/decode streams of data to/from escrambled data format.

The command is meant to compose with pipes by default. You may add echo at the end
to get a newline.

Usage:

$ ./scramble -encode "foo"; echo
$ echo -n "foo" | ./scramble -encode; echo
$ cat encfile | ./scramble -encode; echo

$ ./scramble -decode "[124807030]"; echo
$ echo -n "[124807030]" | ./scramble -decode; echo
$ cat decfile | ./scramble -decode; echo

$ echo -n "foo" | ./scramble -encode | ./scramble -decode; echo
*/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/nitely/scrambled-data-format/scrambler"
)

func main() {
	encodeFlag := flag.Bool("encode", false, "Encode the input text to Base64")
	decodeFlag := flag.Bool("decode", false, "Decode the input Base64 text")
	flag.Parse()
	if *encodeFlag && *decodeFlag {
		fmt.Fprintln(os.Stderr, "Error: You cannot use both -encode and -decode at the same time.")
		flag.Usage()
		os.Exit(1)
	}
	if !*encodeFlag && !*decodeFlag {
		fmt.Fprintln(os.Stderr, "Error: You must specify either -encode or -decode.")
		flag.Usage()
		os.Exit(1)
	}
	args := flag.Args()
	var reader io.Reader
	if len(args) > 0 {
		reader = bytes.NewBuffer([]byte(args[0]))
	} else {
		// pipe
		reader = bufio.NewReader(os.Stdin)
	}
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	if *encodeFlag {
		err := scrambler.PipeEncoder(reader, writer)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error encoding input: ", err)
			os.Exit(1)
		}
	}
	if *decodeFlag {
		err := scrambler.PipeDecoder(reader, writer)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error decoding input: ", err)
			os.Exit(1)
		}
	}
}
